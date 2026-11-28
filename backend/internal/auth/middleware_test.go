package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := []byte("test-secret")
	t.Setenv("JWT_SECRET", "test-secret")

	user := User{ID: "u1", Email: "u@example.com", Role: RoleOwner}
	token, err := GenerateToken(user)
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}

	tests := []struct {
		name       string
		authHeader string
		wantStatus int
	}{
		{name: "valid token", authHeader: "Bearer " + token, wantStatus: http.StatusOK},
		{name: "missing header", authHeader: "", wantStatus: http.StatusUnauthorized},
		{name: "bad format", authHeader: token, wantStatus: http.StatusUnauthorized},
		{name: "invalid token", authHeader: "Bearer bad.token.value", wantStatus: http.StatusUnauthorized},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			r.GET("/protected", AuthMiddleware(secret), func(c *gin.Context) {
				id, _ := UserIDFromContext(c)
				c.JSON(http.StatusOK, gin.H{"id": id})
			})
			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d, body=%s", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

func TestRequireRoles(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := []byte("test-secret")
	t.Setenv("JWT_SECRET", "test-secret")

	ownerToken, _ := GenerateToken(User{ID: "u1", Email: "o@example.com", Role: RoleOwner})
	engineerToken, _ := GenerateToken(User{ID: "u2", Email: "e@example.com", Role: RoleEngineer})

	r := gin.New()
	r.GET("/admin", AuthMiddleware(secret), RequireRoles(RoleOwner), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	reqOwner := httptest.NewRequest(http.MethodGet, "/admin", nil)
	reqOwner.Header.Set("Authorization", "Bearer "+ownerToken)
	wOwner := httptest.NewRecorder()
	r.ServeHTTP(wOwner, reqOwner)
	if wOwner.Code != http.StatusOK {
		t.Fatalf("owner status = %d", wOwner.Code)
	}

	reqEng := httptest.NewRequest(http.MethodGet, "/admin", nil)
	reqEng.Header.Set("Authorization", "Bearer "+engineerToken)
	wEng := httptest.NewRecorder()
	r.ServeHTTP(wEng, reqEng)
	if wEng.Code != http.StatusForbidden {
		t.Fatalf("engineer status = %d, want 403", wEng.Code)
	}
}
