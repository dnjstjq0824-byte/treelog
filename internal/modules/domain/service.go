package domain

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Domain struct {
	ID                 string     `json:"id"`
	BlogID             string     `json:"blog_id"`
	DomainName         string     `json:"domain_name"`
	DomainType         string     `json:"domain_type"`
	IsPrimary          bool       `json:"is_primary"`
	VerificationToken  string     `json:"verification_token"`
	VerificationStatus string     `json:"verification_status"`
	TLSStatus          string     `json:"tls_status"`
	TLSExpiresAt       *time.Time `json:"tls_expires_at,omitempty"`
	RedirectToPrimary  bool       `json:"redirect_to_primary"`
	LastVerifiedAt     *time.Time `json:"last_verified_at,omitempty"`
	FailureReason      *string    `json:"failure_reason,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

type RegisterDomainInput struct {
	DomainName string `json:"domain_name"`
	DomainType string `json:"domain_type"` // CUSTOM_CNAME or CUSTOM_APEX
}

type Fault struct {
	Status  int
	Code    string
	Message string
}

func (f *Fault) Error() string {
	return f.Message
}

type DNSLookupFunc func(ctx context.Context, host string) ([]string, error)

type Service struct {
	pool      *pgxpool.Pool
	dnsLookup DNSLookupFunc
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{
		pool: pool,
		dnsLookup: func(ctx context.Context, host string) ([]string, error) {
			r := net.Resolver{}
			return r.LookupTXT(ctx, host)
		},
	}
}

func (s *Service) SetDNSLookup(lookup DNSLookupFunc) {
	s.dnsLookup = lookup
}

func (s *Service) CheckBlogPermission(ctx context.Context, blogID, userID string, allowedRoles ...string) error {
	var role string
	err := s.pool.QueryRow(ctx, `
		SELECT role FROM blog_members WHERE blog_id = $1 AND user_id = $2
	`, blogID, userID).Scan(&role)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &Fault{Status: http.StatusForbidden, Code: "FORBIDDEN", Message: "Not a member of this blog"}
		}
		return err
	}
	for _, r := range allowedRoles {
		if role == r {
			return nil
		}
	}
	return &Fault{Status: http.StatusForbidden, Code: "INSUFFICIENT_PERMISSIONS", Message: "Requires blog admin privileges"}
}

func (s *Service) ListDomains(ctx context.Context, blogID, userID string) ([]Domain, error) {
	if err := s.CheckBlogPermission(ctx, blogID, userID, "OWNER", "ADMIN"); err != nil {
		return nil, err
	}

	rows, err := s.pool.Query(ctx, `
		SELECT id, blog_id, domain_name, domain_type, is_primary, verification_token,
		       verification_status, tls_status, tls_expires_at, redirect_to_primary,
		       last_verified_at, failure_reason, created_at, updated_at
		FROM blog_domains
		WHERE blog_id = $1
		ORDER BY is_primary DESC, created_at ASC
	`, blogID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	domains := make([]Domain, 0)
	for rows.Next() {
		var d Domain
		if err := rows.Scan(
			&d.ID, &d.BlogID, &d.DomainName, &d.DomainType, &d.IsPrimary, &d.VerificationToken,
			&d.VerificationStatus, &d.TLSStatus, &d.TLSExpiresAt, &d.RedirectToPrimary,
			&d.LastVerifiedAt, &d.FailureReason, &d.CreatedAt, &d.UpdatedAt,
		); err != nil {
			return nil, err
		}
		domains = append(domains, d)
	}
	return domains, rows.Err()
}

func (s *Service) RegisterDomain(ctx context.Context, blogID, userID string, input RegisterDomainInput) (*Domain, error) {
	name := strings.ToLower(strings.TrimSpace(input.DomainName))
	if name == "" || len(name) > 253 {
		return nil, &Fault{Status: http.StatusBadRequest, Code: "INVALID_DOMAIN_NAME", Message: "Domain name is required and must not exceed 253 characters"}
	}
	if strings.Contains(name, " ") || !strings.Contains(name, ".") {
		return nil, &Fault{Status: http.StatusBadRequest, Code: "INVALID_DOMAIN_NAME", Message: "Invalid domain format"}
	}

	dType := strings.ToUpper(strings.TrimSpace(input.DomainType))
	if dType != "CUSTOM_CNAME" && dType != "CUSTOM_APEX" {
		dType = "CUSTOM_CNAME"
	}

	if s.pool != nil {
		if err := s.CheckBlogPermission(ctx, blogID, userID, "OWNER", "ADMIN"); err != nil {
			return nil, err
		}
	}

	// Generate verification token
	tokenBytes := make([]byte, 16)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, err
	}
	verificationToken := "treeblog-verify=" + hex.EncodeToString(tokenBytes)

	var d Domain
	err := s.pool.QueryRow(ctx, `
		INSERT INTO blog_domains (
			blog_id, domain_name, domain_type, is_primary, verification_token,
			verification_status, tls_status, redirect_to_primary
		) VALUES ($1, $2, $3, FALSE, $4, 'PENDING_DNS', 'NONE', TRUE)
		RETURNING id, blog_id, domain_name, domain_type, is_primary, verification_token,
		          verification_status, tls_status, tls_expires_at, redirect_to_primary,
		          last_verified_at, failure_reason, created_at, updated_at
	`, blogID, name, dType, verificationToken).Scan(
		&d.ID, &d.BlogID, &d.DomainName, &d.DomainType, &d.IsPrimary, &d.VerificationToken,
		&d.VerificationStatus, &d.TLSStatus, &d.TLSExpiresAt, &d.RedirectToPrimary,
		&d.LastVerifiedAt, &d.FailureReason, &d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			return nil, &Fault{Status: http.StatusConflict, Code: "DOMAIN_ALREADY_REGISTERED", Message: "Domain is already registered by another blog"}
		}
		return nil, err
	}

	return &d, nil
}

func (s *Service) VerifyDomain(ctx context.Context, blogID, domainID, userID string) (*Domain, error) {
	if err := s.CheckBlogPermission(ctx, blogID, userID, "OWNER", "ADMIN"); err != nil {
		return nil, err
	}

	var d Domain
	err := s.pool.QueryRow(ctx, `
		SELECT id, blog_id, domain_name, domain_type, is_primary, verification_token,
		       verification_status, tls_status, tls_expires_at, redirect_to_primary,
		       last_verified_at, failure_reason, created_at, updated_at
		FROM blog_domains
		WHERE id = $1 AND blog_id = $2
	`, domainID, blogID).Scan(
		&d.ID, &d.BlogID, &d.DomainName, &d.DomainType, &d.IsPrimary, &d.VerificationToken,
		&d.VerificationStatus, &d.TLSStatus, &d.TLSExpiresAt, &d.RedirectToPrimary,
		&d.LastVerifiedAt, &d.FailureReason, &d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &Fault{Status: http.StatusNotFound, Code: "DOMAIN_NOT_FOUND", Message: "Domain record not found"}
		}
		return nil, err
	}

	// Lookup DNS TXT record: _treeblog-challenge.<domain>
	challengeHost := "_treeblog-challenge." + d.DomainName
	records, lookupErr := s.dnsLookup(ctx, challengeHost)

	verified := false
	if lookupErr == nil {
		for _, rec := range records {
			if strings.TrimSpace(rec) == d.VerificationToken {
				verified = true
				break
			}
		}
	}

	now := time.Now().UTC()
	if verified {
		d.VerificationStatus = "VERIFIED"
		d.TLSStatus = "ACTIVE"
		expiry := now.Add(90 * 24 * time.Hour)
		d.TLSExpiresAt = &expiry
		d.LastVerifiedAt = &now
		d.FailureReason = nil

		_, err = s.pool.Exec(ctx, `
			UPDATE blog_domains
			SET verification_status = 'VERIFIED',
			    tls_status = 'ACTIVE',
			    tls_expires_at = $1,
			    last_verified_at = $2,
			    failure_reason = NULL,
			    updated_at = NOW()
			WHERE id = $3
		`, d.TLSExpiresAt, d.LastVerifiedAt, d.ID)
	} else {
		reason := fmt.Sprintf("TXT record %s with expected value not found", challengeHost)
		if lookupErr != nil {
			reason = fmt.Sprintf("DNS lookup failed for %s: %v", challengeHost, lookupErr)
		}
		d.VerificationStatus = "FAILED"
		d.FailureReason = &reason

		_, err = s.pool.Exec(ctx, `
			UPDATE blog_domains
			SET verification_status = 'FAILED',
			    failure_reason = $1,
			    updated_at = NOW()
			WHERE id = $2
		`, reason, d.ID)
	}
	if err != nil {
		return nil, err
	}

	return &d, nil
}

func (s *Service) SetPrimaryDomain(ctx context.Context, blogID, domainID, userID string) (*Domain, error) {
	if err := s.CheckBlogPermission(ctx, blogID, userID, "OWNER", "ADMIN"); err != nil {
		return nil, err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var status string
	err = tx.QueryRow(ctx, `
		SELECT verification_status FROM blog_domains WHERE id = $1 AND blog_id = $2
	`, domainID, blogID).Scan(&status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &Fault{Status: http.StatusNotFound, Code: "DOMAIN_NOT_FOUND", Message: "Domain not found"}
		}
		return nil, err
	}
	if status != "VERIFIED" {
		return nil, &Fault{Status: http.StatusBadRequest, Code: "DOMAIN_NOT_VERIFIED", Message: "Domain must be verified before setting as primary"}
	}

	// Reset all domains for this blog to is_primary = false
	if _, err := tx.Exec(ctx, `UPDATE blog_domains SET is_primary = FALSE, updated_at = NOW() WHERE blog_id = $1`, blogID); err != nil {
		return nil, err
	}

	var d Domain
	err = tx.QueryRow(ctx, `
		UPDATE blog_domains
		SET is_primary = TRUE, updated_at = NOW()
		WHERE id = $1 AND blog_id = $2
		RETURNING id, blog_id, domain_name, domain_type, is_primary, verification_token,
		          verification_status, tls_status, tls_expires_at, redirect_to_primary,
		          last_verified_at, failure_reason, created_at, updated_at
	`, domainID, blogID).Scan(
		&d.ID, &d.BlogID, &d.DomainName, &d.DomainType, &d.IsPrimary, &d.VerificationToken,
		&d.VerificationStatus, &d.TLSStatus, &d.TLSExpiresAt, &d.RedirectToPrimary,
		&d.LastVerifiedAt, &d.FailureReason, &d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &d, nil
}

func (s *Service) DeleteDomain(ctx context.Context, blogID, domainID, userID string) error {
	if err := s.CheckBlogPermission(ctx, blogID, userID, "OWNER", "ADMIN"); err != nil {
		return err
	}

	var isPrimary bool
	err := s.pool.QueryRow(ctx, `SELECT is_primary FROM blog_domains WHERE id = $1 AND blog_id = $2`, domainID, blogID).Scan(&isPrimary)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &Fault{Status: http.StatusNotFound, Code: "DOMAIN_NOT_FOUND", Message: "Domain not found"}
		}
		return err
	}
	if isPrimary {
		return &Fault{Status: http.StatusBadRequest, Code: "CANNOT_DELETE_PRIMARY_DOMAIN", Message: "Cannot delete the primary domain. Set another domain as primary first."}
	}

	res, err := s.pool.Exec(ctx, `DELETE FROM blog_domains WHERE id = $1 AND blog_id = $2`, domainID, blogID)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return &Fault{Status: http.StatusNotFound, Code: "DOMAIN_NOT_FOUND", Message: "Domain not found"}
	}
	return nil
}

func (s *Service) ResolveHost(ctx context.Context, host string) (blogID string, isPrimary bool, redirectTo string, err error) {
	host = strings.ToLower(strings.Split(host, ":")[0]) // strip port if any

	// 1. Direct custom domain match in blog_domains
	var domainName string
	var redirect bool
	err = s.pool.QueryRow(ctx, `
		SELECT blog_id, is_primary, redirect_to_primary, domain_name
		FROM blog_domains
		WHERE domain_name = $1 AND verification_status = 'VERIFIED'
	`, host).Scan(&blogID, &isPrimary, &redirect, &domainName)
	if err == nil {
		if !isPrimary && redirect {
			// Find primary domain for this blog
			var primaryDomain string
			if pErr := s.pool.QueryRow(ctx, `SELECT domain_name FROM blog_domains WHERE blog_id = $1 AND is_primary = TRUE`, blogID).Scan(&primaryDomain); pErr == nil {
				return blogID, false, "https://" + primaryDomain, nil
			}
		}
		return blogID, isPrimary, "", nil
	}

	// 2. Subdomain check (e.g. handle.treeblog.net)
	if strings.HasSuffix(host, ".treeblog.net") {
		handle := strings.TrimSuffix(host, ".treeblog.net")
		var bID string
		err = s.pool.QueryRow(ctx, `SELECT id FROM blogs WHERE handle = $1 AND status = 'ACTIVE'`, handle).Scan(&bID)
		if err == nil {
			// Check if blog has a primary custom domain
			var primaryDomain string
			if pErr := s.pool.QueryRow(ctx, `SELECT domain_name FROM blog_domains WHERE blog_id = $1 AND is_primary = TRUE`, bID).Scan(&primaryDomain); pErr == nil {
				return bID, false, "https://" + primaryDomain, nil
			}
			return bID, true, "", nil
		}
	}

	return "", false, "", errors.New("domain not found")
}
