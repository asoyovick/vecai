package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"testing"
	"time"
)

func TestTokenHashIsStableAndSHA256(t *testing.T) {
	token, err := newToken()
	if err != nil {
		t.Fatalf("newToken: %v", err)
	}
	if len(token) != 64 {
		t.Fatalf("token length = %d, want 64 hex chars", len(token))
	}
	sum := sha256.Sum256([]byte(token))
	if TokenHash(token) != hex.EncodeToString(sum[:]) {
		t.Fatal("TokenHash is not the SHA-256 of the token")
	}
	if TokenHash(token) == TokenHash(token+"x") {
		t.Fatal("TokenHash collided for different tokens")
	}
}

func TestSessionCookieHeaders(t *testing.T) {
	set := SessionCookieHeader("abc", time.Now().Add(sessionTTL))
	if set.Name != SessionCookie {
		t.Fatalf("cookie name = %q", set.Name)
	}
	if !set.HttpOnly {
		t.Fatal("session cookie must be HttpOnly")
	}
	clear := ClearSessionCookieHeader()
	if clear.MaxAge != -1 {
		t.Fatalf("clear cookie MaxAge = %d", clear.MaxAge)
	}
	var _ *http.Cookie = clear
}
