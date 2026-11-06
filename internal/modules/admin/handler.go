package admin

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"treeblog/internal/modules/identity"
	"treeblog/internal/platform/httpserver"
)

// Handler serves platform administration endpoints.
type Handler struct {
	service *Service
}

// NewHandler builds an admin handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Name implements httpserver.Module.
func (h *Handler) Name() string {
	return "admin"
}

// Mount implements httpserver.Module.
func (h *Handler) Mount(r chi.Router) {
	r.Route("/api/v1/admin", func(r chi.Router) {
		r.Use(identity.RequireAuth)
		r.Use(identity.RequirePlatformRole("SUPER_ADMIN"))
		r.Get("/stats", h.handleStats)
		r.Get("/users", h.handleListUsers)
		r.Post("/users/{userId}/suspend", h.handleSuspendUser)
		r.Post("/users/{userId}/unsuspend", h.handleUnsuspendUser)
		r.Get("/blogs", h.handleListBlogs)
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

func (h *Handler) handleStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.Stats(r.Context())
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data": stats, "request_id": httpserver.RequestID(r),
	})
}

func queryPaging(r *http.Request) (limit, offset int) {
	limit, offset = 30, 0
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if raw := r.URL.Query().Get("offset"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed >= 0 {
			offset = parsed
		}
	}
	return limit, offset
}

func (h *Handler) handleListUsers(w http.ResponseWriter, r *http.Request) {
	limit, offset := queryPaging(r)
	users, err := h.service.ListUsers(r.Context(), r.URL.Query().Get("q"), r.URL.Query().Get("status"), limit, offset)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data": users, "request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleSuspendUser(w http.ResponseWriter, r *http.Request) {
	if err := h.service.SuspendUser(r.Context(), chi.URLParam(r, "userId")); err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       map[string]string{"status": "SUSPENDED"},
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleUnsuspendUser(w http.ResponseWriter, r *http.Request) {
	if err := h.service.UnsuspendUser(r.Context(), chi.URLParam(r, "userId")); err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       map[string]string{"status": "ACTIVE"},
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleListBlogs(w http.ResponseWriter, r *http.Request) {
	limit, offset := queryPaging(r)
	blogs, err := h.service.ListBlogs(r.Context(), r.URL.Query().Get("status"), limit, offset)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data": blogs, "request_id": httpserver.RequestID(r),
	})
}
