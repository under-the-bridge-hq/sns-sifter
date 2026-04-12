package xapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

type Client struct {
	httpClient  *http.Client
	bearerToken string
	baseURL     string
	Verbose     bool
}

func NewClient(bearerToken string) *Client {
	return &Client{
		httpClient:  &http.Client{Timeout: 30 * time.Second},
		bearerToken: bearerToken,
		baseURL:     "https://api.x.com",
	}
}

func (c *Client) get(path string, params map[string]string) ([]byte, error) {
	req, err := http.NewRequest("GET", c.baseURL+path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.bearerToken)

	q := req.URL.Query()
	for k, v := range params {
		q.Set(k, v)
	}
	req.URL.RawQuery = q.Encode()

	if c.Verbose {
		fmt.Printf("  API: GET %s\n", req.URL)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		resetStr := resp.Header.Get("x-rate-limit-reset")
		if resetStr != "" {
			resetUnix, _ := strconv.ParseInt(resetStr, 10, 64)
			waitDur := time.Until(time.Unix(resetUnix, 0))
			if waitDur > 0 {
				fmt.Printf("  レート制限に到達。%s 待機します...\n", waitDur.Truncate(time.Second))
				time.Sleep(waitDur + time.Second)
				return c.get(path, params)
			}
		}
	}

	if resp.StatusCode != http.StatusOK {
		var apiErr APIError
		if json.Unmarshal(body, &apiErr) == nil && apiErr.Title != "" {
			return nil, &apiErr
		}
		return nil, fmt.Errorf("API error: %d %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// RawGet はデバッグ用に生のレスポンス body を返す。
// 通常のコードパスでは使わず、APIレスポンスの構造調査専用。
func (c *Client) RawGet(path string, params map[string]string) ([]byte, error) {
	return c.get(path, params)
}

func (c *Client) GetUserByUsername(username string) (*User, error) {
	params := map[string]string{
		"user.fields": "id,username,name,description,public_metrics",
	}
	body, err := c.get("/2/users/by/username/"+username, params)
	if err != nil {
		return nil, err
	}
	var resp UserResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}
