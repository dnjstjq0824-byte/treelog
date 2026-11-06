package admin

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Fault is the module error envelope.
type Fault struct {
	Status  int
	Code    string
	Message string
}

func (f *Fault) Error() string { return f.Message }

// Service implements platform administration queries and user enforcement.
type Service struct {
	pool *pgxpool.Pool
}

// NewService builds an admin service.
func NewService(pool *pgxpool.Pool) (*Service, error) {
	if pool == nil {
		return nil, errors.New("admin service requires a database pool")
	}
	return &Service{pool: pool}, nil
}

func (s *Service) requirePool() error {
	if s.pool == nil {
		return &Fault{Status: http.StatusServiceUnavailable, Code: "SERVICE_UNAVAILABLE", Message: "Database is unavailable"}
	}
	return nil
}

// PlatformStats aggregates headline platform metrics.
type PlatformStats struct {
	UsersTotal       int64 `json:"users_total"`
	UsersActive      int64 `json:"users_active"`
	UsersSuspended   int64 `json:"users_suspended"`
	BlogsTotal       int64 `json:"blogs_total"`
	BlogsSuspended   int64 `json:"blogs_suspended"`
	PostsPublished   int64 `json:"posts_published"`
	PostsToday       int64 `json:"posts_today"`
	CommentsTotal    int64 `json:"comments_total"`
	MediaReady       int64 `json:"media_ready"`
	CasesOpen        int64 `json:"cases_open"`
	AppealsPending   int64 `json:"appeals_pending"`
	OrdersPaidToday  int64 `json:"orders_paid_today"`
	RevenueTodayMinor int64 `json:"revenue_today_minor"`
}

// Stats returns headline platform metrics in one round trip.
func (s *Service) Stats(ctx context.Context) (*PlatformStats, error) {
	if err := s.requirePool(); err != nil {
		return nil, err
	}
	var stats PlatformStats
	err := s.pool.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(*) FROM users),
			(SELECT COUNT(*) FROM users WHERE status = 'ACTIVE'),
			(SELECT COUNT(*) FROM users WHERE status = 'SUSPENDED'),
			(SELECT COUNT(*) FROM blogs),
			(SELECT COUNT(*) FROM blogs WHERE status = 'SUSPENDED'),
			(SELECT COUNT(*) FROM posts WHERE status = 'PUBLISHED'),
			(SELECT COUNT(*) FROM posts WHERE status = 'PUBLISHED' AND published_at >= NOW() - INTERVAL '24 hours'),
			(SELECT COUNT(*) FROM comments WHERE status = 'PUBLISHED'),
			(SELECT COUNT(*) FROM media_assets WHERE status = 'READY'),
			(SELECT COUNT(*) FROM moderation_cases WHERE status IN ('OPEN', 'IN_REVIEW', 'APPEALED')),
			(SELECT COUNT(*) FROM appeals WHERE status IN ('PENDING', 'UNDER_REVIEW')),
			(SELECT COUNT(*) FROM orders WHERE status = 'PAID' AND updated_at >= NOW() - INTERVAL '24 hours'),
			COALESCE((SELECT SUM(gross_amount_minor) FROM orders WHERE status = 'PAID' AND updated_at >= NOW() - INTERVAL '24 hours'), 0)
	`).Scan(
		&stats.UsersTotal, &stats.UsersActive, &stats.UsersSuspended,
		&stats.BlogsTotal, &stats.BlogsSuspended,
		&stats.PostsPublished, &stats.PostsToday, &stats.CommentsTotal, &stats.MediaReady,
		&stats.CasesOpen, &stats.AppealsPending,
		&stats.OrdersPaidToday, &stats.RevenueTodayMinor,
	)
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

// AdminUser is one row of the user directory.
type AdminUser struct {
	ID          string    `json:"id"`
	Email       string    `json:"email"`
	Status      string    `json:"status"`
	Username    string    `json:"username"`
	DisplayName string    `json:"display_name"`
	Roles       []string  `json:"roles"`
	TrustTier   int       `json:"trust_tier"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// ListUsers searches the user directory with offset paging.
