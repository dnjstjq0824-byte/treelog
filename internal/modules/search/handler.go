package search

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"treeblog/internal/platform/httpserver"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Name() string {
	return "search"
}

func (h *Handler) Mount(r chi.Router) {
	// Search endpoints
	r.Route("/api/v1/search", func(r chi.Router) {
		r.Get("/posts", h.handleSearchPosts)
		r.Get("/blogs", h.handleSearchBlogs)
		r.Get("/tags", h.handleSearchTags)
		r.Get("/suggest", h.handleSuggest)
		r.Post("/click", h.handleLogClick)
	})

	// Discover & Recommendation endpoints
	r.Route("/api/v1/discover", func(r chi.Router) {
		r.Get("/trending-posts", h.handleGetTrendingPosts)
		r.Get("/popular-blogs", h.handleGetPopularBlogs)
	})
	r.Get("/api/v1/public/posts/{postId}/related", h.handleGetRelatedPosts)
}

func (h *Handler) handleError(w http.ResponseWriter, r *http.Request, err error) {
	var fault *Fault
	if errors.As(err, &fault) {
		httpserver.WriteError(w, r, fault.Status, fault.Code, fault.Message)
		return
	}
	httpserver.WriteError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
}

func (h *Handler) handleSearchPosts(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	blogScopeID := r.URL.Query().Get("blog_id")

	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	offset := 0
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	results, total, err := h.service.SearchPosts(r.Context(), q, blogScopeID, limit, offset)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"query":   q,
			"items":   results,
			"total":   total,
			"limit":   limit,
			"offset":  offset,
		},
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleSearchBlogs(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	blogs, err := h.service.SearchBlogs(r.Context(), q, limit)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       blogs,
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleSearchTags(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	tags, err := h.service.SearchTags(r.Context(), q, limit)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       tags,
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleSuggest(w http.ResponseWriter, r *http.Request) {
	prefix := r.URL.Query().Get("q")
	limit := 5
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	suggestions, err := h.service.Suggest(r.Context(), prefix, limit)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       suggestions,
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleLogClick(w http.ResponseWriter, r *http.Request) {
	var input ClickEventInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httpserver.WriteError(w, r, http.StatusBadRequest, "INVALID_REQUEST_PAYLOAD", "Invalid request body")
		return
	}

	if err := h.service.LogClickEvent(r.Context(), input); err != nil {
		h.handleError(w, r, err)
		return
	}

	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       map[string]any{"logged": true},
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleGetTrendingPosts(w http.ResponseWriter, r *http.Request) {
	limit := 10
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	posts, err := h.service.GetTrendingPosts(r.Context(), limit)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       posts,
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleGetPopularBlogs(w http.ResponseWriter, r *http.Request) {
	limit := 10
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	blogs, err := h.service.GetPopularBlogs(r.Context(), limit)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       blogs,
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleGetRelatedPosts(w http.ResponseWriter, r *http.Request) {
	postID := chi.URLParam(r, "postId")
	limit := 5
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	posts, err := h.service.GetRelatedPosts(r.Context(), postID, limit)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       posts,
		"request_id": httpserver.RequestID(r),
	})
}
