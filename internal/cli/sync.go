package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/under-the-bridge-hq/sns-sifter/internal/store"
	"github.com/under-the-bridge-hq/sns-sifter/internal/xapi"
)

func (a *App) runSync(args []string) int {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "使い方: sifter sync following <username>")
		return 1
	}
	switch args[0] {
	case "following":
		return a.syncFollowing(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "不明なサブコマンド: sync %s\n", args[0])
		return 1
	}
}

func (a *App) syncFollowing(args []string) int {
	if a.Client == nil {
		fmt.Fprintln(os.Stderr, "Bearer Token が設定されていません (--token または SIFTER_BEARER_TOKEN)")
		return 1
	}

	username := args[0]
	force := false
	for _, arg := range args[1:] {
		if arg == "--force" {
			force = true
		}
	}

	// username → user ID の解決（キャッシュ優先）
	user, err := store.GetUserByUsername(a.DB, username)
	if err != nil {
		fmt.Fprintf(os.Stderr, "DBエラー: %v\n", err)
		return 1
	}
	if user == nil {
		fmt.Printf("@%s のユーザー情報を取得中...\n", username)
		user, err = a.Client.GetUserByUsername(username)
		if err != nil {
			fmt.Fprintf(os.Stderr, "APIエラー: %v\n", err)
			return 1
		}
		if err := store.UpsertUser(a.DB, user); err != nil {
			fmt.Fprintf(os.Stderr, "DB保存エラー: %v\n", err)
			return 1
		}
	}

	// クールダウンチェック
	if !force {
		latest, err := store.LatestCompletedSync(a.DB, user.ID, "following")
		if err != nil {
			fmt.Fprintf(os.Stderr, "DBエラー: %v\n", err)
			return 1
		}
		if latest != nil && latest.CompletedAt != nil {
			since := time.Since(*latest.CompletedAt)
			if since < 24*time.Hour {
				fmt.Printf("@%s のフォロー一覧は %s 前に同期済みです (%d件)。\n",
					username, since.Truncate(time.Minute), latest.TotalFetched)
				fmt.Println("再取得するには --force を指定してください。")
				return 0
			}
		}
	}

	// 同期開始
	syncID, err := store.CreateSync(a.DB, user.ID, "following")
	if err != nil {
		fmt.Fprintf(os.Stderr, "DB同期レコード作成エラー: %v\n", err)
		return 1
	}

	fmt.Printf("@%s (フォロー数: %d) のフォロー一覧を取得中...\n", username, user.PublicMetrics.FollowingCount)

	users, apiCalls, err := a.Client.GetAllFollowing(user.ID, func(page []xapi.User, calls int) {
		fmt.Printf("  %d件取得済み (APIコール: %d)\n", len(page), calls)
	})
	if err != nil {
		store.FailSync(a.DB, syncID, err.Error())
		fmt.Fprintf(os.Stderr, "APIエラー: %v\n", err)
		return 1
	}

	// ユーザー情報をDB保存
	if err := store.UpsertUsers(a.DB, users); err != nil {
		store.FailSync(a.DB, syncID, err.Error())
		fmt.Fprintf(os.Stderr, "DB保存エラー: %v\n", err)
		return 1
	}

	// フォロー関係を保存
	ids := make([]string, len(users))
	for i, u := range users {
		ids[i] = u.ID
	}
	if err := store.InsertFollowing(a.DB, syncID, user.ID, ids); err != nil {
		store.FailSync(a.DB, syncID, err.Error())
		fmt.Fprintf(os.Stderr, "DB保存エラー: %v\n", err)
		return 1
	}

	if err := store.CompleteSync(a.DB, syncID, len(users), apiCalls); err != nil {
		fmt.Fprintf(os.Stderr, "DB更新エラー: %v\n", err)
		return 1
	}

	cost := float64(len(users)) * 0.005
	fmt.Printf("\n完了: %d件同期 (APIコール: %d, 推定コスト: $%.2f)\n", len(users), apiCalls, cost)
	return 0
}
