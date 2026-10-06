// Package middleware provides session resolution, role enforcement and the
// API guards every request flows through.
package middleware

import (
	"context"
	"database/sql"
	"net/http"
	"strings"
	"time"

	"vecai/internal/auth"
	"vecai/internal/httpx"
)

type ctxKey string

const UserKey ctxKey = "user"

func UserFrom(ctx context.Context) *auth.User {
	u, _ := ctx.Value(UserKey).(*auth.User)
	return u
}

// Session attaches the current user (if signed in) to the request context.
func Session(db *sql.DB, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if u := auth.SessionUser(db, r); u != nil {
			r = r.WithContext(context.WithValue(r.Context(), UserKey, u))
		}
		next.ServeHTTP(w, r)
	})
}

func RequireUser(w http.ResponseWriter, r *http.Request) *auth.User {
	u := UserFrom(r.Context())
	if u == nil {
		httpx.JSON(w, 401, map[string]string{"error": "You must be signed in."})
		return nil
	}
	return u
}

func RequireRole(w http.ResponseWriter, r *http.Request, roles ...string) *auth.User {
	u := RequireUser(w, r)
	if u == nil {
		return nil
	}
	for _, role := range roles {
		if u.Role == role {
			return u
		}
	}
	httpx.JSON(w, 403, map[string]string{"error": "You do not have permission to perform this action."})
	return nil
}

// NotViewer rejects read-only accounts for mutation endpoints.
func NotViewer(w http.ResponseWriter, u *auth.User, message string) bool {
	if u.Role == "viewer" {
		httpx.JSON(w, 403, map[string]string{"error": message})
		return false
	}
	return true
}

// JSONGuard mirrors the original CSRF posture: mutations under /api must
// arrive as JSON or multipart from our own frontend.
func JSONGuard(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
		default:
			ct := r.Header.Get("Content-Type")
			isJSON := strings.Contains(ct, "application/json")
			isMultipart := strings.Contains(ct, "multipart/form-data")
			bodyless := ct == "" && r.ContentLength <= 0
			if !isJSON && !isMultipart && !bodyless {
				httpx.JSON(w, 415, map[string]string{"error": "Unsupported content type."})
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

// Logging emits one line per API request.
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		if strings.HasPrefix(r.URL.Path, "/api/") {
			println(r.Method, r.URL.Path, time.Since(start).Round(time.Millisecond).String())
		}
	})
}

// Recover converts panics into 500s instead of dropped connections.
func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				httpx.JSON(w, 500, map[string]string{"error": "Something went wrong on the server. Please try again."})
			}
		}()
		next.ServeHTTP(w, r)
	})
}
