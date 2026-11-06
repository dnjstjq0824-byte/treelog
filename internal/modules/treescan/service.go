package treescan

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Fault is the module error envelope.
type Fault struct {
	Status  int
	Code    string
	Message string
}

func (f *Fault) Error() string { return f.Message }

// Service maintains the TreeScan outbound sync feed.
type Service struct {
	pool *pgxpool.Pool
}

// NewService builds a treescan service.
func NewService(pool *pgxpool.Pool) (*Service, error) {
	if pool == nil {
		return nil, errors.New("treescan service requires a database pool")
	}
	return &Service{pool: pool}, nil
}

func (s *Service) requirePool() error {
	if s.pool == nil {
		return &Fault{Status: http.StatusServiceUnavailable, Code: "SERVICE_UNAVAILABLE", Message: "Database is unavailable"}
	}
	return nil
}

// SyncEntry is one export feed row.
type SyncEntry struct {
	SequenceCursor int64     `json:"sequence_cursor"`
	PostID         string    `json:"post_id"`
	BlogID         string    `json:"blog_id"`
	SyncAction     string    `json:"sync_action"`
	Visibility     string    `json:"visibility_snapshot"`
	PayloadHash    string    `json:"payload_hash"`
	Status         string    `json:"sync_status"`
	Title          string    `json:"title,omitempty"`
	Slug           string    `json:"slug,omitempty"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// MarkPost records or refreshes the sync row for a post after publish,
// update, or unpublish. It is called by the background worker.
func (s *Service) MarkPost(ctx context.Context, postID string, deleted bool) error {
	if err := s.requirePool(); err != nil {
		return err
	}
	var blogID, title, slug, visibility, status string
	err := s.pool.QueryRow(ctx, `SELECT blog_id, title, slug, visibility, status FROM posts WHERE id = $1`, postID).
		Scan(&blogID, &title, &slug, &visibility, &status)
	if errors.Is(err, pgx.ErrNoRows) {
		return &Fault{Status: http.StatusNotFound, Code: "POST_NOT_FOUND", Message: "Post not found"}
	}
	if err != nil {
		return err
	}
	action := "UPSERT"
	syncStatus := "PENDING"
	if deleted || status == "TRASHED" || status == "PURGED" {
		action = "DELETE_TOMBSTONE"
	}
	if visibility != "PUBLIC" || status != "PUBLISHED" {
		if action == "UPSERT" {
			syncStatus = "SKIPPED_NOT_PUBLIC"
		}
	}
	payload, _ := json.Marshal(map[string]any{
		"post_id": postID, "blog_id": blogID, "title": title, "slug": slug,
		"visibility": visibility, "status": status, "action": action,
	})
	sum := sha256.Sum256(payload)
	_, err = s.pool.Exec(ctx, `
		INSERT INTO treescan_sync_state (post_id, blog_id, sync_action, visibility_snapshot, payload_hash, sync_status)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (post_id) DO UPDATE SET
			blog_id = EXCLUDED.blog_id,
			sync_action = EXCLUDED.sync_action,
			visibility_snapshot = EXCLUDED.visibility_snapshot,
			payload_hash = EXCLUDED.payload_hash,
			sync_status = CASE
				WHEN treescan_sync_state.payload_hash = EXCLUDED.payload_hash
				 AND treescan_sync_state.sync_status = 'SYNCED'
				THEN 'SYNCED' ELSE EXCLUDED.sync_status END,
			updated_at = NOW()
	`, postID, blogID, action, visibility, hex.EncodeToString(sum[:]), syncStatus)
	return err
}

// Export returns feed entries after a cursor.
func (s *Service) Export(ctx context.Context, cursor int64, limit int) ([]SyncEntry, int64, error) {
	if err := s.requirePool(); err != nil {
		return nil, cursor, err
	}
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	if cursor < 0 {
		cursor = 0
	}
	rows, err := s.pool.Query(ctx, `
		SELECT t.sequence_cursor, t.post_id, t.blog_id, t.sync_action, t.visibility_snapshot,
		       t.payload_hash, t.sync_status, COALESCE(p.title, ''), COALESCE(p.slug, ''), t.updated_at
		FROM treescan_sync_state t
		LEFT JOIN posts p ON p.id = t.post_id
		WHERE t.sequence_cursor > $1
		ORDER BY t.sequence_cursor ASC
		LIMIT $2
	`, cursor, limit)
	if err != nil {
		return nil, cursor, err
	}
	defer rows.Close()
	entries := make([]SyncEntry, 0)
	next := cursor
	for rows.Next() {
		var entry SyncEntry
		if err := rows.Scan(&entry.SequenceCursor, &entry.PostID, &entry.BlogID, &entry.SyncAction,
			&entry.Visibility, &entry.PayloadHash, &entry.Status,
			&entry.Title, &entry.Slug, &entry.UpdatedAt); err != nil {
			return nil, cursor, err
		}
		entries = append(entries, entry)
		next = entry.SequenceCursor
	}
	if err := rows.Err(); err != nil {
		return nil, cursor, err
	}
	return entries, next, nil
}

// Ack marks every entry up to a cursor as synced.
func (s *Service) Ack(ctx context.Context, cursor int64) (int64, error) {
	if err := s.requirePool(); err != nil {
		return 0, err
	}
	if cursor <= 0 {
		return 0, &Fault{Status: http.StatusBadRequest, Code: "INVALID_CURSOR", Message: "Cursor must be positive"}
	}
	res, err := s.pool.Exec(ctx, `UPDATE treescan_sync_state
		SET sync_status = 'SYNCED', last_synced_at = NOW()
		WHERE sequence_cursor <= $1 AND sync_status IN ('PENDING', 'FAILED')`, cursor)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected(), nil
}

// SyncStatus aggregates feed health.
type SyncStatus struct {
	Pending int64 `json:"pending"`
	Failed  int64 `json:"failed"`
	Synced  int64 `json:"synced"`
	Skipped int64 `json:"skipped_not_public"`
}

// Status returns feed health counters.
func (s *Service) Status(ctx context.Context) (*SyncStatus, error) {
	if err := s.requirePool(); err != nil {
		return nil, err
	}
	var status SyncStatus
	err := s.pool.QueryRow(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE sync_status = 'PENDING'),
			COUNT(*) FILTER (WHERE sync_status = 'FAILED'),
			COUNT(*) FILTER (WHERE sync_status = 'SYNCED'),
			COUNT(*) FILTER (WHERE sync_status = 'SKIPPED_NOT_PUBLIC')
		FROM treescan_sync_state
	`).Scan(&status.Pending, &status.Failed, &status.Synced, &status.Skipped)
	if err != nil {
		return nil, err
	}
	return &status, nil
}
