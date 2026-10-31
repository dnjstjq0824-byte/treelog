package identity

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base32"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
)

// GenerateTOTPSecret creates a new random 20-byte base32 secret.
func GenerateTOTPSecret() (string, error) {
	bytes := make([]byte, 20)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(bytes), nil
}

// GenerateBackupCodes generates n 8-character random alphanumeric backup codes.
func GenerateBackupCodes(count int) ([]string, []string, error) {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	codes := make([]string, count)
	hashes := make([]string, count)

	for i := 0; i < count; i++ {
		bytes := make([]byte, 8)
		if _, err := rand.Read(bytes); err != nil {
			return nil, nil, err
		}
		var sb strings.Builder
		for _, b := range bytes {
			sb.WriteByte(charset[int(b)%len(charset)])
		}
		code := sb.String()
		codes[i] = code
		hash := sha256.Sum256([]byte(code))
		hashes[i] = hex.EncodeToString(hash[:])
	}
	return codes, hashes, nil
}

// ComputeTOTP generates the 6-digit TOTP code for the given secret at counter value.
func ComputeTOTP(secret string, counter uint64) (string, error) {
	key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(strings.ToUpper(strings.TrimSpace(secret)))
	if err != nil {
		return "", fmt.Errorf("invalid base32 secret: %w", err)
	}

	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, counter)

	mac := hmac.New(sha1.New, key)
	mac.Write(buf)
	sum := mac.Sum(nil)

	offset := sum[len(sum)-1] & 0x0f
	binaryCode := (int(sum[offset])&0x7f)<<24 |
		(int(sum[offset+1])&0xff)<<16 |
		(int(sum[offset+2])&0xff)<<8 |
		(int(sum[offset+3]) & 0xff)

	otp := binaryCode % int(math.Pow10(6))
	return fmt.Sprintf("%06d", otp), nil
}

// VerifyTOTP checks a 6-digit TOTP code against a secret within ±1 time step (30s window).
func VerifyTOTP(secret, code string, now time.Time) bool {
	code = strings.TrimSpace(code)
	if len(code) != 6 {
		return false
	}
	currentCounter := uint64(now.Unix() / 30)

	// Check current, -1, and +1 step for clock drift tolerance
	for _, counter := range []uint64{currentCounter, currentCounter - 1, currentCounter + 1} {
		expected, err := ComputeTOTP(secret, counter)
		if err == nil && hmac.Equal([]byte(expected), []byte(code)) {
			return true
		}
	}
	return false
}

// VerifyBackupCode verifies a candidate backup code against the stored list of SHA-256 hashes.
// Returns the index of the matched hash or -1 if not found.
func VerifyBackupCode(code string, storedHashesJSON []byte) (int, error) {
	code = strings.ToUpper(strings.TrimSpace(code))
	if len(code) != 8 {
		return -1, errors.New("invalid backup code length")
	}
	var hashes []string
	if err := json.Unmarshal(storedHashesJSON, &hashes); err != nil {
		return -1, err
	}
	targetHash := sha256.Sum256([]byte(code))
	targetHex := hex.EncodeToString(targetHash[:])

	for i, h := range hashes {
		if hmac.Equal([]byte(h), []byte(targetHex)) {
			return i, nil
		}
	}
	return -1, nil
}

