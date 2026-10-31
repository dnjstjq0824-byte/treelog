package theme

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"treeblog/internal/modules/identity"
	"treeblog/internal/platform/httpserver"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Name() string {
	return "theme"
}

func (h *Handler) Mount(r chi.Router) {
	// Public catalog
	r.Get("/api/v1/themes", h.handleListThemes)

	// Blog theme configuration
	r.Route("/api/v1/blogs/{blogId}/theme", func(r chi.Router) {
		r.Get("/", h.handleGetBlogTheme)

		r.Group(func(r chi.Router) {
			r.Use(identity.RequireAuth)
			r.Put("/", h.handleUpdateBlogTheme)
			r.Post("/reset", h.handleResetBlogTheme)
		})
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

func (h *Handler) handleListThemes(w http.ResponseWriter, r *http.Request) {
	themes, err := h.service.ListThemes(r.Context())
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       themes,
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleGetBlogTheme(w http.ResponseWriter, r *http.Request) {
	blogID := chi.URLParam(r, "blogId")

	cfg, err := h.service.GetBlogTheme(r.Context(), blogID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       cfg,
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleUpdateBlogTheme(w http.ResponseWriter, r *http.Request) {
	user, _ := identity.UserFromContext(r.Context())
	blogID := chi.URLParam(r, "blogId")

	var input UpdateBlogThemeInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httpserver.WriteError(w, r, http.StatusBadRequest, "INVALID_REQUEST_PAYLOAD", "Invalid request body")
		return
	}

	cfg, err := h.service.UpdateBlogTheme(r.Context(), blogID, user.UserID, input)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       cfg,
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleResetBlogTheme(w http.ResponseWriter, r *http.Request) {
	user, _ := identity.UserFromContext(r.Context())
	blogID := chi.URLParam(r, "blogId")

	cfg, err := h.service.ResetBlogTheme(r.Context(), blogID, user.UserID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       cfg,
		"request_id": httpserver.RequestID(r),
	})
}
