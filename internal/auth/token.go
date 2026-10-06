package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"
)

const (
	SessionCookie = "vecai_session"
	sessionTTL    = 7 * 24 * time.Hour
)

func tokenHash(token string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(token)))
}

// TokenHash is the stored form of a session token.
func TokenHash(token string) string { return tokenHash(token) }

func newToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

// SessionCookieHeader builds the cookie set at login.
func SessionCookieHeader(token string, expiresAtTime time.Time) *http.Cookie {
	return &http.Cookie{
		Name: SessionCookie, Value: token, Path: "/",
		Expires: expiresAtTime, HttpOnly: true, SameSite: http.SameSiteLaxMode,
	}
}

// ClearSessionCookieHeader clears the cookie.
func ClearSessionCookieHeader() *http.Cookie {
	return &http.Cookie{Name: SessionCookie, Value: "", Path: "/", MaxAge: -1, HttpOnly: true, SameSite: http.SameSiteLaxMode}
}
