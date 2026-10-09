package xapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

type Client struct {
	httpClient  *http.Client
	bearerToken string
	oauth1      *OAuth1Signer
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

const (
	maxRateLimitRetries = 3
	defaultMaxWait      = 15 * time.Minute
)

// SetOAuth1 を呼ぶと、以降の全リクエストを OAuth 1.0a User Context で署名する。
func (c *Client) SetOAuth1(consumerKey, consumerSecret, accessToken, accessTokenSecret string) {
	c.oauth1 = &OAuth1Signer{
		ConsumerKey:    consumerKey,
		ConsumerSecret: consumerSecret,
		Token:          accessToken,
		TokenSecret:    accessTokenSecret,
	}
}

// HasOAuth1 は OAuth 1.0a が設定されているかを返す。
func (c *Client) HasOAuth1() bool { return c.oauth1 != nil }

func maxRateLimitWait() time.Duration {
	if v := os.Getenv("SIFTER_RATE_LIMIT_MAX_WAIT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			return d
		}
	}
	return defaultMaxWait
}

func (c *Client) get(path string, params map[string]string) ([]byte, error) {
	query := url.Values{}
	for k, v := range params {
		query.Set(k, v)
	}
	baseURL := c.baseURL + path
	maxWait := maxRateLimitWait()

	for attempt := 0; ; attempt++ {
		req, err := http.NewRequest("GET", baseURL, nil)
		if err != nil {
			return nil, err
		}
		req.URL.RawQuery = EncodeQuery(query)
		if c.oauth1 != nil {
			req.Header.Set("Authorization", c.oauth1.AuthorizationHeader("GET", baseURL, query, nil, nil))
		} else {
			req.Header.Set("Authorization", "Bearer "+c.bearerToken)
		}

		if c.Verbose {
			fmt.Printf("  API: GET %s\n", req.URL)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, err
		}
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			if resetStr := resp.Header.Get("x-rate-limit-reset"); resetStr != "" {
				resetUnix, _ := strconv.ParseInt(resetStr, 10, 64)
				waitDur := time.Until(time.Unix(resetUnix, 0)) + time.Second
				if waitDur > maxWait {
					return nil, fmt.Errorf("レート制限: リセットまで %s (上限 %s を超過): %s",
						waitDur.Truncate(time.Second), maxWait, string(body))
				}
				if attempt >= maxRateLimitRetries {
					return nil, fmt.Errorf("レート制限: %d 回再試行しても解消しませんでした: %s", maxRateLimitRetries, string(body))
				}
				if waitDur > 0 {
					fmt.Printf("  レート制限に到達。%s 待機します (再試行 %d/%d)...\n",
						waitDur.Truncate(time.Second), attempt+1, maxRateLimitRetries)
					time.Sleep(waitDur)
				}
				continue
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

// GetLikedPostsPage は GET /2/users/{id}/liked_tweets の 1 ページを取得する (OAuth 1.0a 必須)。
func (c *Client) GetLikedPostsPage(userID string, maxResults int, paginationToken string) (*TweetsResponse, error) {
	params := map[string]string{
		"max_results":  strconv.Itoa(maxResults),
		"tweet.fields": "id,text,author_id,created_at",
		"expansions":   "author_id",
		"user.fields":  "id,username,name,description,public_metrics",
	}
	if paginationToken != "" {
		params["pagination_token"] = paginationToken
	}
	body, err := c.get("/2/users/"+userID+"/liked_tweets", params)
	if err != nil {
		return nil, err
	}
	var resp TweetsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
