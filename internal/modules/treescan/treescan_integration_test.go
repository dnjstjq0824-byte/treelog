package treescan

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

func TestTreescan_Flow(t *testing.T) {
	svc, cleanup := setupIntegration(t)
	defer cleanup()
	ctx := context.Background()

	userID := uuid.NewString()
	handle := "ts-" + strings.ReplaceAll(uuid.NewString()[:8], "-", "x")
	if _, err := svc.pool.Exec(ctx, `INSERT INTO users (id, email, email_verified, status)
		VALUES ($1, $2, TRUE, 'ACTIVE')`, userID, "ts-"+userID[:8]+"@example.com"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.pool.Exec(ctx, `INSERT INTO user_profiles (user_id, username, display_name)
		VALUES ($1, $2, 'Treescan User')`, userID, handle); err != nil {
		t.Fatal(err)
	}
	blogID := uuid.NewString()
	if _, err := svc.pool.Exec(ctx, `INSERT INTO blogs (id, owner_user_id, handle, title)
		VALUES ($1, $2, $3, 'Treescan Blog')`, blogID, userID, handle); err != nil {
		t.Fatal(err)
	}
	postID := uuid.NewString()
	if _, err := svc.pool.Exec(ctx, `INSERT INTO posts (id, blog_id, author_user_id, slug, title, status, visibility, published_at)
		VALUES ($1, $2, $3, 'ts-post', 'Treescan Post', 'PUBLISHED', 'PUBLIC', NOW())`, postID, blogID, userID); err != nil {
		t.Fatal(err)
	}

	if err := svc.MarkPost(ctx, postID, false); err != nil {
		t.Fatalf("MarkPost: %v", err)
	}
	if err := svc.MarkPost(ctx, uuid.NewString(), false); err == nil {
		t.Fatalf("missing post accepted")
	}

	entries, next, err := svc.Export(ctx, 0, 100)
	if err != nil {
		t.Fatalf("Export: %v", err)
	}
	found := false
	for _, entry := range entries {
		if entry.PostID == postID {
			found = true
			if entry.SyncAction != "UPSERT" || entry.Title != "Treescan Post" || entry.PayloadHash == "" {
				t.Fatalf("entry: %+v", entry)
			}
		}
	}
	if !found || next <= 0 {
		t.Fatalf("export missing entry (found=%v next=%d total=%d)", found, next, len(entries))
	}

	status, err := svc.Status(ctx)
	if err != nil || status.Pending < 1 {
		t.Fatalf("status: %+v %v", status, err)
	}
	acked, err := svc.Ack(ctx, next)
	if err != nil || acked < 1 {
		t.Fatalf("ack: %d %v", acked, err)
	}
	if _, err := svc.Ack(ctx, 0); err == nil {
		t.Fatalf("zero ack accepted")
	}
}
