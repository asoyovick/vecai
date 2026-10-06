package handlers

import (
	"net/http"
	"strings"

	"vecai/internal/httpx"
	"vecai/internal/middleware"
)

func (s *Server) ListSiteUpdates(w http.ResponseWriter, r *http.Request) {
	if middleware.RequireUser(w, r) == nil {
		return
	}
	q := `SELECT su.*, p.name AS project_name, u.name AS submitted_by_name
		FROM site_updates su JOIN projects p ON p.id = su.project_id LEFT JOIN users u ON u.id = su.submitted_by`
	var args []any
	var conds []string
	if v := r.URL.Query().Get("projectId"); v != "" && v != "All" {
		conds = append(conds, "su.project_id = ?")
		args = append(args, v)
	}
	if len(conds) > 0 {
		q += " WHERE " + strings.Join(conds, " AND ")
	}
	q += " ORDER BY su.date DESC, su.id DESC LIMIT 50"
	rows, err := httpx.QueryMaps(s.DB, q, args...)
	if err != nil {
		httpx.Error(w, err)
		return
	}
	out := []map[string]any{}
	for _, row := range rows {
		mats, _ := httpx.QueryMaps(s.DB, "SELECT name, quantity, unit FROM site_update_materials WHERE site_update_id = ?", row["id"])
		if mats == nil {
			mats = []map[string]any{}
		}
		row["materialsUsed"] = mats
		out = append(out, row)
	}
	httpx.JSON(w, 200, map[string]any{"siteUpdates": out})
}

func parseMaterials(raw any) ([]map[string]any, error) {
	arr, ok := raw.([]any)
	if !ok {
		return nil, nil
	}
	var out []map[string]any
	for _, item := range arr {
		entry, ok := item.(map[string]any)
		if !ok {
			continue
		}
		name, _ := entry["name"].(string)
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		quantity, err := httpx.AsNumber(entry["quantity"], "Quantity", false, 0, false)
		if err != nil || quantity <= 0 {
			return nil, httpx.Err(400, "Enter a quantity greater than zero for “"+name+"”.")
		}
		unit, _ := entry["unit"].(string)
		if len(unit) > 24 {
			unit = unit[:24]
		}
		if len(name) > 120 {
			name = name[:120]
		}
		out = append(out, map[string]any{"name": name, "quantity": quantity, "unit": strings.TrimSpace(unit)})
	}
	return out, nil
}

func (s *Server) CreateSiteUpdate(w http.ResponseWriter, r *http.Request) {
	u := middleware.RequireUser(w, r)
	if u == nil || !middleware.NotViewer(w, u, "You do not have permission to submit site updates.") {
		return
	}
	var body map[string]any
	if err := httpx.Decode(w, r, &body); err != nil {
		httpx.Error(w, err)
		return
	}
	projectID, err := httpx.AsNumber(body["projectId"], "Project", true, 0, true)
	if err != nil {
		httpx.Error(w, err)
		return
	}
	date, err := httpx.AsDate(body["date"], "Date", true)
	if err != nil {
		httpx.Error(w, err)
		return
	}
	weather, _ := httpx.AsString(body["weather"], "Weather", false, 60)
	temperature, _ := httpx.AsString(body["temperature"], "Temperature", false, 24)
	workers, err := httpx.AsNumber(body["workersOnSite"], "Workers on site", false, 0, true)
	if err != nil {
		httpx.Error(w, err)
		return
	}
	workCompleted, err := httpx.AsString(body["workCompleted"], "Work completed", true, 5000)
	if err != nil {
		httpx.Error(w, err)
		return
	}
	issues, _ := httpx.AsString(body["issues"], "Issues", false, 3000)
	notes, _ := httpx.AsString(body["notes"], "Notes", false, 3000)
	materials, err := parseMaterials(body["materialsUsed"])
	if err != nil {
		httpx.Error(w, err)
		return
	}
	var projectName string
	err = s.DB.QueryRow("SELECT name FROM projects WHERE id = ?", int64(projectID)).Scan(&projectName)
	if err != nil {
		httpx.Error(w, httpx.Err(400, "Select a valid project."))
		return
	}
	res, err := s.DB.Exec(`INSERT INTO site_updates (project_id, submitted_by, date, weather, temperature, workers_on_site, work_completed, issues, notes, photo_count) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 0)`,
		int64(projectID), u.ID, *date, weather, temperature, int64(workers), workCompleted, issues, notes)
	if err != nil {
		httpx.Error(w, err)
		return
	}
	updateID, _ := res.LastInsertId()
	for _, m := range materials {
		s.DB.Exec("INSERT INTO site_update_materials (site_update_id, name, quantity, unit) VALUES (?, ?, ?, ?)", updateID, m["name"], m["quantity"], m["unit"])
	}
	httpx.LogActivity(s.DB, int64(projectID), &httpx.Actor{ID: u.ID, Name: u.Name}, "site-update", "Submitted site update for "+projectName+" — "+*date)
	row, err := httpx.QueryMap(s.DB, `SELECT su.*, p.name AS project_name, u.name AS submitted_by_name FROM site_updates su JOIN projects p ON p.id = su.project_id LEFT JOIN users u ON u.id = su.submitted_by WHERE su.id = ?`, updateID)
	if err != nil {
		httpx.Error(w, err)
		return
	}
	if materials == nil {
		materials = []map[string]any{}
	}
	row["materialsUsed"] = materials
	httpx.JSON(w, 201, map[string]any{"siteUpdate": row})
}
