package seo

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"treeblog/internal/platform/httpserver"
)

// Handler serves SEO endpoints.
type Handler struct {
	service *Service
}

// NewHandler builds an SEO handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Name implements httpserver.Module.
func (h *Handler) Name() string {
	return "seo"
}

// Mount implements httpserver.Module.
func (h *Handler) Mount(r chi.Router) {
	r.Get("/sitemap.xml", h.handleSitemap)
	r.Get("/robots.txt", h.handleRobots)
	r.Get("/api/v1/public/blogs/{blogKey}/seo", h.handleBlogMeta)
}

func (h *Handler) handleError(w http.ResponseWriter, r *http.Request, err error) {
	var fault *Fault
	if errors.As(err, &fault) {
		httpserver.WriteError(w, r, fault.Status, fault.Code, fault.Message)
		return
	}
	httpserver.WriteError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
}

func (h *Handler) handleSitemap(w http.ResponseWriter, r *http.Request) {
	body, err := h.service.Sitemap(r.Context())
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(body))
}

func (h *Handler) handleRobots(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(h.service.Robots()))
}

func (h *Handler) handleBlogMeta(w http.ResponseWriter, r *http.Request) {
	meta, err := h.service.BlogMeta(r.Context(), chi.URLParam(r, "blogKey"))
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data": meta, "request_id": httpserver.RequestID(r),
	})
}
