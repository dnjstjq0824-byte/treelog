package post

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"treeblog/internal/db"
	"treeblog/internal/modules/identity"
)

func setupProtectionIntegration(t *testing.T) (*Service, *pgxpool.Pool, func()) {
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
	return NewService(pools.Primary), pools.Primary, func() { pools.Close() }
}

func seedProtectionDraft(t *testing.T, pool *pgxpool.Pool, ctx context.Context) (authorID, blogID, postID, handle string) {
	t.Helper()
	authorID = uuid.NewString()
	handle = "pp-" + strings.ReplaceAll(uuid.NewString()[:8], "-", "x")
	email := "pp-" + authorID[:8] + "@example.com"
	if _, err := pool.Exec(ctx, `INSERT INTO users (id, email, email_verified, status)
		VALUES ($1, $2, TRUE, 'ACTIVE')`, authorID, email); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO user_profiles (user_id, username, display_name)
		VALUES ($1, $2, 'Protection Author')`, authorID, handle); err != nil {
		t.Fatal(err)
	}
	blogID = uuid.NewString()
	if _, err := pool.Exec(ctx, `INSERT INTO blogs (id, owner_user_id, handle, title)
		VALUES ($1, $2, $3, 'Protection Blog')`, blogID, authorID, handle); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO blog_members (blog_id, user_id, role)
		VALUES ($1, $2, 'OWNER')`, blogID, authorID); err != nil {
		t.Fatal(err)
	}
	postID = uuid.NewString()
	if _, err := pool.Exec(ctx, `INSERT INTO posts (id, blog_id, author_user_id, slug, title, status)
		VALUES ($1, $2, $3, 'secret-post', 'Secret', 'DRAFT')`, postID, blogID, authorID); err != nil {
		t.Fatal(err)
	}
	return authorID, blogID, postID, handle
}

func storedProtectionHash(t *testing.T, pool *pgxpool.Pool, ctx context.Context, postID string) *string {
	t.Helper()
	var hash *string
	if err := pool.QueryRow(ctx, `SELECT protection_password_hash FROM posts WHERE id = $1`, postID).Scan(&hash); err != nil {
		t.Fatal(err)
	}
	return hash
}

func faultCode(err error) string {
	if fault, ok := err.(*Fault); ok {
		return fault.Code
	}
	return ""
}

// TestProtectedPostFlow covers the G-07 lifecycle end to end: publishing
// with and without a password, locked reads, wrong-password rejection,
// author bypass, and hash clearing on visibility change.
func TestProtectedPostFlow(t *testing.T) {
	svc, pool, cleanup := setupProtectionIntegration(t)
	defer cleanup()
	ctx := context.Background()
	authorID, blogID, postID, handle := seedProtectionDraft(t, pool, ctx)
	strangerID := uuid.NewString()

	content := json.RawMessage(`{"type":"doc","content":[]}`)
	base := PublishPostInput{Slug: "secret-post", ContentJSON: content, ContentHTML: "<p>hi</p>", ContentPlain: "hi"}

	// PROTECTED without any password is rejected.
	base.Title, base.Visibility = "Secret", "PROTECTED"
	if _, err := svc.PublishPost(ctx, blogID, postID, authorID, base); faultCode(err) != "PROTECTION_PASSWORD_REQUIRED" {
		t.Fatalf("expected PROTECTION_PASSWORD_REQUIRED, got %v", err)
	}

	// PROTECTED with a password stores an Argon2id hash, never plaintext.
	base.ProtectionPassword = strPtr("s3cret-unlock")
	if _, err := svc.PublishPost(ctx, blogID, postID, authorID, base); err != nil {
		t.Fatalf("publish protected: %v", err)
	}
	hash := storedProtectionHash(t, pool, ctx, postID)
	if hash == nil || *hash == "s3cret-unlock" || !identity.VerifyPassword("s3cret-unlock", *hash) {
		t.Fatal("expected a verifiable non-plaintext hash in storage")
	}

	// Locked reads: missing password, then wrong password.
	if _, err := svc.GetPublicPost(ctx, handle, "secret-post", strangerID, ""); faultCode(err) != "POST_PASSWORD_REQUIRED" {
		t.Fatalf("expected POST_PASSWORD_REQUIRED, got %v", err)
	}
	if _, err := svc.GetPublicPost(ctx, handle, "secret-post", strangerID, "nope-nope"); faultCode(err) != "POST_PASSWORD_INVALID" {
		t.Fatalf("expected POST_PASSWORD_INVALID, got %v", err)
	}

	// Correct password unlocks; the author bypasses the lock.
	unlocked, err := svc.GetPublicPost(ctx, handle, "secret-post", strangerID, "s3cret-unlock")
	if err != nil || unlocked.Title != "Secret" {
		t.Fatalf("expected unlocked post, got %+v / %v", unlocked, err)
	}
	if _, err := svc.GetPublicPost(ctx, handle, "secret-post", authorID, ""); err != nil {
		t.Fatalf("author should bypass the lock: %v", err)
	}

	// Republishing as PROTECTED without a password keeps the stored hash.
	base.Title, base.ProtectionPassword = "Secret v2", nil
	if _, err := svc.PublishPost(ctx, blogID, postID, authorID, base); err != nil {
		t.Fatalf("republish keeping hash: %v", err)
	}
	if _, err := svc.GetPublicPost(ctx, handle, "secret-post", strangerID, "s3cret-unlock"); err != nil {
		t.Fatalf("original password should still unlock: %v", err)
	}

	// Switching away from PROTECTED clears the hash.
	base.Title, base.Visibility = "Secret v3", "PUBLIC"
	if _, err := svc.PublishPost(ctx, blogID, postID, authorID, base); err != nil {
		t.Fatalf("publish public: %v", err)
	}
	if hash := storedProtectionHash(t, pool, ctx, postID); hash != nil {
		t.Fatal("expected protection hash to be cleared on PUBLIC")
	}
	if _, err := svc.GetPublicPost(ctx, handle, "secret-post", strangerID, ""); err != nil {
		t.Fatalf("public post should need no password: %v", err)
	}
}

func strPtr(s string) *string { return &s }

// TestGonePostReturns410 pins the G-15 crawler contract: trashed posts
// report 410 Gone instead of a retried 404.
func TestGonePostReturns410(t *testing.T) {
	svc, pool, cleanup := setupProtectionIntegration(t)
	defer cleanup()
	ctx := context.Background()
	_, _, postID, handle := seedProtectionDraft(t, pool, ctx)

	if _, err := pool.Exec(ctx, `UPDATE posts SET status = 'TRASHED', trashed_at = NOW() WHERE id = $1`, postID); err != nil {
		t.Fatal(err)
	}
	var slug string
	if err := pool.QueryRow(ctx, `SELECT slug FROM posts WHERE id = $1`, postID).Scan(&slug); err != nil {
		t.Fatal(err)
	}
	_, err := svc.GetPublicPost(ctx, handle, slug, "", "")
	if fault, ok := err.(*Fault); !ok || fault.Status != http.StatusGone || fault.Code != "POST_GONE" {
		t.Fatalf("expected 410 POST_GONE, got %v", err)
	}
}
