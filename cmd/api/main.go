package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"treeblog/internal/db"
	"treeblog/internal/platform/config"
	"treeblog/internal/platform/httpserver"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := run(logger); err != nil {
		logger.Error("api stopped", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	pools, err := db.Open(context.Background(), cfg.DatabaseURL, cfg.ReplicaURL)
	if err != nil {
		return err
	}
	defer pools.Close()
	server, err := httpserver.New(cfg.HTTPAddr, nil, pools.Primary.Ping, logger)
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	errCh := make(chan error, 1)
	go func() { errCh <- server.HTTP.ListenAndServe() }()
	logger.Info("api listening", "addr", cfg.HTTPAddr)
	select {
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()
		if err := server.HTTP.Shutdown(shutdownCtx); err != nil {
			return err
		}
		return nil
	}
}
