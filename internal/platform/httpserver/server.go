package httpserver

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

type Module interface {
	Name() string
	Mount(chi.Router)
}

type ReadinessCheck func(context.Context) error

type Server struct {
	HTTP *http.Server
}

type contextKey int

const requestIDKey contextKey = iota

func requestID(r *http.Request) string {
	value, _ := r.Context().Value(requestIDKey).(string)
	return value
}

func withRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var bytes [12]byte
		if _, err := rand.Read(bytes[:]); err != nil {
			writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Request could not be processed")
			return
		}
		id := "req_" + hex.EncodeToString(bytes[:])
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), requestIDKey, id)))
	})
}

func recoverErrors(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if value := recover(); value != nil {
					logger.Error("request panic", "request_id", requestID(r), "panic", value)
					writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Request could not be processed")
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

func New(addr string, modules []Module, ready ReadinessCheck, logger *slog.Logger) (*Server, error) {
	if ready == nil {
		return nil, errors.New("readiness check is required")
	}
	if logger == nil {
		return nil, errors.New("logger is required")
	}
	r := chi.NewRouter()
	r.Use(withRequestID, recoverErrors(logger))
	r.Get("/api/v1/health/live", func(w http.ResponseWriter, req *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"data": map[string]string{"status": "live"}, "request_id": requestID(req)})
	})
	r.Get("/api/v1/health/ready", func(w http.ResponseWriter, req *http.Request) {
		if err := ready(req.Context()); err != nil {
			logger.Warn("readiness check failed", "error", err)
			writeError(w, req, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Service is not ready")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"data": map[string]string{"status": "ready"}, "request_id": requestID(req)})
	})
	names := make(map[string]struct{}, len(modules))
	for _, module := range modules {
		if module == nil || module.Name() == "" {
			return nil, errors.New("module name is required")
		}
		if _, exists := names[module.Name()]; exists {
			return nil, errors.New("duplicate module: " + module.Name())
		}
		names[module.Name()] = struct{}{}
		module.Mount(r)
	}
	r.NotFound(func(w http.ResponseWriter, req *http.Request) {
		writeError(w, req, http.StatusNotFound, "RESOURCE_NOT_FOUND", "Resource not found")
	})
	r.MethodNotAllowed(func(w http.ResponseWriter, req *http.Request) {
		writeError(w, req, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
	})
	return &Server{HTTP: &http.Server{
		Addr: addr, Handler: r, ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout: 60 * time.Second, MaxHeaderBytes: 1 << 20,
	}}, nil
}
