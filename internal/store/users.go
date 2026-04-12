package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/under-the-bridge-hq/sns-sifter/internal/xapi"
)

func UpsertUser(db *sql.DB, u *xapi.User) error {
	metricsJSON, _ := json.Marshal(u.PublicMetrics)
	_, err := db.Exec(`
		INSERT INTO users (id, username, name, description, public_metrics_json, fetched_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			username = excluded.username,
			name = excluded.name,
			description = excluded.description,
			public_metrics_json = excluded.public_metrics_json,
			fetched_at = excluded.fetched_at
	`, u.ID, u.Username, u.Name, u.Description, string(metricsJSON), time.Now().UTC().Format(time.RFC3339))
	return err
}

func UpsertUsers(db *sql.DB, users []xapi.User) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO users (id, username, name, description, public_metrics_json, fetched_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			username = excluded.username,
			name = excluded.name,
			description = excluded.description,
			public_metrics_json = excluded.public_metrics_json,
			fetched_at = excluded.fetched_at
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	now := time.Now().UTC().Format(time.RFC3339)
	for _, u := range users {
		metricsJSON, _ := json.Marshal(u.PublicMetrics)
		if _, err := stmt.Exec(u.ID, u.Username, u.Name, u.Description, string(metricsJSON), now); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func GetUserByUsername(db *sql.DB, username string) (*xapi.User, error) {
	var u xapi.User
	var metricsJSON string
	err := db.QueryRow(`SELECT id, username, name, description, public_metrics_json FROM users WHERE username = ?`, username).
		Scan(&u.ID, &u.Username, &u.Name, &u.Description, &metricsJSON)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	json.Unmarshal([]byte(metricsJSON), &u.PublicMetrics)
	return &u, nil
}

func GetUsersByIDs(db *sql.DB, ids []string) ([]xapi.User, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	query := `SELECT id, username, name, description, public_metrics_json FROM users WHERE id IN (`
	args := make([]any, len(ids))
	for i, id := range ids {
		if i > 0 {
			query += ","
		}
		query += "?"
		args[i] = id
	}
	query += ")"

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []xapi.User
	for rows.Next() {
		var u xapi.User
		var metricsJSON string
		if err := rows.Scan(&u.ID, &u.Username, &u.Name, &u.Description, &metricsJSON); err != nil {
			return nil, err
		}
		json.Unmarshal([]byte(metricsJSON), &u.PublicMetrics)
		users = append(users, u)
	}
	return users, rows.Err()
}

type UserFilter struct {
	Keyword string
}

// CompareOptions は following compare の絞り込み条件
type CompareOptions struct {
	// Keywords は OR マッチ (description / name / username の LIKE)
	Keywords []string
	// SortBy: "followers" | "following" | "tweets" | "listed" (デフォルト followers)
	SortBy string
	// Top: 上位 N 件 (0 で無制限)
	Top int
	// ExcludeBaseUserID: ベースユーザー自身を結果から除外
	ExcludeBaseUserID string
}

// CompareFollowing は refSyncID がフォローしていて baseSyncID がフォローしていないユーザーを返す。
// API コールは発生せず DB のみで完結する。
func CompareFollowing(db *sql.DB, baseSyncID, refSyncID int64, opts *CompareOptions) ([]xapi.User, error) {
	if opts == nil {
		opts = &CompareOptions{}
	}

	query := `
		SELECT u.id, u.username, u.name, u.description, u.public_metrics_json
		FROM following f_ref
		JOIN users u ON u.id = f_ref.target_user_id
		WHERE f_ref.sync_id = ?
		  AND u.id NOT IN (SELECT target_user_id FROM following WHERE sync_id = ?)
	`
	args := []any{refSyncID, baseSyncID}

	if opts.ExcludeBaseUserID != "" {
		query += " AND u.id != ?"
		args = append(args, opts.ExcludeBaseUserID)
	}

	if len(opts.Keywords) > 0 {
		clauses := make([]string, 0, len(opts.Keywords))
		for _, kw := range opts.Keywords {
			clauses = append(clauses, "(u.description LIKE ? OR u.name LIKE ? OR u.username LIKE ?)")
			like := "%" + kw + "%"
			args = append(args, like, like, like)
		}
		query += " AND (" + strings.Join(clauses, " OR ") + ")"
	}

	sortField := "followers_count"
	switch opts.SortBy {
	case "following":
		sortField = "following_count"
	case "tweets":
		sortField = "tweet_count"
	case "listed":
		sortField = "listed_count"
	case "", "followers":
		sortField = "followers_count"
	}
	query += fmt.Sprintf(" ORDER BY CAST(json_extract(u.public_metrics_json, '$.%s') AS INTEGER) DESC", sortField)

	if opts.Top > 0 {
		query += " LIMIT ?"
		args = append(args, opts.Top)
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []xapi.User
	for rows.Next() {
		var u xapi.User
		var metricsJSON string
		if err := rows.Scan(&u.ID, &u.Username, &u.Name, &u.Description, &metricsJSON); err != nil {
			return nil, err
		}
		json.Unmarshal([]byte(metricsJSON), &u.PublicMetrics)
		users = append(users, u)
	}
	return users, rows.Err()
}

func SearchFollowing(db *sql.DB, syncID int64, filter *UserFilter) ([]xapi.User, error) {
	query := `
		SELECT u.id, u.username, u.name, u.description, u.public_metrics_json
		FROM following f
		JOIN users u ON u.id = f.target_user_id
		WHERE f.sync_id = ?
	`
	args := []any{syncID}

	if filter != nil && filter.Keyword != "" {
		query += " AND (u.description LIKE ? OR u.name LIKE ? OR u.username LIKE ?)"
		kw := "%" + filter.Keyword + "%"
		args = append(args, kw, kw, kw)
	}
	query += " ORDER BY u.username"

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []xapi.User
	for rows.Next() {
		var u xapi.User
		var metricsJSON string
		if err := rows.Scan(&u.ID, &u.Username, &u.Name, &u.Description, &metricsJSON); err != nil {
			return nil, err
		}
		json.Unmarshal([]byte(metricsJSON), &u.PublicMetrics)
		users = append(users, u)
	}
	return users, rows.Err()
}
