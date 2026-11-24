package media

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"treeblog/internal/db"
	"treeblog/internal/platform/s3"
)

func setupDeliveryIntegration(t *testing.T) (*Service, func()) {
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
	svc, err := NewService(pools.Primary, StorageConfig{
		S3:           s3.Config{Endpoint: "http://localhost:9000", Region: "us-east-1", AccessKey: "test", SecretKey: "test"},
		BucketRaw:    "treeblog-raw-private",
		BucketPublic: "treeblog-media-public",
	})
	if err != nil {
		pools.Close()
		t.Fatal(err)
	}
	return svc, func() { pools.Close() }
}

func seedReadyAsset(t *testing.T, svc *Service, ctx context.Context) (blogID, userID, assetID string) {
	t.Helper()
	userID = uuid.NewString()
	blogID = uuid.NewString()
	handle := "md-" + strings.ReplaceAll(uuid.NewString()[:8], "-", "x")
	exec := func(query string, args ...any) {
		t.Helper()
		if _, err := svc.pool.Exec(ctx, query, args...); err != nil {
			t.Fatal(err)
		}
	}
	exec(`INSERT INTO users (id, email, email_verified, status) VALUES ($1, $2, TRUE, 'ACTIVE')`,
		userID, "md-"+userID[:8]+"@example.com")
	exec(`INSERT INTO blogs (id, owner_user_id, handle, title) VALUES ($1, $2, $3, 'Media Blog')`,
		blogID, userID, handle)
	exec(`INSERT INTO blog_members (blog_id, user_id, role) VALUES ($1, $2, 'OWNER')`, blogID, userID)
	sessionID := uuid.NewString()
	exec(`INSERT INTO upload_sessions (id, blog_id, user_id, original_filename, mime_type, expected_size_bytes, sha256_hash, storage_bucket, storage_object_key, status, expires_at)
		VALUES ($1, $2, $3, 'a.png', 'image/png', 100, $4, 'treeblog-raw-private', $5, 'COMPLETED', NOW() + INTERVAL '1 hour')`,
		sessionID, blogID, userID, strings.Repeat("a", 64), "raw/"+sessionID+"/a.png")
	assetID = uuid.NewString()
	exec(`INSERT INTO media_assets (id, blog_id, uploader_user_id, upload_session_id, media_type, mime_type, sha256_hash, width, height, status)
		VALUES ($1, $2, $3, $4, 'IMAGE', 'image/png', $5, 640, 480, 'READY')`,
		assetID, blogID, userID, sessionID, strings.Repeat("b", 64))
	exec(`INSERT INTO media_variants (media_asset_id, variant_name, format, width, height, size_bytes, public_cdn_url)
		VALUES ($1, 'W640', 'JPEG', 640, 480, 1000, '/x.jpg'), ($1, 'W640', 'WEBP', 640, 480, 500, '/x.webp')`, assetID)
	return blogID, userID, assetID
}

func faultCode(err error) string {
	if fault, ok := err.(*Fault); ok {
		return fault.Code
	}
	return ""
}

// TestDeliveryFlow covers G-12 serving: format-aware variant resolution,
// the srcset list endpoint data, and the author-only original download.
func TestDeliveryFlow(t *testing.T) {
	svc, cleanup := setupDeliveryIntegration(t)
	defer cleanup()
	ctx := context.Background()
	blogID, userID, assetID := seedReadyAsset(t, svc, ctx)

	bucket, key, contentType, err := svc.PublicObject(ctx, assetID, "W640", "")
	if err != nil || bucket != "treeblog-media-public" || contentType != "image/jpeg" || !strings.HasSuffix(key, ".jpg") {
		t.Fatalf("jpeg default: %s %s %s / %v", bucket, key, contentType, err)
	}
	_, key, contentType, err = svc.PublicObject(ctx, assetID, "w640", "webp")
	if err != nil || contentType != "image/webp" || !strings.HasSuffix(key, ".webp") {
		t.Fatalf("webp twin: %s %s / %v", key, contentType, err)
	}
	if _, _, _, err := svc.PublicObject(ctx, assetID, "W640", "avif"); faultCode(err) != "INVALID_VARIANT_FORMAT" {
		t.Fatalf("expected INVALID_VARIANT_FORMAT, got %v", err)
	}
	if _, _, _, err := svc.PublicObject(ctx, assetID, "W1200", ""); faultCode(err) != "VARIANT_NOT_FOUND" {
		t.Fatalf("expected VARIANT_NOT_FOUND, got %v", err)
	}

	variants, err := svc.ListVariants(ctx, assetID)
	if err != nil || len(variants) != 2 {
		t.Fatalf("variants: %+v / %v", variants, err)
	}

	stranger := uuid.NewString()
	if _, _, _, err := svc.OriginalObject(ctx, blogID, assetID, stranger); faultCode(err) != "FORBIDDEN" && faultCode(err) != "INSUFFICIENT_PERMISSIONS" {
		t.Fatalf("expected author-only original, got %v", err)
	}
	bucket, key, contentType, err = svc.OriginalObject(ctx, blogID, assetID, userID)
	if err != nil || bucket != "treeblog-raw-private" || contentType != "image/png" {
		t.Fatalf("original: %s %s %s / %v", bucket, key, contentType, err)
	}
}
