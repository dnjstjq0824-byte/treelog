package moderation

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"
)

// ActionInput carries the enforcement action payload.
type ActionInput struct {
	ActionType         string `json:"action_type"`
	PolicyClauseCode   string `json:"policy_clause_code"`
	InternalNote       string `json:"internal_note"`
	PublicNoticeReason string `json:"public_notice_reason"`
}

// ApplyAction records an enforcement action and executes its side effects
// in one transaction.
func (s *Service) ApplyAction(ctx context.Context, moderatorID, caseID string, input ActionInput) (*ActionRow, error) {
	actionType := strings.ToUpper(strings.TrimSpace(input.ActionType))
	if !validActionTypes[actionType] {
		return nil, &Fault{Status: http.StatusBadRequest, Code: "INVALID_ACTION_TYPE", Message: "Unknown action type"}
	}
	if strings.TrimSpace(input.PolicyClauseCode) == "" || len(input.PolicyClauseCode) > 40 {
		return nil, &Fault{Status: http.StatusBadRequest, Code: "INVALID_POLICY_CLAUSE", Message: "Policy clause code is required (max 40 chars)"}
	}
	if strings.TrimSpace(input.PublicNoticeReason) == "" || len(input.PublicNoticeReason) > 500 {
		return nil, &Fault{Status: http.StatusBadRequest, Code: "INVALID_PUBLIC_NOTICE", Message: "Public notice reason is required (max 500 chars)"}
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

	var row ActionRow
	err = tx.QueryRow(ctx, `
		INSERT INTO moderation_actions (case_id, actor_moderator_id, action_type, policy_clause_code, internal_note, public_notice_reason)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, actor_moderator_id, action_type, policy_clause_code, public_notice_reason, created_at
	`, caseID, moderatorID, actionType, input.PolicyClauseCode, input.InternalNote, input.PublicNoticeReason).Scan(
		&row.ID, &row.ActorModeratorID, &row.ActionType, &row.PolicyClause, &row.PublicNotice, &row.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	nextStatus := "ACTION_TAKEN"
	switch actionType {
	case "DISMISS":
		nextStatus = "DISMISSED"
	case "WARN_USER":
		if err := notifyTargetUser(ctx, tx, targetType, targetID, caseID, input.PublicNoticeReason); err != nil {
			return nil, err
		}
	case "QUARANTINE_CONTENT":
		if err := quarantineContent(ctx, tx, targetType, targetID); err != nil {
			return nil, err
		}
	case "RESTORE_CONTENT":
		if err := restoreContent(ctx, tx, targetType, targetID); err != nil {
			return nil, err
		}
	case "DELETE_CONTENT":
		if err := deleteContent(ctx, tx, targetType, targetID); err != nil {
			return nil, err
		}
	case "SUSPEND_BLOG_7D", "SUSPEND_BLOG_PERM":
		if err := suspendBlog(ctx, tx, targetType, targetID); err != nil {
			return nil, err
		}
	case "SUSPEND_USER":
		if err := suspendUser(ctx, tx, targetType, targetID); err != nil {
			return nil, err
		}
	}
	if _, err := tx.Exec(ctx, `UPDATE moderation_cases
		SET status = $2, updated_at = NOW() WHERE id = $1`, caseID, nextStatus); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &row, nil
}

func notifyTargetUser(ctx context.Context, tx pgx.Tx, targetType, targetID, caseID, notice string) error {
	userID, err := targetOwner(ctx, tx, targetType, targetID)
	if err != nil || userID == "" {
		return err
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO notifications (recipient_user_id, category, group_key, dedup_key, title, body, action_url)
		VALUES ($1, 'MODERATION', 'moderation:' || $2, 'mod-warn:' || $2, '경고: 커뮤니티 가이드 위반', $3, '/settings/safety')
		ON CONFLICT DO NOTHING`, userID, caseID, notice)
	return err
}

func targetOwner(ctx context.Context, tx pgx.Tx, targetType, targetID string) (string, error) {
	var query string
	switch targetType {
	case "USER":
		return targetID, nil
	case "POST":
		query = `SELECT author_user_id FROM posts WHERE id = $1`
	case "COMMENT":
		query = `SELECT author_user_id FROM comments WHERE id = $1`
	case "GUESTBOOK":
		query = `SELECT author_user_id FROM guestbook_entries WHERE id = $1`
	case "BLOG":
		query = `SELECT owner_user_id FROM blogs WHERE id = $1`
	default:
		return "", nil
	}
	var userID string
	if err := tx.QueryRow(ctx, query, targetID).Scan(&userID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", &Fault{Status: http.StatusNotFound, Code: "REPORT_TARGET_NOT_FOUND", Message: "Action target does not exist"}
		}
		return "", err
	}
	return userID, nil
}

func quarantineContent(ctx context.Context, tx pgx.Tx, targetType, targetID string) error {
	switch targetType {
	case "POST":
		_, err := tx.Exec(ctx, `UPDATE posts SET status = 'QUARANTINED', updated_at = NOW() WHERE id = $1`, targetID)
		return err
	case "COMMENT":
		_, err := tx.Exec(ctx, `UPDATE comments SET status = 'QUARANTINED_SPAM', updated_at = NOW() WHERE id = $1`, targetID)
		return err
	default:
		return &Fault{Status: http.StatusBadRequest, Code: "ACTION_TARGET_MISMATCH", Message: "Quarantine applies to posts and comments"}
	}
}

func restoreContent(ctx context.Context, tx pgx.Tx, targetType, targetID string) error {
	switch targetType {
	case "POST":
		_, err := tx.Exec(ctx, `UPDATE posts SET status = 'PUBLISHED', updated_at = NOW()
			WHERE id = $1 AND status = 'QUARANTINED'`, targetID)
		return err
	case "COMMENT":
		_, err := tx.Exec(ctx, `UPDATE comments SET status = 'PUBLISHED', updated_at = NOW()
			WHERE id = $1 AND status IN ('QUARANTINED_SPAM', 'HIDDEN_BY_OWNER')`, targetID)
		return err
	default:
		return &Fault{Status: http.StatusBadRequest, Code: "ACTION_TARGET_MISMATCH", Message: "Restore applies to posts and comments"}
	}
}

func deleteContent(ctx context.Context, tx pgx.Tx, targetType, targetID string) error {
	switch targetType {
	case "POST":
		_, err := tx.Exec(ctx, `UPDATE posts SET status = 'TRASHED', trashed_at = NOW(),
			purge_after_at = NOW() + INTERVAL '30 days', updated_at = NOW() WHERE id = $1`, targetID)
		return err
	case "COMMENT":
		_, err := tx.Exec(ctx, `UPDATE comments SET status = 'HIDDEN_BY_OWNER', updated_at = NOW() WHERE id = $1`, targetID)
		return err
	case "GUESTBOOK":
		_, err := tx.Exec(ctx, `DELETE FROM guestbook_entries WHERE id = $1`, targetID)
		return err
	default:
		return &Fault{Status: http.StatusBadRequest, Code: "ACTION_TARGET_MISMATCH", Message: "Delete applies to posts, comments, and guestbook entries"}
	}
}

func suspendBlog(ctx context.Context, tx pgx.Tx, targetType, targetID string) error {
	blogID := targetID
	if targetType == "POST" {
		if err := tx.QueryRow(ctx, `SELECT blog_id FROM posts WHERE id = $1`, targetID).Scan(&blogID); err != nil {
			return err
		}
	} else if targetType != "BLOG" {
		return &Fault{Status: http.StatusBadRequest, Code: "ACTION_TARGET_MISMATCH", Message: "Blog suspension applies to blogs and posts"}
	}
	_, err := tx.Exec(ctx, `UPDATE blogs SET status = 'SUSPENDED', updated_at = NOW() WHERE id = $1`, blogID)
	return err
}

func suspendUser(ctx context.Context, tx pgx.Tx, targetType, targetID string) error {
	userID, err := targetOwner(ctx, tx, targetType, targetID)
	if err != nil || userID == "" {
		if err == nil {
			return &Fault{Status: http.StatusBadRequest, Code: "ACTION_TARGET_MISMATCH", Message: "User suspension needs a resolvable user"}
		}
		return err
	}
	if _, err := tx.Exec(ctx, `UPDATE users SET status = 'SUSPENDED', updated_at = NOW() WHERE id = $1`, userID); err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `UPDATE sessions SET is_revoked = TRUE, revoked_reason = 'USER_SUSPENDED'
		WHERE user_id = $1 AND is_revoked = FALSE`, userID)
	return err
}

// ResumeExpiredSuspensions re-activates blogs whose latest enforcement was a
// 7-day suspension older than 7 days. It returns the resumed blog count.
func (s *Service) ResumeExpiredSuspensions(ctx context.Context) (int, error) {
	if err := s.requirePool(); err != nil {
		return 0, err
	}
	rows, err := s.pool.Query(ctx, `
		SELECT b.id
		FROM blogs b
		WHERE b.status = 'SUSPENDED'
		  AND EXISTS (
			SELECT 1 FROM moderation_actions a
			JOIN moderation_cases c ON c.id = a.case_id
			WHERE a.action_type = 'SUSPEND_BLOG_7D'
			  AND a.created_at <= NOW() - INTERVAL '7 days'
			  AND ((c.target_type = 'BLOG' AND c.target_id = b.id)
			    OR (c.target_type = 'POST' AND c.target_id IN (SELECT id FROM posts WHERE blog_id = b.id)))
			  AND NOT EXISTS (
				SELECT 1 FROM moderation_actions newer
				WHERE newer.case_id = a.case_id AND newer.created_at > a.created_at
			  )
		  )
		FOR UPDATE OF b SKIP LOCKED
	`)
	if err != nil {
		return 0, err
	}
	ids := make([]string, 0)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return 0, err
		}
		ids = append(ids, id)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return 0, err
	}
	for _, id := range ids {
		if _, err := s.pool.Exec(ctx, `UPDATE blogs SET status = 'ACTIVE', updated_at = NOW()
			WHERE id = $1 AND status = 'SUSPENDED'`, id); err != nil {
			return 0, err
		}
	}
	return len(ids), nil
}
