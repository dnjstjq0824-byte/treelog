package user

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
	return "user"
}

func (h *Handler) Mount(r chi.Router) {
	r.Route("/api/v1/me", func(r chi.Router) {
		r.Use(identity.RequireAuth)

		r.Get("/", h.handleGetMe)
		r.Patch("/profile", h.handleUpdateProfile)
		r.Patch("/preferences", h.handleUpdatePreferences)
		r.Get("/sessions", h.handleGetSessions)
		r.Delete("/sessions/{sessionId}", h.handleRevokeSession)
		r.Post("/deactivate", h.handleDeactivate)
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

func (h *Handler) handleGetMe(w http.ResponseWriter, r *http.Request) {
	user, _ := identity.UserFromContext(r.Context())
	profile, err := h.service.GetMe(r.Context(), user.UserID, user.SessionID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       profile,
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleUpdateProfile(w http.ResponseWriter, r *http.Request) {
	user, _ := identity.UserFromContext(r.Context())
	var req struct {
		DisplayName string `json:"display_name"`
		Bio         string `json:"bio"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpserver.WriteError(w, r, http.StatusBadRequest, "INVALID_REQUEST_PAYLOAD", "Invalid request body")
		return
	}
	if err := h.service.UpdateProfile(r.Context(), user.UserID, req.DisplayName, req.Bio); err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data": map[string]string{
			"message": "Profile updated successfully",
		},
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleUpdatePreferences(w http.ResponseWriter, r *http.Request) {
	user, _ := identity.UserFromContext(r.Context())
	var req struct {
		PreferredLocale string `json:"preferred_locale"`
		Timezone        string `json:"timezone"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpserver.WriteError(w, r, http.StatusBadRequest, "INVALID_REQUEST_PAYLOAD", "Invalid request body")
		return
	}
	if err := h.service.UpdatePreferences(r.Context(), user.UserID, req.PreferredLocale, req.Timezone); err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data": map[string]string{
			"message": "Preferences updated successfully",
		},
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleGetSessions(w http.ResponseWriter, r *http.Request) {
	user, _ := identity.UserFromContext(r.Context())
	sessions, err := h.service.GetSessions(r.Context(), user.UserID, user.SessionID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       sessions,
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleRevokeSession(w http.ResponseWriter, r *http.Request) {
	user, _ := identity.UserFromContext(r.Context())
	targetID := chi.URLParam(r, "sessionId")
	if targetID == "" {
		httpserver.WriteError(w, r, http.StatusBadRequest, "INVALID_REQUEST_PAYLOAD", "sessionId is required")
		return
	}
	if err := h.service.RevokeSession(r.Context(), user.UserID, targetID); err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data": map[string]string{
			"message": "Session revoked successfully",
		},
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleDeactivate(w http.ResponseWriter, r *http.Request) {
	user, _ := identity.UserFromContext(r.Context())
	if err := h.service.DeactivateAccount(r.Context(), user.UserID); err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data": map[string]string{
			"message": "Account deactivated successfully",
		},
		"request_id": httpserver.RequestID(r),
	})
}

