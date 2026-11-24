package post

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

// TestHandleErrorSlugRedirect pins the G-05 contract: a renamed slug must
// produce a real 301 with a Location header pointing at the current slug,
// not a JSON error body.
func TestHandleErrorSlugRedirect(t *testing.T) {
	h := NewHandler(nil)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/public/blogs/forest/posts/old-slug", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("blogKey", "forest")
	rctx.URLParams.Add("slug", "old-slug")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rec := httptest.NewRecorder()
	h.handleError(rec, req, &Fault{
		Status:  http.StatusMovedPermanently,
		Code:    "MOVED_PERMANENTLY",
		Message: "new-slug",
	})

	res := rec.Result()
	if res.StatusCode != http.StatusMovedPermanently {
		t.Fatalf("expected 301, got %d", res.StatusCode)
	}
	want := "/api/v1/public/blogs/forest/posts/new-slug"
	if got := res.Header.Get("Location"); got != want {
		t.Errorf("expected Location %q, got %q", want, got)
	}
	if ctype := res.Header.Get("Content-Type"); ctype != "" {
		t.Errorf("expected no JSON body on redirect, got Content-Type %q", ctype)
	}
}
