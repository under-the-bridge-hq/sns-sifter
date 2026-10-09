package cli

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/under-the-bridge-hq/sns-sifter/internal/auth"
	"github.com/under-the-bridge-hq/sns-sifter/internal/xapi"
)

func (a *App) runAuth(args []string) int {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "使い方: sifter auth <login|status>")
		return 1
	}
	switch args[0] {
	case "login":
		return a.authLogin(args[1:])
	case "status":
		return a.authStatus()
	default:
		fmt.Fprintf(os.Stderr, "不明なサブコマンド: auth %s\n", args[0])
		return 1
	}
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func (a *App) authLogin(args []string) int {
	consumerKey := os.Getenv("SIFTER_OAUTH_CONSUMER_KEY")
	consumerSecret := os.Getenv("SIFTER_OAUTH_CONSUMER_SECRET")
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--consumer-key":
			if i+1 < len(args) {
				i++
				consumerKey = args[i]
			}
		case "--consumer-secret":
			if i+1 < len(args) {
				i++
				consumerSecret = args[i]
			}
		}
	}
	if consumerKey == "" || consumerSecret == "" {
		fmt.Fprintln(os.Stderr, "consumer key/secret が未設定です (SIFTER_OAUTH_CONSUMER_KEY / SIFTER_OAUTH_CONSUMER_SECRET または --consumer-key / --consumer-secret)")
		return 1
	}

	host := envOr("SIFTER_OAUTH_CALLBACK_HOST", "127.0.0.1")
	port := envOr("SIFTER_OAUTH_CALLBACK_PORT", "8976")
	callbackURL := fmt.Sprintf("http://%s:%s/oauth/callback", host, port)

	signer := &xapi.OAuth1Signer{ConsumerKey: consumerKey, ConsumerSecret: consumerSecret}

	rt, err := signer.PostOAuthForm(xapi.OAuthBaseURL+"/oauth/request_token", map[string]string{"oauth_callback": callbackURL})
	if err != nil {
		fmt.Fprintf(os.Stderr, "request_token 取得エラー: %v\n", err)
		return 1
	}
	reqToken, reqSecret := rt.Get("oauth_token"), rt.Get("oauth_token_secret")
	if reqToken == "" || reqSecret == "" {
		fmt.Fprintln(os.Stderr, "request_token のレスポンスに oauth_token / oauth_token_secret がありません")
		return 1
	}

	type cbResult struct{ token, verifier string }
	resultCh := make(chan cbResult, 1)
	mux := http.NewServeMux()
	mux.HandleFunc("/oauth/callback", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("oauth_token") != reqToken || q.Get("oauth_verifier") == "" {
			http.Error(w, "invalid callback", http.StatusBadRequest)
			return
		}
		fmt.Fprintln(w, "認可を受け付けました。このタブは閉じて構いません。")
		select {
		case resultCh <- cbResult{q.Get("oauth_token"), q.Get("oauth_verifier")}:
		default:
		}
	})
	ln, err := net.Listen("tcp", net.JoinHostPort(host, port))
	if err != nil {
		fmt.Fprintf(os.Stderr, "callback サーバー起動エラー: %v\n", err)
		return 1
	}
	srv := &http.Server{Handler: mux}
	go srv.Serve(ln)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		srv.Shutdown(ctx)
	}()

	fmt.Println("次の URL をブラウザで開いて認可してください:")
	fmt.Printf("  %s/oauth/authorize?oauth_token=%s\n", xapi.OAuthBaseURL, url.QueryEscape(reqToken))
	fmt.Printf("callback 待機中: %s (タイムアウト 300 秒)\n", callbackURL)

	var res cbResult
	select {
	case res = <-resultCh:
	case <-time.After(300 * time.Second):
		fmt.Fprintln(os.Stderr, "タイムアウト: callback を受信できませんでした")
		return 1
	}

	signer.Token = res.token
	signer.TokenSecret = reqSecret
	at, err := signer.PostOAuthForm(xapi.OAuthBaseURL+"/oauth/access_token", map[string]string{"oauth_verifier": res.verifier})
	if err != nil {
		fmt.Fprintf(os.Stderr, "access_token 取得エラー: %v\n", err)
		return 1
	}
	creds := &auth.Credentials{
		ConsumerKey:       consumerKey,
		ConsumerSecret:    consumerSecret,
		AccessToken:       at.Get("oauth_token"),
		AccessTokenSecret: at.Get("oauth_token_secret"),
		UserID:            at.Get("user_id"),
		ScreenName:        at.Get("screen_name"),
	}
	if creds.AccessToken == "" || creds.AccessTokenSecret == "" {
		fmt.Fprintln(os.Stderr, "access_token のレスポンスにトークンがありません")
		return 1
	}
	path, err := auth.Save(creds)
	if err != nil {
		fmt.Fprintf(os.Stderr, "認証情報の保存エラー: %v\n", err)
		return 1
	}
	fmt.Printf("認証完了: @%s (user_id=%s)\n保存先: %s\n", creds.ScreenName, creds.UserID, path)
	return 0
}

func (a *App) authStatus() int {
	path, err := auth.Path()
	if err != nil {
		fmt.Fprintf(os.Stderr, "保存先の解決エラー: %v\n", err)
		return 1
	}
	fmt.Printf("保存先: %s\n", path)
	creds, err := auth.Load()
	if errors.Is(err, auth.ErrNotFound) {
		fmt.Println("状態: 未認証 (sifter auth login を実行してください)")
		return 0
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "読み込みエラー: %v\n", err)
		return 1
	}
	fmt.Printf("screen_name: %s\nuser_id: %s\naccess token: %s\n", creds.ScreenName, creds.UserID, presence(creds.AccessToken != "" && creds.AccessTokenSecret != ""))
	return 0
}

func presence(ok bool) string {
	if ok {
		return "あり"
	}
	return "なし"
}
