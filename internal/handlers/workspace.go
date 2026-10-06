package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	vrand "math/rand/v2"
	"net/http"
	"strings"

	"vecai/internal/auth"
	"vecai/internal/httpx"
	"vecai/internal/middleware"
)

var defaultPrefs = map[string]bool{
	"material-alerts": true, "task-reminders": true, "site-updates": true,
	"document-uploads": false, "budget-thresholds": true, "weekly-summary": true,
}

func initialsFor(name string) string {
	parts := strings.Fields(name)
	out := ""
	for i, p := range parts {
		if i >= 2 {
			break
		}
		if len(p) > 0 {
			out += strings.ToUpper(p[:1])
		}
	}
	if out == "" {
		out = "U"
	}
	return out
}

func (s *Server) Dashboard(w http.ResponseWriter, r *http.Request) {
	if middleware.RequireUser(w, r) == nil {
		return
	}
	projects, err := httpx.QueryMaps(s.DB, "SELECT * FROM projects ORDER BY created_at DESC")
	if err != nil {
		httpx.Error(w, err)
		return
	}
	if projects == nil {
		projects = []map[string]any{}
	}
	stats, err := httpx.QueryMap(s.DB, `SELECT
		(SELECT COUNT(*) FROM projects WHERE status != 'Completed') AS active_projects,
		(SELECT COUNT(*) FROM tasks WHERE status != 'completed' AND due_date IS NOT NULL AND due_date <= date('now', '+7 days')) AS tasks_due,
		(SELECT COUNT(*) FROM tasks WHERE status = 'blocked') AS blocked_tasks,
		(SELECT COUNT(*) FROM materials WHERE status IN ('critical','low')) AS material_alerts`)
	if err != nil {
		httpx.Error(w, err)
		return
	}
	materialAlerts, _ := httpx.QueryMaps(s.DB, `SELECT m.id, m.name, m.available_qty, m.required_qty, m.unit, m.status, p.name AS project_name
		FROM materials m JOIN projects p ON p.id = m.project_id WHERE m.status IN ('critical','low')
		ORDER BY CASE m.status WHEN 'critical' THEN 0 ELSE 1 END, m.name LIMIT 8`)
	if materialAlerts == nil {
		materialAlerts = []map[string]any{}
	}
	deadlines, _ := httpx.QueryMaps(s.DB, `SELECT t.id, t.name, t.due_date, t.status, p.name AS project_name
		FROM tasks t JOIN projects p ON p.id = t.project_id
		WHERE t.status != 'completed' AND t.due_date IS NOT NULL AND t.due_date <= date('now', '+14 days')
		ORDER BY t.due_date LIMIT 6`)
	if deadlines == nil {
		deadlines = []map[string]any{}
	}
	activity, _ := httpx.QueryMaps(s.DB, "SELECT * FROM activity_log ORDER BY created_at DESC, id DESC LIMIT 8")
	if activity == nil {
		activity = []map[string]any{}
	}
	httpx.JSON(w, 200, map[string]any{"stats": stats, "projects": projects, "materialAlerts": materialAlerts, "deadlines": deadlines, "activity": activity})
}

func (s *Server) Activity(w http.ResponseWriter, r *http.Request) {
	if middleware.RequireUser(w, r) == nil {
		return
	}
	var rows []map[string]any
	var err error
	if v := r.URL.Query().Get("projectId"); v != "" {
		rows, err = httpx.QueryMaps(s.DB, "SELECT * FROM activity_log WHERE project_id = ? ORDER BY created_at DESC, id DESC LIMIT 50", v)
	} else {
		rows, err = httpx.QueryMaps(s.DB, "SELECT * FROM activity_log ORDER BY created_at DESC, id DESC LIMIT 50")
	}
	if err != nil {
		httpx.Error(w, err)
		return
	}
	if rows == nil {
		rows = []map[string]any{}
	}
	httpx.JSON(w, 200, map[string]any{"activity": rows})
}

func (s *Server) Team(w http.ResponseWriter, r *http.Request) {
	if middleware.RequireUser(w, r) == nil {
		return
	}
	members, err := httpx.QueryMaps(s.DB, `SELECT u.id, u.name, u.email, u.role, u.job_title, u.initials, u.color, u.phone, u.created_at,
		(SELECT COUNT(*) FROM project_members pm WHERE pm.user_id = u.id) AS project_count
		FROM users u ORDER BY u.name`)
	if err != nil {
		httpx.Error(w, err)
		return
	}
	assignments, _ := httpx.QueryMaps(s.DB, `SELECT pm.user_id, p.id AS project_id, p.name AS project_name FROM project_members pm JOIN projects p ON p.id = pm.project_id`)
	if members == nil {
		members = []map[string]any{}
	}
	if assignments == nil {
		assignments = []map[string]any{}
	}
	httpx.JSON(w, 200, map[string]any{"members": members, "assignments": assignments})
}

