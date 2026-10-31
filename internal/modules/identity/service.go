package identity

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var usernamePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{2,31}$`)

type Fault struct {
	Status     int
	Code       string
	Message    string
	RetryAfter int
}

func (f *Fault) Error() string { return f.Code }

type VerificationMailer interface {
	SendVerification(context.Context, string, string) error
	SendPasswordReset(context.Context, string, string) error
}

type Service struct {
	DB         *pgxpool.Pool
	Signer     *TokenSigner
	Mailer     VerificationMailer
	CodePepper []byte
	DummyHash  string
}

type RegisterInput struct {
	Email          string `json:"email"`
	Username       string `json:"username"`
	DisplayName    string `json:"display_name"`
	Password       string `json:"password"`
	TermsVersion   string `json:"terms_version"`
	PrivacyVersion string `json:"privacy_version"`
}

type Tokens struct {
	Access  string `json:"access_token"`
	Refresh string `json:"-"`
	Expires int64  `json:"expires_in"`
}

func NewService(pool *pgxpool.Pool, signer *TokenSigner, mailer VerificationMailer, pepper []byte) (*Service, error) {
	if pool == nil || signer == nil || mailer == nil || len(pepper) < 32 {
		return nil, errors.New("identity dependencies and a 32-byte code pepper are required")
	}
	dummy, err := HashPassword("DummyPassword123!")
	if err != nil {
		return nil, err
	}
	return &Service{DB: pool, Signer: signer, Mailer: mailer, CodePepper: pepper, DummyHash: dummy}, nil
}

func normalizeEmail(email string) (string, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	address, err := mail.ParseAddress(email)
	if err != nil || address.Address != email || len(email) > 254 {
		return "", &Fault{Status: 400, Code: "INVALID_REQUEST_PAYLOAD", Message: "Invalid email address"}
	}
	return email, nil
}

func randomCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

func randomToken() (string, string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", "", err
	}
	token := base64.RawURLEncoding.EncodeToString(bytes)
	sum := sha256.Sum256([]byte(token))
	return token, hex.EncodeToString(sum[:]), nil
}

func (s *Service) codeHash(userID, code string) string {
	mac := hmac.New(sha256.New, s.CodePepper)
	_, _ = mac.Write([]byte(userID + ":" + code))
	return hex.EncodeToString(mac.Sum(nil))
}

func (s *Service) Register(ctx context.Context, input RegisterInput) (string, error) {
	email, err := normalizeEmail(input.Email)
	if err != nil {
		return "", err
	}
	username := strings.ToLower(strings.TrimSpace(input.Username))
	if !usernamePattern.MatchString(username) || len(input.DisplayName) < 1 || len(input.DisplayName) > 80 ||
		input.TermsVersion == "" || len(input.TermsVersion) > 24 ||
		input.PrivacyVersion == "" || len(input.PrivacyVersion) > 24 {
		return "", &Fault{Status: 400, Code: "INVALID_REQUEST_PAYLOAD", Message: "Invalid registration fields"}
	}
	passwordHash, err := HashPassword(input.Password)
	if err != nil {
		return "", &Fault{Status: 400, Code: "INVALID_REQUEST_PAYLOAD", Message: err.Error()}
	}
	code, err := randomCode()
	if err != nil {
		return "", err
	}
	userID := uuid.NewString()
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `INSERT INTO users(id,email,password_hash,status)
		VALUES($1,$2,$3,'PENDING_VERIFICATION')`, userID, email, passwordHash); err != nil {
		return "", registrationError(err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO user_profiles(user_id,username,display_name)
		VALUES($1,$2,$3)`, userID, username, strings.TrimSpace(input.DisplayName)); err != nil {
		return "", registrationError(err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO identities(user_id,provider,provider_subject,provider_email)
		VALUES($1,'LOCAL',$2,$2)`, userID, email); err != nil {
		return "", err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO user_consents(user_id,consent_type,policy_version,is_agreed)
		VALUES($1,'TERMS_OF_SERVICE',$2,TRUE),($1,'PRIVACY_POLICY',$3,TRUE)`,
		userID, input.TermsVersion, input.PrivacyVersion); err != nil {
		return "", err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO email_verification_challenges(user_id,code_hash,expires_at)
		VALUES($1,$2,NOW()+INTERVAL '10 minutes')`, userID, s.codeHash(userID, code)); err != nil {
		return "", err
	}
	if err := tx.Commit(ctx); err != nil {
		return "", err
	}
	if err := s.Mailer.SendVerification(ctx, email, code); err != nil {
		return userID, fmt.Errorf("verification email delivery failed: %w", err)
	}
	return userID, nil
}

func registrationError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return &Fault{Status: 409, Code: "DUPLICATE_EMAIL_OR_HANDLE", Message: "Email or username already exists"}
	}
	return err
}

func (s *Service) VerifyEmail(ctx context.Context, email, code string) error {
	email, err := normalizeEmail(email)
	if err != nil {
		return err
	}
	if len(code) != 6 {
		return &Fault{Status: 400, Code: "INVALID_REQUEST_PAYLOAD", Message: "Invalid verification code"}
	}
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var userID, storedHash string
	var expires time.Time
	var attempts int
	err = tx.QueryRow(ctx, `SELECT c.user_id,c.code_hash,c.expires_at,c.attempts
		FROM email_verification_challenges c JOIN users u ON u.id=c.user_id
		WHERE u.email=$1 FOR UPDATE OF c`, email).Scan(&userID, &storedHash, &expires, &attempts)
	if errors.Is(err, pgx.ErrNoRows) {
		return &Fault{Status: 400, Code: "INVALID_REQUEST_PAYLOAD", Message: "Invalid or expired verification code"}
	}
	if err != nil {
		return err
	}
	if attempts >= 10 || time.Now().After(expires) {
		return &Fault{Status: 400, Code: "INVALID_REQUEST_PAYLOAD", Message: "Invalid or expired verification code"}
	}
	if !hmac.Equal([]byte(storedHash), []byte(s.codeHash(userID, code))) {
		if _, err := tx.Exec(ctx, `UPDATE email_verification_challenges SET attempts=attempts+1 WHERE user_id=$1`, userID); err != nil {
			return err
		}
		if err := tx.Commit(ctx); err != nil {
			return err
		}
		return &Fault{Status: 400, Code: "INVALID_REQUEST_PAYLOAD", Message: "Invalid or expired verification code"}
	}
	if _, err := tx.Exec(ctx, `UPDATE users SET email_verified=TRUE,status='ACTIVE',updated_at=NOW()
		WHERE id=$1 AND status='PENDING_VERIFICATION'`, userID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM email_verification_challenges WHERE user_id=$1`, userID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *Service) Login(ctx context.Context, email, password string) (Tokens, error) {
	email, err := normalizeEmail(email)
	if err != nil {
		return Tokens{}, err
	}
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return Tokens{}, err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `INSERT INTO login_attempts(email) VALUES($1) ON CONFLICT DO NOTHING`, email); err != nil {
		return Tokens{}, err
	}
	var failed int
	var lockedUntil *time.Time
	if err := tx.QueryRow(ctx, `SELECT failed_count,locked_until FROM login_attempts
		WHERE email=$1 FOR UPDATE`, email).Scan(&failed, &lockedUntil); err != nil {
		return Tokens{}, err
	}
	if lockedUntil != nil && time.Now().Before(*lockedUntil) {
		return Tokens{}, &Fault{Status: 423, Code: "ACCOUNT_TEMPORARILY_LOCKED", Message: "Account temporarily locked"}
	}
	var userID, status string
	var passwordHash *string
	err = tx.QueryRow(ctx, `SELECT id,password_hash,status FROM users WHERE email=$1`, email).Scan(&userID, &passwordHash, &status)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return Tokens{}, err
	}
	valid := false
	if err == nil && passwordHash != nil {
		valid = VerifyPassword(password, *passwordHash)
	} else {
		_ = VerifyPassword(password, s.DummyHash)
	}
	if !valid {
		failed++
		_, err = tx.Exec(ctx, `UPDATE login_attempts SET failed_count=$2,
			locked_until=CASE WHEN $2>=10 THEN NOW()+INTERVAL '15 minutes' ELSE NULL END,
			updated_at=NOW() WHERE email=$1`, email, failed)
		if err != nil {
			return Tokens{}, err
		}
		if err := tx.Commit(ctx); err != nil {
			return Tokens{}, err
		}
		if failed >= 10 {
			return Tokens{}, &Fault{Status: 423, Code: "ACCOUNT_TEMPORARILY_LOCKED", Message: "Account temporarily locked"}
		}
		if failed >= 5 {
			return Tokens{}, &Fault{Status: 429, Code: "RATE_LIMIT_EXCEEDED", Message: "Too many login attempts", RetryAfter: 1 << min(failed-5, 5)}
		}
		return Tokens{}, &Fault{Status: 401, Code: "AUTH_TOKEN_MISSING", Message: "Invalid email or password"}
	}
	if status != "ACTIVE" {
		return Tokens{}, &Fault{Status: 403, Code: "INSUFFICIENT_BLOG_ROLE", Message: "Account is not active"}
	}
	refresh, hash, err := randomToken()
	if err != nil {
		return Tokens{}, err
	}
	sessionID, familyID := uuid.NewString(), uuid.NewString()
	if _, err := tx.Exec(ctx, `INSERT INTO sessions(id,user_id,refresh_token_hash,token_family_id,expires_at)
		VALUES($1,$2,$3,$4,NOW()+INTERVAL '14 days')`, sessionID, userID, hash, familyID); err != nil {
		return Tokens{}, err
	}
	if _, err := tx.Exec(ctx, `UPDATE login_attempts SET failed_count=0,locked_until=NULL,updated_at=NOW()
		WHERE email=$1`, email); err != nil {
		return Tokens{}, err
	}
	if _, err := tx.Exec(ctx, `UPDATE users SET last_login_at=NOW() WHERE id=$1`, userID); err != nil {
		return Tokens{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Tokens{}, err
	}
	access, err := s.Signer.Issue(userID, sessionID, time.Now())
	if err != nil {
		return Tokens{}, err
	}
	return Tokens{Access: access, Refresh: refresh, Expires: 900}, nil
}

