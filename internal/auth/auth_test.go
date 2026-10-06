package auth

import "testing"

func TestHashAndVerifyRoundTrip(t *testing.T) {
	hash, err := HashPassword("hunter2-pass")
	if err != nil {
		t.Fatal(err)
	}
	if !VerifyPassword("hunter2-pass", hash) {
		t.Fatal("expected password to verify")
	}
	if VerifyPassword("wrong", hash) {
		t.Fatal("expected wrong password to fail")
	}
}

func TestVerifyCompatibleWithOriginalFormat(t *testing.T) {
	// Produced by the original Node scryptSync(password, saltHexString, 64)
	// where the hex string's own bytes were the scrypt salt.
	stored := "scrypt:12ca40480567f50305269bf2e2fd5bd5:f121a17ee06fdf59277527002e40adedb0965507410d03febce080229d06ccfc61dad59f1ddfc5cf1b9dd374a5db2842f39492643152879ab319af760515039c"
	if !VerifyPassword("demo1234", stored) {
		t.Fatal("expected legacy hash to verify")
	}
}
