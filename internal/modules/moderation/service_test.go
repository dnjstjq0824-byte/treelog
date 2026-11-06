package moderation

import (
	"context"
	"errors"
	"net/http"
	"testing"
)

func TestNewServiceRequiresPool(t *testing.T) {
	if _, err := NewService(nil); err == nil {
		t.Fatalf("nil pool accepted")
	}
}

func TestReportValidation(t *testing.T) {
	svc := &Service{}
	ctx := context.Background()
	if _, err := svc.Report(ctx, "u", ReportInput{TargetType: "BOGUS", TargetID: "x", ReasonCategory: "SPAM"}, ""); !isFault(err, http.StatusBadRequest, "INVALID_TARGET_TYPE") {
		t.Fatalf("bad target: %v", err)
	}
	if _, err := svc.Report(ctx, "u", ReportInput{TargetType: "POST", TargetID: "x", ReasonCategory: "BOGUS"}, ""); !isFault(err, http.StatusBadRequest, "INVALID_REASON_CATEGORY") {
		t.Fatalf("bad reason: %v", err)
	}
}

func TestActionValidation(t *testing.T) {
	svc := &Service{}
	ctx := context.Background()
	if _, err := svc.ApplyAction(ctx, "m", "c", ActionInput{ActionType: "BOGUS"}); !isFault(err, http.StatusBadRequest, "INVALID_ACTION_TYPE") {
		t.Fatalf("bad action: %v", err)
	}
	if _, err := svc.ApplyAction(ctx, "m", "c", ActionInput{ActionType: "DISMISS"}); !isFault(err, http.StatusBadRequest, "INVALID_POLICY_CLAUSE") {
		t.Fatalf("missing clause: %v", err)
	}
}

func TestAppealValidation(t *testing.T) {
	svc := &Service{}
	ctx := context.Background()
	if _, err := svc.FileAppeal(ctx, "u", "c", ""); !isFault(err, http.StatusBadRequest, "INVALID_APPEAL") {
		t.Fatalf("empty statement: %v", err)
	}
	if _, err := svc.DecideAppeal(ctx, "m", "a", "MAYBE", ""); !isFault(err, http.StatusBadRequest, "INVALID_DECISION") {
		t.Fatalf("bad decision: %v", err)
	}
}

func isFault(err error, status int, code string) bool {
	var fault *Fault
	return errors.As(err, &fault) && fault.Status == status && fault.Code == code
}
