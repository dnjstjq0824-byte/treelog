package treescan

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"treeblog/internal/platform/httpserver"
)

// Handler serves the TreeScan sync feed. External scanners authenticate
// with an API key; operators use the status endpoint with the same key.
type Handler struct {
	service *Service
	apiKeys []string
}

// NewHandler builds a treescan handler. apiKeys authorizes scanner access;
// an empty list disables the feed with 503.
func NewHandler(service *Service, apiKeys []string) *Handler {
	return &Handler{service: service, apiKeys: apiKeys}
}

// Name implements httpserver.Module.
func (h *Handler) Name() string {
	return "treescan"
}

// Mount implements httpserver.Module.
func (h *Handler) Mount(r chi.Router) {
	r.Route("/api/v1/treescan", func(r chi.Router) {
		r.Use(h.requireAPIKey)
		r.Get("/export", h.handleExport)
		r.Post("/ack", h.handleAck)
		r.Get("/status", h.handleStatus)
	})
}

func (h *Handler) requireAPIKey(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(h.apiKeys) == 0 {
			httpserver.WriteError(w, r, http.StatusServiceUnavailable, "TREESCAN_DISABLED", "TreeScan feed is not configured")
			return
		}
		provided := r.Header.Get("X-API-Key")
		authorized := false
		for _, key := range h.apiKeys {
			if subtle.ConstantTimeCompare([]byte(key), []byte(provided)) == 1 {
				authorized = true
				break
			}
		}
		if !authorized {
			httpserver.WriteError(w, r, http.StatusUnauthorized, "INVALID_API_KEY", "Invalid API key")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (h *Handler) handleError(w http.ResponseWriter, r *http.Request, err error) {
	var fault *Fault
	if errors.As(err, &fault) {
		httpserver.WriteError(w, r, fault.Status, fault.Code, fault.Message)
		return
	}
	httpserver.WriteError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
}

func (h *Handler) handleExport(w http.ResponseWriter, r *http.Request) {
	var cursor int64
	if raw := r.URL.Query().Get("cursor"); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || parsed < 0 {
			httpserver.WriteError(w, r, http.StatusBadRequest, "INVALID_CURSOR", "Invalid cursor")
			return
		}
		cursor = parsed
	}
	limit := 100
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	entries, next, err := h.service.Export(r.Context(), cursor, limit)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data":        entries,
		"next_cursor": next,
		"request_id":  httpserver.RequestID(r),
	})
}

func (h *Handler) handleAck(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Cursor int64 `json:"sequence_cursor"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpserver.WriteError(w, r, http.StatusBadRequest, "INVALID_REQUEST_PAYLOAD", "Invalid request body")
		return
	}
	acked, err := h.service.Ack(r.Context(), req.Cursor)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       map[string]any{"acked": acked},
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleStatus(w http.ResponseWriter, r *http.Request) {
	status, err := h.service.Status(r.Context())
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data": status, "request_id": httpserver.RequestID(r),
	})
}
