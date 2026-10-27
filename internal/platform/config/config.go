package config

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
)

type Config struct {
	HTTPAddr        string
	DatabaseURL     string
	ReplicaURL      string
	ShutdownTimeout time.Duration
}

func Load() (Config, error) {
	addr := os.Getenv("TREEBLOG_HTTP_ADDR")
	if addr == "" {
		return Config{}, errors.New("TREEBLOG_HTTP_ADDR is required")
	}
	if _, _, err := net.SplitHostPort(addr); err != nil {
		return Config{}, fmt.Errorf("invalid TREEBLOG_HTTP_ADDR: %w", err)
	}
	databaseURL := os.Getenv("TREEBLOG_DATABASE_URL")
	if databaseURL == "" {
		return Config{}, errors.New("TREEBLOG_DATABASE_URL is required")
	}
	seconds := 15
	if raw := os.Getenv("TREEBLOG_SHUTDOWN_SECONDS"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 1 || value > 300 {
			return Config{}, errors.New("TREEBLOG_SHUTDOWN_SECONDS must be between 1 and 300")
		}
		seconds = value
	}
	return Config{HTTPAddr: addr, DatabaseURL: databaseURL, ReplicaURL: os.Getenv("TREEBLOG_REPLICA_URL"), ShutdownTimeout: time.Duration(seconds) * time.Second}, nil
}