func (s *Service) ListUsers(ctx context.Context, query, status string, limit, offset int) ([]AdminUser, error) {
	if err := s.requirePool(); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 100 {
		limit = 30
	}
	if offset < 0 {
		offset = 0
	}
	sql := `
		SELECT u.id, u.email, u.status, COALESCE(up.username, ''), COALESCE(up.display_name, ''),
		       u.trust_tier, u.last_login_at, u.created_at
		FROM users u LEFT JOIN user_profiles up ON up.user_id = u.id
		WHERE ($1 = '' OR u.email ILIKE '%' || $1 || '%' OR up.username ILIKE '%' || $1 || '%')
		  AND ($2 = '' OR u.status = $2)
		ORDER BY u.created_at DESC LIMIT $3 OFFSET $4
	`
	rows, err := s.pool.Query(ctx, sql, strings.TrimSpace(query), strings.ToUpper(strings.TrimSpace(status)), limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	users := make([]AdminUser, 0)
	for rows.Next() {
		var user AdminUser
		if err := rows.Scan(&user.ID, &user.Email, &user.Status, &user.Username, &user.DisplayName,
			&user.TrustTier, &user.LastLoginAt, &user.CreatedAt); err != nil {
			return nil, err
		}
		user.Roles = s.rolesFor(ctx, user.ID)
		users = append(users, user)
	}
	return users, rows.Err()
}

func (s *Service) rolesFor(ctx context.Context, userID string) []string {
	roles := []string{}
	rows, err := s.pool.Query(ctx, `SELECT r.code FROM user_roles ur
		JOIN roles r ON r.id = ur.role_id WHERE ur.user_id = $1 ORDER BY r.code`, userID)
	if err != nil {
		return roles
	}
	defer rows.Close()
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err == nil {
			roles = append(roles, code)
		}
	}
	return roles
}

// SuspendUser suspends an account and revokes its sessions.
func (s *Service) SuspendUser(ctx context.Context, userID string) error {
	if err := s.requirePool(); err != nil {
		return err
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	res, err := tx.Exec(ctx, `UPDATE users SET status = 'SUSPENDED', updated_at = NOW()
		WHERE id = $1 AND status = 'ACTIVE'`, userID)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		var exists bool
		if err := tx.QueryRow(ctx, `SELECT TRUE FROM users WHERE id = $1`, userID).Scan(&exists); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return &Fault{Status: http.StatusNotFound, Code: "USER_NOT_FOUND", Message: "User not found"}
			}
			return err
		}
		return &Fault{Status: http.StatusConflict, Code: "USER_NOT_ACTIVE", Message: "Only active accounts can be suspended"}
	}
	if _, err := tx.Exec(ctx, `UPDATE sessions SET is_revoked = TRUE, revoked_reason = 'ADMIN_SUSPENDED'
		WHERE user_id = $1 AND is_revoked = FALSE`, userID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// UnsuspendUser re-activates a suspended account.
func (s *Service) UnsuspendUser(ctx context.Context, userID string) error {
	if err := s.requirePool(); err != nil {
		return err
	}
	res, err := s.pool.Exec(ctx, `UPDATE users SET status = 'ACTIVE', updated_at = NOW()
		WHERE id = $1 AND status = 'SUSPENDED'`, userID)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return &Fault{Status: http.StatusConflict, Code: "USER_NOT_SUSPENDED", Message: "User is not suspended"}
	}
	return nil
}

// AdminBlog is one row of the blog directory.
type AdminBlog struct {
	ID        string    `json:"id"`
	Handle    string    `json:"handle"`
	Title     string    `json:"title"`
	Status    string    `json:"status"`
	OwnerID   string    `json:"owner_user_id"`
	PostCount int       `json:"post_count"`
	CreatedAt time.Time `json:"created_at"`
}

// ListBlogs lists blogs with offset paging.
func (s *Service) ListBlogs(ctx context.Context, status string, limit, offset int) ([]AdminBlog, error) {
	if err := s.requirePool(); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 100 {
		limit = 30
	}
	if offset < 0 {
		offset = 0
	}
	rows, err := s.pool.Query(ctx, `
		SELECT id, handle, title, status, owner_user_id, post_count, created_at
		FROM blogs
		WHERE ($1 = '' OR status = $1)
		ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		strings.ToUpper(strings.TrimSpace(status)), limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	blogs := make([]AdminBlog, 0)
	for rows.Next() {
		var blog AdminBlog
		if err := rows.Scan(&blog.ID, &blog.Handle, &blog.Title, &blog.Status, &blog.OwnerID, &blog.PostCount, &blog.CreatedAt); err != nil {
			return nil, err
		}
		blogs = append(blogs, blog)
	}
	return blogs, rows.Err()
}
