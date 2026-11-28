package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrAccountNotVerified = errors.New("account email verification is pending")
)

// UserRepository contract models what structural methods your service needs
type UserRepository interface {
	CreateUser(ctx context.Context, user User) error
	GetUserByEmail(ctx context.Context, email string) (User, error)
	GetUserByID(ctx context.Context, id string) (User, error)
	UpdateProfile(ctx context.Context, id, fullname, phone string) error
	UpdateVerificationStatus(ctx context.Context, email string, isVerified bool) error
	SetVerificationToken(ctx context.Context, email, tokenHash string, expiresAt time.Time) error
	GetVerificationToken(ctx context.Context, email string) (tokenHash string, expiresAt time.Time, err error)
	ClearVerificationToken(ctx context.Context, email string) error
}

type Service struct {
	repo UserRepository
}

func NewService(repo UserRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Register(ctx context.Context, req RegisterRequest) (AuthResponse, error) {
	req.Email = normalizeEmail(req.Email)
	req.Fullname = trimSpace(req.Fullname)

	if err := ValidateRegister(req); err != nil {
		return AuthResponse{}, err
	}

	// Passwords are hashed before storage so plaintext never touches the database.
	hash, err := HashPassword(req.Password)
	if err != nil {
		return AuthResponse{}, err
	}

	user := User{
		ID:           uuid.NewString(),
		Email:        req.Email,
		PasswordHash: hash,
		Fullname:     req.Fullname,
		Role:         req.Role,
		Phone:        req.Phone,
		IsVerified:   false,
		CreatedAt:    time.Now().UTC(),
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		return AuthResponse{}, err
	}

	rawToken, tokenHash, err := generateVerificationToken()
	if err != nil {
		return AuthResponse{}, err
	}
	expiresAt := time.Now().UTC().Add(24 * time.Hour)
	if err := s.repo.SetVerificationToken(ctx, user.Email, tokenHash, expiresAt); err != nil {
		return AuthResponse{}, err
	}

	// Real flow: send email. Default delivery logs token for dev/test environments.
	log.Printf("[AUTH] Verification token for %s: %s (expires %s)", user.Email, rawToken, expiresAt.Format(time.RFC3339))

	return AuthResponse{Token: "", User: user}, nil
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (AuthResponse, error) {
	req.Email = normalizeEmail(req.Email)

	if err := ValidateLogin(req); err != nil {
		return AuthResponse{}, err
	}

	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		// Same error for unknown email and wrong password so logins cannot be probed.
		return AuthResponse{}, ErrInvalidCredentials
	}

	if !CheckPasswordHash(req.Password, user.PasswordHash) {
		return AuthResponse{}, ErrInvalidCredentials
	}

	if !user.IsVerified {
		return AuthResponse{}, ErrAccountNotVerified
	}

	token, err := GenerateToken(user)
	if err != nil {
		return AuthResponse{}, err
	}

	return AuthResponse{Token: token, User: user}, nil
}

// Profile returns the currently authenticated user.
func (s *Service) Profile(ctx context.Context, userID string) (User, error) {
	return s.repo.GetUserByID(ctx, userID)
}

// UpdateProfile changes the editable fields of the authenticated user.
func (s *Service) UpdateProfile(ctx context.Context, userID, fullname, phone string) (User, error) {
	if trimSpace(fullname) == "" {
		return User{}, ErrFullNameEmpty
	}

	if err := s.repo.UpdateProfile(ctx, userID, trimSpace(fullname), trimSpace(phone)); err != nil {
		return User{}, err
	}

	return s.repo.GetUserByID(ctx, userID)
}

func (s *Service) Verify(ctx context.Context, req VerifyRequest) (AuthResponse, error) {
	email := normalizeEmail(req.Email)
	if email == "" || strings.TrimSpace(req.Token) == "" {
		return AuthResponse{}, errors.New("email and verification token are required")
	}

	hash, expiresAt, err := s.repo.GetVerificationToken(ctx, email)
	if err != nil {
		return AuthResponse{}, errors.New("invalid or expired verification code sequence")
	}
	if time.Now().UTC().After(expiresAt) {
		return AuthResponse{}, errors.New("invalid or expired verification code sequence")
	}
	if hashVerificationToken(req.Token) != hash {
		return AuthResponse{}, errors.New("invalid or expired verification code sequence")
	}

	if err := s.repo.UpdateVerificationStatus(ctx, email, true); err != nil {
		return AuthResponse{}, err
	}
	_ = s.repo.ClearVerificationToken(ctx, email)

	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return AuthResponse{}, err
	}

	token, err := GenerateToken(user)
	if err != nil {
		return AuthResponse{}, err
	}

	return AuthResponse{Token: token, User: user}, nil
}

func generateVerificationToken() (raw string, hash string, err error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", fmt.Errorf("generate verification token: %w", err)
	}
	raw = hex.EncodeToString(b)
	return raw, hashVerificationToken(raw), nil
}

func hashVerificationToken(raw string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(raw)))
	return hex.EncodeToString(sum[:])
}
