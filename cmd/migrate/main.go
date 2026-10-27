package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"treeblog/internal/db"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := run(); err != nil {
		logger.Error("migration failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) != 2 || (os.Args[1] != "up" && os.Args[1] != "down-one") {
		return fmt.Errorf("usage: migrate [up|down-one]")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	pools, err := db.Open(ctx, os.Getenv("TREEBLOG_DATABASE_URL"), "")
	if err != nil {
		return err
	}
	defer pools.Close()
	files := os.DirFS("migrations")
	migrator := db.Migrator{Pool: pools.Primary}
	if os.Args[1] == "up" {
		return migrator.Up(ctx, files)
	}
	return migrator.DownOne(ctx, files)
}
