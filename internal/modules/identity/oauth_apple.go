package identity

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"strings"
	"time"
)

// appleClientSecret builds the ES256-signed client_secret JWT Apple requires
// for the authorization-code exchange.
func appleClientSecret(cfg OAuthClientConfig) (string, error) {
	if cfg.AppleTeamID == "" || cfg.AppleKeyID == "" || cfg.AppleKeyPEM == "" || cfg.ClientID == "" {
		return "", &Fault{Status: 501, Code: "OAUTH_NOT_CONFIGURED", Message: "Apple Sign In is not fully configured"}
	}
	key, err := parseApplePrivateKey(cfg.AppleKeyPEM)
	if err != nil {
		return "", err
	}
	now := time.Now()
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"ES256","kid":"` + cfg.AppleKeyID + `"}`))
	claims, err := json.Marshal(map[string]any{
		"iss": cfg.AppleTeamID,
		"iat": now.Unix(),
		"exp": now.Add(10 * time.Minute).Unix(),
		"aud": "https://appleid.apple.com",
		"sub": cfg.ClientID,
	})
	if err != nil {
		return "", err
	}
	payload := base64.RawURLEncoding.EncodeToString(claims)
	signingInput := header + "." + payload
	sum := sha256.Sum256([]byte(signingInput))
	r, s, err := ecdsa.Sign(rand.Reader, key, sum[:])
	if err != nil {
		return "", err
	}
	signature := make([]byte, 0, 64)
	signature = append(signature, leftPad32(r.Bytes())...)
	signature = append(signature, leftPad32(s.Bytes())...)
	return signingInput + "." + base64.RawURLEncoding.EncodeToString(signature), nil
}

func parseApplePrivateKey(pemText string) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemText))
	if block == nil {
		return nil, &Fault{Status: 501, Code: "OAUTH_NOT_CONFIGURED", Message: "Invalid Apple private key"}
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		if ecKey, secErr := x509.ParseECPrivateKey(block.Bytes); secErr == nil {
			return ecKey, nil
		}
		return nil, &Fault{Status: 501, Code: "OAUTH_NOT_CONFIGURED", Message: "Invalid Apple private key"}
	}
	ecKey, ok := key.(*ecdsa.PrivateKey)
	if !ok {
		return nil, &Fault{Status: 501, Code: "OAUTH_NOT_CONFIGURED", Message: "Apple private key is not ECDSA"}
	}
	return ecKey, nil
}

func leftPad32(raw []byte) []byte {
	if len(raw) >= 32 {
		return raw[len(raw)-32:]
	}
	padded := make([]byte, 32)
	copy(padded[32-len(raw):], raw)
	return padded
}

// appleProfileFromIDToken extracts the subject and email from Apple's
// id_token. The token arrives over the server-to-server TLS channel from
// Apple's token endpoint, which authenticates its origin.
func appleProfileFromIDToken(idToken string) (oauthProfile, error) {
	parts := strings.Split(idToken, ".")
	if len(parts) != 3 {
		return oauthProfile{}, &Fault{Status: 502, Code: "OAUTH_PROFILE_FAILED", Message: "Invalid Apple identity token"}
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return oauthProfile{}, &Fault{Status: 502, Code: "OAUTH_PROFILE_FAILED", Message: "Invalid Apple identity token"}
	}
	var claims struct {
		Subject string `json:"sub"`
		Email   string `json:"email"`
	}
	if err := json.Unmarshal(raw, &claims); err != nil {
		return oauthProfile{}, &Fault{Status: 502, Code: "OAUTH_PROFILE_FAILED", Message: "Invalid Apple identity token"}
	}
	if claims.Subject == "" {
		return oauthProfile{}, &Fault{Status: 502, Code: "OAUTH_PROFILE_FAILED", Message: "Apple identity token has no subject"}
	}
	var verifiedRaw map[string]any
	_ = json.Unmarshal(raw, &verifiedRaw)
	verified := false
	switch value := verifiedRaw["email_verified"].(type) {
	case bool:
		verified = value
	case string:
		verified = value == "true"
	}
	return oauthProfile{Subject: claims.Subject, Email: claims.Email, EmailVerified: verified}, nil
}
