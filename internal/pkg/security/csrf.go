package security

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
	"strings"

	"treeblog/internal/platform/httpserver"
)

const (
	CSRFCookieName = "XSRF-TOKEN"
	CSRFHeaderName = "X-CSRF-Token"
	CSRFAltHeader  = "X-XSRF-Token"
)

// GenerateCSRFToken generates a cryptographically secure 32-byte hex token.
func GenerateCSRFToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// EnsureCSRFCookie checks if the client has a CSRF cookie, and sets one if absent.
func EnsureCSRFCookie(w http.ResponseWriter, r *http.Request) string {
	if cookie, err := r.Cookie(CSRFCookieName); err == nil && cookie.Value != "" {
		return cookie.Value
	}
	token, err := GenerateCSRFToken()
	if err != nil {
		return ""
	}
	isSecure := r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https"
	http.SetCookie(w, &http.Cookie{
		Name:     CSRFCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: false, // Accessible to SPA JS for Double-Submit pattern
		Secure:   isSecure,
		SameSite: http.SameSiteLaxMode,
	})
	return token
}

// CSRFMiddleware enforces Double-Submit CSRF token validation on mutating HTTP requests.
func CSRFMiddleware(exemptPaths ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			EnsureCSRFCookie(w, r)

			// Skip safe methods
			switch r.Method {
			case http.MethodGet, http.MethodHead, http.MethodOptions, "TRACE":
				next.ServeHTTP(w, r)
				return
			}

			// Check exempt paths (e.g. webhooks, public auth endpoints using Bearer)
			for _, path := range exemptPaths {
				if strings.HasPrefix(r.URL.Path, path) {
					next.ServeHTTP(w, r)
					return
				}
			}

			cookie, err := r.Cookie(CSRFCookieName)
			if err != nil || cookie.Value == "" {
				httpserver.WriteError(w, r, http.StatusForbidden, "CSRF_TOKEN_MISSING", "CSRF cookie is missing")
				return
			}

			headerToken := r.Header.Get(CSRFHeaderName)
			if headerToken == "" {
				headerToken = r.Header.Get(CSRFAltHeader)
			}
			if headerToken == "" {
				httpserver.WriteError(w, r, http.StatusForbidden, "CSRF_TOKEN_MISSING", "CSRF token header is missing")
				return
			}

			if subtle.ConstantTimeCompare([]byte(cookie.Value), []byte(headerToken)) != 1 {
				httpserver.WriteError(w, r, http.StatusForbidden, "CSRF_TOKEN_INVALID", "Invalid CSRF token")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

