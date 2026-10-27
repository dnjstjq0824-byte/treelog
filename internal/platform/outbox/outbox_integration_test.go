package outbox_test

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"

	"treeblog/internal/db"
	"treeblog/internal/platform/outbox"
	"treeblog/internal/worker"
)

func TestTransactionalOutbox(t *testing.T) {
	databaseURL := os.Getenv("TREEBLOG_TEST_DATABASE_URL")
	redisURL := os.Getenv("TREEBLOG_TEST_REDIS_URL")
	if databaseURL == "" || redisURL == "" {
		t.Skip("disposable PostgreSQL and Redis test services are required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	pools, err := db.Open(ctx, databaseURL, "")
	if err != nil {
		t.Fatal(err)
	}
	defer pools.Close()
	var databaseName string
	if err := pools.Primary.QueryRow(ctx, "SELECT current_database()").Scan(&databaseName); err != nil || databaseName != "treeblog_test" {
		t.Fatalf("refusing to run outside treeblog_test: %q, %v", databaseName, err)
	}
	if err := (db.Migrator{Pool: pools.Primary}).Up(ctx, os.DirFS("../../../migrations")); err != nil {
		t.Fatal(err)
	}
	redis, err := asynq.ParseRedisURI(redisURL)
	if err != nil {
		t.Fatal(err)
	}
	client := asynq.NewClient(redis)
	defer client.Close()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	delivered := make(chan struct{}, 1)
	processor := &worker.Processor{Pool: pools.Primary, Logger: logger,
		Handlers: map[string]worker.Handler{"system.ping": func(context.Context, json.RawMessage) error {
			delivered <- struct{}{}
			return nil
		}}}
	server := worker.NewServer(redis)
	mux := asynq.NewServeMux()
	mux.Handle("treeblog:event", processor)
	if err := server.Start(mux); err != nil {
		t.Fatal(err)
	}
	defer server.Shutdown()
	event := outbox.Event{AggregateType: "system", AggregateID: uuid.NewString(), Type: "system.ping",
		Payload: json.RawMessage(`{"test":true}`), IdempotencyKey: "test-" + uuid.NewString()}
	tx, err := pools.Primary.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := outbox.PublishTx(ctx, tx, event); err != nil {
		t.Fatal(err)
	}
	if err := tx.Rollback(ctx); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := pools.Primary.QueryRow(ctx, "SELECT COUNT(*) FROM outbox_events WHERE idempotency_key=$1", event.IdempotencyKey).Scan(&count); err != nil || count != 0 {
		t.Fatalf("rolled back event persisted: count=%d err=%v", count, err)
	}
	tx, err = pools.Primary.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	id, err := outbox.PublishTx(ctx, tx, event)
	if err != nil {
		t.Fatal(err)
	}
	repeatedID, err := outbox.PublishTx(ctx, tx, event)
	if err != nil || repeatedID != id {
		t.Fatalf("idempotency failed: %q %q %v", id, repeatedID, err)
	}
	if _, err := outbox.PublishTx(ctx, tx, outbox.Event{AggregateType: event.AggregateType,
		AggregateID: event.AggregateID, Type: event.Type, Payload: json.RawMessage(`{"test":false}`),
		IdempotencyKey: event.IdempotencyKey}); err == nil {
		t.Fatal("conflicting idempotency payload accepted")
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	relay := &outbox.Relay{Pool: pools.Primary, Queue: client, Logger: logger}
	if err := relay.DispatchOnce(ctx); err != nil {
		t.Fatal(err)
	}
	select {
	case <-delivered:
	case <-time.After(10 * time.Second):
		t.Fatal("worker did not receive committed event")
	}
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		var status string
		if err := pools.Primary.QueryRow(ctx, "SELECT status FROM background_jobs WHERE id=$1", id).Scan(&status); err == nil && status == "COMPLETED" {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatal("job did not reach COMPLETED status")
}
