package identity

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequireAuth(t *testing.T) {
	handler := RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Without context user -> 401
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}

	// With context user -> 200
	recAuth := httptest.NewRecorder()
	reqAuth := httptest.NewRequest(http.MethodGet, "/test", nil)
	ctx := ContextWithUser(reqAuth.Context(), &AuthUser{UserID: "usr_123"})
	handler.ServeHTTP(recAuth, reqAuth.WithContext(ctx))
	if recAuth.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recAuth.Code)
	}
}

func TestRequirePlatformRole(t *testing.T) {
	handler := RequirePlatformRole("SUPPORT_AGENT", "TRUST_SAFETY_MOD")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Regular user without role -> 403
	recNoRole := httptest.NewRecorder()
	reqNoRole := httptest.NewRequest(http.MethodGet, "/test", nil)
	ctxNoRole := ContextWithUser(reqNoRole.Context(), &AuthUser{UserID: "usr_1", PlatformRoles: []string{"STANDARD_USER"}})
	handler.ServeHTTP(recNoRole, reqNoRole.WithContext(ctxNoRole))
	if recNoRole.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", recNoRole.Code)
	}

	// User with allowed role -> 200
	recAllowed := httptest.NewRecorder()
	reqAllowed := httptest.NewRequest(http.MethodGet, "/test", nil)
	ctxAllowed := ContextWithUser(reqAllowed.Context(), &AuthUser{UserID: "usr_2", PlatformRoles: []string{"SUPPORT_AGENT"}})
	handler.ServeHTTP(recAllowed, reqAllowed.WithContext(ctxAllowed))
	if recAllowed.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recAllowed.Code)
	}

	// Super admin bypass -> 200
	recAdmin := httptest.NewRecorder()
	reqAdmin := httptest.NewRequest(http.MethodGet, "/test", nil)
	ctxAdmin := ContextWithUser(reqAdmin.Context(), &AuthUser{UserID: "usr_3", PlatformRoles: []string{"SUPER_ADMIN"}})
	handler.ServeHTTP(recAdmin, reqAdmin.WithContext(ctxAdmin))
	if recAdmin.Code != http.StatusOK {
		t.Fatalf("expected 200 for SUPER_ADMIN, got %d", recAdmin.Code)
	}
}

