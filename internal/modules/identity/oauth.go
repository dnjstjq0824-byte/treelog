package identity

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"treeblog/internal/pkg/safenet"
)

// OAuthProvider identifies a supported social login provider. Codes match
// the identities.provider CHECK constraint.
type OAuthProvider string

const (
	OAuthGoogle      OAuthProvider = "GOOGLE"
	OAuthApple       OAuthProvider = "APPLE"
	OAuthGitHub      OAuthProvider = "GITHUB"
	OAuthNaver       OAuthProvider = "NAVER"
	OAuthKakao       OAuthProvider = "KAKAO"
	OAuthTreescanOIDC OAuthProvider = "TREESCAN_OIDC"
)

// OAuthClientConfig holds one provider's OAuth client credentials.
type OAuthClientConfig struct {
	ClientID     string
	ClientSecret string
	// RedirectBase is the public API base URL; the callback path is appended.
	RedirectBase string
	// AppBaseURL is the user-web base URL used for post-login redirects.
	AppBaseURL string
	// Apple-specific fields for ES256 client_secret JWTs.
	AppleTeamID string
	AppleKeyID  string
	AppleKeyPEM string
	// TreescanOIDC issuer for generic OIDC discovery.
	TreescanIssuer string
}

type oauthEndpoints struct {
	authorize string
	token     string
	userinfo  string
	scopes    string
}

func staticEndpoints(provider OAuthProvider) (oauthEndpoints, bool) {
	switch provider {
	case OAuthGoogle:
		return oauthEndpoints{
			authorize: "https://accounts.google.com/o/oauth2/v2/auth",
			token:     "https://oauth2.googleapis.com/token",
			userinfo:  "https://openidconnect.googleapis.com/v1/userinfo",
			scopes:    "openid email profile",
		}, true
	case OAuthGitHub:
		return oauthEndpoints{
			authorize: "https://github.com/login/oauth/authorize",
			token:     "https://github.com/login/oauth/access_token",
			userinfo:  "https://api.github.com/user",
			scopes:    "read:user user:email",
		}, true
	case OAuthNaver:
		return oauthEndpoints{
			authorize: "https://nid.naver.com/oauth2.0/authorize",
			token:     "https://nid.naver.com/oauth2.0/token",
			userinfo:  "https://openapi.naver.com/v1/nid/me",
			scopes:    "",
		}, true
	case OAuthKakao:
		return oauthEndpoints{
			authorize: "https://kauth.kakao.com/oauth/authorize",
			token:     "https://kauth.kakao.com/oauth/token",
			userinfo:  "https://kapi.kakao.com/v2/user/me",
			scopes:    "profile_nickname account_email",
		}, true
	case OAuthApple:
		return oauthEndpoints{
			authorize: "https://appleid.apple.com/auth/authorize",
			token:     "https://appleid.apple.com/auth/token",
			userinfo:  "",
			scopes:    "name email",
		}, true
	default:
		return oauthEndpoints{}, false
	}
}

// oauthState is the signed, self-contained login state carried through the
// provider round-trip (double-submitted via query string and cookie).
type oauthState struct {
	Nonce     string `json:"nonce"`
	Provider  string `json:"provider"`
	ExpiresAt int64  `json:"exp"`
	Verifier  string `json:"verifier"`
	Terms     string `json:"terms"`
	Privacy   string `json:"privacy"`
}

func (s *Service) signOAuthState(state oauthState) (string, error) {
	raw, err := json.Marshal(state)
	if err != nil {
		return "", err
	}
	payload := base64.RawURLEncoding.EncodeToString(raw)
	mac := hmac.New(sha256.New, s.CodePepper)
	_, _ = mac.Write([]byte("treeblog/oauth-state/v1\x00" + payload))
	return payload + "." + hex.EncodeToString(mac.Sum(nil)), nil
}

