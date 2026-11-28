package auth

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	ErrUserAlreadyExists = errors.New("user with this email address already exists")
	ErrUserNotFound      = errors.New("requested user account could not be found")
)

// UserRepository defines the strict interface contract for data persistence.
// Both your live SQL database implementation and this memory mock must implement this.

type verificationToken struct {
	hash      string
	expiresAt time.Time
}

// MemoryUserRepository provides a thread-safe, volatile in-memory storage engine.
type MemoryUserRepository struct {
	mu     sync.RWMutex
	users  map[string]User // Keyed by email address for O(1) lookups
	tokens map[string]verificationToken
}

// NewMemoryUserRepository instantiates the database-less mock repository layer.
func NewMemoryUserRepository() *MemoryUserRepository {
	return &MemoryUserRepository{
		users:  make(map[string]User),
		tokens: make(map[string]verificationToken),
	}
}

// CreateUser saves a new domain user model into the volatile map.
func (r *MemoryUserRepository) CreateUser(ctx context.Context, user User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.users[user.Email]; exists {
		return ErrUserAlreadyExists
	}

	r.users[user.Email] = user
	return nil
}

// GetUserByEmail searches for a user by their unique primary email field.
func (r *MemoryUserRepository) GetUserByEmail(ctx context.Context, email string) (User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.users[email]
	if !exists {
		return User{}, ErrUserNotFound
	}

	return user, nil
}

func (r *MemoryUserRepository) GetUserByID(ctx context.Context, id string) (User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, u := range r.users {
		if u.ID == id {
			return u, nil
		}
	}
	return User{}, ErrUserNotFound
}

func (r *MemoryUserRepository) UpdateProfile(ctx context.Context, id, fullname, phone string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for email, u := range r.users {
		if u.ID == id {
			u.Fullname = fullname
			u.Phone = phone
			r.users[email] = u
			return nil
		}
	}
	return ErrUserNotFound
}

// UpdateVerificationStatus flips the onboarding access flag for the target profile.
func (r *MemoryUserRepository) UpdateVerificationStatus(ctx context.Context, email string, isVerified bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	user, exists := r.users[email]
	if !exists {
		return ErrUserNotFound
	}

	user.IsVerified = isVerified
	r.users[email] = user
	return nil
}

func (r *MemoryUserRepository) SetVerificationToken(ctx context.Context, email, tokenHash string, expiresAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.users[email]; !exists {
		return ErrUserNotFound
	}
	r.tokens[email] = verificationToken{hash: tokenHash, expiresAt: expiresAt}
	return nil
}

func (r *MemoryUserRepository) GetVerificationToken(ctx context.Context, email string) (string, time.Time, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.tokens[email]
	if !ok {
		return "", time.Time{}, ErrUserNotFound
	}
	return t.hash, t.expiresAt, nil
}

func (r *MemoryUserRepository) ClearVerificationToken(ctx context.Context, email string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.tokens, email)
	return nil
}
