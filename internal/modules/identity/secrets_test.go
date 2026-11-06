package identity

import (
	"bytes"
	"crypto/rand"
	"testing"
	"time"
)

func testPepper(t *testing.T) []byte {
	t.Helper()
	pepper := make([]byte, 32)
	if _, err := rand.Read(pepper); err != nil {
		t.Fatalf("rand: %v", err)
	}
	return pepper
}

func TestMFASecretRoundTrip(t *testing.T) {
	pepper := testPepper(t)
	secret, err := GenerateTOTPSecret()
	if err != nil {
		t.Fatalf("GenerateTOTPSecret: %v", err)
	}
	sealed, err := EncryptMFASecret(pepper, secret)
	if err != nil {
		t.Fatalf("EncryptMFASecret: %v", err)
	}
	if bytes.Contains(sealed, []byte(secret)) {
		t.Fatalf("sealed payload leaks the plaintext secret")
	}
	opened, err := DecryptMFASecret(pepper, sealed)
	if err != nil {
		t.Fatalf("DecryptMFASecret: %v", err)
	}
	if opened != secret {
		t.Fatalf("round trip mismatch")
	}
	code, err := ComputeTOTP(opened, uint64(time.Now().Unix()/30))
	if err != nil {
		t.Fatalf("ComputeTOTP: %v", err)
	}
	if !VerifyTOTP(opened, code, time.Now()) {
		t.Fatalf("decrypted secret does not verify its own TOTP code")
	}
}

func TestMFASecretTamperRejected(t *testing.T) {
	pepper := testPepper(t)
	sealed, err := EncryptMFASecret(pepper, "JBSWY3DPEHPK3PXP")
	if err != nil {
		t.Fatalf("EncryptMFASecret: %v", err)
	}
	sealed[len(sealed)-1] ^= 0xff
	if _, err := DecryptMFASecret(pepper, sealed); err == nil {
		t.Fatalf("tampered payload accepted")
	}
}

func TestMFASecretWrongPepperRejected(t *testing.T) {
	sealed, err := EncryptMFASecret(testPepper(t), "JBSWY3DPEHPK3PXP")
	if err != nil {
		t.Fatalf("EncryptMFASecret: %v", err)
	}
	if _, err := DecryptMFASecret(testPepper(t), sealed); err == nil {
		t.Fatalf("wrong pepper accepted")
	}
}

func TestMFASecretRejectsShortPepperAndPlaintext(t *testing.T) {
	if _, err := EncryptMFASecret([]byte("short"), "JBSWY3DPEHPK3PXP"); err == nil {
		t.Fatalf("short pepper accepted")
	}
	if _, err := DecryptMFASecret(testPepper(t), []byte("JBSWY3DPEHPK3PXP")); err == nil {
		t.Fatalf("legacy plaintext payload accepted as sealed secret")
	}
}
