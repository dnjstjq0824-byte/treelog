package moderation

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

// Appeal is one appeal row.
type Appeal struct {
	ID            string     `json:"id"`
	CaseID        string     `json:"case_id"`
	AppellantID   string     `json:"appellant_user_id"`
	Statement     string     `json:"appeal_statement"`
	Status        string     `json:"status"`
	DecisionNote  *string    `json:"reviewer_decision_note,omitempty"`
	ResolvedAt    *time.Time `json:"resolved_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

// FileAppeal records an appeal from the affected user.
func (s *Service) FileAppeal(ctx context.Context, userID, caseID, statement string) (*Appeal, error) {
	if strings.TrimSpace(statement) == "" || len(statement) > 5000 {
		return nil, &Fault{Status: http.StatusBadRequest, Code: "INVALID_APPEAL", Message: "Appeal statement is required (max 5000 chars)"}
	}
	if err := s.requirePool(); err != nil {
		return nil, err
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	var targetType, targetID, caseStatus string
	err = tx.QueryRow(ctx, `SELECT target_type, target_id, status FROM moderation_cases
		WHERE id = $1 FOR UPDATE`, caseID).Scan(&targetType, &targetID, &caseStatus)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, &Fault{Status: http.StatusNotFound, Code: "CASE_NOT_FOUND", Message: "Moderation case not found"}
	}
	if err != nil {
		return nil, err
	}
	ownerID, err := targetOwner(ctx, tx, targetType, targetID)
	if err != nil {
		return nil, err
	}
	if ownerID != userID {
		return nil, &Fault{Status: http.StatusForbidden, Code: "NOT_AFFECTED_USER", Message: "Only the affected user can appeal this case"}
	}
	var appeal Appeal
	err = tx.QueryRow(ctx, `
		INSERT INTO appeals (case_id, appellant_user_id, appeal_statement)
		VALUES ($1, $2, $3)
		RETURNING id, case_id, appellant_user_id, appeal_statement, status, reviewer_decision_note, resolved_at, created_at
	`, caseID, userID, strings.TrimSpace(statement)).Scan(
		&appeal.ID, &appeal.CaseID, &appeal.AppellantID, &appeal.Statement,
		&appeal.Status, &appeal.DecisionNote, &appeal.ResolvedAt, &appeal.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	if _, err := tx.Exec(ctx, `UPDATE moderation_cases SET status = 'APPEALED', updated_at = NOW()
		WHERE id = $1`, caseID); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &appeal, nil
}

// ListAppeals returns appeals with offset paging.
func (s *Service) ListAppeals(ctx context.Context, status string, limit, offset int) ([]Appeal, error) {
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
		SELECT id, case_id, appellant_user_id, appeal_statement, status, reviewer_decision_note, resolved_at, created_at
		FROM appeals
	`
	args := []any{}
	if status != "" {
		query += ` WHERE status = $1`
		args = append(args, strings.ToUpper(status))
		query += ` ORDER BY created_at DESC LIMIT $2 OFFSET $3`
		args = append(args, limit, offset)
	} else {
		query += ` ORDER BY created_at DESC LIMIT $1 OFFSET $2`
		args = append(args, limit, offset)
	}
	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]Appeal, 0)
	for rows.Next() {
		var appeal Appeal
		if err := rows.Scan(&appeal.ID, &appeal.CaseID, &appeal.AppellantID, &appeal.Statement,
			&appeal.Status, &appeal.DecisionNote, &appeal.ResolvedAt, &appeal.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, appeal)
	}
	return items, rows.Err()
}

// DecideAppeal resolves an appeal, restoring content on approval.
func (s *Service) DecideAppeal(ctx context.Context, moderatorID, appealID, decision, note string) (*Appeal, error) {
	decision = strings.ToUpper(strings.TrimSpace(decision))
	if decision != "APPROVED_RESTORED" && decision != "REJECTED_UPHELD" {
		return nil, &Fault{Status: http.StatusBadRequest, Code: "INVALID_DECISION", Message: "Decision must be APPROVED_RESTORED or REJECTED_UPHELD"}
	}
	if len(note) > 1000 {
		return nil, &Fault{Status: http.StatusBadRequest, Code: "INVALID_DECISION_NOTE", Message: "Decision note exceeds 1000 characters"}
	}
	if err := s.requirePool(); err != nil {
		return nil, err
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	var appeal Appeal
	var caseID, targetType, targetID string
	err = tx.QueryRow(ctx, `
		SELECT a.id, a.case_id, a.appellant_user_id, a.appeal_statement, a.status,
		       c.target_type, c.target_id
		FROM appeals a JOIN moderation_cases c ON c.id = a.case_id
		WHERE a.id = $1 FOR UPDATE OF a`, appealID).Scan(
		&appeal.ID, &appeal.CaseID, &appeal.AppellantID, &appeal.Statement,
		&appeal.Status, &targetType, &targetID,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, &Fault{Status: http.StatusNotFound, Code: "APPEAL_NOT_FOUND", Message: "Appeal not found"}
	}
	if err != nil {
		return nil, err
	}
	if appeal.Status != "PENDING" && appeal.Status != "UNDER_REVIEW" {
		return nil, &Fault{Status: http.StatusConflict, Code: "APPEAL_ALREADY_DECIDED", Message: "Appeal is already decided"}
	}
	if decision == "APPROVED_RESTORED" {
		if err := restoreContent(ctx, tx, targetType, targetID); err != nil {
			return nil, err
		}
		if targetType == "USER" {
			_, _ = tx.Exec(ctx, `UPDATE users SET status = 'ACTIVE', updated_at = NOW()
				WHERE id = $1 AND status = 'SUSPENDED'`, targetID)
		}
		if targetType == "BLOG" {
			_, _ = tx.Exec(ctx, `UPDATE blogs SET status = 'ACTIVE', updated_at = NOW()
				WHERE id = $1 AND status = 'SUSPENDED'`, targetID)
		}
		_, _ = tx.Exec(ctx, `
			INSERT INTO moderation_actions (case_id, actor_moderator_id, action_type, policy_clause_code, internal_note, public_notice_reason)
			VALUES ($1, $2, 'RESTORE_CONTENT', 'APPEAL', $3, 'Appeal approved; content restored')`,
			caseID, moderatorID, note)
	}
	if _, err := tx.Exec(ctx, `UPDATE appeals SET status = $2, reviewer_moderator_id = $3,
		reviewer_decision_note = NULLIF($4, ''), resolved_at = NOW() WHERE id = $1`,
		appealID, decision, moderatorID, note); err != nil {
		return nil, err
	}
	if _, err := tx.Exec(ctx, `UPDATE moderation_cases SET status = 'APPEAL_RESOLVED', updated_at = NOW()
		WHERE id = $1`, caseID); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	appeal.Status = decision
	if note != "" {
		appeal.DecisionNote = &note
	}
	now := time.Now()
	appeal.ResolvedAt = &now
	return &appeal, nil
}
