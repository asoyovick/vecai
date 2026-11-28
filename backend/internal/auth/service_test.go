package auth

import (
	"context"
	"errors"
	"testing"
)

type fakeUserRepo struct {
	users map[string]User
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{users: map[string]User{}}
}

func (f *fakeUserRepo) CreateUser(_ context.Context, user User) error {
	if _, exists := f.users[user.Email]; exists {
		return errors.New("duplicate email")
	}
	f.users[user.Email] = user
	return nil
}
func (f *fakeUserRepo) GetUserByEmail(_ context.Context, email string) (User, error) {
	u, ok := f.users[email]
	if !ok {
		return User{}, errors.New("not found")
	}
	return u, nil
}
func (f *fakeUserRepo) GetUserByID(_ context.Context, id string) (User, error) {
	for _, u := range f.users {
		if u.ID == id {
			return u, nil
		}
	}
	return User{}, errors.New("not found")
}
func (f *fakeUserRepo) UpdateProfile(_ context.Context, id, fullname, phone string) error {
	for email, u := range f.users {
		if u.ID == id {
			u.Fullname = fullname
			u.Phone = phone
			f.users[email] = u
			return nil
		}
	}
	return errors.New("not found")
}
func (f *fakeUserRepo) UpdateVerificationStatus(_ context.Context, email string, isVerified bool) error {
	u, ok := f.users[email]
	if !ok {
		return errors.New("not found")
	}
	u.IsVerified = isVerified
	f.users[email] = u
	return nil
}

func TestValidateRegister(t *testing.T) {
	tests := []struct {
		name    string
		req     RegisterRequest
		wantErr error
	}{
		{name: "valid", req: RegisterRequest{Email: "a@b.com", Password: "password123", Fullname: "A B", Role: RoleOwner}},
		{name: "bad email", req: RegisterRequest{Email: "bad", Password: "password123", Fullname: "A B", Role: RoleOwner}, wantErr: ErrInvalidEmail},
		{name: "weak password", req: RegisterRequest{Email: "a@b.com", Password: "short", Fullname: "A B", Role: RoleOwner}, wantErr: ErrPasswordTooWeak},
		{name: "missing fullname", req: RegisterRequest{Email: "a@b.com", Password: "password123", Fullname: " ", Role: RoleOwner}, wantErr: ErrFullNameEmpty},
		{name: "bad role", req: RegisterRequest{Email: "a@b.com", Password: "password123", Fullname: "A B", Role: Role("bad")}, wantErr: ErrInvalidRole},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRegister(tt.req)
			if tt.wantErr == nil && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestLogin_WrongPasswordAndUnknownEmailReturnSameError(t *testing.T) {
	repo := newFakeUserRepo()
	svc := NewService(repo)
	_, err := svc.Register(context.Background(), RegisterRequest{
		Email: "known@example.com", Password: "password123", Fullname: "Known User", Role: RoleOwner,
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	_, errUnknown := svc.Login(context.Background(), LoginRequest{Email: "unknown@example.com", Password: "password123"})
	_, errWrong := svc.Login(context.Background(), LoginRequest{Email: "known@example.com", Password: "wrongpassword"})

	if !errors.Is(errUnknown, ErrInvalidCredentials) {
		t.Fatalf("unknown email: expected ErrInvalidCredentials, got %v", errUnknown)
	}
	if !errors.Is(errWrong, ErrInvalidCredentials) {
		t.Fatalf("wrong password: expected ErrInvalidCredentials, got %v", errWrong)
	}
	if errUnknown.Error() != errWrong.Error() {
		t.Fatalf("expected same error message, got %q vs %q", errUnknown.Error(), errWrong.Error())
	}
}

func TestTokenGenerationAndValidation(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	user := User{ID: "user-1", Email: "u@example.com", Role: RoleOwner}
	token, err := GenerateToken(user)
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}
	claims, err := ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken: %v", err)
	}
	if claims.ID != user.ID || claims.Role != user.Role {
		t.Fatalf("unexpected claims: %#v", claims)
	}
	if claims.Subject != user.Email {
		t.Fatalf("unexpected subject: %q", claims.Subject)
	}

	_, err = ValidateToken(token + "tampered")
	if err == nil {
		t.Fatal("expected error for tampered token")
	}
}
