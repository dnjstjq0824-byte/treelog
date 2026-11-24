package seo

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func seedMetaPost(t *testing.T, pool *pgxpool.Pool, ctx context.Context, visibility string) (handle, slug string) {
	t.Helper()
	userID := uuid.NewString()
	handle = "pm-" + strings.ReplaceAll(uuid.NewString()[:8], "-", "x")
	blogID := uuid.NewString()
	slug = "meta-post-" + uuid.NewString()[:8]
	exec := func(query string, args ...any) {
		t.Helper()
		if _, err := pool.Exec(ctx, query, args...); err != nil {
			t.Fatal(err)
		}
	}
	exec(`INSERT INTO users (id, email, email_verified, status) VALUES ($1, $2, TRUE, 'ACTIVE')`,
		userID, "pm-"+userID[:8]+"@example.com")
	exec(`INSERT INTO user_profiles (user_id, username, display_name) VALUES ($1, $2, 'Meta Author')`, userID, handle)
	exec(`INSERT INTO blogs (id, owner_user_id, handle, title) VALUES ($1, $2, $3, 'Meta Blog')`, blogID, userID, handle)
	exec(`INSERT INTO posts (blog_id, author_user_id, slug, title, excerpt, status, visibility, published_at)
		VALUES ($1, $2, $3, 'Meta Title', 'Meta excerpt', 'PUBLISHED', $4, NOW())`, blogID, userID, slug, visibility)
	return handle, slug
}

// TestPostMetaVisibility pins the G-15 crawler matrix: public posts index
// with JSON-LD, restricted posts noindex, private posts 404.
func TestPostMetaVisibility(t *testing.T) {
	svc, cleanup := setupIntegration(t)
	defer cleanup()
	ctx := context.Background()

	handle, slug := seedMetaPost(t, svc.pool, ctx, "PUBLIC")
	meta, err := svc.PostMeta(ctx, handle, slug)
	if err != nil {
		t.Fatalf("public meta: %v", err)
	}
	if meta.Robots != "index,follow" {
		t.Errorf("public robots = %q", meta.Robots)
	}
	if len(meta.JSONLD) != 2 || !strings.HasSuffix(meta.Canonical, "/b/"+handle+"/"+slug) {
		t.Errorf("public meta: %+v", meta)
	}

	unlistedHandle, unlistedSlug := seedMetaPost(t, svc.pool, ctx, "UNLISTED")
	meta, err = svc.PostMeta(ctx, unlistedHandle, unlistedSlug)
	if err != nil {
		t.Fatalf("unlisted meta: %v", err)
	}
	if meta.Robots != "noindex,nofollow" {
		t.Errorf("unlisted robots = %q", meta.Robots)
	}

	privateHandle, privateSlug := seedMetaPost(t, svc.pool, ctx, "PRIVATE")
	if _, err := svc.PostMeta(ctx, privateHandle, privateSlug); err == nil {
		t.Fatal("private post must 404 for crawlers")
	}
	if _, err := svc.PostMeta(ctx, handle, "no-such-slug"); err == nil {
		t.Fatal("missing post must 404")
	}
}
