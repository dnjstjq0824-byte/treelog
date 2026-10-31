package post

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

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
	return "post"
}

func (h *Handler) Mount(r chi.Router) {
	// Public routes
	r.Route("/api/v1/public/blogs/{blogKey}/posts", func(r chi.Router) {
		r.Get("/", h.handleListPublicPosts)
		r.Get("/{slug}", h.handleGetPublicPost)
	})

	// Authenticated post publishing & management routes
	r.Route("/api/v1/blogs/{blogId}/posts/{postId}", func(r chi.Router) {
		r.Use(identity.RequireAuth)

		r.Post("/publish", h.handlePublishPost)
		r.Post("/unpublish", h.handleUnpublishPost)
		r.Post("/restore", h.handleRestorePost)
		r.Delete("/", h.handleTrashPost)
	})
}

func (h *Handler) handleError(w http.ResponseWriter, r *http.Request, err error) {
	var fault *Fault
	if errors.As(err, &fault) {
		if fault.Status == http.StatusMovedPermanently {
			w.Header().Set("Location", "/api/v1/public/blogs/"+chi.URLParam(r, "blogKey")+"/posts/"+fault.Message)
			w.WriteHeader(http.StatusMovedPermanently)
			return
		}
		httpserver.WriteError(w, r, fault.Status, fault.Code, fault.Message)
		return
	}
	httpserver.WriteError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
}

func (h *Handler) handlePublishPost(w http.ResponseWriter, r *http.Request) {
	user, _ := identity.UserFromContext(r.Context())
	blogID := chi.URLParam(r, "blogId")
	postID := chi.URLParam(r, "postId")

	var input PublishPostInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httpserver.WriteError(w, r, http.StatusBadRequest, "INVALID_REQUEST_PAYLOAD", "Invalid request body")
		return
	}

	post, err := h.service.PublishPost(r.Context(), blogID, postID, user.UserID, input)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       post,
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleUnpublishPost(w http.ResponseWriter, r *http.Request) {
	user, _ := identity.UserFromContext(r.Context())
	blogID := chi.URLParam(r, "blogId")
	postID := chi.URLParam(r, "postId")

	err := h.service.UnpublishPost(r.Context(), blogID, postID, user.UserID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"post_id": postID,
			"status":  "UNPUBLISHED",
		},
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleTrashPost(w http.ResponseWriter, r *http.Request) {
	user, _ := identity.UserFromContext(r.Context())
	blogID := chi.URLParam(r, "blogId")
	postID := chi.URLParam(r, "postId")

	err := h.service.TrashPost(r.Context(), blogID, postID, user.UserID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"post_id": postID,
			"status":  "TRASHED",
		},
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleRestorePost(w http.ResponseWriter, r *http.Request) {
	user, _ := identity.UserFromContext(r.Context())
	blogID := chi.URLParam(r, "blogId")
	postID := chi.URLParam(r, "postId")

	err := h.service.RestorePost(r.Context(), blogID, postID, user.UserID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"post_id": postID,
			"status":  "RESTORED",
		},
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleListPublicPosts(w http.ResponseWriter, r *http.Request) {
	blogKey := chi.URLParam(r, "blogKey")
	categoryID := r.URL.Query().Get("category_id")
	tag := r.URL.Query().Get("tag")

	limit := 10
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	cursor := r.URL.Query().Get("cursor")

	posts, err := h.service.ListPublicPosts(r.Context(), blogKey, categoryID, tag, limit, cursor)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       posts,
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleGetPublicPost(w http.ResponseWriter, r *http.Request) {
	blogKey := chi.URLParam(r, "blogKey")
	slug := chi.URLParam(r, "slug")

	viewerUserID := ""
	if user, ok := identity.UserFromContext(r.Context()); ok {
		viewerUserID = user.UserID
	}

	post, err := h.service.GetPublicPost(r.Context(), blogKey, slug, viewerUserID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       post,
		"request_id": httpserver.RequestID(r),
	})
}
