package auth

import (
	"net/http"

	"vecai/internal/api"
)

// Middleware resolves the session cookie to a user on the request context.
func Middleware(svc *Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if c, err := r.Cookie(SessionCookie); err == nil && c.Value != "" {
				if u := userFromToken(svc, c.Value); u != nil {
					r = r.WithContext(WithUser(r.Context(), u))
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

func userFromToken(svc *Service, token string) *User {
	row, err := svc.repo.FindUserBySessionToken(tokenHash(token))
	if err != nil {
		return nil
	}
	return &row.User
}

// RequireUser rejects requests without a signed-in user.
func RequireUser(w http.ResponseWriter, r *http.Request) *User {
	u := UserFrom(r.Context())
	if u == nil {
		api.Error(w, 401, "You must be signed in.")
		return nil
	}
	return u
}

// RequireRole rejects requests whose user lacks one of the roles.
func RequireRole(w http.ResponseWriter, r *http.Request, roles ...string) *User {
	u := RequireUser(w, r)
	if u == nil {
		return nil
	}
	for _, role := range roles {
		if u.Role == role {
			return u
		}
	}
	api.Error(w, 403, "You do not have permission to perform this action.")
	return nil
}

// NotViewer blocks read-only accounts on mutation endpoints.
func NotViewer(w http.ResponseWriter, u *User, message string) bool {
	if u.Role == "viewer" {
		api.Error(w, 403, message)
		return false
	}
	return true
}
