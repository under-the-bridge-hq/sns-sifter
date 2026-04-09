package store

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

const schema = `
CREATE TABLE IF NOT EXISTS users (
    id                  TEXT PRIMARY KEY,
    username            TEXT NOT NULL,
    name                TEXT NOT NULL,
    description         TEXT DEFAULT '',
    public_metrics_json TEXT DEFAULT '{}',
    fetched_at          TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);

CREATE TABLE IF NOT EXISTS sync_history (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    source_user_id  TEXT NOT NULL,
    sync_type       TEXT NOT NULL,
    started_at      TEXT NOT NULL,
    completed_at    TEXT,
    total_fetched   INTEGER DEFAULT 0,
    api_calls       INTEGER DEFAULT 0,
    status          TEXT DEFAULT 'running'
);

CREATE TABLE IF NOT EXISTS following (
    source_user_id TEXT NOT NULL,
    target_user_id TEXT NOT NULL,
    sync_id        INTEGER NOT NULL,
    PRIMARY KEY (source_user_id, target_user_id, sync_id)
);
CREATE INDEX IF NOT EXISTS idx_following_source ON following(source_user_id);
`

func Open(dbPath string) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, err
	}
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}
