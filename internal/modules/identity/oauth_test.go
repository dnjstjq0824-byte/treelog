package identity

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"math/big"
	"net/url"
	"strings"
	"testing"
	"time"
)

func oauthTestService() *Service {
	pepper := make([]byte, 32)
	_, _ = rand.Read(pepper)
	return &Service{
		CodePepper: pepper,
		OAuth: map[OAuthProvider]OAuthClientConfig{
			OAuthGoogle: {ClientID: "google-id", ClientSecret: "google-secret", RedirectBase: "http://localhost:8080", AppBaseURL: "http://localhost:3000"},
			OAuthGitHub: {ClientID: "github-id", ClientSecret: "github-secret", RedirectBase: "http://localhost:8080", AppBaseURL: "http://localhost:3000"},
		},
	}
}

func TestOAuthLoginURL(t *testing.T) {
	svc := oauthTestService()
	authURL, signedState, err := svc.OAuthLoginURL(OAuthGoogle, "v1.2", "v2.0")
	if err != nil {
		t.Fatalf("OAuthLoginURL: %v", err)
	}
	parsed, err := url.Parse(authURL)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if parsed.Host != "accounts.google.com" {
		t.Fatalf("host = %s", parsed.Host)
	}
	query := parsed.Query()
	if query.Get("client_id") != "google-id" || query.Get("response_type") != "code" {
		t.Fatalf("bad params: %s", query.Encode())
	}
	if query.Get("code_challenge_method") != "S256" || query.Get("code_challenge") == "" {
		t.Fatalf("PKCE missing: %s", query.Encode())
	}
	if query.Get("state") != signedState {
		t.Fatalf("state mismatch")
	}
	state, err := svc.verifyOAuthState(signedState, string(OAuthGoogle))
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if state.Terms != "v1.2" || state.Privacy != "v2.0" || state.Verifier == "" {
		t.Fatalf("bad state: %+v", state)
	}
}

func TestOAuthLoginURLUnconfigured(t *testing.T) {
	svc := oauthTestService()
	if _, _, err := svc.OAuthLoginURL(OAuthNaver, "", ""); err == nil {
		t.Fatalf("unconfigured provider accepted")
	}
}

// Red drill: any tampering with the signed state must fail verification.
func TestOAuthStateTamperRejected(t *testing.T) {
	svc := oauthTestService()
	_, signed, err := svc.OAuthLoginURL(OAuthGitHub, "", "")
	if err != nil {
		t.Fatalf("OAuthLoginURL: %v", err)
	}
	tampered := signed[:len(signed)-1] + flipLast(signed[len(signed)-1:])
	if _, err := svc.verifyOAuthState(tampered, string(OAuthGitHub)); err == nil {
		t.Fatalf("tampered signature accepted")
	}
	parts := strings.SplitN(signed, ".", 2)
	forged := parts[0][:len(parts[0])-2] + "AA." + parts[1]
	if _, err := svc.verifyOAuthState(forged, string(OAuthGitHub)); err == nil {
		t.Fatalf("tampered payload accepted")
	}
	if _, err := svc.verifyOAuthState(signed, string(OAuthGoogle)); err == nil {
		t.Fatalf("cross-provider state accepted")
	}
	if _, err := svc.verifyOAuthState("garbage", string(OAuthGitHub)); err == nil {
		t.Fatalf("garbage state accepted")
	}
	expired, err := svc.signOAuthState(oauthState{Nonce: "n", Provider: string(OAuthGitHub), ExpiresAt: time.Now().Add(-time.Minute).Unix(), Verifier: "v"})
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	if _, err := svc.verifyOAuthState(expired, string(OAuthGitHub)); err == nil {
		t.Fatalf("expired state accepted")
	}
}

func flipLast(s string) string {
	if s == "a" {
		return "b"
	}
	return "a"
}

