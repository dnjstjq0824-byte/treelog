package moderation

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

func seedUser(t *testing.T, svc *Service, ctx context.Context, tag string) string {
	t.Helper()
	userID := uuid.NewString()
	handle := "md-" + strings.ReplaceAll(uuid.NewString()[:8], "-", "x")
	if _, err := svc.pool.Exec(ctx, `INSERT INTO users (id, email, email_verified, status)
		VALUES ($1, $2, TRUE, 'ACTIVE')`, userID, tag+"-"+userID[:8]+"@example.com"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.pool.Exec(ctx, `INSERT INTO user_profiles (user_id, username, display_name)
		VALUES ($1, $2, 'Mod User')`, userID, handle); err != nil {
		t.Fatal(err)
	}
	return userID
}

func TestModeration_Flow(t *testing.T) {
	svc, cleanup := setupIntegration(t)
	defer cleanup()
	ctx := context.Background()
	authorID := seedUser(t, svc, ctx, "mdauthor")
	reporterID := seedUser(t, svc, ctx, "mdreporter")
	modID := seedUser(t, svc, ctx, "mdmod")

	blogID := uuid.NewString()
	blogHandle := "mdb-" + strings.ReplaceAll(uuid.NewString()[:8], "-", "x")
	if _, err := svc.pool.Exec(ctx, `INSERT INTO blogs (id, owner_user_id, handle, title)
		VALUES ($1, $2, $3, 'Mod Blog')`, blogID, authorID, blogHandle); err != nil {
		t.Fatal(err)
	}
	postID := uuid.NewString()
	if _, err := svc.pool.Exec(ctx, `INSERT INTO posts (id, blog_id, author_user_id, slug, title, status, published_at)
		VALUES ($1, $2, $3, 'spam-post', 'Spam Post', 'PUBLISHED', NOW())`, postID, blogID, authorID); err != nil {
		t.Fatal(err)
	}

	cse, err := svc.Report(ctx, reporterID, ReportInput{
		TargetType: "POST", TargetID: postID, ReasonCategory: "SPAM", Description: "spammy",
	}, "127.0.0.1")
	if err != nil {
		t.Fatalf("Report: %v", err)
	}
	if cse.ReportCount != 1 || cse.Status != "OPEN" {
		t.Fatalf("case: %+v", cse)
	}
	cse2, err := svc.Report(ctx, reporterID, ReportInput{
		TargetType: "POST", TargetID: postID, ReasonCategory: "SPAM",
	}, "")
	if err != nil {
		t.Fatalf("Report again: %v", err)
	}
	if cse2.ID != cse.ID || cse2.ReportCount != 2 {
		t.Fatalf("case reuse: %+v", cse2)
	}

	cases, err := svc.ListCases(ctx, "OPEN", 10, 0)
	if err != nil || len(cases) == 0 {
		t.Fatalf("ListCases: %v %d", err, len(cases))
	}
	detail, err := svc.GetCase(ctx, cse.ID)
	if err != nil || len(detail.Reports) != 2 {
		t.Fatalf("GetCase: %v %+v", err, detail)
	}

	action, err := svc.ApplyAction(ctx, modID, cse.ID, ActionInput{
		ActionType: "QUARANTINE_CONTENT", PolicyClauseCode: "POL-SPAM-1", PublicNoticeReason: "스팸으로 격리",
	})
	if err != nil {
		t.Fatalf("ApplyAction: %v", err)
	}
	if action.ActionType != "QUARANTINE_CONTENT" {
		t.Fatalf("action: %+v", action)
	}
	var status string
	if err := svc.pool.QueryRow(ctx, `SELECT status FROM posts WHERE id = $1`, postID).Scan(&status); err != nil {
		t.Fatal(err)
	}
	if status != "QUARANTINED" {
		t.Fatalf("post status = %s", status)
	}

	appeal, err := svc.FileAppeal(ctx, authorID, cse.ID, "오탐입니다. 복원 요청합니다.")
	if err != nil {
		t.Fatalf("FileAppeal: %v", err)
	}
	if _, err := svc.FileAppeal(ctx, reporterID, cse.ID, "남의 사건 이의제기"); err == nil {
		t.Fatalf("foreign appeal accepted")
	}
	decided, err := svc.DecideAppeal(ctx, modID, appeal.ID, "APPROVED_RESTORED", "오탐 확인")
	if err != nil {
		t.Fatalf("DecideAppeal: %v", err)
	}
	if decided.Status != "APPROVED_RESTORED" {
		t.Fatalf("appeal: %+v", decided)
	}
	if err := svc.pool.QueryRow(ctx, `SELECT status FROM posts WHERE id = $1`, postID).Scan(&status); err != nil {
		t.Fatal(err)
	}
	if status != "PUBLISHED" {
		t.Fatalf("restored status = %s", status)
	}
}
