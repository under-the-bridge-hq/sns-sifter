package store

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/under-the-bridge-hq/sns-sifter/internal/xapi"
)

type LikedPost struct {
	UserID     string
	TweetID    string
	AuthorID   string
	Text       string
	CreatedAt  string
	FetchedAt  string
	Category   sql.NullString
	ReviewedAt sql.NullString
	RawJSON    string
}

func InsertLikedPosts(db *sql.DB, userID string, tweets []xapi.Tweet, rawJSON string) (int, error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO liked_posts (user_id, tweet_id, author_id, text, created_at, fetched_at, raw_json)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(user_id, tweet_id) DO NOTHING
	`)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	now := time.Now().UTC().Format(time.RFC3339)
	inserted := 0
	for _, t := range tweets {
		res, err := stmt.Exec(userID, t.ID, t.AuthorID, t.Text, t.CreatedAt, now, rawJSON)
		if err != nil {
			return inserted, err
		}
		if n, _ := res.RowsAffected(); n > 0 {
			inserted++
		}
	}
	if err := tx.Commit(); err != nil {
		return inserted, err
	}
	return inserted, nil
}

func LikedPostExists(db *sql.DB, userID, tweetID string) (bool, error) {
	var n int
	err := db.QueryRow(`SELECT 1 FROM liked_posts WHERE user_id = ? AND tweet_id = ?`, userID, tweetID).Scan(&n)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}

type LikedPostsFilter struct {
	UserID     string
	Category   string // "ai" / "work" / "other" / "" (all) / "uncategorized"
	Unreviewed bool
	Keyword    string
	Limit      int
}

type LikedPostRow struct {
	TweetID        string
	AuthorID       string
	AuthorUsername string
	AuthorName     string
	Text           string
	CreatedAt      string
	Category       string
	Reviewed       bool
}

func ListLikedPosts(db *sql.DB, f *LikedPostsFilter) ([]LikedPostRow, error) {
	query := `
		SELECT lp.tweet_id, lp.author_id,
		       COALESCE(u.username, '') AS author_username,
		       COALESCE(u.name, '') AS author_name,
		       lp.text, lp.created_at,
		       COALESCE(lp.category, '') AS category,
		       lp.reviewed_at IS NOT NULL AS reviewed
		FROM liked_posts lp
		LEFT JOIN users u ON u.id = lp.author_id
		WHERE lp.user_id = ?
	`
	args := []any{f.UserID}

	switch f.Category {
	case "":
		// no filter
	case "uncategorized":
		query += " AND lp.category IS NULL"
	default:
		query += " AND lp.category = ?"
		args = append(args, f.Category)
	}

	if f.Unreviewed {
		query += " AND lp.reviewed_at IS NULL"
	}

	if f.Keyword != "" {
		query += " AND lp.text LIKE ?"
		args = append(args, "%"+f.Keyword+"%")
	}

	query += " ORDER BY lp.created_at DESC"

	if f.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, f.Limit)
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []LikedPostRow
	for rows.Next() {
		var r LikedPostRow
		if err := rows.Scan(&r.TweetID, &r.AuthorID, &r.AuthorUsername, &r.AuthorName, &r.Text, &r.CreatedAt, &r.Category, &r.Reviewed); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func GetLikedPost(db *sql.DB, userID, tweetID string) (*LikedPostRow, error) {
	rows, err := ListLikedPosts(db, &LikedPostsFilter{UserID: userID})
	if err != nil {
		return nil, err
	}
	for _, r := range rows {
		if r.TweetID == tweetID {
			return &r, nil
		}
	}
	return nil, nil
}

func UpdateLikedPostCategory(db *sql.DB, userID, tweetID, category string) error {
	_, err := db.Exec(`UPDATE liked_posts SET category = ? WHERE user_id = ? AND tweet_id = ?`, category, userID, tweetID)
	return err
}

// CategorizeUncategorized は未分類の投稿に対して classifier を適用する。
// classifier(text) -> "ai" / "work" / "other" を返す。
// all=true で既に分類済みのものも再分類する。
func CategorizeUncategorized(db *sql.DB, userID string, classifier func(text string) string, all bool) (int, error) {
	query := `SELECT tweet_id, text FROM liked_posts WHERE user_id = ? AND category IS NULL`
	if all {
		query = `SELECT tweet_id, text FROM liked_posts WHERE user_id = ?`
	}
	rows, err := db.Query(query, userID)
	if err != nil {
		return 0, err
	}
	type item struct{ id, text string }
	var items []item
	for rows.Next() {
		var it item
		if err := rows.Scan(&it.id, &it.text); err != nil {
			rows.Close()
			return 0, err
		}
		items = append(items, it)
	}
	rows.Close()

	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	stmt, err := tx.Prepare(`UPDATE liked_posts SET category = ? WHERE user_id = ? AND tweet_id = ?`)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	updated := 0
	for _, it := range items {
		cat := classifier(it.text)
		if _, err := stmt.Exec(cat, userID, it.id); err != nil {
			return updated, err
		}
		updated++
	}
	if err := tx.Commit(); err != nil {
		return updated, err
	}
	return updated, nil
}

func MarkLikedPostsReviewed(db *sql.DB, userID string, tweetIDs []string) (int, error) {
	if len(tweetIDs) == 0 {
		return 0, nil
	}
	now := time.Now().UTC().Format(time.RFC3339)
	placeholders := strings.Repeat("?,", len(tweetIDs))
	placeholders = placeholders[:len(placeholders)-1]
	query := fmt.Sprintf(`UPDATE liked_posts SET reviewed_at = ? WHERE user_id = ? AND tweet_id IN (%s)`, placeholders)
	args := []any{now, userID}
	for _, id := range tweetIDs {
		args = append(args, id)
	}
	res, err := db.Exec(query, args...)
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	return int(n), nil
}

// CountLikedPosts は user_id ごとの統計を返す (total / by_category / unreviewed)
type LikedPostsStats struct {
	Total        int
	AICount      int
	WorkCount    int
	OtherCount   int
	Uncategorized int
	Unreviewed   int
}

func CountLikedPosts(db *sql.DB, userID string) (*LikedPostsStats, error) {
	var s LikedPostsStats
	err := db.QueryRow(`
		SELECT
			COUNT(*),
			SUM(CASE WHEN category = 'ai' THEN 1 ELSE 0 END),
			SUM(CASE WHEN category = 'work' THEN 1 ELSE 0 END),
			SUM(CASE WHEN category = 'other' THEN 1 ELSE 0 END),
			SUM(CASE WHEN category IS NULL THEN 1 ELSE 0 END),
			SUM(CASE WHEN reviewed_at IS NULL THEN 1 ELSE 0 END)
		FROM liked_posts WHERE user_id = ?
	`, userID).Scan(&s.Total, &s.AICount, &s.WorkCount, &s.OtherCount, &s.Uncategorized, &s.Unreviewed)
	if err != nil {
		return nil, err
	}
	return &s, nil
}
