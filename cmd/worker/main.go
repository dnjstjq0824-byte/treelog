package main

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/hibiken/asynq"

	"treeblog/internal/db"
	"treeblog/internal/platform/outbox"
	"treeblog/internal/worker"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := run(logger); err != nil {
		logger.Error("worker stopped", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	databaseURL, redisURL := os.Getenv("TREEBLOG_DATABASE_URL"), os.Getenv("TREEBLOG_REDIS_URL")
	if databaseURL == "" || redisURL == "" {
		return errors.New("TREEBLOG_DATABASE_URL and TREEBLOG_REDIS_URL are required")
	}
	redis, err := asynq.ParseRedisURI(redisURL)
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	pools, err := db.Open(ctx, databaseURL, "")
	if err != nil {
		return err
	}
	defer pools.Close()
	client := asynq.NewClient(redis)
	defer client.Close()
	processor := &worker.Processor{Pool: pools.Primary, Logger: logger,
		Handlers: map[string]worker.Handler{"system.ping": func(context.Context, json.RawMessage) error { return nil }}}
	server := worker.NewServer(redis)
	mux := asynq.NewServeMux()
	mux.Handle("treeblog:event", processor)
	if err := server.Start(mux); err != nil {
		return err
	}
	defer server.Shutdown()
	relay := &outbox.Relay{Pool: pools.Primary, Queue: client, Logger: logger}
	return relay.Run(ctx)
}
