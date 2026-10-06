package auth

import (
	"context"
)

type ctxKey string

const userKey ctxKey = "user"

// UserFrom returns the authenticated user carried by the context.
func UserFrom(ctx context.Context) *User {
	u, _ := ctx.Value(userKey).(*User)
	return u
}

// WithUser stores the user on the context.
func WithUser(ctx context.Context, u *User) context.Context {
	return context.WithValue(ctx, userKey, u)
}
