package notification

import (
	"errors"
	"net/http"
	"strconv"

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
