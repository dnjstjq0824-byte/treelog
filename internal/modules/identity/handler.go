package identity

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"treeblog/internal/platform/httpserver"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Name() string {
	return "identity"
}

func (h *Handler) Mount(r chi.Router) {
	r.Route("/api/v1/auth", func(r chi.Router) {
		r.Post("/register", h.handleRegister)
		r.Post("/verify-email", h.handleVerifyEmail)
		r.Post("/login", h.handleLogin)
		r.Post("/refresh", h.handleRefresh)
		r.Post("/logout", h.handleLogout)
		r.Post("/password/forgot", h.handleForgotPassword)
		r.Post("/password/reset", h.handleResetPassword)

		r.Route("/mfa", func(r chi.Router) {
			r.Use(RequireAuth)
			r.Post("/setup", h.handleMFASetup)
			r.Post("/verify", h.handleMFAVerify)
			r.Delete("/", h.handleMFADisable)
		})
	})
}

func (h *Handler) handleError(w http.ResponseWriter, r *http.Request, err error) {
	var fault *Fault
	if errors.As(err, &fault) {
		if fault.RetryAfter > 0 {
			w.Header().Set("Retry-After", strconv.Itoa(fault.RetryAfter))
		}
		httpserver.WriteError(w, r, fault.Status, fault.Code, fault.Message)
		return
	}
	httpserver.WriteError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
}

func (h *Handler) handleRegister(w http.ResponseWriter, r *http.Request) {
	var input RegisterInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httpserver.WriteError(w, r, http.StatusBadRequest, "INVALID_REQUEST_PAYLOAD", "Invalid request body")
		return
	}
	userID, err := h.service.Register(r.Context(), input)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusCreated, map[string]any{
		"data": map[string]string{
			"user_id": userID,
			"status":  "PENDING_VERIFICATION",
			"email":   strings.ToLower(strings.TrimSpace(input.Email)),
		},
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleVerifyEmail(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
		Code  string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpserver.WriteError(w, r, http.StatusBadRequest, "INVALID_REQUEST_PAYLOAD", "Invalid request body")
		return
	}
	if err := h.service.VerifyEmail(r.Context(), req.Email, req.Code); err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data": map[string]string{
			"message": "Email verified successfully",
			"status":  "ACTIVE",
		},
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpserver.WriteError(w, r, http.StatusBadRequest, "INVALID_REQUEST_PAYLOAD", "Invalid request body")
		return
	}
	tokens, err := h.service.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	setRefreshCookie(w, r, tokens.Refresh, 14*24*3600)
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"access_token": tokens.Access,
			"expires_in":   tokens.Expires,
		},
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleRefresh(w http.ResponseWriter, r *http.Request) {
	refreshToken := ""
	if cookie, err := r.Cookie("refresh_token"); err == nil {
		refreshToken = cookie.Value
	}
	if refreshToken == "" {
		var req struct {
			RefreshToken string `json:"refresh_token"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
			refreshToken = req.RefreshToken
		}
	}
	if refreshToken == "" {
		httpserver.WriteError(w, r, http.StatusUnauthorized, "AUTH_TOKEN_MISSING", "Refresh token missing")
		return
	}
	tokens, err := h.service.Refresh(r.Context(), refreshToken)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	setRefreshCookie(w, r, tokens.Refresh, 14*24*3600)
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"access_token": tokens.Access,
			"expires_in":   tokens.Expires,
		},
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleLogout(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	accessToken := strings.TrimPrefix(authHeader, "Bearer ")
	if accessToken == "" || accessToken == authHeader {
		httpserver.WriteError(w, r, http.StatusUnauthorized, "AUTH_TOKEN_MISSING", "Bearer access token required")
		return
	}
	if err := h.service.Logout(r.Context(), accessToken); err != nil {
		h.handleError(w, r, err)
		return
	}
	setRefreshCookie(w, r, "", -1)
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data": map[string]string{
			"message": "Logged out successfully",
		},
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpserver.WriteError(w, r, http.StatusBadRequest, "INVALID_REQUEST_PAYLOAD", "Invalid request body")
		return
	}
	if err := h.service.ForgotPassword(r.Context(), req.Email); err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data": map[string]string{
			"message": "If the email is registered, a reset link has been sent.",
		},
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleResetPassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token       string `json:"token"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpserver.WriteError(w, r, http.StatusBadRequest, "INVALID_REQUEST_PAYLOAD", "Invalid request body")
		return
	}
	if err := h.service.ResetPassword(r.Context(), req.Token, req.NewPassword); err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data": map[string]string{
			"message": "Password has been successfully reset.",
		},
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleMFASetup(w http.ResponseWriter, r *http.Request) {
	user, _ := UserFromContext(r.Context())
	secret, codes, err := h.service.SetupMFA(r.Context(), user.UserID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"secret":       secret,
			"backup_codes": codes,
		},
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleMFAVerify(w http.ResponseWriter, r *http.Request) {
	user, _ := UserFromContext(r.Context())
	var req struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpserver.WriteError(w, r, http.StatusBadRequest, "INVALID_REQUEST_PAYLOAD", "Invalid request body")
		return
	}
	if err := h.service.VerifyMFASetup(r.Context(), user.UserID, req.Code); err != nil {
		h.handleError(w, r, err)
		return
	}
	if err := h.service.MarkSessionMFAVerified(r.Context(), user.SessionID); err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data": map[string]string{"message": "MFA enabled successfully"},
		"request_id": httpserver.RequestID(r),
	})
}

func (h *Handler) handleMFADisable(w http.ResponseWriter, r *http.Request) {
	user, _ := UserFromContext(r.Context())
	var req struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpserver.WriteError(w, r, http.StatusBadRequest, "INVALID_REQUEST_PAYLOAD", "Invalid request body")
		return
	}
	if err := h.service.DisableMFA(r.Context(), user.UserID, req.Code); err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data": map[string]string{"message": "MFA disabled successfully"},
		"request_id": httpserver.RequestID(r),
	})
}

func setRefreshCookie(w http.ResponseWriter, r *http.Request, value string, maxAge int) {
	isSecure := r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https"
	cookie := &http.Cookie{
		Name:     "refresh_token",
		Value:    value,
		Path:     "/api/v1/auth",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   isSecure,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, cookie)
}