func (s *Service) verifyOAuthState(signed, provider string) (oauthState, error) {
	var state oauthState
	parts := strings.SplitN(signed, ".", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return state, &Fault{Status: 400, Code: "INVALID_OAUTH_STATE", Message: "Invalid OAuth state"}
	}
	mac := hmac.New(sha256.New, s.CodePepper)
	_, _ = mac.Write([]byte("treeblog/oauth-state/v1\x00" + parts[0]))
	expected := mac.Sum(nil)
	actual, err := hex.DecodeString(parts[1])
	if err != nil || !hmac.Equal(expected, actual) {
		return state, &Fault{Status: 400, Code: "INVALID_OAUTH_STATE", Message: "Invalid OAuth state"}
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return state, &Fault{Status: 400, Code: "INVALID_OAUTH_STATE", Message: "Invalid OAuth state"}
	}
	if err := json.Unmarshal(raw, &state); err != nil {
		return state, &Fault{Status: 400, Code: "INVALID_OAUTH_STATE", Message: "Invalid OAuth state"}
	}
	if state.Provider != provider || state.Nonce == "" || state.Verifier == "" {
		return state, &Fault{Status: 400, Code: "INVALID_OAUTH_STATE", Message: "Invalid OAuth state"}
	}
	if time.Now().Unix() > state.ExpiresAt {
		return state, &Fault{Status: 400, Code: "OAUTH_STATE_EXPIRED", Message: "OAuth login session expired; please try again"}
	}
	return state, nil
}

func newPKCEVerifier() (verifier, challenge string, err error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", "", err
	}
	verifier = base64.RawURLEncoding.EncodeToString(raw)
	sum := sha256.Sum256([]byte(verifier))
	challenge = base64.RawURLEncoding.EncodeToString(sum[:])
	return verifier, challenge, nil
}

