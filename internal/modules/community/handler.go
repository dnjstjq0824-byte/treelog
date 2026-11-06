package community

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"treeblog/internal/modules/identity"
	"treeblog/internal/platform/httpserver"
)

// Handler serves community endpoints.
type Handler struct {
	service *Service
}

// NewHandler builds a community handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Name implements httpserver.Module.
func (h *Handler) Name() string {
	return "community"
}

// Mount implements httpserver.Module.
func (h *Handler) Mount(r chi.Router) {
	r.Route("/api/v1/blogs/{blogId}", func(r chi.Router) {
		r.Use(identity.RequireAuth)
		r.Post("/follow", h.handleFollow)
		r.Delete("/follow", h.handleUnfollow)
		r.Post("/subscribe", h.handleSubscribe)
		r.Delete("/subscribe", h.handleUnsubscribe)
	})
	r.Route("/api/v1/posts/{postId}", func(r chi.Router) {
		r.Get("/reactions", h.handleReactionCounts)
		r.Get("/comments", h.handleListComments)
		r.Route("/", func(r chi.Router) {
			r.Use(identity.RequireAuth)
			r.Post("/reactions", h.handleReact)
			r.Delete("/reactions", h.handleUnreact)
			r.Post("/comments", h.handleCreateComment)
			r.Post("/bookmark", h.handleAddBookmark)
			r.Delete("/bookmark", h.handleRemoveBookmark)
		})
	})
	r.Route("/api/v1/me", func(r chi.Router) {
		r.Use(identity.RequireAuth)
		r.Get("/follows", h.handleListFollows)
		r.Get("/subscriptions", h.handleListSubscriptions)
		r.Get("/bookmarks", h.handleListBookmarks)
	})
	r.Route("/api/v1/users/{userId}", func(r chi.Router) {
		r.Use(identity.RequireAuth)
		r.Post("/block", h.handleBlock)
		r.Delete("/block", h.handleUnblock)
	})
	r.Route("/api/v1/comments/{commentId}", func(r chi.Router) {
		r.Use(identity.RequireAuth)
		r.Delete("/", h.handleDeleteComment)
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

func (h *Handler) handleFollow(w http.ResponseWriter, r *http.Request) {
	user, _ := identity.UserFromContext(r.Context())
	if err := h.service.Follow(r.Context(), user.UserID, chi.URLParam(r, "blogId")); err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       map[string]bool{"following": true},
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleUnfollow(w http.ResponseWriter, r *http.Request) {
	user, _ := identity.UserFromContext(r.Context())
	if err := h.service.Unfollow(r.Context(), user.UserID, chi.URLParam(r, "blogId")); err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       map[string]bool{"following": false},
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleSubscribe(w http.ResponseWriter, r *http.Request) {
	user, _ := identity.UserFromContext(r.Context())
	if err := h.service.Subscribe(r.Context(), user.UserID, chi.URLParam(r, "blogId")); err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       map[string]bool{"subscribed": true},
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleUnsubscribe(w http.ResponseWriter, r *http.Request) {
	user, _ := identity.UserFromContext(r.Context())
	if err := h.service.Unsubscribe(r.Context(), user.UserID, chi.URLParam(r, "blogId")); err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       map[string]bool{"subscribed": false},
		"request_id": httpserver.RequestID(r),
	})
}

func queryLimit(r *http.Request, def int) int {
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			return parsed
		}
	}
	return def
}

func (h *Handler) handleListFollows(w http.ResponseWriter, r *http.Request) {
	user, _ := identity.UserFromContext(r.Context())
	items, err := h.service.ListFollows(r.Context(), user.UserID, queryLimit(r, 30))
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data": items, "request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleListSubscriptions(w http.ResponseWriter, r *http.Request) {
	user, _ := identity.UserFromContext(r.Context())
	items, err := h.service.ListSubscriptions(r.Context(), user.UserID, queryLimit(r, 30))
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data": items, "request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleReact(w http.ResponseWriter, r *http.Request) {
	user, _ := identity.UserFromContext(r.Context())
	var req struct {
		ReactionType string `json:"reaction_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpserver.WriteError(w, r, http.StatusBadRequest, "INVALID_REQUEST_PAYLOAD", "Invalid request body")
		return
	}
	if err := h.service.React(r.Context(), user.UserID, chi.URLParam(r, "postId"), req.ReactionType); err != nil {
		h.handleError(w, r, err)
		return
	}
	h.writeReactionSummary(w, r)
}

func (h *Handler) handleUnreact(w http.ResponseWriter, r *http.Request) {
	user, _ := identity.UserFromContext(r.Context())
	if err := h.service.Unreact(r.Context(), user.UserID, chi.URLParam(r, "postId")); err != nil {
		h.handleError(w, r, err)
		return
	}
	h.writeReactionSummary(w, r)
}

func (h *Handler) handleReactionCounts(w http.ResponseWriter, r *http.Request) {
	h.writeReactionSummary(w, r)
}

func (h *Handler) writeReactionSummary(w http.ResponseWriter, r *http.Request) {
	viewerID := ""
	if user, ok := identity.UserFromContext(r.Context()); ok {
		viewerID = user.UserID
	}
	summary, err := h.service.ReactionCounts(r.Context(), viewerID, chi.URLParam(r, "postId"))
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data": summary, "request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleCreateComment(w http.ResponseWriter, r *http.Request) {
	user, _ := identity.UserFromContext(r.Context())
	var input CreateCommentInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httpserver.WriteError(w, r, http.StatusBadRequest, "INVALID_REQUEST_PAYLOAD", "Invalid request body")
		return
	}
	comment, err := h.service.CreateComment(r.Context(), user.UserID, chi.URLParam(r, "postId"), input)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusCreated, map[string]any{
		"data": comment, "request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleListComments(w http.ResponseWriter, r *http.Request) {
	viewerID := ""
	if user, ok := identity.UserFromContext(r.Context()); ok {
		viewerID = user.UserID
	}
	items, err := h.service.ListComments(r.Context(), viewerID, chi.URLParam(r, "postId"), queryLimit(r, 30))
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data": items, "request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleDeleteComment(w http.ResponseWriter, r *http.Request) {
	user, _ := identity.UserFromContext(r.Context())
	if err := h.service.DeleteComment(r.Context(), user.UserID, chi.URLParam(r, "commentId")); err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       map[string]bool{"deleted": true},
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleAddBookmark(w http.ResponseWriter, r *http.Request) {
	user, _ := identity.UserFromContext(r.Context())
	var req struct {
		FolderName string `json:"folder_name"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	if err := h.service.AddBookmark(r.Context(), user.UserID, chi.URLParam(r, "postId"), req.FolderName); err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       map[string]bool{"bookmarked": true},
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleRemoveBookmark(w http.ResponseWriter, r *http.Request) {
	user, _ := identity.UserFromContext(r.Context())
	if err := h.service.RemoveBookmark(r.Context(), user.UserID, chi.URLParam(r, "postId")); err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       map[string]bool{"bookmarked": false},
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleListBookmarks(w http.ResponseWriter, r *http.Request) {
	user, _ := identity.UserFromContext(r.Context())
	items, err := h.service.ListBookmarks(r.Context(), user.UserID, queryLimit(r, 30))
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data": items, "request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleBlock(w http.ResponseWriter, r *http.Request) {
	user, _ := identity.UserFromContext(r.Context())
	var req struct {
		Reason string `json:"reason"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	if err := h.service.BlockUser(r.Context(), user.UserID, chi.URLParam(r, "userId"), req.Reason); err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       map[string]bool{"blocked": true},
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleUnblock(w http.ResponseWriter, r *http.Request) {
	user, _ := identity.UserFromContext(r.Context())
	if err := h.service.UnblockUser(r.Context(), user.UserID, chi.URLParam(r, "userId")); err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       map[string]bool{"blocked": false},
		"request_id": httpserver.RequestID(r),
	})
}
