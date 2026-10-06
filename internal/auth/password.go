package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"strings"

	"golang.org/x/crypto/scrypt"
)

// HashPassword hashes with scrypt in the legacy "scrypt:salt:hash" format.
func HashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	saltHex := hex.EncodeToString(salt)
	hash, err := scrypt.Key([]byte(password), []byte(saltHex), 16384, 8, 1, 64)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("scrypt:%s:%x", saltHex, hash), nil
}

// VerifyPassword checks a password against a stored hash in constant time.
func VerifyPassword(password, stored string) bool {
	parts := strings.Split(stored, ":")
	if len(parts) != 3 || parts[0] != "scrypt" {
		return false
	}
	expected, err := hex.DecodeString(parts[2])
	if err != nil {
		return false
	}
	candidate, err := scrypt.Key([]byte(password), []byte(parts[1]), 16384, 8, 1, len(expected))
	if err != nil {
		return false
	}
	return subtle.ConstantTimeCompare(candidate, expected) == 1
}
