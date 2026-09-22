package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Role defines the explicit domain security classification strings
type Role string

const (
	RoleOwner            Role = "owner"
	RoleArchitect        Role = "architect"
	RoleQuantitySurveyor Role = "quantity_surveyor"
	RoleEngineer         Role = "engineer"
	RoleContractor       Role = "contractor"
	RoleSupplier         Role = "supplier"
)

// User represents the baseline system model stored in the users table.
type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // Hidden from JSON prints
	Fullname     string    `json:"fullname"`
	Role         Role      `json:"role"`
	Phone        string    `json:"phone,omitempty"`
	IsVerified   bool      `json:"is_verified"`
	CreatedAt    time.Time `json:"created_at"`
}

// CustomClaims packages validation contexts inside your signature tokens
type CustomClaims struct {
	ID   string `json:"id"`
	Role Role   `json:"role"`
	jwt.RegisteredClaims
}

// Request and Response JSON DTO structures
type RegisterRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
	Fullname string `json:"fullname" binding:"required"`
	Role     Role   `json:"role" binding:"required"`
	Phone    string `json:"phone"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type VerifyRequest struct {
	Email string `json:"email"`
	Token string `json:"token"`
}
