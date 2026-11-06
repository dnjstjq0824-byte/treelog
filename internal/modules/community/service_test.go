package community

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
)

func TestNewServiceRequiresPool(t *testing.T) {
	if _, err := NewService(nil); err == nil {
		t.Fatalf("nil pool accepted")
	}
}

func TestValidationWithoutPool(t *testing.T) {
	svc := &Service{}
	ctx := context.Background()
	if err := svc.React(ctx, "u", "p", "BOGUS"); !isFault(err, http.StatusBadRequest, "INVALID_REACTION_TYPE") {
		t.Fatalf("bad reaction: %v", err)
	}
	if _, err := svc.CreateComment(ctx, "u", "p", CreateCommentInput{}); !isFault(err, http.StatusBadRequest, "INVALID_COMMENT") {
		t.Fatalf("empty comment: %v", err)
	}
	if _, err := svc.CreateComment(ctx, "u", "p", CreateCommentInput{ContentText: strings.Repeat("x", 3001)}); !isFault(err, http.StatusBadRequest, "INVALID_COMMENT") {
		t.Fatalf("oversize comment accepted")
	}
	if err := svc.BlockUser(ctx, "u", "u", ""); !isFault(err, http.StatusBadRequest, "CANNOT_BLOCK_SELF") {
		t.Fatalf("self block: %v", err)
	}
}

func TestUnavailableWithoutPool(t *testing.T) {
	svc := &Service{}
	ctx := context.Background()
	if err := svc.Follow(ctx, "u", "b"); !isFault(err, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE") {
		t.Fatalf("follow: %v", err)
	}
	if err := svc.Subscribe(ctx, "u", "b"); err == nil {
		t.Fatalf("subscribe accepted without pool")
	}
	if _, err := svc.ReactionCounts(ctx, "u", "p"); err == nil {
		t.Fatalf("counts accepted without pool")
	}
	if _, err := svc.ListComments(ctx, "u", "p", 10); err == nil {
		t.Fatalf("list comments accepted without pool")
	}
	if err := svc.AddBookmark(ctx, "u", "p", ""); err == nil {
		t.Fatalf("bookmark accepted without pool")
	}
}

func isFault(err error, status int, code string) bool {
	var fault *Fault
	return errors.As(err, &fault) && fault.Status == status && fault.Code == code
}
