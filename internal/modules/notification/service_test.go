package notification

import (
	"context"
	"testing"
	"time"
)

func TestNewServiceRequiresPool(t *testing.T) {
	if _, err := NewService(nil); err == nil {
		t.Fatalf("nil pool accepted")
	}
}

func TestCursorRoundTrip(t *testing.T) {
	moment := time.Date(2026, 6, 1, 9, 0, 0, 42, time.UTC)
	cursor := encodeCursor(moment, "id-1")
	backTime, backID, err := decodeCursor(cursor)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !backTime.Equal(moment) || backID != "id-1" {
		t.Fatalf("round trip mismatch")
	}
	for _, bad := range []string{"", "!!!", "bm90LWEtdGltZXQ="} {
		if _, _, err := decodeCursor(bad); err == nil {
			t.Fatalf("bad cursor %q accepted", bad)
		}
	}
}

func TestUnavailableWithoutPool(t *testing.T) {
	svc := &Service{}
	if _, _, err := svc.List(context.Background(), "u", 10, "", false); err == nil {
		t.Fatalf("list accepted without pool")
	}
	if _, err := svc.UnreadCount(context.Background(), "u"); err == nil {
		t.Fatalf("count accepted without pool")
	}
	if err := svc.MarkRead(context.Background(), "u", "n"); err == nil {
		t.Fatalf("mark-read accepted without pool")
	}
	if err := svc.MarkAllRead(context.Background(), "u"); err == nil {
		t.Fatalf("mark-all-read accepted without pool")
	}
}
