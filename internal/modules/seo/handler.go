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
	r.Get("/api/v1/public/blogs/{blogKey}/posts/{slug}/seo", h.handlePostMeta)
	r.Get("/b/{handle}/feed", h.handlePostFeed("rss"))
	r.Get("/b/{handle}/atom", h.handlePostFeed("atom"))
	r.Get("/b/{handle}/tags/{tag}/feed", h.handleTagFeed("rss"))
	r.Get("/b/{handle}/tags/{tag}/atom", h.handleTagFeed("atom"))
	r.Get("/b/{handle}/comments/feed", h.handleCommentFeed("rss"))
	r.Get("/b/{handle}/comments/atom", h.handleCommentFeed("atom"))
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

func (h *Handler) handlePostMeta(w http.ResponseWriter, r *http.Request) {
	meta, err := h.service.PostMeta(r.Context(), chi.URLParam(r, "blogKey"), chi.URLParam(r, "slug"))
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	w.Header().Set("Cache-Control", "public, max-age=300")
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data": meta, "request_id": httpserver.RequestID(r),
	})
}

func feedContentType(format string) string {
	if format == "atom" {
		return "application/atom+xml; charset=utf-8"
	}
	return "application/rss+xml; charset=utf-8"
}

func (h *Handler) writeFeed(w http.ResponseWriter, r *http.Request, format, body string, err error) {
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	w.Header().Set("Content-Type", feedContentType(format))
	w.Header().Set("Cache-Control", "public, max-age=900")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(body))
}

func (h *Handler) handlePostFeed(format string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := h.service.BlogPostFeed(r.Context(), chi.URLParam(r, "handle"), "", format)
		h.writeFeed(w, r, format, body, err)
	}
}

func (h *Handler) handleTagFeed(format string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := h.service.BlogPostFeed(r.Context(), chi.URLParam(r, "handle"), chi.URLParam(r, "tag"), format)
		h.writeFeed(w, r, format, body, err)
	}
}

func (h *Handler) handleCommentFeed(format string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := h.service.BlogCommentFeed(r.Context(), chi.URLParam(r, "handle"), format)
		h.writeFeed(w, r, format, body, err)
	}
}
