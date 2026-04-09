package store

import (
	"database/sql"
	"time"
)

type SyncRecord struct {
	ID            int64
	SourceUserID  string
	SyncType      string
	StartedAt     time.Time
	CompletedAt   *time.Time
	TotalFetched  int
	APICalls      int
	Status        string
}

func CreateSync(db *sql.DB, sourceUserID, syncType string) (int64, error) {
	res, err := db.Exec(`
		INSERT INTO sync_history (source_user_id, sync_type, started_at, status)
		VALUES (?, ?, ?, 'running')
	`, sourceUserID, syncType, time.Now().UTC().Format(time.RFC3339))
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func CompleteSync(db *sql.DB, syncID int64, totalFetched, apiCalls int) error {
	_, err := db.Exec(`
		UPDATE sync_history
		SET completed_at = ?, total_fetched = ?, api_calls = ?, status = 'completed'
		WHERE id = ?
	`, time.Now().UTC().Format(time.RFC3339), totalFetched, apiCalls, syncID)
	return err
}

func FailSync(db *sql.DB, syncID int64, errMsg string) error {
	_, err := db.Exec(`
		UPDATE sync_history
		SET completed_at = ?, status = 'failed'
		WHERE id = ?
	`, time.Now().UTC().Format(time.RFC3339), syncID)
	_ = err
	return err
}

func InsertFollowing(db *sql.DB, syncID int64, sourceUserID string, targetUserIDs []string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`INSERT INTO following (source_user_id, target_user_id, sync_id) VALUES (?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, targetID := range targetUserIDs {
		if _, err := stmt.Exec(sourceUserID, targetID, syncID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func LatestCompletedSync(db *sql.DB, sourceUserID, syncType string) (*SyncRecord, error) {
	var r SyncRecord
	var startedAt, completedAt string
	err := db.QueryRow(`
		SELECT id, source_user_id, sync_type, started_at, completed_at, total_fetched, api_calls, status
		FROM sync_history
		WHERE source_user_id = ? AND sync_type = ? AND status = 'completed'
		ORDER BY id DESC LIMIT 1
	`, sourceUserID, syncType).Scan(
		&r.ID, &r.SourceUserID, &r.SyncType, &startedAt, &completedAt,
		&r.TotalFetched, &r.APICalls, &r.Status,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	r.StartedAt, _ = time.Parse(time.RFC3339, startedAt)
	if completedAt != "" {
		t, _ := time.Parse(time.RFC3339, completedAt)
		r.CompletedAt = &t
	}
	return &r, nil
}

func PreviousCompletedSync(db *sql.DB, sourceUserID, syncType string, beforeSyncID int64) (*SyncRecord, error) {
	var r SyncRecord
	var startedAt, completedAt string
	err := db.QueryRow(`
		SELECT id, source_user_id, sync_type, started_at, completed_at, total_fetched, api_calls, status
		FROM sync_history
		WHERE source_user_id = ? AND sync_type = ? AND status = 'completed' AND id < ?
		ORDER BY id DESC LIMIT 1
	`, sourceUserID, syncType, beforeSyncID).Scan(
		&r.ID, &r.SourceUserID, &r.SyncType, &startedAt, &completedAt,
		&r.TotalFetched, &r.APICalls, &r.Status,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	r.StartedAt, _ = time.Parse(time.RFC3339, startedAt)
	if completedAt != "" {
		t, _ := time.Parse(time.RFC3339, completedAt)
		r.CompletedAt = &t
	}
	return &r, nil
}

func GetFollowingIDs(db *sql.DB, syncID int64) ([]string, error) {
	rows, err := db.Query(`SELECT target_user_id FROM following WHERE sync_id = ?`, syncID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func ListSyncHistory(db *sql.DB, sourceUserID, syncType string) ([]SyncRecord, error) {
	rows, err := db.Query(`
		SELECT id, source_user_id, sync_type, started_at, COALESCE(completed_at, ''), total_fetched, api_calls, status
		FROM sync_history
		WHERE source_user_id = ? AND sync_type = ?
		ORDER BY id DESC
	`, sourceUserID, syncType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []SyncRecord
	for rows.Next() {
		var r SyncRecord
		var startedAt, completedAt string
		if err := rows.Scan(&r.ID, &r.SourceUserID, &r.SyncType, &startedAt, &completedAt, &r.TotalFetched, &r.APICalls, &r.Status); err != nil {
			return nil, err
		}
		r.StartedAt, _ = time.Parse(time.RFC3339, startedAt)
		if completedAt != "" {
			t, _ := time.Parse(time.RFC3339, completedAt)
			r.CompletedAt = &t
		}
		records = append(records, r)
	}
	return records, rows.Err()
}
