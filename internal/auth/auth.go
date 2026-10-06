// Package auth implements password hashing (scrypt, compatible with the
// historical format), session tokens and the session cookie helpers.
package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/hex"
	"fmt"
	"golang.org/x/crypto/scrypt"
	"net/http"
	"strings"
	"time"
)

const (
	SessionCookie = "vecai_session"
	sessionTTL    = 7 * 24 * time.Hour
)

type User struct {
	ID       int64  `json:"id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Role     string `json:"role"`
	JobTitle string `json:"jobTitle"`
	Initials string `json:"initials"`
	Color    string `json:"color"`
	Phone    string `json:"phone"`
}

func HashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	saltHex := hex.EncodeToString(salt)
	// NOTE: the salt used as scrypt input is the hex string's bytes,
	// matching the format produced since the app's first version.
	hash, err := scrypt.Key([]byte(password), []byte(saltHex), 16384, 8, 1, 64)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("scrypt:%s:%x", saltHex, hash), nil
}

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

func CreateSession(db *sql.DB, userID int64) (token string, expiresAt time.Time, err error) {
	buf := make([]byte, 32)
	if _, err = rand.Read(buf); err != nil {
		return "", time.Time{}, err
	}
	token = hex.EncodeToString(buf)
	expiresAt = time.Now().Add(sessionTTL)
	_, err = db.Exec("INSERT INTO sessions (token_hash, user_id, expires_at) VALUES (?, ?, ?)",
		tokenHash(token), userID, expiresAt.UTC().Format("2006-01-02 15:04:05"))
	if err != nil {
		return "", time.Time{}, err
	}
	return token, expiresAt, nil
}

func DestroySession(db *sql.DB, token string) {
	db.Exec("DELETE FROM sessions WHERE token_hash = ?", tokenHash(token))
}

func tokenHash(token string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(token)))
}

// SessionUser resolves the session cookie to a user, or nil.
func SessionUser(db *sql.DB, r *http.Request) *User {
	cookie, err := r.Cookie(SessionCookie)
	if err != nil || cookie.Value == "" {
		return nil
	}
	row := db.QueryRow(`SELECT u.id, u.email, u.name, u.role, u.job_title, u.initials, u.color, u.phone
		FROM sessions s JOIN users u ON u.id = s.user_id
		WHERE s.token_hash = ? AND s.expires_at > datetime('now')`, tokenHash(cookie.Value))
	var u User
	if err := row.Scan(&u.ID, &u.Email, &u.Name, &u.Role, &u.JobTitle, &u.Initials, &u.Color, &u.Phone); err != nil {
		return nil
	}
	return &u
}

func SessionCookieHeader(token string, expiresAt time.Time) *http.Cookie {
	return &http.Cookie{
		Name: SessionCookie, Value: token, Path: "/",
		Expires: expiresAt, HttpOnly: true, SameSite: http.SameSiteLaxMode,
	}
}

func ClearSessionCookieHeader() *http.Cookie {
	return &http.Cookie{Name: SessionCookie, Value: "", Path: "/", MaxAge: -1, HttpOnly: true, SameSite: http.SameSiteLaxMode}
}
