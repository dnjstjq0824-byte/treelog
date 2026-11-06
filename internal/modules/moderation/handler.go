package moderation

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"treeblog/internal/modules/identity"
	"treeblog/internal/platform/httpserver"
	"treeblog/internal/pkg/security"
)

// Handler serves moderation endpoints.
type Handler struct {
	service *Service
}

// NewHandler builds a moderation handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Name implements httpserver.Module.
func (h *Handler) Name() string {
	return "moderation"
}

// Mount implements httpserver.Module.
func (h *Handler) Mount(r chi.Router) {
	r.Route("/api/v1/reports", func(r chi.Router) {
		r.Use(identity.RequireAuth)
		r.Post("/", h.handleReport)
	})
	r.Route("/api/v1/appeals", func(r chi.Router) {
		r.Use(identity.RequireAuth)
		r.Post("/", h.handleFileAppeal)
	})
	r.Route("/api/v1/admin/moderation", func(r chi.Router) {
		r.Use(identity.RequireAuth)
		r.Use(identity.RequirePlatformRole("SUPER_ADMIN", "TRUST_SAFETY_MOD"))
		r.Get("/cases", h.handleListCases)
		r.Get("/cases/{caseId}", h.handleGetCase)
		r.Post("/cases/{caseId}/actions", h.handleApplyAction)
		r.Get("/appeals", h.handleListAppeals)
		r.Post("/appeals/{appealId}/decision", h.handleDecideAppeal)
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

func (h *Handler) handleReport(w http.ResponseWriter, r *http.Request) {
	user, _ := identity.UserFromContext(r.Context())
	var input ReportInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httpserver.WriteError(w, r, http.StatusBadRequest, "INVALID_REQUEST_PAYLOAD", "Invalid request body")
		return
	}
	cse, err := h.service.Report(r.Context(), user.UserID, input, security.ClientIP(r))
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusCreated, map[string]any{
		"data": cse, "request_id": httpserver.RequestID(r),
	})
}

func queryOffset(r *http.Request) (limit, offset int) {
	limit, offset = 30, 0
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if raw := r.URL.Query().Get("offset"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed >= 0 {
			offset = parsed
		}
	}
	return limit, offset
}

func (h *Handler) handleListCases(w http.ResponseWriter, r *http.Request) {
	limit, offset := queryOffset(r)
	items, err := h.service.ListCases(r.Context(), r.URL.Query().Get("status"), limit, offset)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data": items, "request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleGetCase(w http.ResponseWriter, r *http.Request) {
	detail, err := h.service.GetCase(r.Context(), chi.URLParam(r, "caseId"))
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data": detail, "request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleApplyAction(w http.ResponseWriter, r *http.Request) {
	user, _ := identity.UserFromContext(r.Context())
	var input ActionInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httpserver.WriteError(w, r, http.StatusBadRequest, "INVALID_REQUEST_PAYLOAD", "Invalid request body")
		return
	}
	row, err := h.service.ApplyAction(r.Context(), user.UserID, chi.URLParam(r, "caseId"), input)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data": row, "request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleFileAppeal(w http.ResponseWriter, r *http.Request) {
	user, _ := identity.UserFromContext(r.Context())
	var req struct {
		CaseID    string `json:"case_id"`
		Statement string `json:"statement"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpserver.WriteError(w, r, http.StatusBadRequest, "INVALID_REQUEST_PAYLOAD", "Invalid request body")
		return
	}
	appeal, err := h.service.FileAppeal(r.Context(), user.UserID, req.CaseID, req.Statement)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusCreated, map[string]any{
		"data": appeal, "request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleListAppeals(w http.ResponseWriter, r *http.Request) {
	limit, offset := queryOffset(r)
	items, err := h.service.ListAppeals(r.Context(), r.URL.Query().Get("status"), limit, offset)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data": items, "request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleDecideAppeal(w http.ResponseWriter, r *http.Request) {
	user, _ := identity.UserFromContext(r.Context())
	var req struct {
		Decision string `json:"decision"`
		Note     string `json:"note"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpserver.WriteError(w, r, http.StatusBadRequest, "INVALID_REQUEST_PAYLOAD", "Invalid request body")
		return
	}
	appeal, err := h.service.DecideAppeal(r.Context(), user.UserID, chi.URLParam(r, "appealId"), req.Decision, req.Note)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data": appeal, "request_id": httpserver.RequestID(r),
	})
}
