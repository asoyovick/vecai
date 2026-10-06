package auth

import "net/http"

// Routes registers the auth endpoints on the mux under /api.
func (h *Handler) Routes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/auth/login", h.Login)
	mux.HandleFunc("POST /api/auth/register", h.Register)
	mux.HandleFunc("POST /api/auth/logout", h.Logout)
	mux.HandleFunc("GET /api/auth/me", h.Me)
	mux.HandleFunc("GET /api/team", h.ListTeam)
	mux.HandleFunc("POST /api/team/invite", h.InviteMember)
	mux.HandleFunc("PATCH /api/me", h.UpdateMe)
	mux.HandleFunc("POST /api/me/password", h.ChangePassword)
	mux.HandleFunc("GET /api/me/preferences", h.GetPreferences)
	mux.HandleFunc("PATCH /api/me/preferences", h.PatchPreferences)
	mux.HandleFunc("GET /api/organisation", h.GetOrganisation)
	mux.HandleFunc("PATCH /api/organisation", h.UpdateOrganisation)
}
