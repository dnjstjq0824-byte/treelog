package moderation

import (
	"context"
	"encoding/json"
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

// Service implements reports, cases, enforcement actions, and appeals.
type Service struct {
	pool *pgxpool.Pool
}

// NewService builds a moderation service.
func NewService(pool *pgxpool.Pool) (*Service, error) {
	if pool == nil {
		return nil, errors.New("moderation service requires a database pool")
	}
	return &Service{pool: pool}, nil
}

func (s *Service) requirePool() error {
	if s.pool == nil {
		return &Fault{Status: http.StatusServiceUnavailable, Code: "SERVICE_UNAVAILABLE", Message: "Database is unavailable"}
	}
	return nil
}

var validTargetTypes = map[string]bool{
	"POST": true, "COMMENT": true, "GUESTBOOK": true, "BLOG": true, "USER": true, "AD_CREATIVE": true,
}

var validReasonCategories = map[string]bool{
	"SPAM": true, "COPYRIGHT_DMCA": true, "DEFAMATION": true, "PRIVACY_LEAK": true,
	"ADULT_ILLEGAL": true, "PHISHING_MALWARE": true, "IMPERSONATION": true, "HATE_SPEECH": true,
}

var validActionTypes = map[string]bool{
	"DISMISS": true, "WARN_USER": true, "QUARANTINE_CONTENT": true, "RESTORE_CONTENT": true,
	"DELETE_CONTENT": true, "SUSPEND_BLOG_7D": true, "SUSPEND_BLOG_PERM": true, "SUSPEND_USER": true,
}

// ReportInput carries the user report payload.
type ReportInput struct {
	TargetType               string   `json:"target_type"`
	TargetID                 string   `json:"target_id"`
	ReasonCategory           string   `json:"reason_category"`
	Description              string   `json:"description"`
	EvidenceURLs             []string `json:"evidence_urls"`
	CopyrightOwnerConfirmed  bool     `json:"is_copyright_owner_confirmed"`
}

// ModerationCase is one case row.
type ModerationCase struct {
	ID             string    `json:"id"`
	TargetType     string    `json:"target_type"`
	TargetID       string    `json:"target_id"`
	ReasonCategory string    `json:"reason_category"`
	ReportCount    int       `json:"report_count"`
	PriorityScore  int       `json:"priority_score"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
}

// Report files a user report, creating or reusing the matching case.
func (s *Service) Report(ctx context.Context, reporterID string, input ReportInput, reporterIP string) (*ModerationCase, error) {
	targetType := strings.ToUpper(strings.TrimSpace(input.TargetType))
	reason := strings.ToUpper(strings.TrimSpace(input.ReasonCategory))
	if !validTargetTypes[targetType] {
		return nil, &Fault{Status: http.StatusBadRequest, Code: "INVALID_TARGET_TYPE", Message: "Unknown report target type"}
	}
	if !validReasonCategories[reason] {
		return nil, &Fault{Status: http.StatusBadRequest, Code: "INVALID_REASON_CATEGORY", Message: "Unknown reason category"}
	}
	if strings.TrimSpace(input.TargetID) == "" {
		return nil, &Fault{Status: http.StatusBadRequest, Code: "INVALID_TARGET_ID", Message: "Target id is required"}
	}
	if len(input.Description) > 2000 {
		return nil, &Fault{Status: http.StatusBadRequest, Code: "INVALID_DESCRIPTION", Message: "Description exceeds 2000 characters"}
	}
	if err := s.requirePool(); err != nil {
		return nil, err
	}
	if err := s.targetExists(ctx, targetType, input.TargetID); err != nil {
		return nil, err
	}
	evidence, _ := json.Marshal(input.EvidenceURLs)
	if evidence == nil {
		evidence = []byte("[]")
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var cse ModerationCase
	err = tx.QueryRow(ctx, `
		INSERT INTO moderation_cases (target_type, target_id, reason_category)
		VALUES ($1, $2, $3)
		ON CONFLICT (target_type, target_id, reason_category) DO UPDATE
		SET report_count = moderation_cases.report_count + 1,
		    priority_score = LEAST(moderation_cases.priority_score + 5, 100),
		    status = CASE WHEN moderation_cases.status IN ('DISMISSED', 'APPEAL_RESOLVED')
		                  THEN 'OPEN' ELSE moderation_cases.status END,
		    updated_at = NOW()
		RETURNING id, target_type, target_id, reason_category, report_count, priority_score, status, created_at
	`, targetType, input.TargetID, reason).Scan(
		&cse.ID, &cse.TargetType, &cse.TargetID, &cse.ReasonCategory,
		&cse.ReportCount, &cse.PriorityScore, &cse.Status, &cse.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	var ipParam *string
	if reporterIP != "" {
		ipParam = &reporterIP
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO user_reports (case_id, reporter_user_id, reason_category, description, evidence_urls_json, is_copyright_owner_confirmed, reporter_ip)
		VALUES ($1, $2, $3, $4, $5, $6, $7::inet)
	`, cse.ID, reporterID, reason, input.Description, string(evidence), input.CopyrightOwnerConfirmed, ipParam); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &cse, nil
}

