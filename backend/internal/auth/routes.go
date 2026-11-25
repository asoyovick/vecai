package auth

import "github.com/gin-gonic/gin"

// RegisterRoutes mounts public auth endpoints and protected user endpoints.
func (h *Handler) RegisterRoutes(public, protected *gin.RouterGroup) {
	authGroup := public.Group("/auth")
	{
		authGroup.POST("/register", h.Register)
		authGroup.POST("/login", h.Login)
		authGroup.POST("/verify", h.Verify)
	}

	protectedAuth := protected.Group("/auth")
	{
		protectedAuth.POST("/logout", h.Logout)
	}

	users := protected.Group("/users")
	{
		users.GET("/me", h.Me)
		users.PUT("/me", h.UpdateMe)
	}
}