func TestSanitizeOAuthUsername(t *testing.T) {
	cases := map[string]string{
		"Kim Tree":      "kimtree",
		"user_123":      "user_123",
		"김트리blogger": "blogger",
		"ab":            "",
		"9lives":        "u-9lives",
		"A!@#B$%C^D":   "abcd",
		"Gold-Roger_99": "gold-roger_99",
	}
	for raw, want := range cases {
		if got := sanitizeOAuthUsername(raw); got != want {
			t.Errorf("sanitize(%q) = %q, want %q", raw, got, want)
		}
	}
}

func TestParseProfileVariants(t *testing.T) {
	oidc, err := parseOIDCProfile(map[string]any{"sub": "s", "email": "a@b.c", "email_verified": true, "name": "N"})
	if err != nil || !oidc.EmailVerified || oidc.Subject != "s" {
		t.Fatalf("oidc: %+v %v", oidc, err)
	}
	naver, err := parseNaverProfile(map[string]any{"response": map[string]any{"id": "n1", "email": "n@x.y", "nickname": "NN"}})
	if err != nil || naver.Subject != "n1" || !naver.EmailVerified || naver.Name != "NN" {
		t.Fatalf("naver: %+v %v", naver, err)
	}
	if _, err := parseNaverProfile(map[string]any{}); err == nil {
		t.Fatalf("empty naver accepted")
	}
	kakao, err := parseKakaoProfile(map[string]any{
		"id":           float64(12345),
		"kakao_account": map[string]any{"email": "k@x.y", "is_email_verified": true, "profile": map[string]any{"nickname": "K"}},
	})
	if err != nil || kakao.Subject != "12345" || !kakao.EmailVerified {
		t.Fatalf("kakao: %+v %v", kakao, err)
	}
}

func TestAppleClientSecretVerifies(t *testing.T) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	pemText := string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}))
	secret, err := appleClientSecret(OAuthClientConfig{
		ClientID: "com.example.app", AppleTeamID: "TEAM123", AppleKeyID: "KEY456", AppleKeyPEM: pemText,
	})
	if err != nil {
		t.Fatalf("appleClientSecret: %v", err)
	}
	parts := strings.Split(secret, ".")
	if len(parts) != 3 {
		t.Fatalf("bad jwt shape")
	}
	header, _ := base64.RawURLEncoding.DecodeString(parts[0])
	if !strings.Contains(string(header), "KEY456") || !strings.Contains(string(header), "ES256") {
		t.Fatalf("bad header: %s", header)
	}
	sum := sha256.Sum256([]byte(parts[0] + "." + parts[1]))
	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil || len(sig) != 64 {
		t.Fatalf("bad signature encoding")
	}
	r := new(big.Int).SetBytes(sig[:32])
	s := new(big.Int).SetBytes(sig[32:])
	if !ecdsa.Verify(&key.PublicKey, sum[:], r, s) {
		t.Fatalf("ES256 signature does not verify")
	}
	if _, err := appleClientSecret(OAuthClientConfig{}); err == nil {
		t.Fatalf("empty apple config accepted")
	}
}

func TestAppleProfileFromIDToken(t *testing.T) {
	payload := base64.RawURLEncoding.EncodeToString([]byte(`{"sub":"apple-sub-1","email":"a@privaterelay.appleid.com","email_verified":"true"}`))
	profile, err := appleProfileFromIDToken("header." + payload + ".sig")
	if err != nil {
		t.Fatalf("profile: %v", err)
	}
	if profile.Subject != "apple-sub-1" || !profile.EmailVerified {
		t.Fatalf("bad profile: %+v", profile)
	}
	if _, err := appleProfileFromIDToken("broken"); err == nil {
		t.Fatalf("broken token accepted")
	}
}

func TestParseOAuthProvider(t *testing.T) {
	for raw, want := range map[string]OAuthProvider{
		"google": OAuthGoogle, "APPLE": OAuthApple, "github": OAuthGitHub,
		"naver": OAuthNaver, "kakao": OAuthKakao, "treescan": OAuthTreescanOIDC,
	} {
		got, ok := parseOAuthProvider(raw)
		if !ok || got != want {
			t.Errorf("parse(%q) = %v %v", raw, got, ok)
		}
	}
	if _, ok := parseOAuthProvider("facebook"); ok {
		t.Errorf("unknown provider accepted")
	}
}
