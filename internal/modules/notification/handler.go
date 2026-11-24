package notification

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"treeblog/internal/modules/identity"
	"treeblog/internal/platform/httpserver"
)

// Handler serves notification endpoints.
type Handler struct {
	service *Service
}

// NewHandler builds a notification handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Name implements httpserver.Module.
func (h *Handler) Name() string {
	return "notification"
}

// Mount implements httpserver.Module.
func (h *Handler) Mount(r chi.Router) {
	r.Route("/api/v1/me/notifications", func(r chi.Router) {
		r.Use(identity.RequireAuth)

		r.Get("/", h.handleList)
		r.Get("/unread-count", h.handleUnreadCount)
		r.Post("/{id}/read", h.handleMarkRead)
		r.Post("/read-all", h.handleMarkAllRead)
		r.Get("/stream", h.handleStream)
	})
}

// StreamTokenMiddleware copies the stream endpoint's access_token query
// parameter into the Authorization header so the standard authentication
// middleware can verify it. Native EventSource cannot set request headers,
// so the short-lived access token travels in the URL on this endpoint only.
// It must mount before identity.AuthenticateMiddleware.
func StreamTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/me/notifications/stream" && r.Header.Get("Authorization") == "" {
			if token := r.URL.Query().Get("access_token"); token != "" {
				r.Header.Set("Authorization", "Bearer "+token)
			}
		}
		next.ServeHTTP(w, r)
	})
}

// heartbeatInterval keeps intermediaries from closing idle streams.
const heartbeatInterval = 25 * time.Second

// handleStream serves the live notification feed over Server-Sent Events.
// It emits a ready event with the current unread count, then one event per
// new notification until the client disconnects.
func (h *Handler) handleStream(w http.ResponseWriter, r *http.Request) {
	user, _ := identity.UserFromContext(r.Context())
	flusher, ok := w.(http.Flusher)
	if !ok {
		h.handleError(w, r, &Fault{Status: http.StatusInternalServerError, Code: "STREAM_UNSUPPORTED", Message: "Streaming is not supported"})
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)

	unread, _ := h.service.UnreadCount(r.Context(), user.UserID)
	fmt.Fprintf(w, "event: ready\ndata: {\"unread_count\":%d}\n\n", unread)
	flusher.Flush()

	ch := make(chan Notification, streamBuffer)
	h.service.Hub().Subscribe(user.UserID, ch)
	defer h.service.Hub().Unsubscribe(user.UserID, ch)

	heartbeat := time.NewTicker(heartbeatInterval)
	defer heartbeat.Stop()
	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case <-heartbeat.C:
			fmt.Fprint(w, ": heartbeat\n\n")
			flusher.Flush()
		case item := <-ch:
			encoded, err := json.Marshal(item)
			if err != nil {
				continue
			}
			fmt.Fprintf(w, "event: notification\ndata: %s\n\n", encoded)
			flusher.Flush()
		}
	}
}

func (h *Handler) handleError(w http.ResponseWriter, r *http.Request, err error) {
	var fault *Fault
	if errors.As(err, &fault) {
		httpserver.WriteError(w, r, fault.Status, fault.Code, fault.Message)
		return
	}
	httpserver.WriteError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
}

func (h *Handler) handleList(w http.ResponseWriter, r *http.Request) {
	user, _ := identity.UserFromContext(r.Context())
	limit := 30
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	items, nextCursor, err := h.service.List(r.Context(), user.UserID, limit,
		r.URL.Query().Get("cursor"), r.URL.Query().Get("unread_only") == "true")
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data":        items,
		"next_cursor": nextCursor,
		"request_id":  httpserver.RequestID(r),
	})
}

func (h *Handler) handleUnreadCount(w http.ResponseWriter, r *http.Request) {
	user, _ := identity.UserFromContext(r.Context())
	count, err := h.service.UnreadCount(r.Context(), user.UserID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       map[string]any{"unread_count": count},
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleMarkRead(w http.ResponseWriter, r *http.Request) {
	user, _ := identity.UserFromContext(r.Context())
	if err := h.service.MarkRead(r.Context(), user.UserID, chi.URLParam(r, "id")); err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       map[string]bool{"read": true},
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleMarkAllRead(w http.ResponseWriter, r *http.Request) {
	user, _ := identity.UserFromContext(r.Context())
	if err := h.service.MarkAllRead(r.Context(), user.UserID); err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       map[string]bool{"read": true},
		"request_id": httpserver.RequestID(r),
	})
}
