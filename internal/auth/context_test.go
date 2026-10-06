package auth

import (
	"context"
	"testing"
)

func TestUserContextRoundTrip(t *testing.T) {
	ctx := WithUser(context.Background(), &User{ID: 7, Role: "admin"})
	if ctx == nil {
		t.Fatal("WithUser returned a nil context")
	}
	u := UserFrom(ctx)
	if u == nil || u.ID != 7 || u.Role != "admin" {
		t.Fatalf("UserFrom round trip failed: %+v", u)
	}
	if UserFrom(context.Background()) != nil {
		t.Fatal("UserFrom on a plain context returned a user")
	}
}
