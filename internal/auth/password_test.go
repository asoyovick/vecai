package auth

import (
	"encoding/hex"
	"testing"

	"golang.org/x/crypto/scrypt"
)

func legacyScryptHexSalt(t *testing.T, password, saltHex string) string {
	t.Helper()
	hash, err := scrypt.Key([]byte(password), []byte(saltHex), 16384, 8, 1, 64)
	if err != nil {
		t.Fatalf("scrypt: %v", err)
	}
	return hex.EncodeToString(hash)
}

func TestHashAndVerifyPassword(t *testing.T) {
	hash, err := HashPassword("demo1234")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if hash == "demo1234" {
		t.Fatal("hash must not contain the plaintext password")
	}
	if !VerifyPassword("demo1234", hash) {
		t.Fatal("VerifyPassword rejected a correct password")
	}
	if VerifyPassword("wrong", hash) {
		t.Fatal("VerifyPassword accepted a wrong password")
	}
}

func TestVerifyPasswordRejectsMalformedHashes(t *testing.T) {
	cases := []string{"", "plaintext", "scrypt:onlysalt", "scrypt:abc:def", "bcrypt:$2a$xyz"}
	for _, stored := range cases {
		if VerifyPassword("demo1234", stored) {
			t.Fatalf("VerifyPassword accepted malformed hash %q", stored)
		}
	}
}

// The original Node backend salted scrypt with the hex salt string itself;
// keep that behaviour so existing demo hashes still verify.
func TestVerifyLegacyHexSaltHash(t *testing.T) {
	password := "demo1234"
	saltHex := "0123456789abcdef0123456789abcdef"
	hash := legacyScryptHexSalt(t, password, saltHex)
	stored := "scrypt:" + saltHex + ":" + hash
	if !VerifyPassword(password, stored) {
		t.Fatal("VerifyPassword rejected the legacy hex-salt format")
	}
	if VerifyPassword("nope", stored) {
		t.Fatal("VerifyPassword accepted a wrong password against a legacy hash")
	}
}
