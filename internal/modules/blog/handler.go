package blog

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
	return "blog"
}

func (h *Handler) Mount(r chi.Router) {
	r.Get("/api/v1/tags/popular", h.handleListPopularTags)

	r.Route("/api/v1/blogs", func(r chi.Router) {
		r.Get("/check-handle", h.handleCheckHandle)

		// Public blog lookup
		r.Get("/{blogId}", h.handleGetBlog)
		r.Get("/{blogId}/guestbook", h.handleListGuestbook)
		r.Get("/{blogId}/categories", h.handleListCategories)
		r.Get("/{blogId}/series", h.handleListSeries)

		// Authenticated blog operations
		r.Group(func(r chi.Router) {
			r.Use(identity.RequireAuth)

			r.Post("/", h.handleCreateBlog)
			r.Get("/my", h.handleListMyBlogs)
			r.Post("/{blogId}/guestbook", h.handleCreateGuestbook)

			// Categories management
			r.Post("/{blogId}/categories", h.handleCreateCategory)
			r.Put("/{blogId}/categories/reorder", h.handleReorderCategories)
			r.Delete("/{blogId}/categories/{categoryId}", h.handleDeleteCategory)

			// Series management
			r.Post("/{blogId}/series", h.handleCreateSeries)

			// Blog management requiring ADMIN or OWNER role
			r.Get("/{blogId}/settings", h.handleGetSettings)
			r.Patch("/{blogId}/settings", h.handleUpdateSettings)
			r.Get("/{blogId}/members", h.handleListMembers)
			r.Post("/{blogId}/members", h.handleInviteMember)
			r.Delete("/{blogId}/members/{userId}", h.handleRemoveMember)
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

func (h *Handler) handleCheckHandle(w http.ResponseWriter, r *http.Request) {
	handle := r.URL.Query().Get("handle")
	available, message, err := h.service.CheckHandleAvailable(r.Context(), handle)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"handle":    handle,
			"available": available,
			"message":   message,
		},
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleCreateBlog(w http.ResponseWriter, r *http.Request) {
	user, _ := identity.UserFromContext(r.Context())
	var input CreateBlogInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httpserver.WriteError(w, r, http.StatusBadRequest, "INVALID_REQUEST_PAYLOAD", "Invalid request body")
		return
	}
	blog, err := h.service.CreateBlog(r.Context(), user.UserID, input)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusCreated, map[string]any{
		"data":       blog,
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleGetBlog(w http.ResponseWriter, r *http.Request) {
	blogID := chi.URLParam(r, "blogId")
	blog, err := h.service.GetBlog(r.Context(), blogID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       blog,
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleListMyBlogs(w http.ResponseWriter, r *http.Request) {
	user, _ := identity.UserFromContext(r.Context())
	blogs, err := h.service.ListMyBlogs(r.Context(), user.UserID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       blogs,
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleGetSettings(w http.ResponseWriter, r *http.Request) {
	blogID := chi.URLParam(r, "blogId")
	settings, err := h.service.GetBlogSettings(r.Context(), blogID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       settings,
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleUpdateSettings(w http.ResponseWriter, r *http.Request) {
	blogID := chi.URLParam(r, "blogId")
	var st BlogSettings
	if err := json.NewDecoder(r.Body).Decode(&st); err != nil {
		httpserver.WriteError(w, r, http.StatusBadRequest, "INVALID_REQUEST_PAYLOAD", "Invalid request body")
		return
	}
	if err := h.service.UpdateBlogSettings(r.Context(), blogID, st); err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data": map[string]string{
			"message": "Blog settings updated successfully",
		},
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleListMembers(w http.ResponseWriter, r *http.Request) {
	blogID := chi.URLParam(r, "blogId")
	members, err := h.service.ListMembers(r.Context(), blogID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       members,
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleInviteMember(w http.ResponseWriter, r *http.Request) {
	user, _ := identity.UserFromContext(r.Context())
	blogID := chi.URLParam(r, "blogId")
	var req struct {
		Email string `json:"email"`
		Role  string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpserver.WriteError(w, r, http.StatusBadRequest, "INVALID_REQUEST_PAYLOAD", "Invalid request body")
		return
	}
	member, err := h.service.InviteMember(r.Context(), blogID, user.UserID, req.Email, req.Role)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       member,
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleRemoveMember(w http.ResponseWriter, r *http.Request) {
	blogID := chi.URLParam(r, "blogId")
	targetUserID := chi.URLParam(r, "userId")
	if err := h.service.RemoveMember(r.Context(), blogID, targetUserID); err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data": map[string]string{
			"message": "Member removed successfully",
		},
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleCreateGuestbook(w http.ResponseWriter, r *http.Request) {
	user, _ := identity.UserFromContext(r.Context())
	blogID := chi.URLParam(r, "blogId")
	var req struct {
		ContentText   string  `json:"content_text"`
		IsSecret      bool    `json:"is_secret"`
		ParentEntryID *string `json:"parent_entry_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpserver.WriteError(w, r, http.StatusBadRequest, "INVALID_REQUEST_PAYLOAD", "Invalid request body")
		return
	}
	entry, err := h.service.CreateGuestbookEntry(r.Context(), blogID, user.UserID, req.ContentText, req.IsSecret, req.ParentEntryID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusCreated, map[string]any{
		"data":       entry,
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleListGuestbook(w http.ResponseWriter, r *http.Request) {
	blogID := chi.URLParam(r, "blogId")
	viewerUserID := ""
	isStaff := false
	if user, ok := identity.UserFromContext(r.Context()); ok {
		viewerUserID = user.UserID
		// if super admin or blog role
		if user.HasPlatformRole("SUPER_ADMIN") || user.CurrentBlogRole == "ADMIN" || user.CurrentBlogRole == "OWNER" || user.CurrentBlogRole == "EDITOR" {
			isStaff = true
		}
	}
	entries, err := h.service.ListGuestbookEntries(r.Context(), blogID, viewerUserID, isStaff)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       entries,
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleListCategories(w http.ResponseWriter, r *http.Request) {
	blogID := chi.URLParam(r, "blogId")
	categories, err := h.service.ListCategories(r.Context(), blogID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       categories,
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleCreateCategory(w http.ResponseWriter, r *http.Request) {
	blogID := chi.URLParam(r, "blogId")
	var input CreateCategoryInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httpserver.WriteError(w, r, http.StatusBadRequest, "INVALID_REQUEST_PAYLOAD", "Invalid request body")
		return
	}
	category, err := h.service.CreateCategory(r.Context(), blogID, input)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusCreated, map[string]any{
		"data":       category,
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleReorderCategories(w http.ResponseWriter, r *http.Request) {
	blogID := chi.URLParam(r, "blogId")
	var req struct {
		Categories []ReorderCategoryItem `json:"categories"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpserver.WriteError(w, r, http.StatusBadRequest, "INVALID_REQUEST_PAYLOAD", "Invalid request body")
		return
	}
	if err := h.service.ReorderCategories(r.Context(), blogID, req.Categories); err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data": map[string]string{
			"message": "Categories reordered successfully",
		},
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleDeleteCategory(w http.ResponseWriter, r *http.Request) {
	blogID := chi.URLParam(r, "blogId")
	categoryID := chi.URLParam(r, "categoryId")
	var targetCategoryID *string
	targetParam := r.URL.Query().Get("target_category_id")
	if targetParam != "" {
		targetCategoryID = &targetParam
	}
	if err := h.service.DeleteCategory(r.Context(), blogID, categoryID, targetCategoryID); err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data": map[string]string{
			"message": "Category deleted and posts relocated successfully",
		},
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleListSeries(w http.ResponseWriter, r *http.Request) {
	blogID := chi.URLParam(r, "blogId")
	series, err := h.service.ListSeries(r.Context(), blogID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       series,
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleCreateSeries(w http.ResponseWriter, r *http.Request) {
	blogID := chi.URLParam(r, "blogId")
	var req struct {
		Title       string `json:"title"`
		Slug        string `json:"slug"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpserver.WriteError(w, r, http.StatusBadRequest, "INVALID_REQUEST_PAYLOAD", "Invalid request body")
		return
	}
	series, err := h.service.CreateSeries(r.Context(), blogID, req.Title, req.Slug, req.Description)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusCreated, map[string]any{
		"data":       series,
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleListPopularTags(w http.ResponseWriter, r *http.Request) {
	tags, err := h.service.ListPopularTags(r.Context(), 30)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       tags,
		"request_id": httpserver.RequestID(r),
	})
}
