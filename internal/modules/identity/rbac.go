package identity

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"treeblog/internal/platform/httpserver"
)

type contextKey int

const (
	userContextKey contextKey = iota
)

var blogRoleHierarchy = map[string]int{
	"ANALYST": 10,
	"AUTHOR":  20,
	"EDITOR":  30,
	"ADMIN":   40,
	"OWNER":   50,
}

type AuthUser struct {
	UserID         string
	SessionID      string
	PlatformRoles  []string
	IsMFAVerified  bool
	CurrentBlogRole string
}

func (u *AuthUser) HasPlatformRole(role string) bool {
	for _, r := range u.PlatformRoles {
		if r == "SUPER_ADMIN" || r == role {
			return true
		}
	}
	return false
}

func ContextWithUser(ctx context.Context, user *AuthUser) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

func UserFromContext(ctx context.Context) (*AuthUser, bool) {
	u, ok := ctx.Value(userContextKey).(*AuthUser)
	return u, ok && u != nil
}

// AuthenticateMiddleware parses and verifies Bearer JWT and loads user roles.
func AuthenticateMiddleware(signer *TokenSigner, pool *pgxpool.Pool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				next.ServeHTTP(w, r)
				return
			}
			token := strings.TrimPrefix(authHeader, "Bearer ")
			claims, err := signer.Verify(token, time.Now())
			if err != nil {
				httpserver.WriteError(w, r, http.StatusUnauthorized, "AUTH_TOKEN_EXPIRED", "Access token is invalid or expired")
				return
			}

			// Verify session in DB is still active
			var isRevoked bool
			err = pool.QueryRow(r.Context(), "SELECT is_revoked FROM sessions WHERE id=$1 AND user_id=$2",
				claims.SessionID, claims.Subject).Scan(&isRevoked)
			if err != nil || isRevoked {
				httpserver.WriteError(w, r, http.StatusUnauthorized, "AUTH_SESSION_REVOKED", "Session has been revoked")
				return
			}

			// Query platform roles
			rows, err := pool.Query(r.Context(), `SELECT r.code FROM user_roles ur JOIN roles r ON r.id = ur.role_id WHERE ur.user_id=$1`, claims.Subject)
			var roles []string
			if err == nil {
				defer rows.Close()
				for rows.Next() {
					var role string
					if err := rows.Scan(&role); err == nil {
						roles = append(roles, role)
					}
				}
			}

			// Check if session has passed MFA
			var mfaVerified bool
			_ = pool.QueryRow(r.Context(), `SELECT mfa_verified FROM sessions WHERE id=$1`, claims.SessionID).Scan(&mfaVerified)

			authUser := &AuthUser{
				UserID:        claims.Subject,
				SessionID:     claims.SessionID,
				PlatformRoles: roles,
				IsMFAVerified: mfaVerified,
			}

			ctx := ContextWithUser(r.Context(), authUser)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireAuth ensures the user is authenticated.
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := UserFromContext(r.Context())
		if !ok || user.UserID == "" {
			httpserver.WriteError(w, r, http.StatusUnauthorized, "AUTH_TOKEN_MISSING", "Authentication required")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RequirePlatformRole enforces platform-level roles (e.g. SUPER_ADMIN, SUPPORT_AGENT).
func RequirePlatformRole(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := UserFromContext(r.Context())
			if !ok || user.UserID == "" {
				httpserver.WriteError(w, r, http.StatusUnauthorized, "AUTH_TOKEN_MISSING", "Authentication required")
				return
			}
			matched := false
			for _, allowed := range allowedRoles {
				if user.HasPlatformRole(allowed) {
					matched = true
					break
				}
			}
			if !matched {
				httpserver.WriteError(w, r, http.StatusForbidden, "INSUFFICIENT_PLATFORM_ROLE", "Insufficient platform permissions")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireBlogRole checks if the user has at least the required role in the blog.
func RequireBlogRole(pool *pgxpool.Pool, minRole string) func(http.Handler) http.Handler {
	requiredLevel := blogRoleHierarchy[minRole]
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := UserFromContext(r.Context())
			if !ok || user.UserID == "" {
				httpserver.WriteError(w, r, http.StatusUnauthorized, "AUTH_TOKEN_MISSING", "Authentication required")
				return
			}

			// Super admin bypasses blog-level check
			if user.HasPlatformRole("SUPER_ADMIN") {
				next.ServeHTTP(w, r)
				return
			}

			blogID := chi.URLParam(r, "blogId")
			if blogID == "" {
				blogID = chi.URLParam(r, "blog_id")
			}
			if blogID == "" {
				httpserver.WriteError(w, r, http.StatusBadRequest, "INVALID_REQUEST_PAYLOAD", "blogId is required")
				return
			}

			var role string
			err := pool.QueryRow(r.Context(),
				"SELECT role FROM blog_members WHERE blog_id=$1 AND user_id=$2",
				blogID, user.UserID).Scan(&role)
			if errors.Is(err, pgx.ErrNoRows) || blogRoleHierarchy[role] < requiredLevel {
				httpserver.WriteError(w, r, http.StatusForbidden, "INSUFFICIENT_BLOG_ROLE", "Insufficient blog role permissions")
				return
			}
			if err != nil {
				httpserver.WriteError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
				return
			}

			user.CurrentBlogRole = role
			next.ServeHTTP(w, r)
		})
	}
}

// RequireMFA ensures MFA is active and verified for sensitive operations.
func RequireMFA(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := UserFromContext(r.Context())
		if !ok || user.UserID == "" {
			httpserver.WriteError(w, r, http.StatusUnauthorized, "AUTH_TOKEN_MISSING", "Authentication required")
			return
		}
		if !user.IsMFAVerified {
			httpserver.WriteError(w, r, http.StatusForbidden, "MFA_REQUIRED", "MFA verification is required for this action")
			return
		}
		next.ServeHTTP(w, r)
	})
}
