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
