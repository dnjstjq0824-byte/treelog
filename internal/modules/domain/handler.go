package domain

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
	return "domain"
}

func (h *Handler) Mount(r chi.Router) {
	r.Route("/api/v1/blogs/{blogId}/domains", func(r chi.Router) {
		r.Use(identity.RequireAuth)

		r.Get("/", h.handleListDomains)
		r.Post("/", h.handleRegisterDomain)
		r.Post("/{domainId}/verify", h.handleVerifyDomain)
		r.Put("/{domainId}/primary", h.handleSetPrimaryDomain)
		r.Delete("/{domainId}", h.handleDeleteDomain)
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

func (h *Handler) handleListDomains(w http.ResponseWriter, r *http.Request) {
	user, _ := identity.UserFromContext(r.Context())
	blogID := chi.URLParam(r, "blogId")

	domains, err := h.service.ListDomains(r.Context(), blogID, user.UserID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       domains,
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleRegisterDomain(w http.ResponseWriter, r *http.Request) {
	user, _ := identity.UserFromContext(r.Context())
	blogID := chi.URLParam(r, "blogId")

	var input RegisterDomainInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httpserver.WriteError(w, r, http.StatusBadRequest, "INVALID_REQUEST_PAYLOAD", "Invalid request body")
		return
	}

	domain, err := h.service.RegisterDomain(r.Context(), blogID, user.UserID, input)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       domain,
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleVerifyDomain(w http.ResponseWriter, r *http.Request) {
	user, _ := identity.UserFromContext(r.Context())
	blogID := chi.URLParam(r, "blogId")
	domainID := chi.URLParam(r, "domainId")

	domain, err := h.service.VerifyDomain(r.Context(), blogID, domainID, user.UserID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       domain,
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleSetPrimaryDomain(w http.ResponseWriter, r *http.Request) {
	user, _ := identity.UserFromContext(r.Context())
	blogID := chi.URLParam(r, "blogId")
	domainID := chi.URLParam(r, "domainId")

	domain, err := h.service.SetPrimaryDomain(r.Context(), blogID, domainID, user.UserID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       domain,
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleDeleteDomain(w http.ResponseWriter, r *http.Request) {
	user, _ := identity.UserFromContext(r.Context())
	blogID := chi.URLParam(r, "blogId")
	domainID := chi.URLParam(r, "domainId")

	err := h.service.DeleteDomain(r.Context(), blogID, domainID, user.UserID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"deleted":   true,
			"domain_id": domainID,
		},
		"request_id": httpserver.RequestID(r),
	})
}