func newOAuthNonce() (string, error) {
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

// OAuthConfig returns the client config for a provider, or false when the
// provider is unknown or not configured.
func (s *Service) OAuthConfig(provider OAuthProvider) (OAuthClientConfig, bool) {
	if s.OAuth == nil {
		return OAuthClientConfig{}, false
	}
	cfg, ok := s.OAuth[provider]
	if !ok || cfg.ClientID == "" {
		return OAuthClientConfig{}, false
	}
	return cfg, true
}

func oauthRedirectURI(cfg OAuthClientConfig, provider OAuthProvider) string {
	base := strings.TrimRight(cfg.RedirectBase, "/")
	if base == "" {
		base = "http://localhost:8080"
	}
	return base + "/api/v1/auth/oauth/" + strings.ToLower(string(provider)) + "/callback"
}

// OAuthLoginURL builds the provider authorization URL and the signed state
// to double-submit. termsVersion/privacyVersion are recorded as consents
// when the OAuth login creates a new account.
func (s *Service) OAuthLoginURL(provider OAuthProvider, termsVersion, privacyVersion string) (authURL, signedState string, err error) {
	cfg, ok := s.OAuthConfig(provider)
	if !ok {
		return "", "", &Fault{Status: 501, Code: "OAUTH_NOT_CONFIGURED", Message: "Social login is not configured for this provider"}
	}
	endpoints, ok := staticEndpoints(provider)
	if !ok {
		if provider == OAuthTreescanOIDC {
			endpoints, err = s.discoverOIDCEndpoints(cfg)
			if err != nil {
				return "", "", err
			}
		} else {
			return "", "", &Fault{Status: 400, Code: "UNKNOWN_OAUTH_PROVIDER", Message: "Unknown OAuth provider"}
		}
	}
	verifier, challenge, err := newPKCEVerifier()
	if err != nil {
		return "", "", err
	}
	nonce, err := newOAuthNonce()
	if err != nil {
		return "", "", err
	}
	if termsVersion == "" {
		termsVersion = "v1.0"
	}
	if privacyVersion == "" {
		privacyVersion = "v1.0"
	}
	signedState, err = s.signOAuthState(oauthState{
		Nonce:     nonce,
		Provider:  string(provider),
		ExpiresAt: time.Now().Add(10 * time.Minute).Unix(),
		Verifier:  verifier,
		Terms:     termsVersion,
		Privacy:   privacyVersion,
	})
	if err != nil {
		return "", "", err
	}
	params := url.Values{}
	params.Set("client_id", cfg.ClientID)
	params.Set("redirect_uri", oauthRedirectURI(cfg, provider))
	params.Set("response_type", "code")
	if endpoints.scopes != "" {
		params.Set("scope", endpoints.scopes)
	}
	params.Set("state", signedState)
	params.Set("code_challenge", challenge)
	params.Set("code_challenge_method", "S256")
	params.Set("nonce", nonce)
	if provider == OAuthApple {
		params.Set("response_mode", "query")
	}
	return endpoints.authorize + "?" + params.Encode(), signedState, nil
}

func (s *Service) discoverOIDCEndpoints(cfg OAuthClientConfig) (oauthEndpoints, error) {
	if cfg.TreescanIssuer == "" {
		return oauthEndpoints{}, &Fault{Status: 501, Code: "OAUTH_NOT_CONFIGURED", Message: "TreeScan OIDC issuer is not configured"}
	}
	discovery := strings.TrimRight(cfg.TreescanIssuer, "/") + "/.well-known/openid-configuration"
	client := s.httpClient()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, discovery, nil)
	if err != nil {
		return oauthEndpoints{}, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return oauthEndpoints{}, &Fault{Status: 502, Code: "OAUTH_DISCOVERY_FAILED", Message: "OIDC discovery failed"}
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return oauthEndpoints{}, &Fault{Status: 502, Code: "OAUTH_DISCOVERY_FAILED", Message: "OIDC discovery failed"}
	}
	var doc struct {
		AuthorizationEndpoint string `json:"authorization_endpoint"`
		TokenEndpoint         string `json:"token_endpoint"`
		UserinfoEndpoint      string `json:"userinfo_endpoint"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 64*1024)).Decode(&doc); err != nil {
		return oauthEndpoints{}, &Fault{Status: 502, Code: "OAUTH_DISCOVERY_FAILED", Message: "Invalid OIDC discovery document"}
	}
	if doc.AuthorizationEndpoint == "" || doc.TokenEndpoint == "" || doc.UserinfoEndpoint == "" {
		return oauthEndpoints{}, &Fault{Status: 502, Code: "OAUTH_DISCOVERY_FAILED", Message: "Incomplete OIDC discovery document"}
	}
	return oauthEndpoints{
		authorize: doc.AuthorizationEndpoint,
		token:     doc.TokenEndpoint,
		userinfo:  doc.UserinfoEndpoint,
		scopes:    "openid email profile",
	}, nil
}

func (s *Service) httpClient() *http.Client {
	if s.HTTPClient != nil {
		return s.HTTPClient
	}
	return safenet.NewSafeHTTPClient(10 * time.Second)
}

// oauthProfile is the normalized provider identity.
type oauthProfile struct {
	Subject       string
	Email         string
	EmailVerified bool
	Name          string
}

// OAuthCallback exchanges the authorization code, links or creates the
// account, and issues TreeBlog session tokens.
func (s *Service) OAuthCallback(ctx context.Context, provider OAuthProvider, code, signedState string) (Tokens, error) {
	cfg, ok := s.OAuthConfig(provider)
	if !ok {
		return Tokens{}, &Fault{Status: 501, Code: "OAUTH_NOT_CONFIGURED", Message: "Social login is not configured for this provider"}
	}
	state, err := s.verifyOAuthState(signedState, string(provider))
	if err != nil {
		return Tokens{}, err
	}
	if code == "" || len(code) > 2048 {
		return Tokens{}, &Fault{Status: 400, Code: "INVALID_OAUTH_CODE", Message: "Invalid authorization code"}
	}
	endpoints, ok := staticEndpoints(provider)
	if !ok {
		if provider != OAuthTreescanOIDC {
			return Tokens{}, &Fault{Status: 400, Code: "UNKNOWN_OAUTH_PROVIDER", Message: "Unknown OAuth provider"}
		}
		if endpoints, err = s.discoverOIDCEndpoints(cfg); err != nil {
			return Tokens{}, err
		}
	}
	token, idToken, err := s.exchangeCode(ctx, provider, cfg, endpoints, code, state.Verifier)
	if err != nil {
		return Tokens{}, err
	}
	profile, err := s.fetchProfile(ctx, provider, cfg, endpoints, token, idToken)
	if err != nil {
		return Tokens{}, err
	}
	if profile.Subject == "" {
		return Tokens{}, &Fault{Status: 502, Code: "OAUTH_PROFILE_FAILED", Message: "Provider did not return a subject"}
	}
	if profile.Email == "" || !profile.EmailVerified {
		return Tokens{}, &Fault{Status: 400, Code: "OAUTH_EMAIL_REQUIRED", Message: "A verified provider email is required; verify your email with the provider and try again"}
	}
	return s.linkOAuthIdentity(ctx, provider, profile, state.Terms, state.Privacy)
}

func (s *Service) exchangeCode(ctx context.Context, provider OAuthProvider, cfg OAuthClientConfig, endpoints oauthEndpoints, code, verifier string) (accessToken, idToken string, err error) {
	clientSecret := cfg.ClientSecret
	if provider == OAuthApple {
		clientSecret, err = appleClientSecret(cfg)
		if err != nil {
			return "", "", err
		}
	}
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("redirect_uri", oauthRedirectURI(cfg, provider))
	form.Set("client_id", cfg.ClientID)
	if clientSecret != "" {
		form.Set("client_secret", clientSecret)
	}
	form.Set("code_verifier", verifier)
	reqCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, endpoints.token, strings.NewReader(form.Encode()))
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	resp, err := s.httpClient().Do(req)
	if err != nil {
		return "", "", &Fault{Status: 502, Code: "OAUTH_EXCHANGE_FAILED", Message: "Authorization code exchange failed"}
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	if err != nil {
		return "", "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", "", &Fault{Status: 502, Code: "OAUTH_EXCHANGE_FAILED", Message: "Authorization code exchange failed"}
	}
	// GitHub may answer url-encoded unless Accept: application/json is honored.
	var parsed struct {
		AccessToken string `json:"access_token"`
		IDToken     string `json:"id_token"`
		TokenType   string `json:"token_type"`
		Error       string `json:"error"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		if values, parseErr := url.ParseQuery(string(body)); parseErr == nil {
			parsed.AccessToken = values.Get("access_token")
			parsed.IDToken = values.Get("id_token")
			parsed.Error = values.Get("error")
		}
	}
	if parsed.Error != "" || parsed.AccessToken == "" {
		return "", "", &Fault{Status: 502, Code: "OAUTH_EXCHANGE_FAILED", Message: "Authorization code exchange failed"}
	}
	return parsed.AccessToken, parsed.IDToken, nil
}

