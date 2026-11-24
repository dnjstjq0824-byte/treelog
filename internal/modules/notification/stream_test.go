package notification

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"treeblog/internal/modules/identity"
)

func waitForBody(t *testing.T, rec *httptest.ResponseRecorder, want string) string {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if body := rec.Body.String(); strings.Contains(body, want) {
			return body
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("stream never emitted %q, got:\n%s", want, rec.Body.String())
	return ""
}

// TestHandleStream verifies SSE framing end to end: headers, the ready
// event, one forwarded notification, and clean shutdown on disconnect.
func TestHandleStream(t *testing.T) {
	svc := &Service{}
	h := NewHandler(svc)

	ctx, cancel := context.WithCancel(context.Background())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me/notifications/stream", nil)
	req = req.WithContext(identity.ContextWithUser(ctx, &identity.AuthUser{UserID: "stream-user"}))
	rec := httptest.NewRecorder()

	done := make(chan struct{})
	go func() {
		defer close(done)
		h.handleStream(rec, req)
	}()

	waitForBody(t, rec, "event: ready")
	res := rec.Result()
	if ctype := res.Header.Get("Content-Type"); ctype != "text/event-stream" {
		t.Fatalf("content type = %q", ctype)
	}
	if svc.Hub().SubscriberCount("stream-user") != 1 {
		t.Fatal("stream did not subscribe")
	}

	svc.Hub().Publish("stream-user", Notification{ID: "live-1", Category: "COMMENT", Title: "Live"})
	body := waitForBody(t, rec, "event: notification")
	if !strings.Contains(body, `"ID":"live-1"`) && !strings.Contains(body, `"id":"live-1"`) {
		t.Fatalf("notification payload missing, got:\n%s", body)
	}

	cancel()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("stream did not exit on disconnect")
	}
	if got := svc.Hub().SubscriberCount("stream-user"); got != 0 {
		t.Fatalf("subscribers after disconnect = %d", got)
	}
}

// TestStreamRouteMounted proves the stream endpoint is registered behind
// authentication: anonymous requests are rejected, and authenticated ones
// open a stream while unknown paths 404.
func TestStreamRouteMounted(t *testing.T) {
	h := NewHandler(&Service{})
	r := chi.NewRouter()
	h.Mount(r)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me/notifications/stream", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for anonymous stream, got %d", rec.Code)
	}

	ctx, cancel := context.WithCancel(context.Background())
	req = httptest.NewRequest(http.MethodGet, "/api/v1/me/notifications/stream", nil)
	req = req.WithContext(identity.ContextWithUser(ctx, &identity.AuthUser{UserID: "u"}))
	rec = httptest.NewRecorder()
	done := make(chan struct{})
	go func() {
		defer close(done)
		r.ServeHTTP(rec, req)
	}()
	waitForBody(t, rec, "event: ready")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for authenticated stream, got %d", rec.Code)
	}
	cancel()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("stream did not exit")
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/me/notifications/nope", nil)
	req = req.WithContext(identity.ContextWithUser(context.Background(), &identity.AuthUser{UserID: "u"}))
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for unknown route, got %d", rec.Code)
	}
}

func TestStreamTokenMiddleware(t *testing.T) {
	var gotAuth string
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
	})
	handler := StreamTokenMiddleware(next)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me/notifications/stream?access_token=tok123", nil)
	handler.ServeHTTP(httptest.NewRecorder(), req)
	if gotAuth != "Bearer tok123" {
		t.Fatalf("expected token copied to header, got %q", gotAuth)
	}

	// Other paths and explicit headers are untouched.
	gotAuth = ""
	req = httptest.NewRequest(http.MethodGet, "/api/v1/me/notifications/?access_token=tok123", nil)
	handler.ServeHTTP(httptest.NewRecorder(), req)
	if gotAuth != "" {
		t.Fatalf("expected no copy on list endpoint, got %q", gotAuth)
	}
	req = httptest.NewRequest(http.MethodGet, "/api/v1/me/notifications/stream?access_token=tok123", nil)
	req.Header.Set("Authorization", "Bearer direct")
	handler.ServeHTTP(httptest.NewRecorder(), req)
	if gotAuth != "Bearer direct" {
		t.Fatalf("expected explicit header kept, got %q", gotAuth)
	}
}
