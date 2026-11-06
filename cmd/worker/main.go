package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"

	"treeblog/internal/db"
	"treeblog/internal/modules/media"
	"treeblog/internal/modules/moderation"
	"treeblog/internal/modules/post"
	"treeblog/internal/modules/treescan"
	"treeblog/internal/platform/opensearch"
	"treeblog/internal/platform/outbox"
	"treeblog/internal/platform/s3"
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

	osClient := opensearch.New(opensearch.Config{
		Address:     os.Getenv("TREEBLOG_OPENSEARCH_URL"),
		IndexPrefix: envOr("TREEBLOG_OPENSEARCH_PREFIX", "treeblog"),
	})
	if osClient.Enabled() {
		indexCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		if err := osClient.EnsureIndices(indexCtx); err != nil {
			logger.Warn("opensearch indices not ready; search falls back to SQL", "error", err)
		}
		cancel()
	}

	var mediaProcessor *media.Processor
	if accessKey := os.Getenv("TREEBLOG_S3_ACCESS_KEY"); accessKey != "" {
		s3Client, err := s3.New(s3.Config{
			Endpoint:  envOr("TREEBLOG_S3_ENDPOINT", "http://localhost:9000"),
			Region:    envOr("TREEBLOG_S3_REGION", "us-east-1"),
			AccessKey: accessKey,
			SecretKey: os.Getenv("TREEBLOG_S3_SECRET_KEY"),
		})
		if err != nil {
			return err
		}
		mediaProcessor = &media.Processor{
			Pool:         pools.Primary,
			S3:           s3Client,
			BucketRaw:    envOr("TREEBLOG_S3_BUCKET_RAW", "treeblog-raw-private"),
			BucketPublic: envOr("TREEBLOG_S3_BUCKET_PUBLIC", "treeblog-media-public"),
			Logger:       logger,
		}
	} else {
		logger.Warn("s3 credentials missing; media.process jobs will fail until configured")
	}

	treescanService, err := treescan.NewService(pools.Primary)
	if err != nil {
		return err
	}
	processor := &worker.Processor{Pool: pools.Primary, Logger: logger,
		Handlers: worker.DefaultHandlers(worker.Dependencies{
			Pool:       pools.Primary,
			Logger:     logger,
			OpenSearch: osClient,
			Media:      mediaProcessor,
			Treescan:   treescanService,
			AppBaseURL: envOr("TREEBLOG_APP_BASE_URL", "http://localhost:3000"),
		})}
	server := worker.NewServer(redis)
	mux := asynq.NewServeMux()
	mux.Handle("treeblog:event", processor)
	if err := server.Start(mux); err != nil {
		return err
	}
	defer server.Shutdown()

	modService, err := moderation.NewService(pools.Primary)
	if err != nil {
		return err
	}
	go runSchedulerLoop(ctx, pools.Primary, modService, logger)

	relay := &outbox.Relay{Pool: pools.Primary, Queue: client, Logger: logger}
	return relay.Run(ctx)
}

// runSchedulerLoop executes due scheduled publications and expired
// suspension resumptions every 30 seconds.
func runSchedulerLoop(ctx context.Context, pool *pgxpool.Pool, modService *moderation.Service, logger *slog.Logger) {
	scheduler := post.NewScheduler(pool, logger)
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			runCtx, cancel := context.WithTimeout(ctx, 25*time.Second)
			count, err := scheduler.ProcessScheduledPublications(runCtx)
			resumed, resumeErr := modService.ResumeExpiredSuspensions(runCtx)
			cancel()
			if err != nil {
				logger.Warn("scheduled publication sweep failed", "error", err)
			} else if count > 0 {
				logger.Info("scheduled publications executed", "count", count)
			}
			if resumeErr != nil {
				logger.Warn("suspension resume sweep failed", "error", resumeErr)
			} else if resumed > 0 {
				logger.Info("expired suspensions resumed", "count", resumed)
			}
		}
	}
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
