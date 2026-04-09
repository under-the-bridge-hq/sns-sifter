package store

import (
	"database/sql"
	"encoding/json"
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
