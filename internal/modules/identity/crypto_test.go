package identity

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestPasswordHashAndPolicy(t *testing.T) {
	for _, password := range []string{"short", "onlyletters", "1234567890", "          "} {
		if ValidatePassword(password) == nil {
			t.Fatalf("accepted weak password %q", password)
		}
	}
	hash, err := HashPassword("ThisIsSafe123!")
	if err != nil {
		t.Fatal(err)
	}
	if !VerifyPassword("ThisIsSafe123!", hash) || VerifyPassword("wrongpassword123!", hash) ||
		VerifyPassword("ThisIsSafe123!", strings.Replace(hash, "m=65536", "m=9999999", 1)) {
		t.Fatal("password hash verification failed")
	}
}

func TestJWTSignatureClaimsAndExpiry(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "jwt.pem")
	if err := os.WriteFile(path, pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}), 0600); err != nil {
		t.Fatal(err)
	}
	signer, err := LoadTokenSigner(path, "treeblog-test", "treeblog-web")
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	userID, sessionID := uuid.NewString(), uuid.NewString()
	token, err := signer.Issue(userID, sessionID, now)
	if err != nil {
		t.Fatal(err)
	}
	claims, err := signer.Verify(token, now.Add(time.Minute))
	if err != nil || claims.Subject != userID || claims.SessionID != sessionID {
		t.Fatalf("valid token rejected: %v", err)
	}
	if _, err := signer.Verify(token+"x", now); err == nil {
		t.Fatal("tampered token accepted")
	}
	if _, err := signer.Verify(token, now.Add(16*time.Minute)); err == nil {
		t.Fatal("expired token accepted")
	}
}
