// Package routes wires every endpoint onto the standard mux.
package routes

import (
	"net/http"

	"vecai/internal/handlers"
	"vecai/internal/middleware"
)

func New(s *handlers.Server) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/auth/login", s.Login)
	mux.HandleFunc("POST /api/auth/logout", s.Logout)
	mux.HandleFunc("GET /api/auth/me", s.Me)

	mux.HandleFunc("GET /api/projects", s.ListProjects)
	mux.HandleFunc("POST /api/projects", s.CreateProject)
	mux.HandleFunc("GET /api/projects/{id}", s.GetProject)
	mux.HandleFunc("PATCH /api/projects/{id}", s.PatchProject)
	mux.HandleFunc("GET /api/projects/{id}/budget", s.ProjectBudget)

	mux.HandleFunc("GET /api/tasks", s.ListTasks)
	mux.HandleFunc("POST /api/tasks", s.CreateTask)
	mux.HandleFunc("PATCH /api/tasks/{id}", s.PatchTask)

	mux.HandleFunc("GET /api/materials", s.ListMaterials)
	mux.HandleFunc("POST /api/materials", s.CreateMaterial)
	mux.HandleFunc("POST /api/materials/{id}/deliveries", s.LogDelivery)

	mux.HandleFunc("GET /api/documents", s.ListDocuments)
	mux.HandleFunc("POST /api/documents", s.UploadDocument)
	mux.HandleFunc("GET /api/documents/{id}/download", s.DownloadDocument)

	mux.HandleFunc("GET /api/site-updates", s.ListSiteUpdates)
	mux.HandleFunc("POST /api/site-updates", s.CreateSiteUpdate)

	mux.HandleFunc("GET /api/dashboard", s.Dashboard)
	mux.HandleFunc("GET /api/activity", s.Activity)
	mux.HandleFunc("GET /api/team", s.Team)
	mux.HandleFunc("POST /api/team/invite", s.InviteMember)
	mux.HandleFunc("GET /api/suppliers", s.ListSuppliers)
	mux.HandleFunc("POST /api/suppliers", s.CreateSupplier)
	mux.HandleFunc("PATCH /api/me", s.UpdateMe)
	mux.HandleFunc("POST /api/me/password", s.ChangePassword)
	mux.HandleFunc("GET /api/me/preferences", s.GetPreferences)
	mux.HandleFunc("PATCH /api/me/preferences", s.PatchPreferences)
	mux.HandleFunc("GET /api/organisation", s.GetOrganisation)
	mux.HandleFunc("PATCH /api/organisation", s.UpdateOrganisation)

	mux.HandleFunc("POST /api/assistant", s.AskAssistant)

	return middleware.Recover(middleware.Logging(middleware.Session(s.DB, middleware.JSONGuard(mux))))
}
