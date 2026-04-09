package cli

import (
	"fmt"
	"os"

	"github.com/under-the-bridge-hq/sns-sifter/internal/store"
)

func (a *App) runHistory(args []string) int {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "使い方: sifter history <username>")
		return 1
	}
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

	records, err := store.ListSyncHistory(a.DB, user.ID, "following")
	if err != nil {
		fmt.Fprintf(os.Stderr, "DBエラー: %v\n", err)
		return 1
	}

	if len(records) == 0 {
		fmt.Println("同期履歴がありません。")
		return 0
	}

	fmt.Printf("@%s の同期履歴:\n\n", username)
	fmt.Printf("  %-4s  %-20s  %-10s  %-6s  %-10s  %s\n", "ID", "日時", "ステータス", "件数", "APIコール", "推定コスト")
	fmt.Printf("  %-4s  %-20s  %-10s  %-6s  %-10s  %s\n", "----", "--------------------", "----------", "------", "----------", "----------")

	for _, r := range records {
		cost := fmt.Sprintf("$%.2f", float64(r.TotalFetched)*0.005)
		fmt.Printf("  %-4d  %-20s  %-10s  %-6d  %-10d  %s\n",
			r.ID,
			r.StartedAt.Local().Format("2006-01-02 15:04:05"),
			r.Status,
			r.TotalFetched,
			r.APICalls,
			cost,
		)
	}
	return 0
}
