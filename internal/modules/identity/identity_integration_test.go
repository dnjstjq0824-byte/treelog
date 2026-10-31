package identity

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"os"
	"testing"
	"time"

	"treeblog/internal/db"
)

type memoryMailer struct {
	lastEmail string
	lastCode  string
	lastToken string
}

func (m *memoryMailer) SendVerification(ctx context.Context, email, code string) error {
	m.lastEmail = email
	m.lastCode = code
	return nil
}

func (m *memoryMailer) SendPasswordReset(ctx context.Context, email, token string) error {
	m.lastEmail = email
	m.lastToken = token
	return nil
}

func setupTestService(t *testing.T) (*Service, *memoryMailer, func()) {
	t.Helper()
	url := os.Getenv("TREEBLOG_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TREEBLOG_TEST_DATABASE_URL is not configured")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pools, err := db.Open(ctx, url, "")
	if err != nil {
		t.Fatal(err)
	}

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		pools.Close()
		t.Fatal(err)
	}
	signer, err := NewTokenSigner(key, "https://treeblog.test", "treeblog-web")
	if err != nil {
		pools.Close()
		t.Fatal(err)
	}

	mailer := &memoryMailer{}
	pepper := make([]byte, 32)
	_, _ = rand.Read(pepper)

	svc, err := NewService(pools.Primary, signer, mailer, pepper)
	if err != nil {
		pools.Close()
		t.Fatal(err)
	}

	cleanup := func() {
		pools.Close()
	}
	return svc, mailer, cleanup
}

func TestIdentity_FullFlow(t *testing.T) {
	svc, mailer, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()
	email := "testuser@example.com"
	username := "testuser123"

	// 1. Register
	userID, err := svc.Register(ctx, RegisterInput{
		Email:          email,
		Username:       username,
		DisplayName:    "Test User",
		Password:       "P@ssword123!",
		TermsVersion:   "2026-01-01",
		PrivacyVersion: "2026-01-01",
	})
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	if userID == "" {
		t.Fatal("expected non-empty user ID")
	}

	// 2. Verify Email
	if err := svc.VerifyEmail(ctx, email, mailer.lastCode); err != nil {
		t.Fatalf("VerifyEmail failed: %v", err)
	}

	// 3. Login
	tokens, err := svc.Login(ctx, email, "P@ssword123!")
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	if tokens.Access == "" || tokens.Refresh == "" {
		t.Fatal("expected access and refresh tokens")
	}

	// 4. Refresh token rotation
	newTokens, err := svc.Refresh(ctx, tokens.Refresh)
	if err != nil {
		t.Fatalf("Refresh failed: %v", err)
	}
	if newTokens.Refresh == tokens.Refresh {
		t.Fatal("expected refresh token to be rotated")
	}

	// 5. Detect reuse of previous refresh token (after grace period)
	time.Sleep(6 * time.Second)
	_, err = svc.Refresh(ctx, tokens.Refresh)
	if err == nil {
		t.Fatal("expected error on token reuse")
	}

	// 6. Logout
	if err := svc.Logout(ctx, newTokens.Access); err != nil {
		t.Fatalf("Logout failed: %v", err)
	}
}

func TestIdentity_PasswordReset(t *testing.T) {
	svc, mailer, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()
	email := "resetuser@example.com"

	_, err := svc.Register(ctx, RegisterInput{
		Email:          email,
		Username:       "resetuser",
		DisplayName:    "Reset User",
		Password:       "InitialPassword123!",
		TermsVersion:   "2026-01-01",
		PrivacyVersion: "2026-01-01",
	})
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	if err := svc.VerifyEmail(ctx, email, mailer.lastCode); err != nil {
		t.Fatalf("VerifyEmail failed: %v", err)
	}

	if err := svc.ForgotPassword(ctx, email); err != nil {
		t.Fatalf("ForgotPassword failed: %v", err)
	}
	if mailer.lastToken == "" {
		t.Fatal("expected password reset token to be sent")
	}

	newPassword := "NewBrandPassword123!"
	if err := svc.ResetPassword(ctx, mailer.lastToken, newPassword); err != nil {
		t.Fatalf("ResetPassword failed: %v", err)
	}

	// Verify login with new password
	tokens, err := svc.Login(ctx, email, newPassword)
	if err != nil {
		t.Fatalf("Login with new password failed: %v", err)
	}
	if tokens.Access == "" {
		t.Fatal("expected access token after login")
	}
}