func (s *Server) InviteMember(w http.ResponseWriter, r *http.Request) {
	u := middleware.RequireRole(w, r, "admin")
	if u == nil {
		return
	}
	var body map[string]any
	if err := httpx.Decode(w, r, &body); err != nil {
		httpx.Error(w, err)
		return
	}
	name, err := httpx.AsString(body["name"], "Full name", true, 120)
	if err != nil {
		httpx.Error(w, err)
		return
	}
	email, err := httpx.AsString(body["email"], "Email", true, 200)
	if err != nil {
		httpx.Error(w, err)
		return
	}
	email = strings.ToLower(email)
	jobTitle, err := httpx.AsString(body["jobTitle"], "Job title", true, 120)
	if err != nil {
		httpx.Error(w, err)
		return
	}
	role := "member"
	if body["role"] != nil {
		role, err = httpx.AsOneOf(body["role"], "Role", []string{"member", "viewer"})
		if err != nil {
			httpx.Error(w, err)
			return
		}
	}
	var projectID any
	if body["projectId"] != nil && body["projectId"] != "" {
		v, err := httpx.AsNumber(body["projectId"], "Project", false, 0, true)
		if err != nil {
			httpx.Error(w, err)
			return
		}
		projectID = int64(v)
	}
	phone, _ := httpx.AsString(body["phone"], "Phone", false, 40)
	if !httpx.EmailRe.MatchString(email) {
		httpx.Error(w, httpx.Err(400, "Enter a valid email address."))
		return
	}
	var c int
	s.DB.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", email).Scan(&c)
	if c > 0 {
		httpx.Error(w, httpx.Err(409, "A team member with that email already exists."))
		return
	}
	tokenBuf := make([]byte, 4)
	rand.Read(tokenBuf)
	tempPassword := "vecai-" + hex.EncodeToString(tokenBuf)
	colors := []string{"#1e3a5f", "#0891b2", "#7c3aed", "#db2777", "#059669", "#d97706", "#dc2626", "#0d9488"}
	color := colors[vrand.IntN(len(colors))]
	hash, err := auth.HashPassword(tempPassword)
	if err != nil {
		httpx.Error(w, err)
		return
	}
	res, err := s.DB.Exec(`INSERT INTO users (email, password_hash, name, role, job_title, initials, color, phone) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		email, hash, name, role, jobTitle, initialsFor(name), color, phone)
	if err != nil {
		httpx.Error(w, err)
		return
	}
	newID, _ := res.LastInsertId()
	if projectID != nil {
		s.DB.QueryRow("SELECT COUNT(*) FROM projects WHERE id = ?", projectID).Scan(&c)
		if c > 0 {
			s.DB.Exec("INSERT OR IGNORE INTO project_members (project_id, user_id) VALUES (?, ?)", projectID, newID)
			httpx.LogActivity(s.DB, projectID, &httpx.Actor{ID: u.ID, Name: u.Name}, "project", "Invited "+name+" to the project team")
		}
	}
	member, _ := httpx.QueryMap(s.DB, "SELECT id, name, email, role, job_title, initials, color, phone FROM users WHERE id = ?", newID)
	httpx.JSON(w, 201, map[string]any{"member": member, "temporaryPassword": tempPassword})
}

func (s *Server) ListSuppliers(w http.ResponseWriter, r *http.Request) {
	if middleware.RequireUser(w, r) == nil {
		return
	}
	q := "SELECT * FROM suppliers"
	var args []any
	var conds []string
	if v := r.URL.Query().Get("category"); v != "" && v != "All" {
		conds = append(conds, "category = ?")
		args = append(args, v)
	}
	if v := r.URL.Query().Get("q"); v != "" {
		conds = append(conds, "(name LIKE ? OR category LIKE ?)")
		args = append(args, "%"+v+"%", "%"+v+"%")
	}
	if len(conds) > 0 {
		q += " WHERE " + strings.Join(conds, " AND ")
	}
	q += " ORDER BY name"
	rows, err := httpx.QueryMaps(s.DB, q, args...)
	if err != nil {
		httpx.Error(w, err)
		return
	}
	if rows == nil {
		rows = []map[string]any{}
	}
	httpx.JSON(w, 200, map[string]any{"suppliers": rows})
}

func (s *Server) CreateSupplier(w http.ResponseWriter, r *http.Request) {
	u := middleware.RequireUser(w, r)
	if u == nil || !middleware.NotViewer(w, u, "You do not have permission to add suppliers.") {
		return
	}
	var body map[string]any
	if err := httpx.Decode(w, r, &body); err != nil {
		httpx.Error(w, err)
		return
	}
	name, err := httpx.AsString(body["name"], "Supplier name", true, 160)
	if err != nil {
		httpx.Error(w, err)
		return
	}
	category, err := httpx.AsString(body["category"], "Category", true, 80)
	if err != nil {
		httpx.Error(w, err)
		return
	}
	contact, _ := httpx.AsString(body["contact"], "Contact person", false, 120)
	phone, _ := httpx.AsString(body["phone"], "Phone", false, 40)
	email, _ := httpx.AsString(body["email"], "Email", false, 200)
	location, _ := httpx.AsString(body["location"], "Location", false, 160)
	rating, err := httpx.AsNumber(body["rating"], "Rating", false, 0, false)
	if err != nil {
		httpx.Error(w, err)
		return
	}
	if rating > 5 {
		httpx.Error(w, httpx.Err(400, "Rating must be between 0 and 5."))
		return
	}
	res, err := s.DB.Exec(`INSERT INTO suppliers (name, category, contact, phone, email, location, rating) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		name, category, contact, phone, email, location, rating)
	if err != nil {
		httpx.Error(w, err)
		return
	}
	id, _ := res.LastInsertId()
	row, err := httpx.QueryMap(s.DB, "SELECT * FROM suppliers WHERE id = ?", id)
	if err != nil {
		httpx.Error(w, err)
		return
	}
	httpx.JSON(w, 201, map[string]any{"supplier": row})
}

