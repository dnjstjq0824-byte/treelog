package media

import (
	"encoding/base64"
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
	return "media"
}

func (h *Handler) Mount(r chi.Router) {
	r.Route("/api/v1/blogs/{blogId}/media", func(r chi.Router) {
		r.Use(identity.RequireAuth)

		r.Post("/upload-sessions", h.handleCreateUploadSession)
		r.Post("/upload-sessions/{sessionId}/complete", h.handleCompleteUpload)
		r.Get("/", h.handleListMedia)
		r.Patch("/{assetId}", h.handleUpdateMediaMetadata)
		r.Post("/video-embed", h.handleCreateVideoEmbed)
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

func (h *Handler) handleCreateUploadSession(w http.ResponseWriter, r *http.Request) {
	user, _ := identity.UserFromContext(r.Context())
	blogID := chi.URLParam(r, "blogId")

	var input CreateUploadSessionInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httpserver.WriteError(w, r, http.StatusBadRequest, "INVALID_REQUEST_PAYLOAD", "Invalid request body")
		return
	}

	session, err := h.service.CreateUploadSession(r.Context(), blogID, user.UserID, input)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       session,
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleCompleteUpload(w http.ResponseWriter, r *http.Request) {
	user, _ := identity.UserFromContext(r.Context())
	blogID := chi.URLParam(r, "blogId")
	sessionID := chi.URLParam(r, "sessionId")

	var input CompleteUploadInput
	_ = json.NewDecoder(r.Body).Decode(&input)

	var sampleBytes []byte
	if input.MagicBytesBase64 != "" {
		sampleBytes, _ = base64.StdEncoding.DecodeString(input.MagicBytesBase64)
	}

	asset, err := h.service.CompleteUpload(r.Context(), blogID, sessionID, user.UserID, sampleBytes)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       asset,
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleListMedia(w http.ResponseWriter, r *http.Request) {
	user, _ := identity.UserFromContext(r.Context())
	blogID := chi.URLParam(r, "blogId")

	limit := 30
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	cursor := r.URL.Query().Get("cursor")

	assets, err := h.service.ListMedia(r.Context(), blogID, user.UserID, limit, cursor)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       assets,
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleUpdateMediaMetadata(w http.ResponseWriter, r *http.Request) {
	user, _ := identity.UserFromContext(r.Context())
	blogID := chi.URLParam(r, "blogId")
	assetID := chi.URLParam(r, "assetId")

	var input UpdateMediaAssetInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httpserver.WriteError(w, r, http.StatusBadRequest, "INVALID_REQUEST_PAYLOAD", "Invalid request body")
		return
	}

	asset, err := h.service.UpdateMediaMetadata(r.Context(), blogID, assetID, user.UserID, input)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       asset,
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleCreateVideoEmbed(w http.ResponseWriter, r *http.Request) {
	user, _ := identity.UserFromContext(r.Context())
	blogID := chi.URLParam(r, "blogId")

	var input VideoEmbedInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httpserver.WriteError(w, r, http.StatusBadRequest, "INVALID_REQUEST_PAYLOAD", "Invalid request body")
		return
	}

	asset, err := h.service.CreateVideoEmbed(r.Context(), blogID, user.UserID, input)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       asset,
		"request_id": httpserver.RequestID(r),
	})
}
