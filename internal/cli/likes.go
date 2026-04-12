package cli

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/under-the-bridge-hq/sns-sifter/internal/domain"
	"github.com/under-the-bridge-hq/sns-sifter/internal/store"
)

func (a *App) runLikes(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "使い方: sifter likes <list|show|categorize|review|stats> ...")
		return 1
	}
	switch args[0] {
	case "list":
		return a.likesList(args[1:])
	case "show":
		return a.likesShow(args[1:])
	case "categorize":
		return a.likesCategorize(args[1:])
	case "review":
		return a.likesReview(args[1:])
	case "stats":
		return a.likesStats(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "不明なサブコマンド: likes %s\n", args[0])
		return 1
	}
}

func (a *App) likesList(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "使い方: sifter likes list <username> [--category ai|work|other|uncategorized] [--unreviewed] [--keyword <kw>] [--limit N]")
		return 1
	}
	username := args[0]

	f := &store.LikedPostsFilter{Limit: 50}
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--category":
			if i+1 < len(args) {
				i++
				f.Category = args[i]
			}
		case "--unreviewed":
			f.Unreviewed = true
		case "--keyword":
			if i+1 < len(args) {
				i++
				f.Keyword = args[i]
			}
		case "--limit":
			if i+1 < len(args) {
				i++
				if n, err := strconv.Atoi(args[i]); err == nil {
					f.Limit = n
				}
			}
		}
	}

	user, err := store.GetUserByUsername(a.DB, username)
	if err != nil || user == nil {
		fmt.Fprintf(os.Stderr, "@%s のデータがありません。先に sync を実行してください。\n", username)
		return 1
	}
	f.UserID = user.ID

	rows, err := store.ListLikedPosts(a.DB, f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "DBエラー: %v\n", err)
		return 1
	}

	header := fmt.Sprintf("@%s のいいね投稿 (%d件", username, len(rows))
	if f.Category != "" {
		header += ", category=" + f.Category
	}
	if f.Unreviewed {
		header += ", unreviewed"
	}
	if f.Keyword != "" {
		header += fmt.Sprintf(", keyword=%q", f.Keyword)
	}
	header += ")"
	fmt.Println(header)
	fmt.Println()

	for _, r := range rows {
		text := strings.ReplaceAll(r.Text, "\n", " ")
		if len([]rune(text)) > 200 {
			text = string([]rune(text)[:200]) + "..."
		}
		mark := " "
		if r.Reviewed {
			mark = "✓"
		}
		cat := r.Category
		if cat == "" {
			cat = "?"
		}
		author := "@" + r.AuthorUsername
		if r.AuthorUsername == "" {
			author = r.AuthorID
		}
		fmt.Printf("[%s] %-4s %s  %s\n", mark, cat, r.TweetID, author)
		fmt.Printf("       %s\n", r.CreatedAt)
		fmt.Printf("       %s\n\n", text)
	}
	return 0
}

func (a *App) likesShow(args []string) int {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "使い方: sifter likes show <username> <tweet_id>")
		return 1
	}
	username := args[0]
	tweetID := args[1]

	user, err := store.GetUserByUsername(a.DB, username)
	if err != nil || user == nil {
		fmt.Fprintf(os.Stderr, "@%s のデータがありません。\n", username)
		return 1
	}
	r, err := store.GetLikedPost(a.DB, user.ID, tweetID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "DBエラー: %v\n", err)
		return 1
	}
	if r == nil {
		fmt.Fprintf(os.Stderr, "tweet %s が見つかりません。\n", tweetID)
		return 1
	}
	fmt.Printf("Tweet ID: %s\n", r.TweetID)
	fmt.Printf("Author:   @%s (%s)\n", r.AuthorUsername, r.AuthorName)
	fmt.Printf("Date:     %s\n", r.CreatedAt)
	fmt.Printf("Category: %s\n", strOr(r.Category, "(uncategorized)"))
	fmt.Printf("Reviewed: %v\n", r.Reviewed)
	fmt.Println()
	fmt.Println(r.Text)
	return 0
}

func (a *App) likesCategorize(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "使い方: sifter likes categorize <username> [--all]")
		return 1
	}
	username := args[0]
	all := false
	for _, arg := range args[1:] {
		if arg == "--all" {
			all = true
		}
	}
	user, err := store.GetUserByUsername(a.DB, username)
	if err != nil || user == nil {
		fmt.Fprintf(os.Stderr, "@%s のデータがありません。\n", username)
		return 1
	}

	updated, err := store.CategorizeUncategorized(a.DB, user.ID, domain.ClassifyByKeyword, all)
	if err != nil {
		fmt.Fprintf(os.Stderr, "DBエラー: %v\n", err)
		return 1
	}
	fmt.Printf("分類完了: %d件\n", updated)

	stats, err := store.CountLikedPosts(a.DB, user.ID)
	if err == nil && stats != nil {
		fmt.Printf("内訳: ai=%d, work=%d, other=%d, uncategorized=%d\n",
			stats.AICount, stats.WorkCount, stats.OtherCount, stats.Uncategorized)
	}
	return 0
}

func (a *App) likesReview(args []string) int {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "使い方: sifter likes review <username> <tweet_id> [<tweet_id>...]")
		return 1
	}
	username := args[0]
	user, err := store.GetUserByUsername(a.DB, username)
	if err != nil || user == nil {
		fmt.Fprintf(os.Stderr, "@%s のデータがありません。\n", username)
		return 1
	}
	n, err := store.MarkLikedPostsReviewed(a.DB, user.ID, args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "DBエラー: %v\n", err)
		return 1
	}
	fmt.Printf("レビュー済みマーク: %d件\n", n)
	return 0
}

func (a *App) likesStats(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "使い方: sifter likes stats <username>")
		return 1
	}
	username := args[0]
	user, err := store.GetUserByUsername(a.DB, username)
	if err != nil || user == nil {
		fmt.Fprintf(os.Stderr, "@%s のデータがありません。\n", username)
		return 1
	}
	stats, err := store.CountLikedPosts(a.DB, user.ID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "DBエラー: %v\n", err)
		return 1
	}
	fmt.Printf("@%s のいいね統計\n", username)
	fmt.Printf("  total:         %d\n", stats.Total)
	fmt.Printf("  ai:            %d\n", stats.AICount)
	fmt.Printf("  work:          %d\n", stats.WorkCount)
	fmt.Printf("  other:         %d\n", stats.OtherCount)
	fmt.Printf("  uncategorized: %d\n", stats.Uncategorized)
	fmt.Printf("  unreviewed:    %d\n", stats.Unreviewed)
	return 0
}

func strOr(s, fallback string) string {
	if s == "" {
		return fallback
	}
	return s
}
