package cli

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/under-the-bridge-hq/sns-sifter/internal/store"
	"github.com/under-the-bridge-hq/sns-sifter/internal/xapi"
)

type App struct {
	DB      *sql.DB
	Client  *xapi.Client
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
	if token != "" {
		client := xapi.NewClient(token)
		client.Verbose = verbose
		app.Client = client
	}

	switch remaining[0] {
	case "sync":
		return app.runSync(remaining[1:])
	case "following":
		return app.runFollowing(remaining[1:])
	case "history":
		return app.runHistory(remaining[1:])
	default:
		fmt.Fprintf(os.Stderr, "不明なコマンド: %s\n", remaining[0])
		printUsage()
		return 1
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `使い方: sifter [flags] <command>

コマンド:
  sync following <username>       フォロー一覧を取得しDBにキャッシュ
  following list <username>       キャッシュ済みフォロー一覧を表示
  following diff <username>       前回との差分を表示
  history <username>              同期履歴を表示

フラグ:
  --db <path>     SQLiteパス (デフォルト: ~/.sifter/data.db)
  --token <token> Bearer Token (環境変数: SIFTER_BEARER_TOKEN)
  --verbose, -v   詳細出力
`)
}
