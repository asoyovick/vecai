package project

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"vecai/internal/api"
	"vecai/internal/auth"
)

// Handler serves the project and task endpoints.
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// Routes registers the project endpoints on the mux.
func (h *Handler) Routes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/projects", h.ListProjects)
	mux.HandleFunc("POST /api/projects", h.CreateProject)
	mux.HandleFunc("GET /api/projects/{id}", h.GetProject)
	mux.HandleFunc("PATCH /api/projects/{id}", h.PatchProject)
	mux.HandleFunc("GET /api/projects/{id}/budget", h.ProjectBudget)
	mux.HandleFunc("GET /api/tasks", h.ListTasks)
	mux.HandleFunc("POST /api/tasks", h.CreateTask)
	mux.HandleFunc("PATCH /api/tasks/{id}", h.PatchTask)
}

func (h *Handler) ListProjects(w http.ResponseWriter, r *http.Request) {
	if auth.RequireUser(w, r) == nil {
		return
	}
	projects, err := h.svc.ListProjects()
	if err != nil {
		api.Error(w, 500, "Could not load projects.")
		return
	}
	api.JSON(w, 200, map[string]any{"projects": projects})
}

func (h *Handler) CreateProject(w http.ResponseWriter, r *http.Request) {
	u := auth.RequireRole(w, r, "admin")
	if u == nil {
		return
	}
	var body map[string]any
	if api.HandleDecodeError(w, api.Decode(w, r, &body)) {
		return
	}
	project, err := h.svc.CreateProject(u, body)
	if err != nil {
		api.Error(w, 400, err.Error())
		return
	}
	api.JSON(w, 201, map[string]any{"project": project})
}

func (h *Handler) GetProject(w http.ResponseWriter, r *http.Request) {
	if auth.RequireUser(w, r) == nil {
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		api.Error(w, 400, "Invalid project id.")
		return
	}
	detail, err := h.svc.ProjectDetail(id)
	if err != nil {
		if errors.Is(err, ErrProjectNotFound) {
			api.Error(w, 404, err.Error())
			return
		}
		api.Error(w, 500, "Could not load the project.")
		return
	}
	api.JSON(w, 200, detail)
}

func (h *Handler) PatchProject(w http.ResponseWriter, r *http.Request) {
	u := auth.RequireRole(w, r, "admin")
	if u == nil {
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		api.Error(w, 400, "Invalid project id.")
		return
	}
	var body map[string]any
	if api.HandleDecodeError(w, api.Decode(w, r, &body)) {
		return
	}
	project, err := h.svc.PatchProject(u, id, body)
	if err != nil {
		switch {
		case errors.Is(err, ErrProjectNotFound):
			api.Error(w, 404, err.Error())
		case errors.Is(err, ErrNoFields):
			api.Error(w, 400, err.Error())
		default:
			api.Error(w, 400, err.Error())
		}
		return
	}
	api.JSON(w, 200, map[string]any{"project": project})
}

func (h *Handler) ProjectBudget(w http.ResponseWriter, r *http.Request) {
	if auth.RequireUser(w, r) == nil {
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		api.Error(w, 400, "Invalid project id.")
		return
	}
	budget, err := h.svc.Budget(id)
	if err != nil {
		api.Error(w, 500, "Could not load the budget.")
		return
	}
	api.JSON(w, 200, map[string]any{"budget": budget})
}

func (h *Handler) ListTasks(w http.ResponseWriter, r *http.Request) {
	if auth.RequireUser(w, r) == nil {
		return
	}
	var projectID, status, q any
	if v := r.URL.Query().Get("projectId"); v != "" {
		projectID = v
	}
	if v := r.URL.Query().Get("status"); v != "" && v != "All" {
		status = v
	}
	if v := r.URL.Query().Get("q"); v != "" {
		q = "%" + v + "%"
	}
	tasks, err := h.svc.ListTasks(projectID, status, q)
	if err != nil {
		api.Error(w, 500, "Could not load tasks.")
		return
	}
	api.JSON(w, 200, map[string]any{"tasks": tasks})
}

func (h *Handler) CreateTask(w http.ResponseWriter, r *http.Request) {
	u := auth.RequireUser(w, r)
	if u == nil {
		return
	}
	if !auth.NotViewer(w, u, "You do not have permission to create tasks.") {
		return
	}
	var body map[string]any
	if api.HandleDecodeError(w, api.Decode(w, r, &body)) {
		return
	}
	task, err := h.svc.CreateTask(u, body)
	if err != nil {
		api.Error(w, 400, err.Error())
		return
	}
	api.JSON(w, 201, map[string]any{"task": task})
}

func (h *Handler) PatchTask(w http.ResponseWriter, r *http.Request) {
	u := auth.RequireUser(w, r)
	if u == nil {
		return
	}
	if !auth.NotViewer(w, u, "You do not have permission to update tasks.") {
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		api.Error(w, 400, "Invalid task id.")
		return
	}
	var body map[string]any
	if api.HandleDecodeError(w, api.Decode(w, r, &body)) {
		return
	}
	task, err := h.svc.PatchTask(u, id, body)
	if err != nil {
		switch {
		case errors.Is(err, ErrTaskNotFound):
			api.Error(w, 404, err.Error())
		case errors.Is(err, ErrNoFields):
			api.Error(w, 400, err.Error())
		default:
			if strings.Contains(err.Error(), "permission") {
				api.Error(w, 403, err.Error())
				return
			}
			api.Error(w, 400, err.Error())
		}
		return
	}
	api.JSON(w, 200, map[string]any{"task": task})
}
