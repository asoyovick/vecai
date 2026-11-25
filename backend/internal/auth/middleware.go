package auth

import (
	"errors"
	"log"
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

const (
	minJWTSecretLength = 32
	devJWTSecret       = "vecai-dev-secret-change-me-please"
)

// SecretFromEnv reads the JWT signing secret from the environment.
// In production (APP_ENV=production), JWT_SECRET is required and must be at
// least 32 characters. In other environments, a development-only default is
// permitted with a loud warning.
func SecretFromEnv() ([]byte, error) {
	secret := strings.TrimSpace(os.Getenv("JWT_SECRET"))
	appEnv := strings.ToLower(strings.TrimSpace(os.Getenv("APP_ENV")))

	if appEnv == "production" {
		if secret == "" {
			return nil, errors.New("JWT_SECRET is required when APP_ENV=production")
		}
		if len(secret) < minJWTSecretLength {
			return nil, errors.New("JWT_SECRET must be at least 32 characters when APP_ENV=production")
		}
		return []byte(secret), nil
	}

	if secret == "" {
		log.Println("[WARN] JWT_SECRET is not set; using an insecure development default. Set JWT_SECRET before deploying.")
		return []byte(devJWTSecret), nil
	}
	if len(secret) < minJWTSecretLength {
		log.Println("[WARN] JWT_SECRET is shorter than 32 characters; use a stronger secret before deploying.")
	}
	return []byte(secret), nil
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
