package admin

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"treeblog/internal/db"
)

func setupIntegration(t *testing.T) (*Service, func()) {
	t.Helper()
	url := os.Getenv("TREEBLOG_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TREEBLOG_TEST_DATABASE_URL is not configured")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	pools, err := db.Open(ctx, url, "")
	if err != nil {
		t.Fatal(err)
	}
	if err := (db.Migrator{Pool: pools.Primary}).Up(ctx, os.DirFS("../../../migrations")); err != nil {
		pools.Close()
		t.Fatal(err)
	}
	svc, err := NewService(pools.Primary)
	if err != nil {
		pools.Close()
		t.Fatal(err)
	}
	return svc, func() { pools.Close() }
}

func TestAdmin_Flow(t *testing.T) {
	svc, cleanup := setupIntegration(t)
	defer cleanup()
	ctx := context.Background()

	userID := uuid.NewString()
	handle := "ad-" + strings.ReplaceAll(uuid.NewString()[:8], "-", "x")
	email := "admin-" + userID[:8] + "@example.com"
	if _, err := svc.pool.Exec(ctx, `INSERT INTO users (id, email, email_verified, status)
		VALUES ($1, $2, TRUE, 'ACTIVE')`, userID, email); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.pool.Exec(ctx, `INSERT INTO user_profiles (user_id, username, display_name)
		VALUES ($1, $2, 'Admin Target')`, userID, handle); err != nil {
		t.Fatal(err)
	}

	stats, err := svc.Stats(ctx)
	if err != nil {
		t.Fatalf("Stats: %v", err)
	}
	if stats.UsersTotal < 1 || stats.UsersActive < 1 {
		t.Fatalf("stats: %+v", stats)
	}

	users, err := svc.ListUsers(ctx, email, "", 10, 0)
	if err != nil || len(users) != 1 || users[0].ID != userID {
		t.Fatalf("ListUsers: %v %+v", err, users)
	}

	if err := svc.SuspendUser(ctx, userID); err != nil {
		t.Fatalf("SuspendUser: %v", err)
	}
	if err := svc.SuspendUser(ctx, userID); err == nil {
		t.Fatalf("double suspend accepted")
	}
	suspended, err := svc.ListUsers(ctx, email, "SUSPENDED", 10, 0)
	if err != nil || len(suspended) != 1 {
		t.Fatalf("ListUsers suspended: %v %d", err, len(suspended))
	}
	if err := svc.UnsuspendUser(ctx, userID); err != nil {
		t.Fatalf("UnsuspendUser: %v", err)
	}
	if err := svc.UnsuspendUser(ctx, userID); err == nil {
		t.Fatalf("double unsuspend accepted")
	}
	if err := svc.SuspendUser(ctx, uuid.NewString()); err == nil {
		t.Fatalf("missing user suspend accepted")
	}

	blogs, err := svc.ListBlogs(ctx, "", 10, 0)
	if err != nil {
		t.Fatalf("ListBlogs: %v", err)
	}
	_ = blogs
}
