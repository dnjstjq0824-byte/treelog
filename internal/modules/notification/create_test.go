package notification

import (
	"context"
	"testing"
	"time"
)

// TestCreateFanout verifies persistence, hub delivery, and the per-category
// in-app mute switch.
func TestCreateFanout(t *testing.T) {
	svc, cleanup := setupIntegration(t)
	defer cleanup()
	ctx := context.Background()
	userID := seedUser(t, svc, ctx)

	ch := make(chan Notification, 4)
	svc.Hub().Subscribe(userID, ch)
	defer svc.Hub().Unsubscribe(userID, ch)

	item, err := svc.Create(ctx, userID, "COMMENT", "COMMENT:p1", "COMMENT:c1",
		"새 댓글", "본문", "/b/h/s#c1")
	if err != nil || item == nil {
		t.Fatalf("create: %+v / %v", item, err)
	}
	select {
	case got := <-ch:
		if got.ID != item.ID {
			t.Fatalf("hub delivered %+v", got)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("hub delivered nothing")
	}
	items, _, err := svc.List(ctx, userID, 10, "", false)
	if err != nil || len(items) != 1 || items[0].ID != item.ID {
		t.Fatalf("list: %+v / %v", items, err)
	}

	// Muting the category suppresses storage and delivery.
	if _, err := svc.pool.Exec(ctx, `INSERT INTO notification_preferences (user_id, comments_in_app)
		VALUES ($1, FALSE) ON CONFLICT (user_id) DO UPDATE SET comments_in_app = FALSE`, userID); err != nil {
		t.Fatal(err)
	}
	muted, err := svc.Create(ctx, userID, "COMMENT", "COMMENT:p1", "COMMENT:c2",
		"조용한 댓글", "본문", "/b/h/s#c2")
	if err != nil || muted != nil {
		t.Fatalf("muted create: %+v / %v", muted, err)
	}
	items, _, _ = svc.List(ctx, userID, 10, "", false)
	if len(items) != 1 {
		t.Fatalf("muted notification stored: %+v", items)
	}
	select {
	case got := <-ch:
		t.Fatalf("muted notification delivered: %+v", got)
	case <-time.After(100 * time.Millisecond):
	}

	// Other categories are unaffected by the mute.
	if _, err := svc.Create(ctx, userID, "REACTION", "REACTION:p1", "REACTION:p1:u",
		"공감", "본문", "/b/h/s"); err != nil {
		t.Fatalf("reaction create: %v", err)
	}
	items, _, _ = svc.List(ctx, userID, 10, "", false)
	if len(items) != 2 {
		t.Fatalf("expected 2 notifications, got %+v", items)
	}
}
