package main

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// TestSchedulerIntervalContract pins the J_JOBS_SCHEDULING timing contract:
// due publications must be swept every 5 seconds.
func TestSchedulerIntervalContract(t *testing.T) {
	if schedulerInterval != 5*time.Second {
		t.Errorf("expected 5s scheduler interval, got %v", schedulerInterval)
	}
}

// TestRunSchedulerLoopExitsOnCancel verifies the loop shuts down promptly.
// No tick fires before cancellation, so no database access happens.
func TestRunSchedulerLoopExitsOnCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		runSchedulerLoop(ctx, nil, nil, testLogger(), time.Hour)
	}()
	cancel()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("scheduler loop did not exit after context cancellation")
	}
}
