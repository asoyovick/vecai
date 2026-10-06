package reports

import (
	"net/http"

	"vecai/internal/api"
	"vecai/internal/auth"
)

// Handler serves the dashboard, activity, daily report and assistant endpoints.
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// Routes registers the report endpoints on the mux.
func (h *Handler) Routes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/dashboard", h.Dashboard)
	mux.HandleFunc("GET /api/activity", h.Activity)
	mux.HandleFunc("GET /api/site-updates", h.ListSiteUpdates)
	mux.HandleFunc("POST /api/site-updates", h.CreateSiteUpdate)
	mux.HandleFunc("POST /api/assistant", h.Ask)
}

func (h *Handler) Dashboard(w http.ResponseWriter, r *http.Request) {
	if auth.RequireUser(w, r) == nil {
		return
	}
	dashboard, err := h.svc.Dashboard()
	if err != nil {
		api.Error(w, 500, "Could not load the dashboard.")
		return
	}
	api.JSON(w, 200, dashboard)
}

func (h *Handler) Activity(w http.ResponseWriter, r *http.Request) {
	if auth.RequireUser(w, r) == nil {
		return
	}
	var projectID any
	if v := r.URL.Query().Get("projectId"); v != "" {
		projectID = v
	}
	activity, err := h.svc.Activity(projectID)
	if err != nil {
		api.Error(w, 500, "Could not load activity.")
		return
	}
	api.JSON(w, 200, map[string]any{"activity": activity})
}

func (h *Handler) ListSiteUpdates(w http.ResponseWriter, r *http.Request) {
	if auth.RequireUser(w, r) == nil {
		return
	}
	var projectID any
	if v := r.URL.Query().Get("projectId"); v != "" && v != "All" {
		projectID = v
	}
	rows, err := h.svc.ListDailyReports(projectID)
	if err != nil {
		api.Error(w, 500, "Could not load site updates.")
		return
	}
	api.JSON(w, 200, map[string]any{"siteUpdates": rows})
}

func (h *Handler) CreateSiteUpdate(w http.ResponseWriter, r *http.Request) {
	u := auth.RequireUser(w, r)
	if u == nil {
		return
	}
	if !auth.NotViewer(w, u, "You do not have permission to submit site updates.") {
		return
	}
	var body map[string]any
	if api.HandleDecodeError(w, api.Decode(w, r, &body)) {
		return
	}
	report, err := h.svc.CreateDailyReport(u, body)
	if err != nil {
		api.Error(w, 400, err.Error())
		return
	}
	api.JSON(w, 201, map[string]any{"siteUpdate": report})
}

func (h *Handler) Ask(w http.ResponseWriter, r *http.Request) {
	if auth.RequireUser(w, r) == nil {
		return
	}
	var body map[string]any
	if api.HandleDecodeError(w, api.Decode(w, r, &body)) {
		return
	}
	question, err := api.AsString(body["question"], "Question", true, 500)
	if err != nil {
		api.Error(w, 400, err.Error())
		return
	}
	var projectID *int64
	if body["projectId"] != nil && body["projectId"] != "" {
		v, err := api.AsNumber(body["projectId"], "Project", false, 0, true)
		if err != nil {
			api.Error(w, 400, err.Error())
			return
		}
		pv := int64(v)
		projectID = &pv
	}
	api.JSON(w, 200, map[string]any{"answer": h.svc.Answer(question, projectID)})
}
