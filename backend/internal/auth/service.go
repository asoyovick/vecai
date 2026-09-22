package auth

import (
	"context"
	"errors"
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
		IsVerified:   true, // MVP: verification flow lands in a later sprint.
		CreatedAt:    time.Now().UTC(),
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		return AuthResponse{}, err
	}

	token, err := GenerateToken(user)
	if err != nil {
		return AuthResponse{}, err
	}

	return AuthResponse{Token: token, User: user}, nil
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
	// Simple mock verification loop check: we'll automatically pass if token is "123456"
	if req.Token != "123456" {
		return AuthResponse{}, errors.New("invalid or expired verification code sequence")
	}

	if err := s.repo.UpdateVerificationStatus(ctx, req.Email, true); err != nil {
		return AuthResponse{}, err
	}

	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return AuthResponse{}, err
	}

	token, err := GenerateToken(user)
	if err != nil {
		return AuthResponse{}, err
	}

	return AuthResponse{Token: token, User: user}, nil
}
