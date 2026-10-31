package blog

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"treeblog/internal/modules/identity"
)

func TestValidateHandle(t *testing.T) {
	tests := []struct {
		handle string
		valid  bool
	}{
		{"myblog", true},
		{"tech-log", true},
		{"coder123", true},
		{"a-b", true},
		{"ab", false},                      // too short (< 3)
		{"-myblog", false},                  // starts with hyphen
		{"myblog-", false},                  // ends with hyphen
		{"my_blog", false},                  // contains underscore
		{"admin", false},                    // reserved keyword
		{"api", false},                      // reserved keyword
		{"explore", false},                  // reserved keyword
		{"treeblog", false},                 // reserved keyword
		{"this-handle-is-definitely-way-too-long-for-the-system", false}, // > 30 chars
	}

	for _, tc := range tests {
		valid, msg := ValidateHandle(tc.handle)
		if valid != tc.valid {
			t.Errorf("ValidateHandle(%q) = %v (msg: %s); want %v", tc.handle, valid, msg, tc.valid)
		}
	}
}

func TestBlogHandler_Routing(t *testing.T) {
	h := &Handler{service: nil}
	r := chi.NewRouter()
	h.Mount(r)

	t.Run("create blog unauthenticated", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/blogs/", bytes.NewBufferString("{}"))
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rec.Code)
		}
	})

	t.Run("create blog invalid json", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/blogs/", bytes.NewBufferString("{invalid"))
		ctx := identity.ContextWithUser(req.Context(), &identity.AuthUser{UserID: "usr_1"})
		r.ServeHTTP(rec, req.WithContext(ctx))

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("create category unauthenticated", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/blogs/blog_123/categories", bytes.NewBufferString("{}"))
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rec.Code)
		}
	})

	t.Run("create series unauthenticated", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/blogs/blog_123/series", bytes.NewBufferString("{}"))
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rec.Code)
		}
	})
}
