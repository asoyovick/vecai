package auth

import (
	"context"
	"errors"

	"github.com/gin-gonic/gin"
)

// UpdateProfileRequest is the payload for PUT /api/v1/users/me
type UpdateProfileRequest struct {
	Fullname string `json:"fullname" binding:"required"`
	Phone    string `json:"phone"`
}

// userIDContextKey stores the authenticated user's ID inside the Gin context.
type userIDContextKey struct{}

// WithUser injects an authenticated user ID into a context (used in tests).
func WithUser(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDContextKey{}, userID)
}

// UserIDFromContext extracts the authenticated user ID inside a Gin handler.
func UserIDFromContext(c *gin.Context) (string, bool) {
	if raw, ok := c.Get(string(UserIDContextKey)); ok {
		if id, ok := raw.(string); ok && id != "" {
			return id, true
		}
	}
	return "", false
}

// RoleFromContext extracts the authenticated user's role inside a Gin handler.
func RoleFromContext(c *gin.Context) (Role, bool) {
	if raw, ok := c.Get(string(UserRoleContextKey)); ok {
		if role, ok := raw.(Role); ok && role != "" {
			return role, true
		}
	}
	return "", false
}

var ErrMissingContext = errors.New("authentication context missing")
