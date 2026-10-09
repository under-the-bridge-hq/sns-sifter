package xapi

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const OAuthBaseURL = "https://api.x.com"

// PostOAuthForm は request_token / access_token 用の署名付き POST を行い、form レスポンスを返す。
// extraOAuth には oauth_callback / oauth_verifier を渡す。
func (s *OAuth1Signer) PostOAuthForm(endpoint string, extraOAuth map[string]string) (url.Values, error) {
	req, err := http.NewRequest("POST", endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", s.AuthorizationHeader("POST", endpoint, nil, nil, extraOAuth))
	resp, err := (&http.Client{Timeout: 30 * time.Second}).Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: %d %s", endpoint, resp.StatusCode, string(body))
	}
	return url.ParseQuery(string(body))
}
