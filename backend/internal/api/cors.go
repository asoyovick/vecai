package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CORSMiddleware allows the statically-exported Next.js frontend to call the
// API from a different origin. Allowed origins are configurable so production
// can be locked down without code changes.
func CORSMiddleware(allowedOrigins []string) gin.HandlerFunc {
	allowed := make(map[string]bool, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		allowed[origin] = true
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		// 1. Check if the incoming Origin header matches allowed origins
		if origin != "" && allowed[origin] {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Vary", "Origin")
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
			c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		}

		// 2. Handle preflight OPTIONS requests
		if c.Request.Method == "OPTIONS" {
			// If origin wasn't allowed or present, reject the preflight request
			if origin == "" || !allowed[origin] {
				c.AbortWithStatus(http.StatusForbidden)
				return
			}
			// Respond with 204 No Content (with CORS headers attached above)
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}