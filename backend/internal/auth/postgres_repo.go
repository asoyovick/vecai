package auth

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"
)

// PostgresUserRepository persists users in the users table (migration 000003).
type PostgresUserRepository struct {
	db *sql.DB
}

// Compile-time check that the Postgres repository satisfies the service contract.
var _ UserRepository = (*PostgresUserRepository)(nil)

func NewPostgresUserRepository(db *sql.DB) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) CreateUser(ctx context.Context, user User) error {
	query := `
		INSERT INTO users (id, name, email, password_hash, role, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.ExecContext(ctx, query,
		user.ID, user.Fullname, user.Email, user.PasswordHash, string(user.Role), user.CreatedAt,
	)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" { // unique_violation on email
			return ErrUserAlreadyExists
		}
		return err
	}
	return nil
}

func (r *PostgresUserRepository) scanUser(row interface{ Scan(dest ...any) error }) (User, error) {
	user := User{}
	err := row.Scan(
		&user.ID, &user.Fullname, &user.Email, &user.PasswordHash,
		&user.Role, &user.Phone, &user.IsVerified, &user.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrUserNotFound
	}
	if err != nil {
		return User{}, err
	}
	return user, nil
}

func (r *PostgresUserRepository) GetUserByEmail(ctx context.Context, email string) (User, error) {
	query := `
		SELECT id, name, email, password_hash, role, COALESCE(phone, ''), is_verified, created_at
		FROM users
		WHERE email = $1
	`
	return r.scanUser(r.db.QueryRowContext(ctx, query, email))
}

func (r *PostgresUserRepository) GetUserByID(ctx context.Context, id string) (User, error) {
	query := `
		SELECT id, name, email, password_hash, role, COALESCE(phone, ''), is_verified, created_at
		FROM users
		WHERE id = $1
	`
	return r.scanUser(r.db.QueryRowContext(ctx, query, id))
}

func (r *PostgresUserRepository) UpdateProfile(ctx context.Context, id, fullname, phone string) error {
	query := `
		UPDATE users
		SET name = $2, phone = $3
		WHERE id = $1
	`
	result, err := r.db.ExecContext(ctx, query, id, fullname, phone)
	if err != nil {
		return err
	}
	if rows, err := result.RowsAffected(); err == nil && rows == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (r *PostgresUserRepository) UpdateVerificationStatus(ctx context.Context, email string, isVerified bool) error {
	query := `
		UPDATE users
		SET is_verified = $2
		WHERE email = $1
	`
	result, err := r.db.ExecContext(ctx, query, email, isVerified)
	if err != nil {
		return err
	}
	if rows, err := result.RowsAffected(); err == nil && rows == 0 {
		return ErrUserNotFound
	}
	return nil
}
