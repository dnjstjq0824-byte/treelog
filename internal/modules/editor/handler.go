package editor

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
	return "editor"
}

func (h *Handler) Mount(r chi.Router) {
	r.Route("/api/v1/blogs/{blogId}", func(r chi.Router) {
		r.Use(identity.RequireAuth)

		r.Post("/posts", h.handleCreateDraft)
		r.Put("/posts/{postId}/autosave", h.handleAutosave)
		r.Get("/posts/{postId}/revisions", h.handleListRevisions)
		r.Post("/posts/{postId}/revisions/{revisionId}/restore", h.handleRestoreRevision)

		r.Post("/editor/link-preview", h.handleLinkPreview)
		r.Post("/editor/markdown-convert", h.handleMarkdownConvert)
	})
}

func (h *Handler) handleError(w http.ResponseWriter, r *http.Request, err error) {
	var fault *Fault
	if errors.As(err, &fault) {
		if fault.Status == http.StatusConflict && fault.Data != nil {
			httpserver.WriteJSON(w, http.StatusConflict, map[string]any{
				"error": map[string]any{
					"code":       fault.Code,
					"message":    fault.Message,
					"data":       fault.Data,
					"request_id": httpserver.RequestID(r),
				},
			})
			return
		}
		httpserver.WriteError(w, r, fault.Status, fault.Code, fault.Message)
		return
	}
	httpserver.WriteError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
}

func (h *Handler) handleCreateDraft(w http.ResponseWriter, r *http.Request) {
	user, _ := identity.UserFromContext(r.Context())
	blogID := chi.URLParam(r, "blogId")

	var input CreateDraftInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httpserver.WriteError(w, r, http.StatusBadRequest, "INVALID_REQUEST_PAYLOAD", "Invalid request body")
		return
	}

	draft, err := h.service.CreateDraft(r.Context(), blogID, user.UserID, input)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       draft,
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleAutosave(w http.ResponseWriter, r *http.Request) {
	user, _ := identity.UserFromContext(r.Context())
	blogID := chi.URLParam(r, "blogId")
	postID := chi.URLParam(r, "postId")

	var input AutosaveInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httpserver.WriteError(w, r, http.StatusBadRequest, "INVALID_REQUEST_PAYLOAD", "Invalid request body")
		return
	}

	res, err := h.service.AutosaveDraft(r.Context(), blogID, postID, user.UserID, input)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       res,
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleListRevisions(w http.ResponseWriter, r *http.Request) {
	user, _ := identity.UserFromContext(r.Context())
	blogID := chi.URLParam(r, "blogId")
	postID := chi.URLParam(r, "postId")

	revisions, err := h.service.ListRevisions(r.Context(), blogID, postID, user.UserID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       revisions,
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleRestoreRevision(w http.ResponseWriter, r *http.Request) {
	user, _ := identity.UserFromContext(r.Context())
	blogID := chi.URLParam(r, "blogId")
	postID := chi.URLParam(r, "postId")
	revID := chi.URLParam(r, "revisionId")

	restored, err := h.service.RestoreRevision(r.Context(), blogID, postID, revID, user.UserID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       restored,
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleLinkPreview(w http.ResponseWriter, r *http.Request) {
	var input LinkPreviewInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httpserver.WriteError(w, r, http.StatusBadRequest, "INVALID_REQUEST_PAYLOAD", "Invalid request body")
		return
	}

	preview, err := h.service.FetchLinkPreview(r.Context(), input.URL)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       preview,
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleMarkdownConvert(w http.ResponseWriter, r *http.Request) {
	var input MarkdownConvertInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httpserver.WriteError(w, r, http.StatusBadRequest, "INVALID_REQUEST_PAYLOAD", "Invalid request body")
		return
	}

	converted, err := h.service.ConvertMarkdown(r.Context(), input)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       converted,
		"request_id": httpserver.RequestID(r),
	})
}
