package monetization

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Membership is one membership tier row.
type Membership struct {
	ID                 string     `json:"id"`
	BlogID             string     `json:"blog_id"`
	Name               string     `json:"name"`
	Description        string     `json:"description"`
	MonthlyPriceMinor  int64      `json:"monthly_price_minor"`
	AnnualPriceMinor   *int64     `json:"annual_price_minor,omitempty"`
	BenefitsJSON       string     `json:"benefits_json"`
	IsActive           bool       `json:"is_active"`
	CreatedAt          time.Time  `json:"created_at"`
}

// MembershipInput carries the tier payload.
type MembershipInput struct {
	Name              string  `json:"name"`
	Description       string  `json:"description"`
	MonthlyPriceMinor int64   `json:"monthly_price_minor"`
	AnnualPriceMinor  *int64  `json:"annual_price_minor,omitempty"`
	BenefitsJSON      string  `json:"benefits_json"`
}

// ListMemberships returns a blog's active tiers (public).
func (s *Service) ListMemberships(ctx context.Context, blogID string) ([]Membership, error) {
	if err := s.requirePool(); err != nil {
		return nil, err
	}
	rows, err := s.pool.Query(ctx, `
		SELECT id, blog_id, name, description, monthly_price_minor, annual_price_minor,
		       benefits_json, is_active, created_at
		FROM memberships WHERE blog_id = $1 AND is_active = TRUE
		ORDER BY monthly_price_minor ASC`, blogID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]Membership, 0)
	for rows.Next() {
		var item Membership
		if err := rows.Scan(&item.ID, &item.BlogID, &item.Name, &item.Description,
			&item.MonthlyPriceMinor, &item.AnnualPriceMinor, &item.BenefitsJSON,
			&item.IsActive, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// CreateMembership adds a tier. Callers must hold a blog admin role.
func (s *Service) CreateMembership(ctx context.Context, blogID, userID string, input MembershipInput) (*Membership, error) {
	if strings.TrimSpace(input.Name) == "" || len(input.Name) > 80 {
		return nil, &Fault{Status: http.StatusBadRequest, Code: "INVALID_MEMBERSHIP", Message: "Name is required (max 80 chars)"}
	}
	if input.MonthlyPriceMinor < 1000 {
		return nil, &Fault{Status: http.StatusBadRequest, Code: "INVALID_MEMBERSHIP_PRICE", Message: "Minimum monthly price is 1000 KRW"}
	}
	if input.AnnualPriceMinor != nil && *input.AnnualPriceMinor < input.MonthlyPriceMinor {
		return nil, &Fault{Status: http.StatusBadRequest, Code: "INVALID_MEMBERSHIP_PRICE", Message: "Annual price must cover the monthly price"}
	}
	if err := s.requirePool(); err != nil {
		return nil, err
	}
	if err := s.requireBlogAdmin(ctx, blogID, userID); err != nil {
		return nil, err
	}
	benefits := strings.TrimSpace(input.BenefitsJSON)
	if benefits == "" {
		benefits = "[]"
	}
	var item Membership
	err := s.pool.QueryRow(ctx, `
		INSERT INTO memberships (blog_id, name, description, monthly_price_minor, annual_price_minor, benefits_json)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, blog_id, name, description, monthly_price_minor, annual_price_minor,
		          benefits_json, is_active, created_at
	`, blogID, strings.TrimSpace(input.Name), input.Description, input.MonthlyPriceMinor, input.AnnualPriceMinor, benefits).Scan(
		&item.ID, &item.BlogID, &item.Name, &item.Description, &item.MonthlyPriceMinor,
		&item.AnnualPriceMinor, &item.BenefitsJSON, &item.IsActive, &item.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// DeactivateMembership retires a tier without breaking existing orders.
func (s *Service) DeactivateMembership(ctx context.Context, blogID, membershipID, userID string) error {
	if err := s.requirePool(); err != nil {
		return err
	}
	if err := s.requireBlogAdmin(ctx, blogID, userID); err != nil {
		return err
	}
	res, err := s.pool.Exec(ctx, `UPDATE memberships SET is_active = FALSE, updated_at = NOW()
		WHERE id = $1 AND blog_id = $2`, membershipID, blogID)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return &Fault{Status: http.StatusNotFound, Code: "MEMBERSHIP_NOT_FOUND", Message: "Membership not found"}
	}
	return nil
}

func (s *Service) requireBlogAdmin(ctx context.Context, blogID, userID string) error {
	var role string
	err := s.pool.QueryRow(ctx, `SELECT role FROM blog_members WHERE blog_id = $1 AND user_id = $2`,
		blogID, userID).Scan(&role)
	if errors.Is(err, pgx.ErrNoRows) {
		return &Fault{Status: http.StatusForbidden, Code: "NOT_BLOG_MEMBER", Message: "Blog membership is required"}
	}
	if err != nil {
		return err
	}
	if role != "OWNER" && role != "ADMIN" {
		return &Fault{Status: http.StatusForbidden, Code: "INSUFFICIENT_PERMISSIONS", Message: "Requires blog admin privileges"}
	}
	return nil
}

// Pool exposes the pool for tests.
func (s *Service) Pool() *pgxpool.Pool {
	return s.pool
}
