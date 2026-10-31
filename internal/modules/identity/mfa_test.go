package identity

import (
	"encoding/json"
	"testing"
	"time"
)

func TestTOTP_Lifecycle(t *testing.T) {
	secret, err := GenerateTOTPSecret()
	if err != nil {
		t.Fatalf("GenerateTOTPSecret failed: %v", err)
	}
	if len(secret) == 0 {
		t.Fatal("expected non-empty secret")
	}

	now := time.Now()
	counter := uint64(now.Unix() / 30)
	code, err := ComputeTOTP(secret, counter)
	if err != nil {
		t.Fatalf("ComputeTOTP failed: %v", err)
	}
	if len(code) != 6 {
		t.Fatalf("expected 6-digit code, got %q", code)
	}

	// Correct code verification
	if !VerifyTOTP(secret, code, now) {
		t.Fatal("expected code to verify at current time")
	}

	// 1 step (30s) drift tolerance check
	if !VerifyTOTP(secret, code, now.Add(25*time.Second)) {
		t.Fatal("expected code to verify within +25s")
	}
	if !VerifyTOTP(secret, code, now.Add(-25*time.Second)) {
		t.Fatal("expected code to verify within -25s")
	}

	// Invalid code check
	if VerifyTOTP(secret, "000000", now) && code != "000000" {
		t.Fatal("expected wrong code to fail")
	}
}

func TestBackupCodes(t *testing.T) {
	codes, hashes, err := GenerateBackupCodes(8)
	if err != nil {
		t.Fatalf("GenerateBackupCodes failed: %v", err)
	}
	if len(codes) != 8 || len(hashes) != 8 {
		t.Fatalf("expected 8 codes, got %d", len(codes))
	}

	storedJSON, err := json.Marshal(hashes)
	if err != nil {
		t.Fatal(err)
	}

	// Verify valid code
	idx, err := VerifyBackupCode(codes[3], storedJSON)
	if err != nil {
		t.Fatalf("VerifyBackupCode failed: %v", err)
	}
	if idx != 3 {
		t.Fatalf("expected index 3, got %d", idx)
	}

	// Verify invalid code
	idx, err = VerifyBackupCode("XXXXXXXX", storedJSON)
	if err != nil {
		t.Fatalf("VerifyBackupCode failed: %v", err)
	}
	if idx != -1 {
		t.Fatalf("expected index -1 for invalid code, got %d", idx)
	}
}

