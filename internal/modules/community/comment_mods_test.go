package community

import (
	"context"
	"testing"

	"treeblog/internal/modules/moderation"
	"treeblog/internal/modules/notification"
)

// TestCommentModerationFlow covers the G-08 review loop: approval-required
// posting, the pending queue, approve/reject decisions, pinning, and
// reporting a comment into the moderation pipeline.
func TestCommentModerationFlow(t *testing.T) {
	svc, cleanup := setupIntegration(t)
	defer cleanup()
	ctx := context.Background()
	authorID, blogID, postID := seedAuthor(t, svc, ctx, "cmauthor")
	readerID := seedReader(t, svc, ctx, "cmreader")

	if _, err := svc.pool.Exec(ctx, `UPDATE posts SET allow_comments = 'REQUIRE_APPROVAL' WHERE id = $1`, postID); err != nil {
		t.Fatal(err)
	}

	first, err := svc.CreateComment(ctx, readerID, postID, CreateCommentInput{ContentText: "first!"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	visible, err := svc.ListComments(ctx, readerID, postID, 30)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range visible {
		if c.ID == first.ID {
			t.Fatal("pending comment is publicly visible")
		}
	}

	pending, err := svc.ListPendingComments(ctx, authorID, blogID, 30)
	if err != nil {
		t.Fatalf("pending queue: %v", err)
	}
	if len(pending) != 1 || pending[0].ID != first.ID {
		t.Fatalf("pending queue: %+v", pending)
	}
	if pending[0].PostSlug == nil || pending[0].PostTitle == nil {
		t.Fatal("pending queue must carry the post slug and title")
	}
	if _, err := svc.ListPendingComments(ctx, readerID, blogID, 30); faultCode(err) != "NOT_COMMENT_MODERATOR" {
		t.Fatalf("expected NOT_COMMENT_MODERATOR, got %v", err)
	}
	if _, err := svc.ApproveComment(ctx, readerID, first.ID); faultCode(err) != "NOT_COMMENT_MODERATOR" {
		t.Fatalf("expected NOT_COMMENT_MODERATOR, got %v", err)
	}

	if _, err := svc.ApproveComment(ctx, authorID, first.ID); err != nil {
		t.Fatalf("approve: %v", err)
	}
	visible, err = svc.ListComments(ctx, readerID, postID, 30)
	if err != nil || len(visible) != 1 || visible[0].ID != first.ID {
		t.Fatalf("approved comment not listed: %+v / %v", visible, err)
	}
	var count int
	if err := svc.pool.QueryRow(ctx, `SELECT comment_count FROM posts WHERE id = $1`, postID).Scan(&count); err != nil || count != 1 {
		t.Fatalf("comment_count = %d / %v", count, err)
	}
	if _, err := svc.ApproveComment(ctx, authorID, first.ID); faultCode(err) != "COMMENT_NOT_PENDING" {
		t.Fatalf("expected COMMENT_NOT_PENDING, got %v", err)
	}

	second, err := svc.CreateComment(ctx, readerID, postID, CreateCommentInput{ContentText: "spammy"})
	if err != nil {
		t.Fatalf("create second: %v", err)
	}
	if _, err := svc.RejectComment(ctx, authorID, second.ID); err != nil {
		t.Fatalf("reject: %v", err)
	}
	visible, _ = svc.ListComments(ctx, readerID, postID, 30)
	if len(visible) != 1 {
		t.Fatalf("rejected comment leaked into list: %+v", visible)
	}

	third, err := svc.CreateComment(ctx, readerID, postID, CreateCommentInput{ContentText: "pin me"})
	if err != nil {
		t.Fatalf("create third: %v", err)
	}
	if err := svc.SetCommentPinned(ctx, authorID, third.ID, true); faultCode(err) != "COMMENT_NOT_PINNABLE" {
		t.Fatalf("expected COMMENT_NOT_PINNABLE, got %v", err)
	}
	if _, err := svc.ApproveComment(ctx, authorID, third.ID); err != nil {
		t.Fatalf("approve third: %v", err)
	}
	if err := svc.SetCommentPinned(ctx, authorID, third.ID, true); err != nil {
		t.Fatalf("pin: %v", err)
	}
	visible, _ = svc.ListComments(ctx, readerID, postID, 30)
	if len(visible) != 2 || visible[0].ID != third.ID || !visible[0].IsPinned {
		t.Fatalf("pinned comment not first: %+v", visible)
	}
	if err := svc.SetCommentPinned(ctx, authorID, third.ID, false); err != nil {
		t.Fatalf("unpin: %v", err)
	}

	// Reporting a comment routes into the shared moderation pipeline.
	modSvc, err := moderation.NewService(svc.pool)
	if err != nil {
		t.Fatal(err)
	}
	cse, err := modSvc.Report(ctx, readerID, moderation.ReportInput{
		TargetType: "COMMENT", TargetID: first.ID, ReasonCategory: "SPAM",
	}, "")
	if err != nil {
		t.Fatalf("report comment: %v", err)
	}
	if cse.TargetType != "COMMENT" || cse.TargetID != first.ID {
		t.Fatalf("case: %+v", cse)
	}
}

func faultCode(err error) string {
	if fault, ok := err.(*Fault); ok {
		return fault.Code
	}
	return ""
}

// TestCommentFanoutNotifies verifies the G-09 wiring: a fresh comment and a
// first-time reaction produce in-app notifications for the post author,
// while the author's own actions stay silent.
func TestCommentFanoutNotifies(t *testing.T) {
	svc, cleanup := setupIntegration(t)
	defer cleanup()
	ctx := context.Background()
	authorID, _, postID := seedAuthor(t, svc, ctx, "cfauthor")
	readerID := seedReader(t, svc, ctx, "cfreader")

	notifier, err := notification.NewService(svc.pool)
	if err != nil {
		t.Fatal(err)
	}
	svc.SetNotifier(notifier)

	comment, err := svc.CreateComment(ctx, readerID, postID, CreateCommentInput{ContentText: "notify me"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	items, _, err := notifier.List(ctx, authorID, 10, "", false)
	if err != nil || len(items) != 1 || items[0].Category != "COMMENT" {
		t.Fatalf("author notifications: %+v / %v", items, err)
	}
	if _, err := svc.CreateComment(ctx, authorID, postID, CreateCommentInput{ContentText: "self note"}); err != nil {
		t.Fatalf("self comment: %v", err)
	}
	items, _, _ = notifier.List(ctx, authorID, 10, "", false)
	if len(items) != 1 {
		t.Fatalf("self action must not notify: %+v", items)
	}

	if err := svc.React(ctx, readerID, postID, "LIKE"); err != nil {
		t.Fatalf("react: %v", err)
	}
	items, _, _ = notifier.List(ctx, authorID, 10, "", false)
	if len(items) != 2 || items[0].Category != "REACTION" {
		t.Fatalf("reaction notification missing: %+v", items)
	}
	_ = comment
}
