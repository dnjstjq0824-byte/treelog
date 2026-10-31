package post

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Scheduler struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

func NewScheduler(pool *pgxpool.Pool, logger *slog.Logger) *Scheduler {
	return &Scheduler{pool: pool, logger: logger}
}

// ProcessScheduledPublications runs a batch scan for posts due for publication.
func (s *Scheduler) ProcessScheduledPublications(ctx context.Context) (int, error) {
	if s.pool == nil {
		return 0, nil
	}

	rows, err := s.pool.Query(ctx, `
		SELECT id, post_id, blog_id, scheduled_by
		FROM scheduled_publications
		WHERE status = 'PENDING' AND scheduled_for <= NOW()
		FOR UPDATE SKIP LOCKED
		LIMIT 50
	`)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	type duePub struct {
		ID          string
		PostID      string
		BlogID      string
		ScheduledBy string
	}
	dues := make([]duePub, 0)
	for rows.Next() {
		var d duePub
		if err := rows.Scan(&d.ID, &d.PostID, &d.BlogID, &d.ScheduledBy); err == nil {
			dues = append(dues, d)
		}
	}
	rows.Close()

	count := 0
	now := time.Now().UTC()

	for _, d := range dues {
		tx, err := s.pool.Begin(ctx)
		if err != nil {
			continue
		}

		// Update post status to PUBLISHED
		var title, slug string
		var revNo int
		var cJSON json.RawMessage
		var cHTML string
		err = tx.QueryRow(ctx, `
			UPDATE posts
			SET status = 'PUBLISHED', published_at = $1, revision_no = revision_no + 1, updated_at = NOW()
			WHERE id = $2 AND blog_id = $3
			RETURNING title, slug, revision_no, content_json, content_html
		`, now, d.PostID, d.BlogID).Scan(&title, &slug, &revNo, &cJSON, &cHTML)
		if err != nil {
			tx.Rollback(ctx)
			continue
		}

		// Update scheduled_publications status to COMPLETED
		_, err = tx.Exec(ctx, `
			UPDATE scheduled_publications
			SET status = 'COMPLETED', executed_at = $1, updated_at = NOW()
			WHERE id = $2
		`, now, d.ID)
		if err != nil {
			tx.Rollback(ctx)
			continue
		}

		// Save revision snapshot
		_, _ = tx.Exec(ctx, `
			INSERT INTO post_revisions (
				post_id, revision_no, editor_user_id, title, subtitle,
				content_json, content_html, change_summary, is_published_snapshot
			) VALUES ($1, $2, $3, $4, '', $5, $6, 'Scheduled publication snapshot', TRUE)
			ON CONFLICT (post_id, revision_no) DO NOTHING
		`, d.PostID, revNo, d.ScheduledBy, title, cJSON, cHTML)

		// Search index job
		_, _ = tx.Exec(ctx, `
			INSERT INTO search_index_jobs (target_index, document_id, operation, priority)
			VALUES ('POSTS', $1, 'UPSERT', 80)
		`, d.PostID)

		// Outbox event
		payload, _ := json.Marshal(map[string]any{
			"post_id":      d.PostID,
			"blog_id":      d.BlogID,
			"slug":         slug,
			"title":        title,
			"published_at": now.Format(time.RFC3339),
		})
		_, _ = tx.Exec(ctx, `
			INSERT INTO outbox_events (aggregate_type, aggregate_id, event_type, payload_json, idempotency_key)
			VALUES ('POST', $1, 'post.published', $2, 'evt-post-sched-' || $1 || '-' || $3)
		`, d.PostID, payload, revNo)

		if err := tx.Commit(ctx); err == nil {
			count++
			if s.logger != nil {
				s.logger.Info("scheduled post published successfully", "post_id", d.PostID, "blog_id", d.BlogID)
			}
		}
	}

	return count, nil
}
