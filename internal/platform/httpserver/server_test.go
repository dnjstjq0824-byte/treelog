package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthAndErrors(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	ready := true
	server, err := New(":0", nil, func(context.Context) error {
		if !ready {
			return errors.New("database unavailable")
		}
		return nil
	}, logger)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		path, code string
		status     int
	}{
		{"/api/v1/health/live", "", http.StatusOK},
		{"/api/v1/health/ready", "", http.StatusOK},
		{"/missing", "RESOURCE_NOT_FOUND", http.StatusNotFound},
	} {
		response := httptest.NewRecorder()
		server.HTTP.Handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, tc.path, nil))
		if response.Code != tc.status {
			t.Fatalf("%s: got %d, want %d", tc.path, response.Code, tc.status)
		}
		var body map[string]json.RawMessage
		if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		if response.Header().Get("X-Request-ID") == "" {
			t.Fatalf("%s: missing request id", tc.path)
		}
		if tc.code != "" {
			var fault errorBody
			if err := json.Unmarshal(body["error"], &fault); err != nil || fault.Code != tc.code || fault.RequestID == "" {
				t.Fatalf("%s: bad error envelope: %s", tc.path, response.Body.String())
			}
		}
	}
	ready = false
	response := httptest.NewRecorder()
	server.HTTP.Handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/health/ready", nil))
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("unready status: got %d", response.Code)
	}
}

func TestMetricsEndpoint(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	server, err := New(":0", nil, func(context.Context) error { return nil }, logger)
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"/api/v1/health/live", "/missing"} {
		response := httptest.NewRecorder()
		server.HTTP.Handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
	}
	response := httptest.NewRecorder()
	server.HTTP.Handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/health/metrics", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("metrics status: got %d", response.Code)
	}
	var body struct {
		Data struct {
			TotalRequests int64 `json:"total_requests"`
			Routes        map[string]struct {
				Count     int64 `json:"count"`
				Errors4xx int64 `json:"errors_4xx"`
			} `json:"routes"`
		} `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Data.TotalRequests < 2 {
		t.Fatalf("total = %d", body.Data.TotalRequests)
	}
	found404 := false
	for _, stats := range body.Data.Routes {
		if stats.Errors4xx > 0 && stats.Count > 0 {
			found404 = true
		}
	}
	if !found404 {
		t.Fatalf("4xx not tracked: %+v", body.Data.Routes)
	}
}

func TestGlobalMiddlewareOrdering(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	var order []string
	mark := func(name string) func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, name)
				next.ServeHTTP(w, r)
			})
		}
	}
	server, err := New(":0", nil, func(context.Context) error { return nil }, logger,
		WithGlobalMiddleware(mark("outer"), mark("inner")))
	if err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()
	server.HTTP.Handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/health/live", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("status: got %d", response.Code)
	}
	if len(order) != 2 || order[0] != "outer" || order[1] != "inner" {
		t.Fatalf("middleware ran out of order: %v", order)
	}
}
