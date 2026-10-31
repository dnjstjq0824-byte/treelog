package user

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Fault struct {
	Status  int
	Code    string
	Message string
}

func (f *Fault) Error() string { return f.Code }

type UserProfile struct {
	ID             string    `json:"id"`
	Email          string    `json:"email"`
	Username       string    `json:"username"`
	DisplayName    string    `json:"display_name"`
	Bio            string    `json:"bio"`
	AvatarURL      *string   `json:"avatar_url"`
	PreferredLocale string   `json:"preferred_locale"`
	Timezone       string    `json:"timezone"`
	MFAEnabled     bool      `json:"mfa_enabled"`
	Roles          []string  `json:"roles"`
	CreatedAt      time.Time `json:"created_at"`
}

type SessionInfo struct {
	ID           string    `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	LastActiveAt time.Time `json:"last_active_at"`
	ExpiresAt    time.Time `json:"expires_at"`
	IsCurrent    bool      `json:"is_current"`
}

type Service struct {
	db *pgxpool.Pool
}

func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

func (s *Service) GetMe(ctx context.Context, userID, currentSessionID string) (*UserProfile, error) {
	var profile UserProfile
	profile.ID = userID

	err := s.db.QueryRow(ctx, `SELECT u.email, u.preferred_locale, u.timezone, u.mfa_enabled, u.created_at,
		p.username, p.display_name, p.bio
		FROM users u
		JOIN user_profiles p ON p.user_id = u.id
		WHERE u.id = $1`, userID).Scan(
		&profile.Email, &profile.PreferredLocale, &profile.Timezone, &profile.MFAEnabled, &profile.CreatedAt,
		&profile.Username, &profile.DisplayName, &profile.Bio,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, &Fault{Status: 404, Code: "USER_NOT_FOUND", Message: "User not found"}
	}
	if err != nil {
		return nil, err
	}

	// Roles
	rows, err := s.db.Query(ctx, `SELECT r.code FROM user_roles ur JOIN roles r ON r.id = ur.role_id WHERE ur.user_id = $1`, userID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var role string
			if err := rows.Scan(&role); err == nil {
				profile.Roles = append(profile.Roles, role)
			}
		}
	}

	return &profile, nil
}

func (s *Service) UpdateProfile(ctx context.Context, userID string, displayName, bio string) error {
	displayName = strings.TrimSpace(displayName)
	if len(displayName) < 1 || len(displayName) > 80 {
		return &Fault{Status: 400, Code: "INVALID_REQUEST_PAYLOAD", Message: "Display name must be between 1 and 80 characters"}
	}
	if len(bio) > 500 {
		return &Fault{Status: 400, Code: "INVALID_REQUEST_PAYLOAD", Message: "Bio must not exceed 500 characters"}
	}

	_, err := s.db.Exec(ctx, `UPDATE user_profiles
		SET display_name=$2, bio=$3, updated_at=NOW()
		WHERE user_id=$1`, userID, displayName, bio)
	return err
}

func (s *Service) UpdatePreferences(ctx context.Context, userID, locale, timezone string) error {
	locale = strings.TrimSpace(locale)
	timezone = strings.TrimSpace(timezone)
	if locale == "" || len(locale) > 16 || timezone == "" || len(timezone) > 64 {
		return &Fault{Status: 400, Code: "INVALID_REQUEST_PAYLOAD", Message: "Invalid locale or timezone"}
	}

	_, err := s.db.Exec(ctx, `UPDATE users
		SET preferred_locale=$2, timezone=$3, updated_at=NOW()
		WHERE id=$1`, userID, locale, timezone)
	return err
}

func (s *Service) GetSessions(ctx context.Context, userID, currentSessionID string) ([]SessionInfo, error) {
	rows, err := s.db.Query(ctx, `SELECT id, created_at, last_active_at, expires_at
		FROM sessions
		WHERE user_id=$1 AND is_revoked=FALSE AND expires_at > NOW()
		ORDER BY last_active_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []SessionInfo
	for rows.Next() {
		var item SessionInfo
		if err := rows.Scan(&item.ID, &item.CreatedAt, &item.LastActiveAt, &item.ExpiresAt); err != nil {
			return nil, err
		}
		item.IsCurrent = (item.ID == currentSessionID)
		sessions = append(sessions, item)
	}
	return sessions, nil
}

func (s *Service) RevokeSession(ctx context.Context, userID, targetSessionID string) error {
	result, err := s.db.Exec(ctx, `UPDATE sessions
		SET is_revoked=TRUE, revoked_reason='USER_REVOKED'
		WHERE id=$1 AND user_id=$2 AND is_revoked=FALSE`, targetSessionID, userID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return &Fault{Status: 404, Code: "RESOURCE_NOT_FOUND", Message: "Session not found or already revoked"}
	}
	return nil
}

// DeactivateAccount verifies the user does not solely own any active blogs, then marks account deactivated.
func (s *Service) DeactivateAccount(ctx context.Context, userID string) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Check if user is sole owner of any blog
	var ownedBlogs int
	err = tx.QueryRow(ctx, `SELECT COUNT(*) FROM blog_members bm
		JOIN blogs b ON b.id = bm.blog_id
		WHERE bm.user_id = $1 AND bm.role = 'OWNER' AND b.status != 'CLOSED'`, userID).Scan(&ownedBlogs)
	if err != nil {
		return err
	}
	if ownedBlogs > 0 {
		return &Fault{
			Status:  409,
			Code:    "OWNED_BLOGS_EXIST",
			Message: "You must transfer or close your owned blogs before deactivating your account",
		}
	}

	// Deactivate user and revoke all sessions
	if _, err := tx.Exec(ctx, `UPDATE users
		SET status='DEACTIVATED', deactivation_requested_at=NOW(), updated_at=NOW()
		WHERE id=$1`, userID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `UPDATE sessions
		SET is_revoked=TRUE, revoked_reason='ACCOUNT_DEACTIVATED'
		WHERE user_id=$1 AND is_revoked=FALSE`, userID); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