func (s *Service) targetExists(ctx context.Context, targetType, targetID string) error {
	var query string
	switch targetType {
	case "POST":
		query = `SELECT TRUE FROM posts WHERE id = $1`
	case "COMMENT":
		query = `SELECT TRUE FROM comments WHERE id = $1`
	case "GUESTBOOK":
		query = `SELECT TRUE FROM guestbook_entries WHERE id = $1`
	case "BLOG":
		query = `SELECT TRUE FROM blogs WHERE id = $1`
	case "USER":
		query = `SELECT TRUE FROM users WHERE id = $1`
	case "AD_CREATIVE":
		query = `SELECT TRUE FROM ad_campaigns WHERE id = $1`
	default:
		return &Fault{Status: http.StatusBadRequest, Code: "INVALID_TARGET_TYPE", Message: "Unknown report target type"}
	}
	var exists bool
	if err := s.pool.QueryRow(ctx, query, targetID).Scan(&exists); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &Fault{Status: http.StatusNotFound, Code: "REPORT_TARGET_NOT_FOUND", Message: "Reported target does not exist"}
		}
		return err
	}
	return nil
}

// ListCases returns open-queue cases newest-report-first with offset paging.
func (s *Service) ListCases(ctx context.Context, status string, limit, offset int) ([]ModerationCase, error) {
	if err := s.requirePool(); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 100 {
		limit = 30
	}
	if offset < 0 {
		offset = 0
	}
	query := `
		SELECT id, target_type, target_id, reason_category, report_count, priority_score, status, created_at
		FROM moderation_cases
	`
	args := []any{}
	if status != "" {
		query += ` WHERE status = $1`
		args = append(args, strings.ToUpper(status))
		query += ` ORDER BY priority_score DESC, created_at DESC LIMIT $2 OFFSET $3`
		args = append(args, limit, offset)
	} else {
		query += ` ORDER BY priority_score DESC, created_at DESC LIMIT $1 OFFSET $2`
		args = append(args, limit, offset)
	}
	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]ModerationCase, 0)
	for rows.Next() {
		var cse ModerationCase
		if err := rows.Scan(&cse.ID, &cse.TargetType, &cse.TargetID, &cse.ReasonCategory,
			&cse.ReportCount, &cse.PriorityScore, &cse.Status, &cse.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, cse)
	}
	return items, rows.Err()
}

// CaseDetail bundles a case with its reports and actions.
type CaseDetail struct {
	Case    ModerationCase   `json:"case"`
	Reports []ReportRow      `json:"reports"`
	Actions []ActionRow      `json:"actions"`
}

// ReportRow is one user report.
type ReportRow struct {
	ID              string    `json:"id"`
	ReporterUserID  *string   `json:"reporter_user_id"`
	ReasonCategory  string    `json:"reason_category"`
	Description     string    `json:"description"`
	CreatedAt       time.Time `json:"created_at"`
}

// ActionRow is one enforcement action.
type ActionRow struct {
	ID              string    `json:"id"`
	ActorModeratorID string   `json:"actor_moderator_id"`
	ActionType      string    `json:"action_type"`
	PolicyClause    string    `json:"policy_clause_code"`
	PublicNotice    string    `json:"public_notice_reason"`
	CreatedAt       time.Time `json:"created_at"`
}

// GetCase returns a case with reports and actions.
func (s *Service) GetCase(ctx context.Context, caseID string) (*CaseDetail, error) {
	if err := s.requirePool(); err != nil {
		return nil, err
	}
	var cse ModerationCase
	err := s.pool.QueryRow(ctx, `
		SELECT id, target_type, target_id, reason_category, report_count, priority_score, status, created_at
		FROM moderation_cases WHERE id = $1`, caseID).Scan(
		&cse.ID, &cse.TargetType, &cse.TargetID, &cse.ReasonCategory,
		&cse.ReportCount, &cse.PriorityScore, &cse.Status, &cse.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, &Fault{Status: http.StatusNotFound, Code: "CASE_NOT_FOUND", Message: "Moderation case not found"}
	}
	if err != nil {
		return nil, err
	}
	detail := &CaseDetail{Case: cse, Reports: []ReportRow{}, Actions: []ActionRow{}}
	rows, err := s.pool.Query(ctx, `
		SELECT id, reporter_user_id, reason_category, description, created_at
		FROM user_reports WHERE case_id = $1 ORDER BY created_at ASC`, caseID)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var row ReportRow
		if err := rows.Scan(&row.ID, &row.ReporterUserID, &row.ReasonCategory, &row.Description, &row.CreatedAt); err != nil {
			rows.Close()
			return nil, err
		}
		detail.Reports = append(detail.Reports, row)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}
	arows, err := s.pool.Query(ctx, `
		SELECT id, actor_moderator_id, action_type, policy_clause_code, public_notice_reason, created_at
		FROM moderation_actions WHERE case_id = $1 ORDER BY created_at ASC`, caseID)
	if err != nil {
		return nil, err
	}
	for arows.Next() {
		var row ActionRow
		if err := arows.Scan(&row.ID, &row.ActorModeratorID, &row.ActionType, &row.PolicyClause, &row.PublicNotice, &row.CreatedAt); err != nil {
			arows.Close()
			return nil, err
		}
		detail.Actions = append(detail.Actions, row)
	}
	arows.Close()
	return detail, arows.Err()
}