func (s *Service) Refresh(ctx context.Context, token string) (Tokens, error) {
	if token == "" || len(token) > 256 {
		return Tokens{}, &Fault{Status: 401, Code: "AUTH_REFRESH_REVOKED", Message: "Refresh token invalid"}
	}
	hash := sha256.Sum256([]byte(token))
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return Tokens{}, err
	}
	defer tx.Rollback(ctx)
	var userID, familyID string
	var revoked bool
	var reason *string
	var expires, lastActive time.Time
	err = tx.QueryRow(ctx, `SELECT user_id,token_family_id,is_revoked,revoked_reason,expires_at,last_active_at
		FROM sessions WHERE refresh_token_hash=$1 FOR UPDATE`, hex.EncodeToString(hash[:])).Scan(
		&userID, &familyID, &revoked, &reason, &expires, &lastActive)
	if errors.Is(err, pgx.ErrNoRows) {
		return Tokens{}, &Fault{Status: 401, Code: "AUTH_REFRESH_REVOKED", Message: "Refresh token invalid"}
	}
	if err != nil {
		return Tokens{}, err
	}
	if revoked || time.Now().After(expires) {
		if revoked && reason != nil && *reason == "ROTATED" && time.Since(lastActive) > 5*time.Second {
			if _, err := tx.Exec(ctx, `UPDATE sessions SET is_revoked=TRUE,revoked_reason='TOKEN_REUSE'
				WHERE token_family_id=$1 AND is_revoked=FALSE`, familyID); err != nil {
				return Tokens{}, err
			}
			if err := tx.Commit(ctx); err != nil {
				return Tokens{}, err
			}
		}
		return Tokens{}, &Fault{Status: 401, Code: "AUTH_REFRESH_REVOKED", Message: "Refresh token revoked"}
	}
	newToken, newHash, err := randomToken()
	if err != nil {
		return Tokens{}, err
	}
	newSessionID := uuid.NewString()
	if _, err := tx.Exec(ctx, `UPDATE sessions SET is_revoked=TRUE,revoked_reason='ROTATED',
		last_active_at=NOW() WHERE refresh_token_hash=$1`, hex.EncodeToString(hash[:])); err != nil {
		return Tokens{}, err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO sessions(id,user_id,refresh_token_hash,token_family_id,expires_at)
		VALUES($1,$2,$3,$4,NOW()+INTERVAL '14 days')`, newSessionID, userID, newHash, familyID); err != nil {
		return Tokens{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Tokens{}, err
	}
	access, err := s.Signer.Issue(userID, newSessionID, time.Now())
	if err != nil {
		return Tokens{}, err
	}
	return Tokens{Access: access, Refresh: newToken, Expires: 900}, nil
}

func (s *Service) Logout(ctx context.Context, accessToken string) error {
	claims, err := s.Signer.Verify(accessToken, time.Now())
	if err != nil {
		return &Fault{Status: 401, Code: "AUTH_TOKEN_EXPIRED", Message: "Access token invalid"}
	}
	_, err = s.DB.Exec(ctx, `UPDATE sessions SET is_revoked=TRUE,revoked_reason='LOGOUT'
		WHERE id=$1 AND user_id=$2`, claims.SessionID, claims.Subject)
	return err
}

func (s *Service) ForgotPassword(ctx context.Context, rawEmail string) error {
	email, err := normalizeEmail(rawEmail)
	if err != nil {
		return err
	}
	var userID string
	err = s.DB.QueryRow(ctx, "SELECT id FROM users WHERE email=$1", email).Scan(&userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	rawToken, tokenHash, err := randomToken()
	if err != nil {
		return err
	}
	if _, err := s.DB.Exec(ctx, `INSERT INTO password_reset_challenges(user_id, token_hash, expires_at)
		VALUES($1, $2, NOW() + INTERVAL '1 hour')`, userID, tokenHash); err != nil {
		return err
	}
	if s.Mailer != nil {
		_ = s.Mailer.SendPasswordReset(ctx, email, rawToken)
	}
	return nil
}

func (s *Service) ResetPassword(ctx context.Context, token, newPassword string) error {
	token = strings.TrimSpace(token)
	if token == "" || len(token) > 256 {
		return &Fault{Status: 400, Code: "INVALID_REQUEST_PAYLOAD", Message: "Invalid reset token"}
	}
	newHash, err := HashPassword(newPassword)
	if err != nil {
		return &Fault{Status: 400, Code: "INVALID_REQUEST_PAYLOAD", Message: err.Error()}
	}
	tokenSum := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(tokenSum[:])

	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var challengeID, userID string
	var expiresAt time.Time
	var consumedAt *time.Time
	err = tx.QueryRow(ctx, `SELECT id, user_id, expires_at, consumed_at
		FROM password_reset_challenges WHERE token_hash=$1 FOR UPDATE`, tokenHash).
		Scan(&challengeID, &userID, &expiresAt, &consumedAt)
	if errors.Is(err, pgx.ErrNoRows) || consumedAt != nil {
		return &Fault{Status: 400, Code: "INVALID_REQUEST_PAYLOAD", Message: "Invalid or consumed reset token"}
	}
	if err != nil {
		return err
	}
	if time.Now().After(expiresAt) {
		return &Fault{Status: 410, Code: "TOKEN_EXPIRED", Message: "Password reset token has expired"}
	}

	if _, err := tx.Exec(ctx, `UPDATE users SET password_hash=$2, updated_at=NOW() WHERE id=$1`, userID, newHash); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `UPDATE password_reset_challenges SET consumed_at=NOW() WHERE id=$1`, challengeID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `UPDATE sessions SET is_revoked=TRUE, revoked_reason='PASSWORD_RESET'
		WHERE user_id=$1 AND is_revoked=FALSE`, userID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}



// SetupMFA generates a new TOTP secret and backup codes, storing them (inactive) on the user.
func (s *Service) SetupMFA(ctx context.Context, userID string) (string, []string, error) {
	if _, err := uuid.Parse(userID); err != nil {
		return "", nil, &Fault{Status: 400, Code: "INVALID_REQUEST_PAYLOAD", Message: "Invalid user id"}
	}

	var alreadyEnabled bool
	err := s.DB.QueryRow(ctx, `SELECT mfa_enabled FROM users WHERE id=$1`, userID).Scan(&alreadyEnabled)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil, &Fault{Status: 404, Code: "USER_NOT_FOUND", Message: "User not found"}
	}
	if err != nil {
		return "", nil, err
	}
	if alreadyEnabled {
		return "", nil, &Fault{Status: 409, Code: "INVALID_REQUEST_PAYLOAD", Message: "MFA is already enabled"}
	}

	secret, err := GenerateTOTPSecret()
	if err != nil {
		return "", nil, err
	}
	codes, hashes, err := GenerateBackupCodes(10)
	if err != nil {
		return "", nil, err
	}
	hashJSON, err := json.Marshal(hashes)
	if err != nil {
		return "", nil, err
	}

	if _, err := s.DB.Exec(ctx, `UPDATE users
		SET mfa_secret_enc=$2, mfa_backup_codes_hash=$3, updated_at=NOW()
		WHERE id=$1`, userID, []byte(secret), string(hashJSON)); err != nil {
		return "", nil, err
	}
	return secret, codes, nil
}

// VerifyMFASetup validates the first TOTP code and activates MFA for the user.
func (s *Service) VerifyMFASetup(ctx context.Context, userID, code string) error {
	code = strings.TrimSpace(code)
	if len(code) != 6 {
		return &Fault{Status: 400, Code: "INVALID_REQUEST_PAYLOAD", Message: "Invalid TOTP code"}
	}

	var secretBytes []byte
	var enabled bool
	err := s.DB.QueryRow(ctx, `SELECT mfa_secret_enc, mfa_enabled FROM users WHERE id=$1`, userID).
		Scan(&secretBytes, &enabled)
	if errors.Is(err, pgx.ErrNoRows) {
		return &Fault{Status: 404, Code: "USER_NOT_FOUND", Message: "User not found"}
	}
	if err != nil {
		return err
	}
	if enabled {
		return &Fault{Status: 409, Code: "INVALID_REQUEST_PAYLOAD", Message: "MFA is already enabled"}
	}
	if len(secretBytes) == 0 {
		return &Fault{Status: 400, Code: "INVALID_REQUEST_PAYLOAD", Message: "MFA setup has not been started"}
	}

	if !VerifyTOTP(string(secretBytes), code, time.Now()) {
		return &Fault{Status: 400, Code: "INVALID_REQUEST_PAYLOAD", Message: "Invalid TOTP code"}
	}

	_, err = s.DB.Exec(ctx, `UPDATE users SET mfa_enabled=TRUE, updated_at=NOW() WHERE id=$1`, userID)
	return err
}

// DisableMFA verifies a TOTP or backup code and disables MFA.
func (s *Service) DisableMFA(ctx context.Context, userID, code string) error {
	code = strings.TrimSpace(code)
	if code == "" {
		return &Fault{Status: 400, Code: "INVALID_REQUEST_PAYLOAD", Message: "Verification code is required"}
	}

	var secretBytes []byte
	var enabled bool
	var backupJSON []byte
	err := s.DB.QueryRow(ctx, `SELECT mfa_secret_enc, mfa_enabled, mfa_backup_codes_hash FROM users WHERE id=$1`, userID).
		Scan(&secretBytes, &enabled, &backupJSON)
	if errors.Is(err, pgx.ErrNoRows) {
		return &Fault{Status: 404, Code: "USER_NOT_FOUND", Message: "User not found"}
	}
	if err != nil {
		return err
	}
	if !enabled {
		return &Fault{Status: 400, Code: "INVALID_REQUEST_PAYLOAD", Message: "MFA is not enabled"}
	}

	ok := false
	if len(code) == 6 && len(secretBytes) > 0 {
		ok = VerifyTOTP(string(secretBytes), code, time.Now())
	}
	if !ok && len(code) == 8 {
		idx, err := VerifyBackupCode(code, backupJSON)
		if err == nil && idx >= 0 {
			ok = true
		}
	}
	if !ok {
		return &Fault{Status: 401, Code: "MFA_REQUIRED", Message: "Invalid MFA code"}
	}

	_, err = s.DB.Exec(ctx, `UPDATE users
		SET mfa_enabled=FALSE, mfa_secret_enc=NULL, mfa_backup_codes_hash='[]'::jsonb, updated_at=NOW()
		WHERE id=$1`, userID)
	return err
}

// MarkSessionMFAVerified marks the given session as having passed MFA.
func (s *Service) MarkSessionMFAVerified(ctx context.Context, sessionID string) error {
	_, err := s.DB.Exec(ctx, `UPDATE sessions SET mfa_verified=TRUE WHERE id=$1`, sessionID)
	return err
}
