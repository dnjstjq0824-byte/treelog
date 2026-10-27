package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"

	"treeblog/internal/platform/outbox"
)

type Handler func(context.Context, json.RawMessage) error

type Processor struct {
	Pool     *pgxpool.Pool
	Logger   *slog.Logger
	Handlers map[string]Handler
}

func (p *Processor) ProcessTask(ctx context.Context, task *asynq.Task) error {
	if p.Pool == nil || p.Logger == nil {
		return errors.New("worker dependencies are required")
	}
	var event outbox.Event
	if err := json.Unmarshal(task.Payload(), &event); err != nil {
		return fmt.Errorf("invalid event task: %w", err)
	}
	if event.ID == "" || event.Type == "" {
		return errors.New("event id and type are required")
	}
	var status string
	if err := p.Pool.QueryRow(ctx, "SELECT status FROM background_jobs WHERE id=$1", event.ID).Scan(&status); err != nil {
		return err
	}
	if status == "COMPLETED" {
		return nil
	}
	if _, err := p.Pool.Exec(ctx, `UPDATE background_jobs
		SET status='EXECUTING', attempt_count=attempt_count+1, started_at=NOW()
		WHERE id=$1`, event.ID); err != nil {
		return err
	}
	handler, ok := p.Handlers[event.Type]
	var err error
	if !ok {
		err = fmt.Errorf("no handler for event type %s", event.Type)
	} else {
		err = handler(ctx, event.Payload)
	}
	if err == nil {
		_, updateErr := p.Pool.Exec(ctx, `UPDATE background_jobs
			SET status='COMPLETED', completed_at=NOW(), last_error=NULL WHERE id=$1`, event.ID)
		return updateErr
	}
	status = "FAILED"
	retryCount, hasRetry := asynq.GetRetryCount(ctx)
	maxRetry, hasMax := asynq.GetMaxRetry(ctx)
	if hasRetry && hasMax && retryCount >= maxRetry {
		status = "DEAD_LETTER"
	}
	if _, updateErr := p.Pool.Exec(ctx, `UPDATE background_jobs
		SET status=$2, last_error=$3 WHERE id=$1`, event.ID, status, err.Error()); updateErr != nil {
		return errors.Join(err, updateErr)
	}
	p.Logger.Warn("job failed", "event_id", event.ID, "event_type", event.Type, "error", err)
	return err
}

func NewServer(redis asynq.RedisConnOpt) *asynq.Server {
	return asynq.NewServer(redis, asynq.Config{
		Concurrency: 10,
		Queues:      map[string]int{"critical": 8, "default": 4, "media": 2, "bulk": 1},
	})
}
