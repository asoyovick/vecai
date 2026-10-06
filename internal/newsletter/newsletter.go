// Package newsletter handles public newsletter subscriptions.
package newsletter

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/lib/pq"

	"vecai/internal/api"
)

type Handler struct {
	database *sql.DB
}

func NewHandler(database *sql.DB) *Handler { return &Handler{database: database} }

func (h *Handler) Routes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/newsletter/subscribe", h.Subscribe)
}

// Subscribe validates and stores a newsletter email.
func (h *Handler) Subscribe(w http.ResponseWriter, r *http.Request) {
	var body map[string]any
	if api.HandleDecodeError(w, api.Decode(w, r, &body)) {
		return
	}
	email, err := api.AsString(body["email"], "Email", true, 200)
	if err != nil {
		api.Error(w, 400, err.Error())
		return
	}
	if !api.EmailRe.MatchString(email) {
		api.Error(w, 400, "Enter a valid email address.")
		return
	}
	_, err = h.database.Exec(`INSERT INTO newsletter_subscribers (email) VALUES ($1)`, email)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			api.Error(w, 409, "That email is already subscribed.")
			return
		}
		api.Error(w, 500, "Could not save the subscription.")
		return
	}
	api.JSON(w, 201, map[string]any{"subscribed": true})
}
