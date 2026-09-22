package auth

import (
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// ContextKey defines a custom type for Gin context keys.
type ContextKey string

const (
	UserIDContextKey   ContextKey = "user_id"
	UserRoleContextKey ContextKey = "user_role"
)

// SecretFromEnv reads the JWT signing secret from the environment,
// falling back to a development-only value when unset.
func SecretFromEnv() []byte {
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		return []byte(secret)
	}
	return []byte("vecai-dev-secret-change-me")
}

// AuthMiddleware validates JWTs and injects user information into the request context.
func AuthMiddleware(jwtSecret []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := parseBearerToken(c, jwtSecret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   err.Error(),
			})
			return
		}

		c.Set(string(UserIDContextKey), claims.ID)
		c.Set(string(UserRoleContextKey), claims.Role)

		c.Next()
	}
}

// OptionalAuthMiddleware attaches user identity when a valid token is present
// but lets anonymous requests through. Endpoints still enforce authentication
// themselves where it is required.
func OptionalAuthMiddleware(jwtSecret []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		if claims, err := parseBearerToken(c, jwtSecret); err == nil {
			c.Set(string(UserIDContextKey), claims.ID)
			c.Set(string(UserRoleContextKey), claims.Role)
		}
		c.Next()
	}
}

func parseBearerToken(c *gin.Context, jwtSecret []byte) (*CustomClaims, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return nil, errors.New("authorization header is required")
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return nil, errors.New("authorization header must be in the format: Bearer <token>")
	}

	claims := &CustomClaims{}
	token, err := jwt.ParseWithClaims(parts[1], claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid or expired token")
	}

	if claims.ID == "" || claims.Role == "" {
		return nil, errors.New("token payload is incomplete")
	}

	return claims, nil
}

// RequireRoles authorizes users based on their assigned roles.
func RequireRoles(allowedRoles ...Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, ok := RoleFromContext(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "authentication context missing",
			})
			return
		}

		for _, allowed := range allowedRoles {
			if role == allowed {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "access denied",
		})
	}
}
