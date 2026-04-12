package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/under-the-bridge-hq/sns-sifter/internal/store"
)

func (a *App) runDebug(args []string) int {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "使い方: sifter debug following-page <username> [--max-results N] [--pagination-token TOKEN]")
		return 1
	}
	switch args[0] {
	case "following-page":
		return a.debugFollowingPage(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "不明なサブコマンド: debug %s\n", args[0])
		return 1
	}
}

func (a *App) debugFollowingPage(args []string) int {
	if a.Client == nil {
		fmt.Fprintln(os.Stderr, "Bearer Token が設定されていません (--token または SIFTER_BEARER_TOKEN)")
		return 1
	}
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "使い方: sifter debug following-page <username> [--max-results N] [--pagination-token TOKEN]")
		return 1
	}

	username := args[0]
	maxResults := 1
	paginationToken := ""

	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--max-results":
			if i+1 < len(args) {
				i++
				if n, err := strconv.Atoi(args[i]); err == nil {
					maxResults = n
				}
			}
		case "--pagination-token":
			if i+1 < len(args) {
				i++
				paginationToken = args[i]
			}
		}
	}

	// username → user ID は DB キャッシュから(API コールを節約)
	user, err := store.GetUserByUsername(a.DB, username)
	if err != nil {
		fmt.Fprintf(os.Stderr, "DBエラー: %v\n", err)
		return 1
	}
	if user == nil {
		fmt.Fprintf(os.Stderr, "@%s が DB にありません。先に sync を実行してください。\n", username)
		return 1
	}

	estCost := float64(maxResults) * 0.005
	fmt.Printf("デバッグ実行: @%s (id=%s) の following を取得します\n", username, user.ID)
	fmt.Printf("  max_results=%d  pagination_token=%q\n", maxResults, paginationToken)
	fmt.Printf("  推定コスト: $%.4f (%d件 × $0.005)\n\n", estCost, maxResults)

	params := map[string]string{
		"user.fields": "id,username,name,description,public_metrics",
		"max_results": strconv.Itoa(maxResults),
	}
	if paginationToken != "" {
		params["pagination_token"] = paginationToken
	}

	body, err := a.Client.RawGet("/2/users/"+user.ID+"/following", params)
	if err != nil {
		fmt.Fprintf(os.Stderr, "APIエラー: %v\n", err)
		return 1
	}

	// 生 JSON を pretty-print
	var pretty map[string]any
	if err := json.Unmarshal(body, &pretty); err != nil {
		fmt.Println("--- raw body (parse failed) ---")
		fmt.Println(string(body))
		return 1
	}
	prettyBytes, _ := json.MarshalIndent(pretty, "", "  ")
	fmt.Println("--- response JSON ---")
	fmt.Println(string(prettyBytes))

	// meta だけ抜き出して目立たせる
	if meta, ok := pretty["meta"].(map[string]any); ok {
		fmt.Println("\n--- meta 抜粋 ---")
		fmt.Printf("  result_count: %v\n", meta["result_count"])
		if nt, ok := meta["next_token"]; ok && nt != nil {
			fmt.Printf("  next_token  : %v  ← 続きあり\n", nt)
		} else {
			fmt.Printf("  next_token  : (なし)  ← ここで終わり\n")
		}
	} else {
		fmt.Println("\n--- meta が存在しません ---")
	}

	return 0
}
