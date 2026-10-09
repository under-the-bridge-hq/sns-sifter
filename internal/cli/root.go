package cli

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"errors"
	"github.com/under-the-bridge-hq/sns-sifter/internal/auth"
	"github.com/under-the-bridge-hq/sns-sifter/internal/store"
	"github.com/under-the-bridge-hq/sns-sifter/internal/xapi"
)

type App struct {
	DB      *sql.DB
	Client  *xapi.Client
	Creds   *auth.Credentials
	Verbose bool
}

func Run(args []string) int {
	dbPath := os.Getenv("SIFTER_DB")
	if dbPath == "" {
		home, _ := os.UserHomeDir()
		dbPath = filepath.Join(home, ".sifter", "data.db")
	}
	token := os.Getenv("SIFTER_BEARER_TOKEN")
	verbose := false

	// グローバルフラグの解析
	remaining := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--db":
			if i+1 < len(args) {
				i++
				dbPath = args[i]
			}
		case "--token":
			if i+1 < len(args) {
				i++
				token = args[i]
			}
		case "--verbose", "-v":
			verbose = true
		default:
			remaining = append(remaining, args[i])
		}
	}

	if len(remaining) == 0 {
		printUsage()
		return 1
	}

	db, err := store.Open(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "DB初期化エラー: %v\n", err)
		return 1
	}
	defer db.Close()

	app := &App{
		DB:      db,
		Verbose: verbose,
	}
	creds, err := auth.Load()
	switch {
	case err == nil:
		client := xapi.NewClient(token)
		client.SetOAuth1(creds.ConsumerKey, creds.ConsumerSecret, creds.AccessToken, creds.AccessTokenSecret)
		client.Verbose = verbose
		app.Client = client
		app.Creds = creds
	case errors.Is(err, auth.ErrNotFound):
		if token != "" {
			client := xapi.NewClient(token)
			client.Verbose = verbose
			app.Client = client
		}
	default:
		fmt.Fprintf(os.Stderr, "認証情報の読み込みエラー: %v\n", err)
		return 1
	}

	switch remaining[0] {
	case "sync":
		return app.runSync(remaining[1:])
	case "following":
		return app.runFollowing(remaining[1:])
	case "likes":
		return app.runLikes(remaining[1:])
	case "history":
		return app.runHistory(remaining[1:])
	case "auth":
		return app.runAuth(remaining[1:])
	case "debug":
		return app.runDebug(remaining[1:])
	default:
		fmt.Fprintf(os.Stderr, "不明なコマンド: %s\n", remaining[0])
		printUsage()
		return 1
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `使い方: sifter [flags] <command>

コマンド:
  sync following <username>           フォロー一覧を取得しDBにキャッシュ
  sync likes <username> [--full]      いいね投稿を取得しDBにキャッシュ (incremental が既定)
  following list <username>           キャッシュ済みフォロー一覧を表示
  following diff <username>           前回との差分を表示
  following compare <base> <ref>      ref がフォローしていて base がフォローしていないユーザーを抽出
  likes list <username>               いいね投稿一覧を表示 (--category, --unreviewed, --keyword)
  likes show <username> <tweet_id>    いいね投稿の詳細表示
  likes categorize <username>         未分類投稿をキーワードで分類
  likes review <username> <ids...>    1on1セッション後にレビュー済みとしてマーク
  likes stats <username>              統計表示
  history <username>                  同期履歴を表示
  auth login                          OAuth 1.0a 3-legged 認証 (認証情報を保存)
  auth status                         保存済み認証情報の状態を表示

フラグ:
  --db <path>      SQLiteパス (デフォルト: ~/.sifter/data.db)
  --token <token>  Bearer Token (環境変数: SIFTER_BEARER_TOKEN)
  --verbose, -v    詳細出力

環境変数:
  SIFTER_DB                     SQLiteパス
  SIFTER_BEARER_TOKEN           Bearer Token (OAuth 1.0a 認証情報が無い場合に使用)
  SIFTER_CREDENTIALS            認証情報 JSON のパス (デフォルト: ~/.sifter/credentials.json)
  SIFTER_OAUTH_CONSUMER_KEY     auth login 用 consumer key
  SIFTER_OAUTH_CONSUMER_SECRET  auth login 用 consumer secret
  SIFTER_OAUTH_CALLBACK_HOST    callback ホスト (デフォルト: 127.0.0.1)
  SIFTER_OAUTH_CALLBACK_PORT    callback ポート (デフォルト: 8976)
  SIFTER_RATE_LIMIT_MAX_WAIT    429 時の 1 回の待機上限 (デフォルト: 15m)
`)
}
