// Package newsletter handles email subscriptions submitted by the frontend
// blog "Get new posts by email" form.
package newsletter

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/mail"
	"strings"

	"github.com/gin-gonic/gin"

	"vecai/backend/internal/api"
)

// ErrInvalidEmail is returned when the submitted address is not parseable.
var ErrInvalidEmail = errors.New("invalid email address format")

// Repository persists newsletter subscriptions.
type Repository interface {
	Subscribe(ctx context.Context, email string) error
}

// PostgresRepository stores subscriptions in the newsletter_subscribers table.
type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// Subscribe inserts the email; duplicate subscriptions are idempotent.
func (r *PostgresRepository) Subscribe(ctx context.Context, email string) error {
	query := `
		INSERT INTO newsletter_subscribers (id, email, subscribed_at)
		VALUES (gen_random_uuid(), $1, NOW())
		ON CONFLICT (email) DO NOTHING
	`
	_, err := r.db.ExecContext(ctx, query, email)
	return err
}

type Handler struct {
	repo Repository
}

func NewHandler(repo Repository) *Handler {
	return &Handler{repo: repo}
}

// SubscribeRequest is the payload from the frontend newsletter form.
type SubscribeRequest struct {
	Email string `json:"email" binding:"required"`
}

// Subscribe handles POST /api/v1/newsletter/subscribe
func (h *Handler) Subscribe(c *gin.Context) {
	var req SubscribeRequest
	if !api.BindJSON(c, &req) {
		return
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	if _, err := mail.ParseAddress(email); err != nil {
		api.Fail(c, http.StatusBadRequest, ErrInvalidEmail.Error())
		return
	}

	if err := h.repo.Subscribe(c.Request.Context(), email); err != nil {
		api.Fail(c, http.StatusInternalServerError, "failed to subscribe")
		return
	}

	api.OK(c, http.StatusCreated, "Subscribed successfully", gin.H{"email": email})
}

// RegisterRoutes mounts the newsletter endpoints.
func (h *Handler) RegisterRoutes(api *gin.RouterGroup) {
	api.POST("/newsletter/subscribe", h.Subscribe)
}
