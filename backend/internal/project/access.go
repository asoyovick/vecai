package project

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"vecai/backend/internal/api"
	"vecai/backend/internal/auth"
)

const projectContextKey = "project"

// ProjectFromContext returns the authorized project loaded by RequireProjectAccess.
func ProjectFromContext(c *gin.Context) (*Project, bool) {
	v, ok := c.Get(projectContextKey)
	if !ok {
		return nil, false
	}
	p, ok := v.(*Project)
	return p, ok
}

// RequireProjectAccess ensures the authenticated user owns or is a member of the
// project identified by the :id route parameter.
func RequireProjectAccess(service *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := auth.UserIDFromContext(c)
		if !ok {
			api.Fail(c, http.StatusUnauthorized, "authentication required")
			c.Abort()
			return
		}

		projectID := c.Param("id")
		p, err := service.AuthorizeProjectAccess(c.Request.Context(), projectID, userID)
		if err != nil {
			switch {
			case errors.Is(err, ErrProjectNotFound):
				api.Fail(c, http.StatusNotFound, "project not found")
			case errors.Is(err, ErrAccessDenied):
				api.Fail(c, http.StatusForbidden, "access denied")
			default:
				api.Fail(c, http.StatusInternalServerError, "internal server error")
			}
			c.Abort()
			return
		}

		c.Set(projectContextKey, p)
		c.Next()
	}
}

// RequireProjectOwner restricts an action to the project owner.
func RequireProjectOwner() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := auth.UserIDFromContext(c)
		if !ok {
			api.Fail(c, http.StatusUnauthorized, "authentication required")
			c.Abort()
			return
		}
		p, ok := ProjectFromContext(c)
		if !ok {
			api.Fail(c, http.StatusInternalServerError, "project context missing")
			c.Abort()
			return
		}
		if p.OwnerID != userID {
			api.Fail(c, http.StatusForbidden, "only the project owner can perform this action")
			c.Abort()
			return
		}
		c.Next()
	}
}
