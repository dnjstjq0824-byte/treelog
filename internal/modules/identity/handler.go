package identity

import (
	"crypto/hmac"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
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
		r.Post("/verify-email/resend", h.handleResendVerification)
		r.Post("/login", h.handleLogin)
		r.Post("/refresh", h.handleRefresh)
		r.Post("/logout", h.handleLogout)
		r.Post("/password/forgot", h.handleForgotPassword)
		r.Post("/password/reset", h.handleResetPassword)
		r.Get("/oauth/{provider}/login", h.handleOAuthLogin)
		r.Get("/oauth/{provider}/callback", h.handleOAuthCallback)

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

func (h *Handler) handleResendVerification(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpserver.WriteError(w, r, http.StatusBadRequest, "INVALID_REQUEST_PAYLOAD", "Invalid request body")
		return
	}
	if err := h.service.ResendVerification(r.Context(), req.Email); err != nil {
		h.handleError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"data": map[string]string{
			"message": "If the account is pending verification, a new code has been sent.",
		},
		"request_id": httpserver.RequestID(r),
	})
}

func parseOAuthProvider(raw string) (OAuthProvider, bool) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "google":
		return OAuthGoogle, true
	case "apple":
		return OAuthApple, true
	case "github":
		return OAuthGitHub, true
	case "naver":
		return OAuthNaver, true
	case "kakao":
		return OAuthKakao, true
	case "treescan", "treescan_oidc":
		return OAuthTreescanOIDC, true
	default:
		return "", false
	}
}

func (h *Handler) handleOAuthLogin(w http.ResponseWriter, r *http.Request) {
	provider, ok := parseOAuthProvider(chi.URLParam(r, "provider"))
	if !ok {
		httpserver.WriteError(w, r, http.StatusBadRequest, "UNKNOWN_OAUTH_PROVIDER", "Unknown OAuth provider")
		return
	}
	authURL, signedState, err := h.service.OAuthLoginURL(provider,
		r.URL.Query().Get("terms_version"), r.URL.Query().Get("privacy_version"))
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	isSecure := r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https"
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    signedState,
		Path:     "/api/v1/auth/oauth/",
		MaxAge:   600,
		HttpOnly: true,
		Secure:   isSecure,
		SameSite: http.SameSiteLaxMode,
	})
	http.Redirect(w, r, authURL, http.StatusFound)
}

func (h *Handler) handleOAuthCallback(w http.ResponseWriter, r *http.Request) {
	provider, ok := parseOAuthProvider(chi.URLParam(r, "provider"))
	if !ok {
		httpserver.WriteError(w, r, http.StatusBadRequest, "UNKNOWN_OAUTH_PROVIDER", "Unknown OAuth provider")
		return
	}
	appBase := "http://localhost:3000"
	if cfg, configured := h.service.OAuthConfig(provider); configured && cfg.AppBaseURL != "" {
		appBase = strings.TrimRight(cfg.AppBaseURL, "/")
	}
	fail := func(code string) {
		target := appBase + "/oauth/callback?error=" + code
		http.Redirect(w, r, target, http.StatusFound)
	}
	if providerErr := r.URL.Query().Get("error"); providerErr != "" {
		fail("OAUTH_PROVIDER_DENIED")
		return
	}
	queryState := r.URL.Query().Get("state")
	cookie, err := r.Cookie("oauth_state")
	if err != nil || cookie.Value == "" || !hmac.Equal([]byte(cookie.Value), []byte(queryState)) {
		fail("INVALID_OAUTH_STATE")
		return
	}
	http.SetCookie(w, &http.Cookie{Name: "oauth_state", Path: "/api/v1/auth/oauth/", MaxAge: -1})
	tokens, err := h.service.OAuthCallback(r.Context(), provider, r.URL.Query().Get("code"), queryState)
	if err != nil {
		var fault *Fault
		if errors.As(err, &fault) {
			fail(fault.Code)
			return
		}
		fail("OAUTH_FAILED")
		return
	}
	setRefreshCookie(w, r, tokens.Refresh, 14*24*3600)
	fragment := url.Values{}
	fragment.Set("access_token", tokens.Access)
	fragment.Set("expires_in", strconv.FormatInt(tokens.Expires, 10))
	http.Redirect(w, r, appBase+"/oauth/callback#"+fragment.Encode(), http.StatusFound)
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