func (s *Server) UpdateMe(w http.ResponseWriter, r *http.Request) {
	u := middleware.RequireUser(w, r)
	if u == nil {
		return
	}
	var body map[string]any
	if err := httpx.Decode(w, r, &body); err != nil {
		httpx.Error(w, err)
		return
	}
	name, err := httpx.AsString(body["name"], "Name", true, 120)
	if err != nil {
		httpx.Error(w, err)
		return
	}
	email, err := httpx.AsString(body["email"], "Email", true, 200)
	if err != nil {
		httpx.Error(w, err)
		return
	}
	email = strings.ToLower(email)
	phone, _ := httpx.AsString(body["phone"], "Phone", false, 40)
	jobTitle, _ := httpx.AsString(body["jobTitle"], "Job title", false, 120)
	if !httpx.EmailRe.MatchString(email) {
		httpx.Error(w, httpx.Err(400, "Enter a valid email address."))
		return
	}
	var c int
	s.DB.QueryRow("SELECT COUNT(*) FROM users WHERE email = ? AND id != ?", email, u.ID).Scan(&c)
	if c > 0 {
		httpx.Error(w, httpx.Err(409, "That email address is already in use."))
		return
	}
	if _, err := s.DB.Exec("UPDATE users SET name = ?, email = ?, phone = ?, job_title = ?, initials = ?, updated_at = datetime('now') WHERE id = ?",
		name, email, phone, jobTitle, initialsFor(name), u.ID); err != nil {
		httpx.Error(w, err)
		return
	}
	row, err := httpx.QueryMap(s.DB, "SELECT id, email, name, role, job_title, initials, color, phone FROM users WHERE id = ?", u.ID)
	if err != nil {
		httpx.Error(w, err)
		return
	}
	httpx.JSON(w, 200, map[string]any{"user": map[string]any{
		"id": row["id"], "email": row["email"], "name": row["name"], "role": row["role"],
		"jobTitle": row["job_title"], "initials": row["initials"], "color": row["color"], "phone": row["phone"],
	}})
}

