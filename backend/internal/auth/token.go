package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// jwtSecret returns the signing secret shared by token generation and
// validation. It must match the secret used by the auth middleware, so both
// go through SecretFromEnv instead of separate hardcoded values.
func jwtSecret() ([]byte, error) {
	return SecretFromEnv()
}

// GenerateToken wraps user data and metadata into a signed string valid for 24 hours
func GenerateToken(user User) (string, error) {
	secret, err := jwtSecret()
	if err != nil {
		return "", err
	}

	claims := CustomClaims{
		ID:   user.ID,
		Role: user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.Email,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

// ValidateToken parses the incoming JWT and ensures the cryptographic integrity remains clean
func ValidateToken(tokenString string) (*CustomClaims, error) {
	secret, err := jwtSecret()
	if err != nil {
		return nil, err
	}

	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected token signing algorithm configuration")
		}
		return secret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid signature parsing outcome")
}
