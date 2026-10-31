package user

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"treeblog/internal/modules/identity"
)

func TestUserHandler_Unauthenticated(t *testing.T) {
	h := &Handler{service: nil}
	r := chi.NewRouter()
	h.Mount(r)

	// Call /api/v1/me without auth -> 401
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me/", nil)
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized, got %d", rec.Code)
	}
}

func TestUserHandler_Validation(t *testing.T) {
	h := &Handler{service: nil}
	r := chi.NewRouter()
	h.Mount(r)

	t.Run("invalid json on profile update", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/me/profile", bytes.NewBufferString("{invalid"))
		ctx := identity.ContextWithUser(req.Context(), &identity.AuthUser{UserID: "usr_1"})
		r.ServeHTTP(rec, req.WithContext(ctx))

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 Bad Request, got %d", rec.Code)
		}
	})

	t.Run("invalid json on preferences update", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/me/preferences", bytes.NewBufferString("{invalid"))
		ctx := identity.ContextWithUser(req.Context(), &identity.AuthUser{UserID: "usr_1"})
		r.ServeHTTP(rec, req.WithContext(ctx))

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 Bad Request, got %d", rec.Code)
		}
	})
}

