package identity

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
)

// mfaSecretVersion prefixes every sealed MFA secret so future key rotations
// can distinguish storage formats.
var mfaSecretVersion = []byte("v1")

// deriveMFAKey derives a dedicated AES-256 key from the service code pepper.
// The pepper itself is never used directly as a cipher key.
func deriveMFAKey(pepper []byte) [32]byte {
	h := sha256.New()
	h.Write([]byte("treeblog/mfa-storage/v1"))
	h.Write([]byte{0x00})
	h.Write(pepper)
	var key [32]byte
	copy(key[:], h.Sum(nil))
	return key
}

// EncryptMFASecret seals a TOTP secret with AES-256-GCM. The returned bytes
// are stored in users.mfa_secret_enc (BYTEA).
func EncryptMFASecret(pepper []byte, secret string) ([]byte, error) {
	if len(pepper) < 32 {
		return nil, errors.New("a 32-byte code pepper is required to seal MFA secrets")
	}
	if secret == "" {
		return nil, errors.New("MFA secret must not be empty")
	}
	key := deriveMFAKey(pepper)
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	sealed := gcm.Seal(nil, nonce, []byte(secret), nil)
	out := make([]byte, 0, len(mfaSecretVersion)+len(nonce)+len(sealed))
	out = append(out, mfaSecretVersion...)
	out = append(out, nonce...)
	out = append(out, sealed...)
	return out, nil
}

// DecryptMFASecret opens a sealed MFA secret previously sealed with
// EncryptMFASecret. Authentication failures return an error and no secret.
func DecryptMFASecret(pepper, sealed []byte) (string, error) {
	if len(pepper) < 32 {
		return "", errors.New("a 32-byte code pepper is required to open MFA secrets")
	}
	if len(sealed) <= len(mfaSecretVersion) {
		return "", errors.New("sealed MFA secret is too short")
	}
	for i, b := range mfaSecretVersion {
		if sealed[i] != b {
			return "", errors.New("unsupported MFA secret version")
		}
	}
	key := deriveMFAKey(pepper)
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	rest := sealed[len(mfaSecretVersion):]
	if len(rest) <= gcm.NonceSize() {
		return "", errors.New("sealed MFA secret is too short")
	}
	nonce, ciphertext := rest[:gcm.NonceSize()], rest[gcm.NonceSize():]
	plain, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", errors.New("MFA secret authentication failed")
	}
	return string(plain), nil
}
