package identity

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/crypto/argon2"
)

const (
	argonTime    = 3
	argonMemory  = 64 * 1024
	argonThreads = 4
	argonKeyLen  = 32
)

func ValidatePassword(password string) error {
	if len(password) > 1024 || !utf8.ValidString(password) ||
		utf8.RuneCountInString(password) < 10 || strings.TrimSpace(password) == "" {
		return errors.New("password must contain at least 10 characters")
	}
	if utf8.RuneCountInString(password) >= 15 {
		return nil
	}
	var letters, digits, symbols bool
	for _, r := range password {
		letters = letters || unicode.IsLetter(r)
		digits = digits || unicode.IsDigit(r)
		symbols = symbols || unicode.IsPunct(r) || unicode.IsSymbol(r)
	}
	if !letters || !digits || !symbols {
		return errors.New("short passwords need a letter, number, and symbol")
	}
	return nil
}

func HashPassword(password string) (string, error) {
	if err := ValidatePassword(password); err != nil {
		return "", err
	}
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	key := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)
	return fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s", argonMemory, argonTime,
		argonThreads, base64.RawStdEncoding.EncodeToString(salt), base64.RawStdEncoding.EncodeToString(key)), nil
}

func VerifyPassword(password, encoded string) bool {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[0] != "" || parts[1] != "argon2id" || parts[2] != "v=19" {
		return false
	}
	var memory, timeCost, threads uint32
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &timeCost, &threads); err != nil ||
		memory != argonMemory || timeCost != argonTime || threads != argonThreads {
		return false
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil || len(salt) != 16 {
		return false
	}
	stored, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil || len(stored) != argonKeyLen {
		return false
	}
	calculated := argon2.IDKey([]byte(password), salt, timeCost, memory, uint8(threads), argonKeyLen)
	return subtle.ConstantTimeCompare(stored, calculated) == 1
}
