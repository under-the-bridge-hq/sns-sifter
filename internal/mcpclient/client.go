// Package mcpclient は xmcp (FastMCP HTTP) を呼び出す最小クライアント。
//
// FastMCP の HTTP transport は MCP プロトコル (JSON-RPC 2.0) を以下の流れで扱う:
//  1. POST /mcp method=initialize → レスポンスヘッダ mcp-session-id を取得
//  2. POST /mcp method=notifications/initialized (session id ヘッダ付き)
//  3. POST /mcp method=tools/call (session id ヘッダ付き)
//
// レスポンスは Server-Sent Events 形式で `event: message\ndata: {json}\n\n` 形式。
package mcpclient

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"time"
)

type Client struct {
	endpoint  string
	httpClient *http.Client
	sessionID string
	idCounter int64
	Verbose   bool
}

func New(endpoint string) *Client {
	return &Client{
		endpoint:   endpoint,
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}
}

type rpcRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      int64  `json:"id,omitempty"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *rpcError) Error() string { return fmt.Sprintf("MCP error %d: %s", e.Code, e.Message) }

type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

func (c *Client) nextID() int64 {
	return atomic.AddInt64(&c.idCounter, 1)
}

func (c *Client) Initialize() error {
	id := c.nextID()
	req := rpcRequest{
		JSONRPC: "2.0",
		ID:      id,
		Method:  "initialize",
		Params: map[string]any{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]any{},
			"clientInfo": map[string]any{
				"name":    "sifter",
				"version": "0.1",
			},
		},
	}
	body, _ := json.Marshal(req)
	httpReq, err := http.NewRequest("POST", c.endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json, text/event-stream")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("MCP 接続失敗 (%s): %w", c.endpoint, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("MCP initialize 失敗: HTTP %d: %s", resp.StatusCode, string(raw))
	}

	c.sessionID = resp.Header.Get("mcp-session-id")
	if c.sessionID == "" {
		return fmt.Errorf("MCP initialize: session id がヘッダにありません")
	}

	if _, err := readSSEMessage(resp.Body); err != nil {
		return fmt.Errorf("MCP initialize レスポンス解析: %w", err)
	}

	notif := rpcRequest{
		JSONRPC: "2.0",
		Method:  "notifications/initialized",
		Params:  map[string]any{},
	}
	notifBody, _ := json.Marshal(notif)
	httpReq2, _ := http.NewRequest("POST", c.endpoint, bytes.NewReader(notifBody))
	httpReq2.Header.Set("Content-Type", "application/json")
	httpReq2.Header.Set("Accept", "application/json, text/event-stream")
	httpReq2.Header.Set("mcp-session-id", c.sessionID)
	resp2, err := c.httpClient.Do(httpReq2)
	if err != nil {
		return fmt.Errorf("MCP notifications/initialized 失敗: %w", err)
	}
	resp2.Body.Close()

	return nil
}

// CallTool は tools/call を実行し、result.structuredContent を JSON にデコードする。
func (c *Client) CallTool(name string, arguments map[string]any, into any) error {
	if c.sessionID == "" {
		if err := c.Initialize(); err != nil {
			return err
		}
	}
	id := c.nextID()
	req := rpcRequest{
		JSONRPC: "2.0",
		ID:      id,
		Method:  "tools/call",
		Params: map[string]any{
			"name":      name,
			"arguments": arguments,
		},
	}
	body, _ := json.Marshal(req)

	if c.Verbose {
		fmt.Printf("  MCP: tools/call %s\n", name)
	}

	httpReq, _ := http.NewRequest("POST", c.endpoint, bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json, text/event-stream")
	httpReq.Header.Set("mcp-session-id", c.sessionID)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("MCP tools/call %s 失敗: %w", name, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("MCP tools/call %s: HTTP %d: %s", name, resp.StatusCode, string(raw))
	}

	rpcResp, err := readSSEMessage(resp.Body)
	if err != nil {
		return fmt.Errorf("MCP tools/call %s レスポンス解析: %w", name, err)
	}
	if rpcResp.Error != nil {
		return rpcResp.Error
	}

	var result struct {
		StructuredContent json.RawMessage `json:"structuredContent"`
		Content           []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		IsError bool `json:"isError"`
	}
	if err := json.Unmarshal(rpcResp.Result, &result); err != nil {
		return fmt.Errorf("MCP tools/call %s result 解析: %w", name, err)
	}
	if result.IsError {
		text := ""
		if len(result.Content) > 0 {
			text = result.Content[0].Text
		}
		return fmt.Errorf("MCP tool %s エラー: %s", name, text)
	}

	if into == nil {
		return nil
	}

	// structuredContent を優先、なければ content[0].text を JSON とみなす
	if len(result.StructuredContent) > 0 && string(result.StructuredContent) != "null" {
		return json.Unmarshal(result.StructuredContent, into)
	}
	if len(result.Content) > 0 {
		return json.Unmarshal([]byte(result.Content[0].Text), into)
	}
	return fmt.Errorf("MCP tool %s: 空のレスポンス", name)
}

// readSSEMessage は SSE ストリームから 1 件の `event: message\ndata: {...}\n\n` を読む。
func readSSEMessage(r io.Reader) (*rpcResponse, error) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 64*1024), 4*1024*1024)
	var dataLines []string
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			if len(dataLines) > 0 {
				break
			}
			continue
		}
		if strings.HasPrefix(line, "data:") {
			dataLines = append(dataLines, strings.TrimPrefix(line, "data:"))
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(dataLines) == 0 {
		return nil, fmt.Errorf("SSE ストリームから data 行が読めません")
	}
	payload := strings.TrimSpace(strings.Join(dataLines, ""))
	var rpc rpcResponse
	if err := json.Unmarshal([]byte(payload), &rpc); err != nil {
		return nil, fmt.Errorf("SSE payload JSON 解析: %w (payload=%s)", err, payload)
	}
	return &rpc, nil
}