func (s *Server) ChangePassword(w http.ResponseWriter, r *http.Request) {
	u := middleware.RequireUser(w, r)
	if u == nil {
		return
	}
	var body map[string]any
	if err := httpx.Decode(w, r, &body); err != nil {
		httpx.Error(w, err)
		return
	}
	current, err := httpx.AsString(body["currentPassword"], "Current password", true, 0)
	if err != nil {
		httpx.Error(w, err)
		return
	}
	newPassword, err := httpx.AsString(body["newPassword"], "New password", true, 200)
	if err != nil {
		httpx.Error(w, err)
		return
	}
	if len(newPassword) < 8 {
		httpx.Error(w, httpx.Err(400, "New password must be at least 8 characters."))
		return
	}
	if !strings.ContainsAny(newPassword, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ") || !strings.ContainsAny(newPassword, "0123456789") {
		httpx.Error(w, httpx.Err(400, "New password must include at least one letter and one number."))
		return
	}
	var hash string
	if err := s.DB.QueryRow("SELECT password_hash FROM users WHERE id = ?", u.ID).Scan(&hash); err != nil {
		httpx.Error(w, err)
		return
	}
	if !auth.VerifyPassword(current, hash) {
		httpx.Error(w, httpx.Err(400, "Your current password is incorrect."))
		return
	}
	newHash, err := auth.HashPassword(newPassword)
	if err != nil {
		httpx.Error(w, err)
		return
	}
	s.DB.Exec("UPDATE users SET password_hash = ?, updated_at = datetime('now') WHERE id = ?", newHash, u.ID)
	s.DB.Exec("DELETE FROM sessions WHERE user_id = ?", u.ID)
	http.SetCookie(w, auth.ClearSessionCookieHeader())
	httpx.JSON(w, 200, map[string]any{"ok": true})
}

func (s *Server) GetPreferences(w http.ResponseWriter, r *http.Request) {
	u := middleware.RequireUser(w, r)
	if u == nil {
		return
	}
	prefs := map[string]bool{}
	for k, v := range defaultPrefs {
		prefs[k] = v
	}
	var raw string
	err := s.DB.QueryRow("SELECT prefs FROM user_preferences WHERE user_id = ?", u.ID).Scan(&raw)
	if err == nil {
		var stored map[string]bool
		if json.Unmarshal([]byte(raw), &stored) == nil {
			for k, v := range stored {
				prefs[k] = v
			}
		}
	}
	httpx.JSON(w, 200, map[string]any{"preferences": prefs})
}

func (s *Server) PatchPreferences(w http.ResponseWriter, r *http.Request) {
	u := middleware.RequireUser(w, r)
	if u == nil {
		return
	}
	var body map[string]any
	if err := httpx.Decode(w, r, &body); err != nil {
		httpx.Error(w, err)
		return
	}
	next := map[string]bool{}
	var raw string
	err := s.DB.QueryRow("SELECT prefs FROM user_preferences WHERE user_id = ?", u.ID).Scan(&raw)
	if err == nil {
		json.Unmarshal([]byte(raw), &next)
	}
	for key := range defaultPrefs {
		if v, ok := body[key].(bool); ok {
			next[key] = v
		}
	}
	data, _ := json.Marshal(next)
	s.DB.Exec(`INSERT INTO user_preferences (user_id, prefs) VALUES (?, ?) ON CONFLICT (user_id) DO UPDATE SET prefs = excluded.prefs`, u.ID, string(data))
	merged := map[string]bool{}
	for k, v := range defaultPrefs {
		merged[k] = v
	}
	for k, v := range next {
		merged[k] = v
	}
	httpx.JSON(w, 200, map[string]any{"preferences": merged})
}

func (s *Server) GetOrganisation(w http.ResponseWriter, r *http.Request) {
	if middleware.RequireUser(w, r) == nil {
		return
	}
	row, err := httpx.QueryMap(s.DB, "SELECT name, business_type, country, currency FROM organisation WHERE id = 1")
	if err != nil {
		httpx.Error(w, err)
		return
	}
	httpx.JSON(w, 200, map[string]any{"organisation": map[string]any{
		"name": row["name"], "businessType": row["business_type"], "country": row["country"], "currency": row["currency"],
	}})
}

func (s *Server) UpdateOrganisation(w http.ResponseWriter, r *http.Request) {
	if middleware.RequireRole(w, r, "admin") == nil {
		return
	}
	var body map[string]any
	if err := httpx.Decode(w, r, &body); err != nil {
		httpx.Error(w, err)
		return
	}
	name, err := httpx.AsString(body["name"], "Organisation name", true, 160)
	if err != nil {
		httpx.Error(w, err)
		return
	}
	businessType, err := httpx.AsString(body["businessType"], "Business type", true, 80)
	if err != nil {
		httpx.Error(w, err)
		return
	}
	country, err := httpx.AsString(body["country"], "Country", true, 80)
	if err != nil {
		httpx.Error(w, err)
		return
	}
	currency, err := httpx.AsString(body["currency"], "Currency", true, 80)
	if err != nil {
		httpx.Error(w, err)
		return
	}
	s.DB.Exec("UPDATE organisation SET name = ?, business_type = ?, country = ?, currency = ? WHERE id = 1", name, businessType, country, currency)
	httpx.JSON(w, 200, map[string]any{"organisation": map[string]string{"name": name, "businessType": businessType, "country": country, "currency": currency}})
}
