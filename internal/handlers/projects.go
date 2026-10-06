package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"vecai/internal/httpx"
	"vecai/internal/middleware"
)

func (s *Server) ListProjects(w http.ResponseWriter, r *http.Request) {
	if middleware.RequireUser(w, r) == nil {
		return
	}
	rows, err := httpx.QueryMaps(s.DB, "SELECT * FROM projects ORDER BY created_at DESC")
	if err != nil {
		httpx.Error(w, err)
		return
	}
	out := []map[string]any{}
	for _, row := range rows {
		out = append(out, s.mapProject(row))
	}
	httpx.JSON(w, 200, map[string]any{"projects": out})
}

func (s *Server) CreateProject(w http.ResponseWriter, r *http.Request) {
	u := middleware.RequireRole(w, r, "admin")
	if u == nil {
		return
	}
	var body map[string]any
	if err := httpx.Decode(w, r, &body); err != nil {
		httpx.Error(w, err)
		return
	}
	name, err := httpx.AsString(body["name"], "Project name", true, 160)
	if err != nil {
		httpx.Error(w, err)
		return
	}
	location, _ := httpx.AsString(body["location"], "Location", false, 160)
	status := "Tendering"
	if body["status"] != nil {
		status, err = httpx.AsOneOf(body["status"], "Status", projectStatuses)
		if err != nil {
			httpx.Error(w, err)
			return
		}
	}
	startDate, err := httpx.AsDate(body["startDate"], "Start date", false)
	if err != nil {
		httpx.Error(w, err)
		return
	}
	completion, err := httpx.AsDate(body["expectedCompletion"], "Expected completion", false)
	if err != nil {
		httpx.Error(w, err)
		return
	}
	client, _ := httpx.AsString(body["client"], "Client", false, 160)
	contractor, _ := httpx.AsString(body["contractor"], "Contractor", false, 160)
	budgetTotal, err := httpx.AsNumber(body["budgetTotal"], "Total budget", false, 0, true)
	if err != nil {
		httpx.Error(w, err)
		return
	}
	description, _ := httpx.AsString(body["description"], "Description", false, 4000)
	var sd, cd any
	if startDate != nil {
		sd = *startDate
	}
	if completion != nil {
		cd = *completion
	}
	res, err := s.DB.Exec(`INSERT INTO projects (name, location, status, start_date, expected_completion, client, contractor, budget_total, description) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		name, location, status, sd, cd, client, contractor, int64(budgetTotal), description)
	if err != nil {
		httpx.Error(w, err)
		return
	}
	id, _ := res.LastInsertId()
	s.DB.Exec("INSERT INTO project_members (project_id, user_id) VALUES (?, ?)", id, u.ID)
	httpx.LogActivity(s.DB, id, &httpx.Actor{ID: u.ID, Name: u.Name}, "project", "Created project “"+name+"”")
	row, err := httpx.QueryMap(s.DB, "SELECT * FROM projects WHERE id = ?", id)
	if err != nil {
		httpx.Error(w, err)
		return
	}
	httpx.JSON(w, 201, map[string]any{"project": s.mapProject(row)})
}

func (s *Server) GetProject(w http.ResponseWriter, r *http.Request) {
	if middleware.RequireUser(w, r) == nil {
		return
	}
	id := r.PathValue("id")
	row, err := httpx.QueryMap(s.DB, "SELECT * FROM projects WHERE id = ?", id)
	if errors.Is(err, sql.ErrNoRows) {
		httpx.Error(w, httpx.Err(404, "Project not found."))
		return
	} else if err != nil {
		httpx.Error(w, err)
		return
	}
	tasks, _ := httpx.QueryMaps(s.DB, `SELECT t.*, u.name AS assignee_name, u.initials AS assignee_initials, u.color AS assignee_color
		FROM tasks t LEFT JOIN users u ON u.id = t.assignee_id WHERE t.project_id = ? ORDER BY t.due_date IS NULL, t.due_date, t.id`, id)
	materials, _ := httpx.QueryMaps(s.DB, `SELECT m.*, s.name AS supplier_name FROM materials m LEFT JOIN suppliers s ON s.id = m.supplier_id WHERE m.project_id = ? ORDER BY m.name`, id)
	documents, _ := httpx.QueryMaps(s.DB, `SELECT d.*, u.name AS uploaded_by_name FROM documents d LEFT JOIN users u ON u.id = d.uploaded_by WHERE d.project_id = ? ORDER BY d.created_at DESC`, id)
	activity, _ := httpx.QueryMaps(s.DB, "SELECT * FROM activity_log WHERE project_id = ? ORDER BY created_at DESC, id DESC LIMIT 30", id)
	budget, _ := httpx.QueryMaps(s.DB, "SELECT ym, planned, actual FROM budget_months WHERE project_id = ? ORDER BY ym", id)
	for _, arr := range [][]map[string]any{tasks, materials, documents, activity, budget} {
		if arr == nil {
			arr = []map[string]any{}
		}
	}
	if tasks == nil {
		tasks = []map[string]any{}
	}
	if materials == nil {
		materials = []map[string]any{}
	}
	if documents == nil {
		documents = []map[string]any{}
	}
	if activity == nil {
		activity = []map[string]any{}
	}
	if budget == nil {
		budget = []map[string]any{}
	}
	httpx.JSON(w, 200, map[string]any{
		"project": s.mapProject(row), "tasks": tasks, "materials": materials,
		"documents": documents, "activity": activity, "budget": budget,
	})
}

func (s *Server) PatchProject(w http.ResponseWriter, r *http.Request) {
	u := middleware.RequireRole(w, r, "admin")
	if u == nil {
		return
	}
	id := r.PathValue("id")
	existing, err := httpx.QueryMap(s.DB, "SELECT * FROM projects WHERE id = ?", id)
	if errors.Is(err, sql.ErrNoRows) {
		httpx.Error(w, httpx.Err(404, "Project not found."))
		return
	} else if err != nil {
		httpx.Error(w, err)
		return
	}
	var body map[string]any
	if err := httpx.Decode(w, r, &body); err != nil {
		httpx.Error(w, err)
		return
	}
	var sets []string
	var args []any
	if v, ok := body["name"]; ok {
		n, err := httpx.AsString(v, "Project name", true, 160)
		if err != nil {
			httpx.Error(w, err)
			return
		}
		sets = append(sets, "name = ?")
		args = append(args, n)
	}
	if v, ok := body["location"]; ok {
		l, _ := httpx.AsString(v, "Location", false, 160)
		sets = append(sets, "location = ?")
		args = append(args, l)
	}
	if v, ok := body["status"]; ok {
		st, err := httpx.AsOneOf(v, "Status", projectStatuses)
		if err != nil {
			httpx.Error(w, err)
			return
		}
		sets = append(sets, "status = ?")
		args = append(args, st)
	}
	if v, ok := body["progress"]; ok {
		n, err := httpx.AsNumber(v, "Progress", false, 0, true)
		if err != nil {
			httpx.Error(w, err)
			return
		}
		if n > 100 {
			httpx.Error(w, httpx.Err(400, "Progress must be at most 100."))
			return
		}
		sets = append(sets, "progress = ?")
		args = append(args, int64(n))
	}
	if v, ok := body["budgetSpent"]; ok {
		n, err := httpx.AsNumber(v, "Budget spent", false, 0, true)
		if err != nil {
			httpx.Error(w, err)
			return
		}
		sets = append(sets, "budget_spent = ?")
		args = append(args, int64(n))
	}
	if v, ok := body["description"]; ok {
		d, _ := httpx.AsString(v, "Description", false, 4000)
		sets = append(sets, "description = ?")
		args = append(args, d)
	}
	if len(sets) == 0 {
		httpx.Error(w, httpx.Err(400, "No fields to update."))
		return
	}
	args = append(args, id)
	if _, err := s.DB.Exec("UPDATE projects SET "+strings.Join(sets, ", ")+", updated_at = datetime('now') WHERE id = ?", args...); err != nil {
		httpx.Error(w, err)
		return
	}
	if v, ok := body["progress"]; ok {
		httpx.LogActivity(s.DB, existing["id"], &httpx.Actor{ID: u.ID, Name: u.Name}, "project", "Updated progress to "+toPlain(v)+"%")
	} else {
		httpx.LogActivity(s.DB, existing["id"], &httpx.Actor{ID: u.ID, Name: u.Name}, "project", "Updated project details")
	}
	httpx.TouchProject(s.DB, existing["id"].(int64))
	row, _ := httpx.QueryMap(s.DB, "SELECT * FROM projects WHERE id = ?", id)
	httpx.JSON(w, 200, map[string]any{"project": s.mapProject(row)})
}

func toPlain(v any) string {
	if f, ok := v.(float64); ok {
		return strconv.FormatFloat(f, 'f', -1, 64)
	}
	return ""
}

func (s *Server) ProjectBudget(w http.ResponseWriter, r *http.Request) {
	if middleware.RequireUser(w, r) == nil {
		return
	}
	rows, err := httpx.QueryMaps(s.DB, "SELECT ym, planned, actual FROM budget_months WHERE project_id = ? ORDER BY ym", r.PathValue("id"))
	if err != nil {
		httpx.Error(w, err)
		return
	}
	if rows == nil {
		rows = []map[string]any{}
	}
	httpx.JSON(w, 200, map[string]any{"budget": rows})
}
