package admin

import (
	"context"
	"testing"
)

func TestNewServiceRequiresPool(t *testing.T) {
	if _, err := NewService(nil); err == nil {
		t.Fatalf("nil pool accepted")
	}
}

func TestUnavailableWithoutPool(t *testing.T) {
	svc := &Service{}
	ctx := context.Background()
	if _, err := svc.Stats(ctx); err == nil {
		t.Fatalf("stats accepted without pool")
	}
	if _, err := svc.ListUsers(ctx, "", "", 10, 0); err == nil {
		t.Fatalf("list users accepted without pool")
	}
	if err := svc.SuspendUser(ctx, "u"); err == nil {
		t.Fatalf("suspend accepted without pool")
	}
	if _, err := svc.ListBlogs(ctx, "", 10, 0); err == nil {
		t.Fatalf("list blogs accepted without pool")
	}
}
