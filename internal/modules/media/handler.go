package media

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

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
		r.Get("/{assetId}/original", h.handleOriginalDownload)
	})
	r.Get("/api/v1/public/media/{assetId}", h.handlePublicMedia)
	r.Get("/api/v1/public/media/{assetId}/variants", h.handleListVariants)
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

	assets, nextCursor, err := h.service.ListMedia(r.Context(), blogID, user.UserID, limit, cursor)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data":        assets,
		"next_cursor": nextCursor,
		"request_id":  httpserver.RequestID(r),
	})
}

func (h *Handler) handlePublicMedia(w http.ResponseWriter, r *http.Request) {
	assetID := chi.URLParam(r, "assetId")
	variant := r.URL.Query().Get("variant")

	if !h.service.HotlinkAllowed(r.Header.Get("Referer"), r.Host) {
		httpserver.WriteError(w, r, http.StatusForbidden, "HOTLINK_BLOCKED", "Hotlinking from this referer is not allowed")
		return
	}

	// Explicit ?format= wins; otherwise browsers advertising WebP get the
	// smaller twin when the pipeline emitted one, with JPEG fallback.
	format := r.URL.Query().Get("format")
	if format == "" && variant != "" && strings.Contains(r.Header.Get("Accept"), "image/webp") {
		format = "webp"
	}
	bucket, key, _, err := h.service.PublicObject(r.Context(), assetID, variant, format)
	if err != nil {
		var fault *Fault
		if variant != "" && strings.EqualFold(format, "webp") && r.URL.Query().Get("format") == "" &&
			errors.As(err, &fault) && fault.Code == "VARIANT_NOT_FOUND" {
			bucket, key, _, err = h.service.PublicObject(r.Context(), assetID, variant, "jpeg")
		}
		if err != nil {
			h.handleError(w, r, err)
			return
		}
	}
	target, err := h.service.PresignDelivery(bucket, key)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	http.Redirect(w, r, target, http.StatusFound)
}

func (h *Handler) handleListVariants(w http.ResponseWriter, r *http.Request) {
	variants, err := h.service.ListVariants(r.Context(), chi.URLParam(r, "assetId"))
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       variants,
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleOriginalDownload(w http.ResponseWriter, r *http.Request) {
	user, _ := identity.UserFromContext(r.Context())
	bucket, key, _, err := h.service.OriginalObject(r.Context(),
		chi.URLParam(r, "blogId"), chi.URLParam(r, "assetId"), user.UserID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	target, err := h.service.PresignDelivery(bucket, key)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	http.Redirect(w, r, target, http.StatusFound)
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
