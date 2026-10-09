package cli

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/under-the-bridge-hq/sns-sifter/internal/store"
	"github.com/under-the-bridge-hq/sns-sifter/internal/xapi"
)

func (a *App) runSync(args []string) int {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "使い方: sifter sync <following|likes> <username>")
		return 1
	}
	switch args[0] {
	case "following":
		return a.syncFollowing(args[1:])
	case "likes":
		return a.syncLikes(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "不明なサブコマンド: sync %s\n", args[0])
		return 1
	}
}

func (a *App) syncLikes(args []string) int {
	if a.Client == nil || !a.Client.HasOAuth1() {
		fmt.Fprintln(os.Stderr, "liked_tweets には OAuth 1.0a User Context が必要です。`sifter auth login` を実行してください。")
		return 1
	}
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "使い方: sifter sync likes <username> [--full] [--max <N>] [--max-pages <N>]")
		return 1
	}

	username := args[0]
	full := false
	maxResults := 20
	maxPages := 10
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--full":
			full = true
		case "--max":
			if i+1 < len(args) {
				i++
				if n, err := strconv.Atoi(args[i]); err == nil && n >= 5 && n <= 100 {
					maxResults = n
				}
			}
		case "--max-pages":
			if i+1 < len(args) {
				i++
				if n, err := strconv.Atoi(args[i]); err == nil && n > 0 {
					maxPages = n
				}
			}
		}
	}

	user, err := a.resolveUser(username)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}

	syncID, err := store.CreateSync(a.DB, user.ID, "likes")
	if err != nil {
		fmt.Fprintf(os.Stderr, "DB同期レコード作成エラー: %v\n", err)
		return 1
	}

	mode := "incremental"
	if full {
		mode = "full"
	}
	fmt.Printf("@%s の liked_tweets を取得中 (mode=%s, max_results=%d, max_pages=%d)...\n",
		username, mode, maxResults, maxPages)

	var (
		totalFetched  int // 新規として取り込んだ件数
		totalReturned int // API が返した件数 (課金対象)
		totalNew      int
		apiCalls      int
		token         string
		hitKnown      bool
	)

	for page := 0; page < maxPages; page++ {
		resp, err := a.Client.GetLikedPostsPage(user.ID, maxResults, token)
		if err != nil {
			store.FailSync(a.DB, syncID, err.Error())
			fmt.Fprintf(os.Stderr, "APIエラー: %v\n", err)
			return 1
		}
		apiCalls++
		totalReturned += len(resp.Data)

		var newTweets []xapi.Tweet
		for _, t := range resp.Data {
			if !full {
				exists, err := store.LikedPostExists(a.DB, user.ID, t.ID)
				if err != nil {
					store.FailSync(a.DB, syncID, err.Error())
					fmt.Fprintf(os.Stderr, "DBエラー (LikedPostExists): %v\n", err)
					return 1
				}
				if exists {
					hitKnown = true
					break
				}
			}
			newTweets = append(newTweets, t)
		}
		totalFetched += len(newTweets)

		if len(newTweets) > 0 {
			if err := store.UpsertUsers(a.DB, resp.Includes.Users); err != nil {
				store.FailSync(a.DB, syncID, err.Error())
				fmt.Fprintf(os.Stderr, "DB保存エラー (authors): %v\n", err)
				return 1
			}
			n, err := store.InsertLikedPosts(a.DB, user.ID, newTweets, "")
			if err != nil {
				store.FailSync(a.DB, syncID, err.Error())
				fmt.Fprintf(os.Stderr, "DB保存エラー (tweets): %v\n", err)
				return 1
			}
			totalNew += n
		}

		fmt.Printf("  page %d: 取得 %d 件 (新規累計: %d, next=%s)\n",
			page+1, len(resp.Data), totalFetched, truncToken(resp.Meta.NextToken))

		if hitKnown || resp.Meta.NextToken == "" {
			break
		}
		token = resp.Meta.NextToken
	}

	if err := store.CompleteSync(a.DB, syncID, totalNew, apiCalls); err != nil {
		fmt.Fprintf(os.Stderr, "DB更新エラー: %v\n", err)
		return 1
	}

	unit := xapi.PricePostRead
	if a.Creds != nil && a.Creds.UserID == user.ID {
		unit = xapi.PriceOwnedRead
	}
	cost := float64(totalReturned) * unit
	fmt.Printf("\n完了: 新規 %d 件 / 取得 %d 件 (APIコール: %d, 推定コスト: $%.3f = 返却 %d件 × $%.3f)\n",
		totalNew, totalFetched, apiCalls, cost, totalReturned, unit)
	return 0
}

// resolveUser は username からユーザーを解決する。DB キャッシュ → 必要なら API で取得。
func (a *App) resolveUser(username string) (*xapi.User, error) {
	user, err := store.GetUserByUsername(a.DB, username)
	if err != nil {
		return nil, fmt.Errorf("DBエラー: %w", err)
	}
	if user != nil {
		return user, nil
	}
	if a.Client != nil {
		fmt.Printf("@%s のユーザー情報を取得中 ...\n", username)
		u, err := a.Client.GetUserByUsername(username)
		if err != nil {
			return nil, fmt.Errorf("APIエラー: %w", err)
		}
		if err := store.UpsertUser(a.DB, u); err != nil {
			return nil, fmt.Errorf("DB保存エラー: %w", err)
		}
		return u, nil
	}
	return nil, fmt.Errorf("@%s のユーザー情報がありません。Bearer Token または `sifter auth login` を設定してください", username)
}

func truncToken(t string) string {
	if len(t) > 12 {
		return t[:12] + "..."
	}
	if t == "" {
		return "(end)"
	}
	return t
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

	cost := float64(len(users)) * xapi.PriceFollowingRead
	fmt.Printf("\n完了: %d件同期 (APIコール: %d, 推定コスト: $%.2f = %d件 × $%.3f)\n", len(users), apiCalls, cost, len(users), xapi.PriceFollowingRead)
	return 0
}
