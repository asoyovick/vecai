package auth

import "github.com/gin-gonic/gin"

// RegisterRoutes mounts the public auth endpoints and the /users/me profile
// endpoints onto the versioned API group from docs/api.md.
func (h *Handler) RegisterRoutes(api *gin.RouterGroup) {
	authGroup := api.Group("/auth")
	{
		authGroup.POST("/register", h.Register)
		authGroup.POST("/login", h.Login)
		authGroup.POST("/verify", h.Verify)
		authGroup.POST("/logout", h.Logout)
	}

	users := api.Group("/users")
	{
		users.GET("/me", h.Me)
		users.PUT("/me", h.UpdateMe)
	}
}
