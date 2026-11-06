package notification

import (
	"context"
	"os"
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

func seedUser(t *testing.T, svc *Service, ctx context.Context) string {
	t.Helper()
	userID := uuid.NewString()
	email := "notify-" + userID[:8] + "@example.com"
	username := "notify-" + userID[:8]
	if _, err := svc.pool.Exec(ctx, `INSERT INTO users (id, email, email_verified, status)
		VALUES ($1, $2, TRUE, 'ACTIVE')`, userID, email); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.pool.Exec(ctx, `INSERT INTO user_profiles (user_id, username, display_name)
		VALUES ($1, $2, 'Notify User')`, userID, username); err != nil {
		t.Fatal(err)
	}
	return userID
}

func TestNotification_Flow(t *testing.T) {
	svc, cleanup := setupIntegration(t)
	defer cleanup()
	ctx := context.Background()
	userID := seedUser(t, svc, ctx)
	suffix := uuid.NewString()[:8]

	for i := 0; i < 3; i++ {
		if _, err := svc.pool.Exec(ctx, `INSERT INTO notifications
			(recipient_user_id, category, group_key, dedup_key, title, body, action_url)
			VALUES ($1, 'SYSTEM', 'test', $2, 'title', 'body', '/')`,
			userID, "notify-test-"+suffix+"-"+string(rune('a'+i))); err != nil {
			t.Fatal(err)
		}
	}

	count, err := svc.UnreadCount(ctx, userID)
	if err != nil {
		t.Fatalf("UnreadCount: %v", err)
	}
	if count != 3 {
		t.Fatalf("unread = %d, want 3", count)
	}

	page, next, err := svc.List(ctx, userID, 2, "", false)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(page) != 2 || next == "" {
		t.Fatalf("page = %d items, next = %q", len(page), next)
	}
	page2, next2, err := svc.List(ctx, userID, 2, next, false)
	if err != nil {
		t.Fatalf("List page 2: %v", err)
	}
	if len(page2) != 1 || next2 != "" {
		t.Fatalf("page2 = %d items, next = %q", len(page2), next2)
	}

	if err := svc.MarkRead(ctx, userID, page[0].ID); err != nil {
		t.Fatalf("MarkRead: %v", err)
	}
	count, _ = svc.UnreadCount(ctx, userID)
	if count != 2 {
		t.Fatalf("unread after mark = %d, want 2", count)
	}
	if err := svc.MarkRead(ctx, userID, uuid.NewString()); err == nil {
		t.Fatalf("missing notification accepted")
	}

	if err := svc.MarkAllRead(ctx, userID); err != nil {
		t.Fatalf("MarkAllRead: %v", err)
	}
	count, _ = svc.UnreadCount(ctx, userID)
	if count != 0 {
		t.Fatalf("unread after mark-all = %d, want 0", count)
	}
}
