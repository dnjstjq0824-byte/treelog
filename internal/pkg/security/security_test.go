package security

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSecurityHeaders(t *testing.T) {
	handler := SecurityHeadersMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	handler.ServeHTTP(rec, req)

	if rec.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Fatal("missing or invalid X-Content-Type-Options")
	}
	if rec.Header().Get("X-Frame-Options") != "DENY" {
		t.Fatal("missing or invalid X-Frame-Options")
	}
}

func TestCSRFMiddleware(t *testing.T) {
	handler := CSRFMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Safe method should pass and set cookie
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on GET, got %d", rec.Code)
	}
	cookie := rec.Result().Cookies()
	var csrfCookie *http.Cookie
	for _, c := range cookie {
		if c.Name == CSRFCookieName {
			csrfCookie = c
			break
		}
	}
	if csrfCookie == nil || csrfCookie.Value == "" {
		t.Fatal("expected CSRF cookie to be set on GET")
	}

	// Mutating method without token should fail
	recMutateFail := httptest.NewRecorder()
	reqMutateFail := httptest.NewRequest(http.MethodPost, "/api/v1/test", nil)
	handler.ServeHTTP(recMutateFail, reqMutateFail)
	if recMutateFail.Code != http.StatusForbidden {
		t.Fatalf("expected 403 on POST without token, got %d", recMutateFail.Code)
	}

	// Mutating method with matching token should succeed
	recMutateSuccess := httptest.NewRecorder()
	reqMutateSuccess := httptest.NewRequest(http.MethodPost, "/api/v1/test", nil)
	reqMutateSuccess.AddCookie(csrfCookie)
	reqMutateSuccess.Header.Set(CSRFHeaderName, csrfCookie.Value)
	handler.ServeHTTP(recMutateSuccess, reqMutateSuccess)
	if recMutateSuccess.Code != http.StatusOK {
		t.Fatalf("expected 200 on POST with valid token, got %d", recMutateSuccess.Code)
	}
}

func TestRateLimiter(t *testing.T) {
	limiter := NewMemoryRateLimiter()
	key := "test-ip"
	limit := 3
	window := 100 * time.Millisecond

	// First 3 should succeed
	for i := 0; i < limit; i++ {
		allowed, _ := limiter.Allow(key, limit, window)
		if !allowed {
			t.Fatalf("expected request %d to be allowed", i+1)
		}
	}

	// 4th should be rejected
	allowed, retryAfter := limiter.Allow(key, limit, window)
	if allowed {
		t.Fatal("expected 4th request to be rejected")
	}
	if retryAfter <= 0 {
		t.Fatal("expected non-zero retryAfter")
	}

	// Wait for window to expire
	time.Sleep(120 * time.Millisecond)
	allowedAfter, _ := limiter.Allow(key, limit, window)
	if !allowedAfter {
		t.Fatal("expected request to be allowed after window expires")
	}
}

func TestComputeAuditHash(t *testing.T) {
	entry1 := AuditEntry{
		ActorRole:  "USER",
		ActionCode: "POST_PUBLISH",
		TargetType: "POST",
		TargetID:   "post_123",
		CreatedAt:  time.Date(2026, 10, 28, 1, 0, 0, 0, time.UTC),
	}
	prev := "0000000000000000000000000000000000000000000000000000000000000000"
	hash1 := ComputeAuditHash(prev, entry1)
	if len(hash1) != 64 {
		t.Fatalf("expected 64-char hex hash, got %d", len(hash1))
	}

	// Next chained entry
	entry2 := AuditEntry{
		ActorRole:  "ADMIN",
		ActionCode: "USER_BAN",
		TargetType: "USER",
		TargetID:   "user_456",
		CreatedAt:  time.Date(2026, 10, 28, 1, 1, 0, 0, time.UTC),
	}
	hash2 := ComputeAuditHash(hash1, entry2)
	if hash1 == hash2 {
		t.Fatal("expected different hashes for chained entries")
	}
}
