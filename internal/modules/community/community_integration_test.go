package community

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

func seedAuthor(t *testing.T, svc *Service, ctx context.Context, tag string) (userID, blogID, postID string) {
	t.Helper()
	userID = uuid.NewString()
	handle := "cb-" + strings.ReplaceAll(uuid.NewString()[:8], "-", "x")
	email := tag + "-" + userID[:8] + "@example.com"
	if _, err := svc.pool.Exec(ctx, `INSERT INTO users (id, email, email_verified, status)
		VALUES ($1, $2, TRUE, 'ACTIVE')`, userID, email); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.pool.Exec(ctx, `INSERT INTO user_profiles (user_id, username, display_name)
		VALUES ($1, $2, 'Community Author')`, userID, handle); err != nil {
		t.Fatal(err)
	}
	blogID = uuid.NewString()
	if _, err := svc.pool.Exec(ctx, `INSERT INTO blogs (id, owner_user_id, handle, title)
		VALUES ($1, $2, $3, 'Community Blog')`, blogID, userID, handle); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.pool.Exec(ctx, `INSERT INTO blog_members (blog_id, user_id, role)
		VALUES ($1, $2, 'OWNER')`, blogID, userID); err != nil {
		t.Fatal(err)
	}
	postID = uuid.NewString()
	if _, err := svc.pool.Exec(ctx, `INSERT INTO posts (id, blog_id, author_user_id, slug, title, status, published_at)
		VALUES ($1, $2, $3, 'hello', 'Hello', 'PUBLISHED', NOW())`, postID, blogID, userID); err != nil {
		t.Fatal(err)
	}
	return userID, blogID, postID
}

func seedReader(t *testing.T, svc *Service, ctx context.Context, tag string) string {
	t.Helper()
	userID := uuid.NewString()
	handle := "cr-" + strings.ReplaceAll(uuid.NewString()[:8], "-", "x")
	email := tag + "-" + userID[:8] + "@example.com"
	if _, err := svc.pool.Exec(ctx, `INSERT INTO users (id, email, email_verified, status)
		VALUES ($1, $2, TRUE, 'ACTIVE')`, userID, email); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.pool.Exec(ctx, `INSERT INTO user_profiles (user_id, username, display_name)
		VALUES ($1, $2, 'Community Reader')`, userID, handle); err != nil {
		t.Fatal(err)
	}
	return userID
}

func TestCommunity_Flow(t *testing.T) {
	svc, cleanup := setupIntegration(t)
	defer cleanup()
	ctx := context.Background()
	authorID, blogID, postID := seedAuthor(t, svc, ctx, "cbauthor")
	readerID := seedReader(t, svc, ctx, "cbreader")

	if err := svc.Follow(ctx, readerID, blogID); err != nil {
		t.Fatalf("Follow: %v", err)
	}
	if err := svc.Follow(ctx, readerID, blogID); err != nil {
		t.Fatalf("Follow idempotent: %v", err)
	}
	follows, err := svc.ListFollows(ctx, readerID, 10)
	if err != nil || len(follows) != 1 {
		t.Fatalf("ListFollows: %v %d", err, len(follows))
	}

	if err := svc.Subscribe(ctx, readerID, blogID); err != nil {
		t.Fatalf("Subscribe: %v", err)
	}
	subs, err := svc.ListSubscriptions(ctx, readerID, 10)
	if err != nil || len(subs) != 1 {
		t.Fatalf("ListSubscriptions: %v %d", err, len(subs))
	}

	if err := svc.React(ctx, readerID, postID, "like"); err != nil {
		t.Fatalf("React: %v", err)
	}
	if err := svc.React(ctx, readerID, postID, "LOVE"); err != nil {
		t.Fatalf("React change: %v", err)
	}
	summary, err := svc.ReactionCounts(ctx, readerID, postID)
	if err != nil {
		t.Fatalf("ReactionCounts: %v", err)
	}
	if summary.Total != 1 || summary.Counts["LOVE"] != 1 || summary.Mine == nil || *summary.Mine != "LOVE" {
		t.Fatalf("summary: %+v", summary)
	}
	if err := svc.Unreact(ctx, readerID, postID); err != nil {
		t.Fatalf("Unreact: %v", err)
	}

	comment, err := svc.CreateComment(ctx, readerID, postID, CreateCommentInput{ContentText: "좋은 글이네요"})
	if err != nil {
		t.Fatalf("CreateComment: %v", err)
	}
	reply, err := svc.CreateComment(ctx, authorID, postID, CreateCommentInput{ContentText: "감사합니다", ParentCommentID: &comment.ID})
	if err != nil {
		t.Fatalf("CreateReply: %v", err)
	}
	if reply.Depth != 1 {
		t.Fatalf("reply depth = %d", reply.Depth)
	}
	if _, err := svc.CreateComment(ctx, readerID, postID, CreateCommentInput{ContentText: "3단계", ParentCommentID: &reply.ID}); err == nil {
		t.Fatalf("depth-2 reply accepted")
	}
	comments, err := svc.ListComments(ctx, readerID, postID, 10)
	if err != nil || len(comments) != 2 {
		t.Fatalf("ListComments: %v %d", err, len(comments))
	}
	if err := svc.DeleteComment(ctx, readerID, comment.ID); err != nil {
		t.Fatalf("DeleteComment: %v", err)
	}
	if err := svc.DeleteComment(ctx, readerID, reply.ID); err == nil {
		t.Fatalf("foreign comment delete accepted")
	}

	if err := svc.AddBookmark(ctx, readerID, postID, "읽기 목록"); err != nil {
		t.Fatalf("AddBookmark: %v", err)
	}
	bookmarks, err := svc.ListBookmarks(ctx, readerID, 10)
	if err != nil || len(bookmarks) != 1 || bookmarks[0].FolderName != "읽기 목록" || bookmarks[0].Slug != "hello" {
		t.Fatalf("ListBookmarks: %v %+v", err, bookmarks)
	}
	if err := svc.RemoveBookmark(ctx, readerID, postID); err != nil {
		t.Fatalf("RemoveBookmark: %v", err)
	}

	if err := svc.BlockUser(ctx, readerID, authorID, "test"); err != nil {
		t.Fatalf("BlockUser: %v", err)
	}
	if err := svc.UnblockUser(ctx, readerID, authorID); err != nil {
		t.Fatalf("UnblockUser: %v", err)
	}

	if err := svc.Unsubscribe(ctx, readerID, blogID); err != nil {
		t.Fatalf("Unsubscribe: %v", err)
	}
	if err := svc.Unfollow(ctx, readerID, blogID); err != nil {
		t.Fatalf("Unfollow: %v", err)
	}
}
