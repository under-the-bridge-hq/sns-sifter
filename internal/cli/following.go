package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/under-the-bridge-hq/sns-sifter/internal/domain"
	"github.com/under-the-bridge-hq/sns-sifter/internal/store"
)

func (a *App) runFollowing(args []string) int {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "使い方: sifter following <list|diff> <username>")
		return 1
	}
	switch args[0] {
	case "list":
		return a.followingList(args[1:])
	case "diff":
		return a.followingDiff(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "不明なサブコマンド: following %s\n", args[0])
		return 1
	}
}

func (a *App) followingList(args []string) int {
	username := args[0]
	filter := ""
	for i, arg := range args {
		if arg == "--filter" && i+1 < len(args) {
			filter = args[i+1]
		}
	}

	user, err := store.GetUserByUsername(a.DB, username)
	if err != nil {
		fmt.Fprintf(os.Stderr, "DBエラー: %v\n", err)
		return 1
	}
	if user == nil {
		fmt.Fprintf(os.Stderr, "@%s のデータがありません。先に sync を実行してください。\n", username)
		return 1
	}

	latest, err := store.LatestCompletedSync(a.DB, user.ID, "following")
	if err != nil || latest == nil {
		fmt.Fprintf(os.Stderr, "@%s の同期データがありません。先に sync を実行してください。\n", username)
		return 1
	}

	var uf *store.UserFilter
	if filter != "" {
		uf = &store.UserFilter{Keyword: filter}
	}

	users, err := store.SearchFollowing(a.DB, latest.ID, uf)
	if err != nil {
		fmt.Fprintf(os.Stderr, "DBエラー: %v\n", err)
		return 1
	}

	if filter != "" {
		fmt.Printf("@%s のフォロー一覧 (フィルタ: %q, %d/%d件)\n\n", username, filter, len(users), latest.TotalFetched)
	} else {
		fmt.Printf("@%s のフォロー一覧 (%d件)\n\n", username, len(users))
	}

	for _, u := range users {
		desc := u.Description
		desc = strings.ReplaceAll(desc, "\n", " ")
		if len([]rune(desc)) > 80 {
			desc = string([]rune(desc)[:80]) + "..."
		}
		fmt.Printf("  @%-20s %s\n", u.Username, u.Name)
		if desc != "" {
			fmt.Printf("  %-21s %s\n", "", desc)
		}
		fmt.Println()
	}
	return 0
}

func (a *App) followingDiff(args []string) int {
	username := args[0]

	user, err := store.GetUserByUsername(a.DB, username)
	if err != nil {
		fmt.Fprintf(os.Stderr, "DBエラー: %v\n", err)
		return 1
	}
	if user == nil {
		fmt.Fprintf(os.Stderr, "@%s のデータがありません。\n", username)
		return 1
	}

	latest, err := store.LatestCompletedSync(a.DB, user.ID, "following")
	if err != nil || latest == nil {
		fmt.Fprintln(os.Stderr, "同期データがありません。")
		return 1
	}

	previous, err := store.PreviousCompletedSync(a.DB, user.ID, "following", latest.ID)
	if err != nil || previous == nil {
		fmt.Fprintln(os.Stderr, "比較対象の同期データがありません。2回以上 sync を実行してください。")
		return 1
	}

	oldIDs, err := store.GetFollowingIDs(a.DB, previous.ID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "DBエラー: %v\n", err)
		return 1
	}
	newIDs, err := store.GetFollowingIDs(a.DB, latest.ID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "DBエラー: %v\n", err)
		return 1
	}

	added, removed := domain.ComputeDiff(oldIDs, newIDs)

	fmt.Printf("@%s のフォロー差分 (sync #%d → #%d)\n\n", username, previous.ID, latest.ID)

	if len(added) == 0 && len(removed) == 0 {
		fmt.Println("  変更なし")
		return 0
	}

	if len(added) > 0 {
		addedUsers, _ := store.GetUsersByIDs(a.DB, added)
		fmt.Printf("  新規フォロー (+%d):\n", len(added))
		for _, u := range addedUsers {
			fmt.Printf("    + @%-20s %s\n", u.Username, u.Name)
		}
		fmt.Println()
	}

	if len(removed) > 0 {
		removedUsers, _ := store.GetUsersByIDs(a.DB, removed)
		fmt.Printf("  フォロー解除 (-%d):\n", len(removed))
		for _, u := range removedUsers {
			fmt.Printf("    - @%-20s %s\n", u.Username, u.Name)
		}
		fmt.Println()
	}

	return 0
}