func (s *Service) fetchProfile(ctx context.Context, provider OAuthProvider, cfg OAuthClientConfig, endpoints oauthEndpoints, accessToken, idToken string) (oauthProfile, error) {
	switch provider {
	case OAuthApple:
		return appleProfileFromIDToken(idToken)
	case OAuthGitHub:
		return s.githubProfile(ctx, accessToken)
	}
	raw, err := s.getUserinfo(ctx, endpoints.userinfo, accessToken)
	if err != nil {
		return oauthProfile{}, err
	}
	switch provider {
	case OAuthGoogle, OAuthTreescanOIDC:
		return parseOIDCProfile(raw)
	case OAuthNaver:
		return parseNaverProfile(raw)
	case OAuthKakao:
		return parseKakaoProfile(raw)
	default:
		return oauthProfile{}, &Fault{Status: 400, Code: "UNKNOWN_OAUTH_PROVIDER", Message: "Unknown OAuth provider"}
	}
}

func (s *Service) getUserinfo(ctx context.Context, userinfoURL, accessToken string) (map[string]any, error) {
	reqCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, userinfoURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")
	resp, err := s.httpClient().Do(req)
	if err != nil {
		return nil, &Fault{Status: 502, Code: "OAUTH_PROFILE_FAILED", Message: "Provider profile request failed"}
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, &Fault{Status: 502, Code: "OAUTH_PROFILE_FAILED", Message: "Provider profile request failed"}
	}
	var raw map[string]any
	if err := json.NewDecoder(io.LimitReader(resp.Body, 64*1024)).Decode(&raw); err != nil {
		return nil, &Fault{Status: 502, Code: "OAUTH_PROFILE_FAILED", Message: "Invalid provider profile"}
	}
	return raw, nil
}

func parseOIDCProfile(raw map[string]any) (oauthProfile, error) {
	subject, _ := raw["sub"].(string)
	email, _ := raw["email"].(string)
	name, _ := raw["name"].(string)
	verified, _ := raw["email_verified"].(bool)
	return oauthProfile{Subject: subject, Email: email, EmailVerified: verified, Name: name}, nil
}

func parseNaverProfile(raw map[string]any) (oauthProfile, error) {
	response, _ := raw["response"].(map[string]any)
	if response == nil {
		return oauthProfile{}, &Fault{Status: 502, Code: "OAUTH_PROFILE_FAILED", Message: "Invalid Naver profile"}
	}
	id, _ := response["id"].(string)
	email, _ := response["email"].(string)
	name, _ := response["name"].(string)
	if name == "" {
		name, _ = response["nickname"].(string)
	}
	return oauthProfile{Subject: id, Email: email, EmailVerified: email != "", Name: name}, nil
}

func parseKakaoProfile(raw map[string]any) (oauthProfile, error) {
	var subject string
	switch id := raw["id"].(type) {
	case float64:
		subject = fmt.Sprintf("%.0f", id)
	case string:
		subject = id
	}
	account, _ := raw["kakao_account"].(map[string]any)
	profile, _ := account["profile"].(map[string]any)
	email, _ := account["email"].(string)
	verified, _ := account["is_email_verified"].(bool)
	name, _ := profile["nickname"].(string)
	return oauthProfile{Subject: subject, Email: email, EmailVerified: verified && email != "", Name: name}, nil
}

