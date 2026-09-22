// Package api provides cross-cutting HTTP concerns shared by all domain
// modules: the standard JSON response envelope defined in docs/api.md,
// CORS handling for the frontend, and small request-parsing helpers.
package api

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

// Envelope is the standard response format from docs/api.md:
//
//	{"success": true, "message": "...", "data": {...}}
type Envelope struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// OK writes a successful response with a message and optional payload.
func OK(c *gin.Context, status int, message string, data interface{}) {
	c.JSON(status, Envelope{Success: true, Message: message, Data: data})
}

// Fail writes an error response. The message is safe to show to users;
// err is logged server-side and never leaked to the client.
func Fail(c *gin.Context, status int, message string) {
	c.JSON(status, Envelope{Success: false, Error: message})
}

// BindJSON parses the request body into dst and writes a 400 response on failure.
// It reports whether binding succeeded.
func BindJSON(c *gin.Context, dst interface{}) bool {
	if err := c.ShouldBindJSON(dst); err != nil {
		Fail(c, http.StatusBadRequest, "invalid request body")
		return false
	}
	return true
}

// MapRepositoryError converts common database failures into the HTTP status
// codes defined in docs/api.md (409 for duplicates, 500 otherwise).
func MapRepositoryError(c *gin.Context, err error, conflictMessage string) {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "23505" { // unique_violation
		Fail(c, http.StatusConflict, conflictMessage)
		return
	}
	Fail(c, http.StatusInternalServerError, "internal server error")
}

// Trimmed returns the request context.
func Trimmed(s string) string {
	return strings.TrimSpace(s)
}
