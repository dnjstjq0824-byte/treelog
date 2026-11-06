package treescan

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestNewServiceRequiresPool(t *testing.T) {
	if _, err := NewService(nil); err == nil {
		t.Fatalf("nil pool accepted")
	}
}

func TestUnavailableWithoutPool(t *testing.T) {
	svc := &Service{}
	ctx := context.Background()
	if err := svc.MarkPost(ctx, "p", false); err == nil {
		t.Fatalf("mark accepted without pool")
	}
	if _, _, err := svc.Export(ctx, 0, 10); err == nil {
		t.Fatalf("export accepted without pool")
	}
	if _, err := svc.Ack(ctx, 1); err == nil {
		t.Fatalf("ack accepted without pool")
	}
	if _, err := svc.Status(ctx); err == nil {
		t.Fatalf("status accepted without pool")
	}
}

func TestAPIKeyMiddleware(t *testing.T) {
	newRouter := func(keys []string) *chi.Mux {
		r := chi.NewRouter()
		NewHandler(&Service{}, keys).Mount(r)
		return r
	}
	t.Run("disabled without keys", func(t *testing.T) {
		rec := httptest.NewRecorder()
		newRouter(nil).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/treescan/status", nil))
		if rec.Code != http.StatusServiceUnavailable {
			t.Fatalf("code = %d", rec.Code)
		}
	})
	t.Run("rejects bad key", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/treescan/status", nil)
		req.Header.Set("X-API-Key", "wrong")
		newRouter([]string{"correct"}).ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("code = %d", rec.Code)
		}
	})
	t.Run("accepts correct key", func(t *testing.T) {
		// The key passes and the handler reaches the pool guard (503),
		// proving authentication succeeded.
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/treescan/status", nil)
		req.Header.Set("X-API-Key", "correct")
		newRouter([]string{"correct"}).ServeHTTP(rec, req)
		if rec.Code != http.StatusServiceUnavailable {
			t.Fatalf("code = %d", rec.Code)
		}
	})
}