func (s *Service) githubProfile(ctx context.Context, accessToken string) (oauthProfile, error) {
	raw, err := s.getUserinfo(ctx, "https://api.github.com/user", accessToken)
	if err != nil {
		return oauthProfile{}, err
	}
	var subject string
	switch id := raw["id"].(type) {
	case float64:
		subject = fmt.Sprintf("%.0f", id)
	case string:
		subject = id
	}
	name, _ := raw["name"].(string)
	if name == "" {
		name, _ = raw["login"].(string)
	}
	profile := oauthProfile{Subject: subject, Name: name}
	if email, _ := raw["email"].(string); email != "" {
		profile.Email = email
		profile.EmailVerified = true
		return profile, nil
	}
	// /user/emails returns an array; fetch it with the array decoder.
	return s.githubEmails(ctx, accessToken, profile)
}

func (s *Service) githubEmails(ctx context.Context, accessToken string, profile oauthProfile) (oauthProfile, error) {
	reqCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, "https://api.github.com/user/emails", nil)
	if err != nil {
		return profile, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")
	resp, err := s.httpClient().Do(req)
	if err != nil {
		return profile, &Fault{Status: 502, Code: "OAUTH_PROFILE_FAILED", Message: "GitHub email lookup failed"}
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return profile, &Fault{Status: 502, Code: "OAUTH_PROFILE_FAILED", Message: "GitHub email lookup failed"}
	}
	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 64*1024)).Decode(&emails); err != nil {
		return profile, &Fault{Status: 502, Code: "OAUTH_PROFILE_FAILED", Message: "Invalid GitHub email list"}
	}
	for _, entry := range emails {
		if entry.Primary && entry.Verified {
			profile.Email = entry.Email
			profile.EmailVerified = true
			return profile, nil
		}
	}
	for _, entry := range emails {
		if entry.Verified {
			profile.Email = entry.Email
			profile.EmailVerified = true
			return profile, nil
		}
	}
	return profile, nil
}

