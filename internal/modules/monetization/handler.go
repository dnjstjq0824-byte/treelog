package monetization

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"treeblog/internal/modules/identity"
	"treeblog/internal/platform/httpserver"
)

// Handler serves billing endpoints.
type Handler struct {
	service       *Service
	tossClientKey string
}

// NewHandler builds a billing handler. tossClientKey is the publishable key
// the frontend widget needs; empty disables the widget config.
func NewHandler(service *Service, tossClientKey string) *Handler {
	return &Handler{service: service, tossClientKey: tossClientKey}
}

// Name implements httpserver.Module.
func (h *Handler) Name() string {
	return "monetization"
}

// Mount implements httpserver.Module.
func (h *Handler) Mount(r chi.Router) {
	r.Get("/api/v1/billing/config", h.handleConfig)
	r.Post("/api/v1/billing/webhooks/toss", h.handleTossWebhook)
	r.Route("/api/v1/billing", func(r chi.Router) {
		r.Use(identity.RequireAuth)
		r.Post("/orders", h.handleCreateOrder)
		r.Get("/orders", h.handleListOrders)
		r.Get("/orders/{orderId}", h.handleGetOrder)
		r.Post("/payments/confirm", h.handleConfirmPayment)
		r.Post("/payouts", h.handleRequestPayout)
	})
	r.Route("/api/v1/blogs/{blogId}/memberships", func(r chi.Router) {
		r.Get("/", h.handleListMemberships)
		r.Route("/", func(r chi.Router) {
			r.Use(identity.RequireAuth)
			r.Post("/", h.handleCreateMembership)
			r.Delete("/{membershipId}", h.handleDeactivateMembership)
		})
	})
	r.Route("/api/v1/admin/payouts", func(r chi.Router) {
		r.Use(identity.RequireAuth)
		r.Use(identity.RequirePlatformRole("SUPER_ADMIN", "FINANCE_ADMIN"))
		r.Get("/", h.handleListPayouts)
		r.Post("/{payoutId}/decision", h.handleDecidePayout)
		r.Post("/{payoutId}/mark-paid", h.handleMarkPayoutPaid)
		r.Get("/{payoutId}/account", h.handleRevealAccount)
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

func (h *Handler) handleConfig(w http.ResponseWriter, r *http.Request) {
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"toss_client_key": h.tossClientKey,
			"payments_enabled": h.tossClientKey != "",
		},
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleTossWebhook(w http.ResponseWriter, r *http.Request) {
	raw, err := io.ReadAll(io.LimitReader(r.Body, 256*1024))
	if err != nil {
		httpserver.WriteError(w, r, http.StatusBadRequest, "INVALID_WEBHOOK", "Invalid webhook payload")
		return
	}
	if err := h.service.HandleWebhook(r.Context(), json.RawMessage(raw)); err != nil {
		var fault *Fault
		if errors.As(err, &fault) && fault.Status >= 500 {
			// 5xx asks Toss to retry delivery.
			httpserver.WriteError(w, r, fault.Status, fault.Code, fault.Message)
			return
		}
		if errors.As(err, &fault) {
			// 4xx would still retry; log-style acknowledge instead.
			httpserver.WriteJSON(w, http.StatusOK, map[string]any{
				"data": map[string]string{"result": "ignored", "code": fault.Code},
				"request_id": httpserver.RequestID(r),
			})
			return
		}
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       map[string]string{"result": "ok"},
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleCreateOrder(w http.ResponseWriter, r *http.Request) {
	user, _ := identity.UserFromContext(r.Context())
	var input CreateOrderInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httpserver.WriteError(w, r, http.StatusBadRequest, "INVALID_REQUEST_PAYLOAD", "Invalid request body")
		return
	}
	order, err := h.service.CreateOrder(r.Context(), user.UserID, input)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusCreated, map[string]any{
		"data": order, "request_id": httpserver.RequestID(r),
	})
}

func queryPaging(r *http.Request) (limit, offset int) {
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

func (h *Handler) handleListOrders(w http.ResponseWriter, r *http.Request) {
	user, _ := identity.UserFromContext(r.Context())
	limit, offset := queryPaging(r)
	orders, err := h.service.ListOrders(r.Context(), user.UserID, limit, offset)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data": orders, "request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleGetOrder(w http.ResponseWriter, r *http.Request) {
	user, _ := identity.UserFromContext(r.Context())
	order, err := h.service.GetOrder(r.Context(), user.UserID, chi.URLParam(r, "orderId"))
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data": order, "request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleConfirmPayment(w http.ResponseWriter, r *http.Request) {
	user, _ := identity.UserFromContext(r.Context())
	var input ConfirmPaymentInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httpserver.WriteError(w, r, http.StatusBadRequest, "INVALID_REQUEST_PAYLOAD", "Invalid request body")
		return
	}
	order, err := h.service.ConfirmPayment(r.Context(), user.UserID, input)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data": order, "request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleListMemberships(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.ListMemberships(r.Context(), chi.URLParam(r, "blogId"))
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data": items, "request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleCreateMembership(w http.ResponseWriter, r *http.Request) {
	user, _ := identity.UserFromContext(r.Context())
	var input MembershipInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httpserver.WriteError(w, r, http.StatusBadRequest, "INVALID_REQUEST_PAYLOAD", "Invalid request body")
		return
	}
	item, err := h.service.CreateMembership(r.Context(), chi.URLParam(r, "blogId"), user.UserID, input)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusCreated, map[string]any{
		"data": item, "request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleDeactivateMembership(w http.ResponseWriter, r *http.Request) {
	user, _ := identity.UserFromContext(r.Context())
	if err := h.service.DeactivateMembership(r.Context(), chi.URLParam(r, "blogId"), chi.URLParam(r, "membershipId"), user.UserID); err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data":       map[string]bool{"deactivated": true},
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleRequestPayout(w http.ResponseWriter, r *http.Request) {
	user, _ := identity.UserFromContext(r.Context())
	var input RequestPayoutInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httpserver.WriteError(w, r, http.StatusBadRequest, "INVALID_REQUEST_PAYLOAD", "Invalid request body")
		return
	}
	payout, err := h.service.RequestPayout(r.Context(), user.UserID, input)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusCreated, map[string]any{
		"data": payout, "request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleListPayouts(w http.ResponseWriter, r *http.Request) {
	limit, offset := queryPaging(r)
	items, err := h.service.ListPayouts(r.Context(), r.URL.Query().Get("status"), limit, offset)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data": items, "request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleDecidePayout(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Decision string `json:"decision"`
		Reason   string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpserver.WriteError(w, r, http.StatusBadRequest, "INVALID_REQUEST_PAYLOAD", "Invalid request body")
		return
	}
	payout, err := h.service.DecidePayout(r.Context(), chi.URLParam(r, "payoutId"), req.Decision, req.Reason)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data": payout, "request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleMarkPayoutPaid(w http.ResponseWriter, r *http.Request) {
	payout, err := h.service.MarkPayoutPaid(r.Context(), chi.URLParam(r, "payoutId"))
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data": payout, "request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleRevealAccount(w http.ResponseWriter, r *http.Request) {
	bankCode, account, holder, err := h.service.RevealPayoutAccount(r.Context(), chi.URLParam(r, "payoutId"))
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data": map[string]string{
			"bank_code": bankCode, "bank_account": account, "account_holder": holder,
		},
		"request_id": httpserver.RequestID(r),
	})
}
