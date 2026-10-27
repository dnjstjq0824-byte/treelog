package outbox

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Enqueuer interface {
	EnqueueContext(context.Context, *asynq.Task, ...asynq.Option) (*asynq.TaskInfo, error)
}

type Relay struct {
	Pool     *pgxpool.Pool
	Queue    Enqueuer
	Logger   *slog.Logger
	Interval time.Duration
}

type claimedEvent struct {
	Event
	RetryCount int
	QueueName  string
}

func QueueFor(eventType string) string {
	switch {
	case strings.HasPrefix(eventType, "post.deleted"), strings.HasPrefix(eventType, "post.visibility"),
		strings.HasPrefix(eventType, "moderation."), strings.HasPrefix(eventType, "search.delete"):
		return "critical"
	case strings.HasPrefix(eventType, "media."):
		return "media"
	case strings.HasPrefix(eventType, "import."), strings.HasPrefix(eventType, "export."),
		strings.HasPrefix(eventType, "analytics."):
		return "bulk"
	default:
		return "default"
	}
}

func (r *Relay) Run(ctx context.Context) error {
	if r.Pool == nil || r.Queue == nil || r.Logger == nil {
		return errors.New("relay dependencies are required")
	}
	interval := r.Interval
	if interval <= 0 {
		interval = 100 * time.Millisecond
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		if err := r.DispatchOnce(ctx); err != nil && ctx.Err() == nil {
			r.Logger.Error("outbox dispatch failed", "error", err)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

func (r *Relay) DispatchOnce(ctx context.Context) error {
	events, err := r.claim(ctx, 20)
	if err != nil {
		return err
	}
	for _, event := range events {
		payload, err := json.Marshal(event.Event)
		if err != nil {
			return err
		}
		task := asynq.NewTask("treeblog:event", payload)
		_, err = r.Queue.EnqueueContext(ctx, task,
			asynq.TaskID(event.ID), asynq.Queue(event.QueueName), asynq.MaxRetry(7))
		if errors.Is(err, asynq.ErrTaskIDConflict) {
			err = nil
		}
		if err != nil {
			if updateErr := r.recordFailure(ctx, event, err); updateErr != nil {
				return fmt.Errorf("enqueue: %v; record failure: %w", err, updateErr)
			}
			continue
		}
		if _, err := r.Pool.Exec(ctx, `UPDATE outbox_events SET status='PUBLISHED', published_at=NOW()
			WHERE id=$1 AND status='PROCESSING'`, event.ID); err != nil {
			return err
		}
	}
	return nil
}

func (r *Relay) claim(ctx context.Context, limit int) ([]claimedEvent, error) {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	rows, err := tx.Query(ctx, `WITH candidate AS (
		SELECT id FROM outbox_events
		WHERE status IN ('PENDING','PROCESSING') AND next_attempt_at <= NOW()
		ORDER BY next_attempt_at, created_at FOR UPDATE SKIP LOCKED LIMIT $1
	) UPDATE outbox_events o SET status='PROCESSING', next_attempt_at=NOW()+INTERVAL '1 minute'
	FROM candidate c WHERE o.id=c.id
	RETURNING o.id, o.aggregate_type, o.aggregate_id, o.event_type, o.payload_json,
		o.idempotency_key, o.retry_count`, limit)
	if err != nil {
		return nil, err
	}
	var events []claimedEvent
	for rows.Next() {
		var event claimedEvent
		if err := rows.Scan(&event.ID, &event.AggregateType, &event.AggregateID,
			&event.Type, &event.Payload, &event.IdempotencyKey, &event.RetryCount); err != nil {
			rows.Close()
			return nil, err
		}
		event.QueueName = QueueFor(event.Type)
		events = append(events, event)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	for _, event := range events {
		if _, err := tx.Exec(ctx, `INSERT INTO background_jobs
			(id, queue_name, job_type, payload_json, idempotency_key)
			VALUES($1,$2,$3,$4,$5) ON CONFLICT (idempotency_key) DO NOTHING`,
			event.ID, event.QueueName, event.Type, event.Payload, event.IdempotencyKey); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return events, nil
}

func (r *Relay) recordFailure(ctx context.Context, event claimedEvent, cause error) error {
	attempt := event.RetryCount + 1
	status := "PENDING"
	if attempt >= 8 {
		status = "DEAD_LETTER"
	}
	delay := time.Duration(1<<min(attempt, 8)) * time.Second
	_, err := r.Pool.Exec(ctx, `UPDATE outbox_events
		SET status=$2, retry_count=$3, next_attempt_at=NOW()+$4::interval, last_error=$5
		WHERE id=$1 AND status='PROCESSING'`, event.ID, status, attempt, delay.String(), cause.Error())
	if err != nil {
		return err
	}
	if status == "DEAD_LETTER" {
		_, err = r.Pool.Exec(ctx, `UPDATE background_jobs SET status='DEAD_LETTER',
			last_error=$2 WHERE id=$1 AND status='QUEUED'`, event.ID, cause.Error())
	}
	return err
}