// linkOAuthIdentity finds the account linked to a provider subject, links a
// matching verified email, or creates a new account, then issues tokens.
func (s *Service) linkOAuthIdentity(ctx context.Context, provider OAuthProvider, profile oauthProfile, termsVersion, privacyVersion string) (Tokens, error) {
	email, err := normalizeEmail(profile.Email)
	if err != nil {
		return Tokens{}, err
	}
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return Tokens{}, err
	}
	defer tx.Rollback(ctx)

	var userID string
	err = tx.QueryRow(ctx, `SELECT user_id FROM identities WHERE provider = $1 AND provider_subject = $2`,
		string(provider), profile.Subject).Scan(&userID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return Tokens{}, err
	}
	if userID == "" {
		err = tx.QueryRow(ctx, `SELECT id FROM users WHERE email = $1`, email).Scan(&userID)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return Tokens{}, err
		}
		if userID != "" {
			if _, err := tx.Exec(ctx, `INSERT INTO identities (user_id, provider, provider_subject, provider_email)
				VALUES ($1, $2, $3, $4) ON CONFLICT (provider, provider_subject) DO NOTHING`,
				userID, string(provider), profile.Subject, email); err != nil {
				return Tokens{}, err
			}
		}
	}
	if userID == "" {
		userID, err = s.createOAuthUser(ctx, tx, provider, profile, email, termsVersion, privacyVersion)
		if err != nil {
			return Tokens{}, err
		}
	} else {
		if _, err := tx.Exec(ctx, `UPDATE identities SET last_used_at = NOW()
			WHERE user_id = $1 AND provider = $2`, userID, string(provider)); err != nil {
			return Tokens{}, err
		}
	}

	var status string
	if err := tx.QueryRow(ctx, `SELECT status FROM users WHERE id = $1`, userID).Scan(&status); err != nil {
		return Tokens{}, err
	}
	if status != "ACTIVE" {
		return Tokens{}, &Fault{Status: 403, Code: "OAUTH_ACCOUNT_INACTIVE", Message: "Account is not active"}
	}
	refresh, sessionID, err := insertSessionTx(ctx, tx, userID)
	if err != nil {
		return Tokens{}, err
	}
	if _, err := tx.Exec(ctx, `UPDATE users SET last_login_at = NOW() WHERE id = $1`, userID); err != nil {
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

func (s *Service) createOAuthUser(ctx context.Context, tx pgx.Tx, provider OAuthProvider, profile oauthProfile, email, termsVersion, privacyVersion string) (string, error) {
	_ = ctx
	base := sanitizeOAuthUsername(profile.Name)
	if base == "" {
		base = sanitizeOAuthUsername(strings.SplitN(email, "@", 2)[0])
	}
	if base == "" {
		base = "user"
	}
	userID := ""
	for attempt := 0; attempt < 6; attempt++ {
		candidate := base
		if attempt > 0 {
			suffix := make([]byte, 3)
			if _, err := rand.Read(suffix); err != nil {
				return "", err
			}
			candidate = fmt.Sprintf("%s-%s", base, hex.EncodeToString(suffix)[:5])
		}
		if len(candidate) > 32 {
			candidate = candidate[:32]
		}
		tryID := uuid.NewString()
		displayName := strings.TrimSpace(profile.Name)
		if displayName == "" {
			displayName = candidate
		}
		if len(displayName) > 80 {
			displayName = displayName[:80]
		}
		_, err := tx.Exec(ctx, `INSERT INTO users (id, email, email_verified, status)
			VALUES ($1, $2, TRUE, 'ACTIVE')`, tryID, email)
		if err != nil {
			if isUniqueViolation(err) {
				return "", &Fault{Status: 409, Code: "DUPLICATE_EMAIL_OR_HANDLE", Message: "Email is already registered"}
			}
			return "", err
		}
		_, err = tx.Exec(ctx, `INSERT INTO user_profiles (user_id, username, display_name)
			VALUES ($1, $2, $3)`, tryID, candidate, displayName)
		if err != nil {
			if isUniqueViolation(err) {
				_, _ = tx.Exec(ctx, `DELETE FROM users WHERE id = $1`, tryID)
				continue
			}
			return "", err
		}
		if _, err := tx.Exec(ctx, `INSERT INTO identities (user_id, provider, provider_subject, provider_email)
			VALUES ($1, $2, $3, $4)`, tryID, string(provider), profile.Subject, email); err != nil {
			return "", err
		}
		if _, err := tx.Exec(ctx, `INSERT INTO user_consents (user_id, consent_type, policy_version, is_agreed)
			VALUES ($1, 'TERMS_OF_SERVICE', $2, TRUE), ($1, 'PRIVACY_POLICY', $3, TRUE)`,
			tryID, termsVersion, privacyVersion); err != nil {
			return "", err
		}
		if _, err := tx.Exec(ctx, `INSERT INTO user_roles (user_id, role_id)
			SELECT $1, id FROM roles WHERE code = 'USER'`, tryID); err != nil {
			return "", err
		}
		userID = tryID
		break
	}
	if userID == "" {
		return "", &Fault{Status: 409, Code: "DUPLICATE_EMAIL_OR_HANDLE", Message: "Could not allocate a username; please try again"}
	}
	return userID, nil
}

// sanitizeOAuthUsername keeps lowercase alphanumerics plus _ and -, matching
// usernamePattern, and guarantees the minimum length when possible.
func sanitizeOAuthUsername(raw string) string {
	lowered := strings.ToLower(strings.TrimSpace(raw))
	var sb strings.Builder
	for _, r := range lowered {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
			sb.WriteRune(r)
		}
	}
	cleaned := strings.Trim(sb.String(), "_-")
	if len(cleaned) > 24 {
		cleaned = cleaned[:24]
	}
	if len(cleaned) < 3 {
		return ""
	}
	if cleaned[0] >= '0' && cleaned[0] <= '9' {
		cleaned = "u-" + cleaned
	}
	return cleaned
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

// insertSessionTx creates a session row with a fresh refresh token family.
func insertSessionTx(ctx context.Context, tx pgx.Tx, userID string) (refresh, sessionID string, err error) {
	refresh, hash, err := randomToken()
	if err != nil {
		return "", "", err
	}
	sessionID, familyID := uuid.NewString(), uuid.NewString()
	if _, err := tx.Exec(ctx, `INSERT INTO sessions(id,user_id,refresh_token_hash,token_family_id,expires_at)
		VALUES($1,$2,$3,$4,NOW()+INTERVAL '14 days')`, sessionID, userID, hash, familyID); err != nil {
		return "", "", err
	}
	return refresh, sessionID, nil
}
