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
		r.Post("/subscribe/pause", h.handlePauseSubscription)
		r.Post("/subscribe/resume", h.handleResumeSubscription)
		r.Patch("/subscribe/channels", h.handleUpdateSubscriptionChannels)
		r.Post("/neighbors/request", h.handleRequestNeighbor)
		r.Get("/neighbors/requests", h.handleListNeighborRequests)
		r.Post("/neighbors/group", h.handleMoveFollowToGroup)
		r.Post("/neighbors/favorite", h.handleFavoriteFollow)
		r.Delete("/neighbors/favorite", h.handleUnfavoriteFollow)
	})
	r.Route("/api/v1/neighbors/{followId}", func(r chi.Router) {
		r.Use(identity.RequireAuth)
		r.Post("/respond", h.handleRespondNeighborRequest)
	})
	r.Route("/api/v1/me/neighbor-groups", func(r chi.Router) {
		r.Use(identity.RequireAuth)
		r.Get("/", h.handleListNeighborGroups)
		r.Post("/", h.handleCreateNeighborGroup)
		r.Patch("/{groupId}", h.handleRenameNeighborGroup)
		r.Delete("/{groupId}", h.handleDeleteNeighborGroup)
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
		r.Get("/feed", h.handleFollowingFeed)
	})
	r.Route("/api/v1/users/{userId}", func(r chi.Router) {
		r.Use(identity.RequireAuth)
		r.Post("/block", h.handleBlock)
		r.Delete("/block", h.handleUnblock)
	})
	r.Route("/api/v1/comments/{commentId}", func(r chi.Router) {
		r.Use(identity.RequireAuth)
		r.Delete("/", h.handleDeleteComment)
		r.Post("/approve", h.handleApproveComment)
		r.Post("/reject", h.handleRejectComment)
		r.Post("/pin", h.handlePinComment)
		r.Delete("/pin", h.handleUnpinComment)
	})
	r.Route("/api/v1/blogs/{blogId}/comments", func(r chi.Router) {
		r.Use(identity.RequireAuth)
		r.Get("/pending", h.handleListPendingComments)
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

func (h *Handler) handleApproveComment(w http.ResponseWriter, r *http.Request) {
	user, _ := identity.UserFromContext(r.Context())
	comment, err := h.service.ApproveComment(r.Context(), user.UserID, chi.URLParam(r, "commentId"))
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       comment,
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleRejectComment(w http.ResponseWriter, r *http.Request) {
	user, _ := identity.UserFromContext(r.Context())
	comment, err := h.service.RejectComment(r.Context(), user.UserID, chi.URLParam(r, "commentId"))
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       comment,
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handlePinComment(w http.ResponseWriter, r *http.Request) {
	user, _ := identity.UserFromContext(r.Context())
	if err := h.service.SetCommentPinned(r.Context(), user.UserID, chi.URLParam(r, "commentId"), true); err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       map[string]bool{"pinned": true},
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleUnpinComment(w http.ResponseWriter, r *http.Request) {
	user, _ := identity.UserFromContext(r.Context())
	if err := h.service.SetCommentPinned(r.Context(), user.UserID, chi.URLParam(r, "commentId"), false); err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       map[string]bool{"pinned": false},
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleListPendingComments(w http.ResponseWriter, r *http.Request) {
	user, _ := identity.UserFromContext(r.Context())
	limit := 30
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	comments, err := h.service.ListPendingComments(r.Context(), user.UserID, chi.URLParam(r, "blogId"), limit)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       comments,
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleRequestNeighbor(w http.ResponseWriter, r *http.Request) {
	user, _ := identity.UserFromContext(r.Context())
	var req struct {
		Message string `json:"message"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	if err := h.service.RequestNeighbor(r.Context(), user.UserID, chi.URLParam(r, "blogId"), req.Message); err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       map[string]bool{"requested": true},
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleListNeighborRequests(w http.ResponseWriter, r *http.Request) {
	user, _ := identity.UserFromContext(r.Context())
	items, err := h.service.ListNeighborRequests(r.Context(), user.UserID, chi.URLParam(r, "blogId"))
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       items,
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleRespondNeighborRequest(w http.ResponseWriter, r *http.Request) {
	user, _ := identity.UserFromContext(r.Context())
	var req struct {
		Accept bool `json:"accept"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpserver.WriteError(w, r, http.StatusBadRequest, "INVALID_REQUEST_PAYLOAD", "Invalid request body")
		return
	}
	if err := h.service.RespondNeighborRequest(r.Context(), user.UserID, chi.URLParam(r, "followId"), req.Accept); err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       map[string]bool{"accepted": req.Accept},
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleCreateNeighborGroup(w http.ResponseWriter, r *http.Request) {
	user, _ := identity.UserFromContext(r.Context())
	var req struct {
		Name string `json:"name"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	group, err := h.service.CreateNeighborGroup(r.Context(), user.UserID, req.Name)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       group,
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleListNeighborGroups(w http.ResponseWriter, r *http.Request) {
	user, _ := identity.UserFromContext(r.Context())
	groups, err := h.service.ListNeighborGroups(r.Context(), user.UserID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       groups,
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleRenameNeighborGroup(w http.ResponseWriter, r *http.Request) {
	user, _ := identity.UserFromContext(r.Context())
	var req struct {
		Name string `json:"name"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	if err := h.service.RenameNeighborGroup(r.Context(), user.UserID, chi.URLParam(r, "groupId"), req.Name); err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       map[string]bool{"renamed": true},
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleDeleteNeighborGroup(w http.ResponseWriter, r *http.Request) {
	user, _ := identity.UserFromContext(r.Context())
	if err := h.service.DeleteNeighborGroup(r.Context(), user.UserID, chi.URLParam(r, "groupId")); err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       map[string]bool{"deleted": true},
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleMoveFollowToGroup(w http.ResponseWriter, r *http.Request) {
	user, _ := identity.UserFromContext(r.Context())
	var req struct {
		GroupID *string `json:"group_id"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	if err := h.service.MoveFollowToGroup(r.Context(), user.UserID, chi.URLParam(r, "blogId"), req.GroupID); err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       map[string]bool{"moved": true},
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleFavoriteFollow(w http.ResponseWriter, r *http.Request) {
	user, _ := identity.UserFromContext(r.Context())
	if err := h.service.SetFollowFavorite(r.Context(), user.UserID, chi.URLParam(r, "blogId"), true); err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       map[string]bool{"favorite": true},
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleUnfavoriteFollow(w http.ResponseWriter, r *http.Request) {
	user, _ := identity.UserFromContext(r.Context())
	if err := h.service.SetFollowFavorite(r.Context(), user.UserID, chi.URLParam(r, "blogId"), false); err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       map[string]bool{"favorite": false},
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handlePauseSubscription(w http.ResponseWriter, r *http.Request) {
	user, _ := identity.UserFromContext(r.Context())
	if err := h.service.PauseSubscription(r.Context(), user.UserID, chi.URLParam(r, "blogId")); err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       map[string]bool{"paused": true},
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleResumeSubscription(w http.ResponseWriter, r *http.Request) {
	user, _ := identity.UserFromContext(r.Context())
	if err := h.service.ResumeSubscription(r.Context(), user.UserID, chi.URLParam(r, "blogId")); err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       map[string]bool{"paused": false},
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleUpdateSubscriptionChannels(w http.ResponseWriter, r *http.Request) {
	user, _ := identity.UserFromContext(r.Context())
	var req struct {
		NotifyInApp  *bool `json:"notify_in_app"`
		NotifyWebPush *bool `json:"notify_web_push"`
		NotifyEmail  *bool `json:"notify_email"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	if err := h.service.UpdateSubscriptionChannels(r.Context(), user.UserID, chi.URLParam(r, "blogId"), req.NotifyInApp, req.NotifyWebPush, req.NotifyEmail); err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       map[string]bool{"updated": true},
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleFollowingFeed(w http.ResponseWriter, r *http.Request) {
	user, _ := identity.UserFromContext(r.Context())
	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	items, nextCursor, err := h.service.FollowingFeed(r.Context(), user.UserID, limit, r.URL.Query().Get("cursor"))
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
