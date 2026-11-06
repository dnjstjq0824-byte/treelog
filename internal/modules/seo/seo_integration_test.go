package seo

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
	svc, err := NewService(pools.Primary, "https://seo-test.example.com")
	if err != nil {
		pools.Close()
		t.Fatal(err)
	}
	return svc, func() { pools.Close() }
}

func TestSEO_Flow(t *testing.T) {
	svc, cleanup := setupIntegration(t)
	defer cleanup()
	ctx := context.Background()

	userID := uuid.NewString()
	handle := "seo-" + strings.ReplaceAll(uuid.NewString()[:8], "-", "x")
	if _, err := svc.pool.Exec(ctx, `INSERT INTO users (id, email, email_verified, status)
		VALUES ($1, $2, TRUE, 'ACTIVE')`, userID, "seo-"+userID[:8]+"@example.com"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.pool.Exec(ctx, `INSERT INTO user_profiles (user_id, username, display_name)
		VALUES ($1, $2, 'SEO User')`, userID, handle); err != nil {
		t.Fatal(err)
	}
	blogID := uuid.NewString()
	if _, err := svc.pool.Exec(ctx, `INSERT INTO blogs (id, owner_user_id, handle, title, description)
		VALUES ($1, $2, $3, 'SEO & Blog <Test>', 'desc')`, blogID, userID, handle); err != nil {
		t.Fatal(err)
	}
	postID := uuid.NewString()
	if _, err := svc.pool.Exec(ctx, `INSERT INTO posts (id, blog_id, author_user_id, slug, title, status, visibility, published_at)
		VALUES ($1, $2, $3, 'seo-post', 'SEO Post', 'PUBLISHED', 'PUBLIC', NOW())`, postID, blogID, userID); err != nil {
		t.Fatal(err)
	}

	sitemap, err := svc.Sitemap(ctx)
	if err != nil {
		t.Fatalf("Sitemap: %v", err)
	}
	for _, want := range []string{
		"https://seo-test.example.com/",
		"https://seo-test.example.com/b/" + handle,
		"https://seo-test.example.com/b/" + handle + "/seo-post",
	} {
		if !strings.Contains(sitemap, want) {
			t.Fatalf("sitemap missing %s", want)
		}
	}

	meta, err := svc.BlogMeta(ctx, handle)
	if err != nil {
		t.Fatalf("BlogMeta: %v", err)
	}
	if meta.Canonical != "https://seo-test.example.com/b/"+handle || meta.Title != "SEO & Blog <Test>" {
		t.Fatalf("meta: %+v", meta)
	}
	if _, err := svc.BlogMeta(ctx, "missing-"+handle); err == nil {
		t.Fatalf("missing blog accepted")
	}
}
