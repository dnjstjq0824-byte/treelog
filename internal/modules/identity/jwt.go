package identity

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
)

type AccessClaims struct {
	Issuer    string `json:"iss"`
	Audience  string `json:"aud"`
	Subject   string `json:"sub"`
	SessionID string `json:"sid"`
	ID        string `json:"jti"`
	IssuedAt  int64  `json:"iat"`
	ExpiresAt int64  `json:"exp"`
}

type TokenSigner struct {
	private  *rsa.PrivateKey
	public   *rsa.PublicKey
	issuer   string
	audience string
}

func LoadTokenSigner(path, issuer, audience string) (*TokenSigner, error) {
	if path == "" || issuer == "" || audience == "" {
		return nil, errors.New("JWT key path, issuer, and audience are required")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("invalid JWT PEM key")
	}
	var key *rsa.PrivateKey
	switch block.Type {
	case "PRIVATE KEY":
		parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		var ok bool
		key, ok = parsed.(*rsa.PrivateKey)
		if !ok {
			return nil, errors.New("JWT key must be RSA")
		}
	case "RSA PRIVATE KEY":
		key, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("unsupported JWT key type")
	}
	if key.N.BitLen() < 2048 {
		return nil, errors.New("JWT RSA key must be at least 2048 bits")
	}
	return &TokenSigner{private: key, public: &key.PublicKey, issuer: issuer, audience: audience}, nil
}

func (s *TokenSigner) Issue(userID, sessionID string, now time.Time) (string, error) {
	if _, err := uuid.Parse(userID); err != nil {
		return "", errors.New("invalid user id")
	}
	if _, err := uuid.Parse(sessionID); err != nil {
		return "", errors.New("invalid session id")
	}
	claims := AccessClaims{Issuer: s.issuer, Audience: s.audience, Subject: userID,
		SessionID: sessionID, ID: uuid.NewString(), IssuedAt: now.Unix(), ExpiresAt: now.Add(15 * time.Minute).Unix()}
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256","typ":"JWT"}`))
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	signingInput := header + "." + base64.RawURLEncoding.EncodeToString(payload)
	hash := sha256.Sum256([]byte(signingInput))
	signature, err := rsa.SignPKCS1v15(rand.Reader, s.private, crypto.SHA256, hash[:])
	if err != nil {
		return "", err
	}
	return signingInput + "." + base64.RawURLEncoding.EncodeToString(signature), nil
}

func (s *TokenSigner) Verify(token string, now time.Time) (AccessClaims, error) {
	var claims AccessClaims
	parts := strings.Split(token, ".")
	if len(parts) != 3 || len(token) > 8192 {
		return claims, errors.New("malformed access token")
	}
	headerJSON, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil || string(headerJSON) != `{"alg":"RS256","typ":"JWT"}` {
		return claims, errors.New("unsupported token header")
	}
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return claims, errors.New("invalid token signature")
	}
	hash := sha256.Sum256([]byte(parts[0] + "." + parts[1]))
	if err := rsa.VerifyPKCS1v15(s.public, crypto.SHA256, hash[:], signature); err != nil {
		return claims, errors.New("invalid token signature")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil || json.Unmarshal(payload, &claims) != nil {
		return AccessClaims{}, errors.New("invalid token claims")
	}
	if claims.Issuer != s.issuer || claims.Audience != s.audience ||
		claims.ExpiresAt <= now.Unix() || claims.IssuedAt > now.Add(30*time.Second).Unix() ||
		claims.ExpiresAt-claims.IssuedAt > int64((15*time.Minute).Seconds()) ||
		claims.ID == "" {
		return AccessClaims{}, errors.New("access token expired or invalid")
	}
	if _, err := uuid.Parse(claims.Subject); err != nil {
		return AccessClaims{}, fmt.Errorf("invalid subject: %w", err)
	}
	if _, err := uuid.Parse(claims.SessionID); err != nil {
		return AccessClaims{}, fmt.Errorf("invalid session: %w", err)
	}
	return claims, nil
}
