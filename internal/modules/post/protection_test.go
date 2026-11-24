package post

import (
	"context"
	"strings"
	"testing"

	"treeblog/internal/modules/identity"
)

func TestValidateProtectionPassword(t *testing.T) {
	valid := []string{"1234", "correct horse", "비밀번호1234", strings.Repeat("x", 128)}
	for _, pw := range valid {
		if err := ValidateProtectionPassword(pw); err != nil {
			t.Errorf("expected %q to be valid, got %v", pw, err)
		}
	}
	invalid := []string{"", "abc", strings.Repeat("y", 129), "bad\xffutf8"}
	for _, pw := range invalid {
		if err := ValidateProtectionPassword(pw); err == nil {
			t.Errorf("expected %q to be invalid", pw)
		}
	}
}

func TestProtectionPasswordHashRoundTrip(t *testing.T) {
	hash, err := identity.HashPassword("post-secret-42")
	if err != nil {
		t.Fatal(err)
	}
	if hash == "post-secret-42" || !identity.VerifyPassword("post-secret-42", hash) {
		t.Fatal("argon2 round-trip failed")
	}
	if identity.VerifyPassword("wrong-password", hash) {
		t.Fatal("wrong password verified")
	}
}

func TestPublishPostProtectionValidation(t *testing.T) {
	svc := &Service{}
	ctx := context.Background()

	short := "abc"
	_, err := svc.PublishPost(ctx, "blog-id", "post-id", "user-id", PublishPostInput{
		Title: "t", Visibility: "PROTECTED", ProtectionPassword: &short,
	})
	if fault, ok := err.(*Fault); !ok || fault.Code != "PROTECTION_PASSWORD_INVALID" {
		t.Fatalf("expected PROTECTION_PASSWORD_INVALID, got %v", err)
	}

	// A valid password passes validation; without a pool the call then
	// fails at the database gate, proving validation ran first.
	good := "correct-horse-42"
	_, err = svc.PublishPost(ctx, "blog-id", "post-id", "user-id", PublishPostInput{
		Title: "t", Visibility: "PROTECTED", ProtectionPassword: &good,
	})
	if fault, ok := err.(*Fault); !ok || fault.Code != "SERVICE_UNAVAILABLE" {
		t.Fatalf("expected SERVICE_UNAVAILABLE after passing validation, got %v", err)
	}
}
