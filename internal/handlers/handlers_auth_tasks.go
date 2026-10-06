package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"vecai/internal/auth"
	"vecai/internal/httpx"
	"vecai/internal/middleware"
)

func (s *Server) Login(w http.ResponseWriter, r *http.Request) {
	var body map[string]any
	if err := httpx.Decode(w, r, &body); err != nil {
		httpx.Error(w, err)
		return
	}
	email, err := httpx.AsString(body["email"], "Email", true, 200)
	if err != nil {
		httpx.Error(w, err)
		return
	}
	password, err := httpx.AsString(body["password"], "Password", true, 200)
	if err != nil {
		httpx.Error(w, err)
		return
	}
	row := s.DB.QueryRow(`SELECT id, email, password_hash, name, role, job_title, initials, color, phone FROM users WHERE email = ?`, email)
	var u auth.User
	var hash string
	err = row.Scan(&u.ID, &u.Email, &hash, &u.Name, &u.Role, &u.JobTitle, &u.Initials, &u.Color, &u.Phone)
	if errors.Is(err, sql.ErrNoRows) || err == nil && !auth.VerifyPassword(password, hash) {
		httpx.Error(w, httpx.Err(401, "Incorrect email or password."))
		return
	}
	if err != nil {
		httpx.Error(w, err)
		return
	}
	token, expiresAt, err := auth.CreateSession(s.DB, u.ID)
	if err != nil {
		httpx.Error(w, err)
		return
	}
	http.SetCookie(w, auth.SessionCookieHeader(token, expiresAt))
	httpx.JSON(w, 200, map[string]any{"user": u})
}

func (s *Server) Logout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(auth.SessionCookie); err == nil && c.Value != "" {
		auth.DestroySession(s.DB, c.Value)
	}
	http.SetCookie(w, auth.ClearSessionCookieHeader())
	httpx.JSON(w, 200, map[string]any{"ok": true})
}

func (s *Server) Me(w http.ResponseWriter, r *http.Request) {
	u := middleware.RequireUser(w, r)
	if u == nil {
		return
	}
	httpx.JSON(w, 200, map[string]any{"user": u})
}

var taskStatuses = []string{"not-started", "in-progress", "blocked", "completed"}
var taskPriorities = []string{"high", "medium", "low"}

const taskSelect = `SELECT t.*, p.name AS project_name, u.name AS assignee_name, u.initials AS assignee_initials, u.color AS assignee_color
	FROM tasks t JOIN projects p ON p.id = t.project_id LEFT JOIN users u ON u.id = t.assignee_id`

func (s *Server) ListTasks(w http.ResponseWriter, r *http.Request) {
	if middleware.RequireUser(w, r) == nil {
		return
	}
	q := taskSelect
	var args []any
	var conds []string
	if v := r.URL.Query().Get("projectId"); v != "" {
		conds = append(conds, "t.project_id = ?")
		args = append(args, v)
	}
	if v := r.URL.Query().Get("status"); v != "" && v != "All" {
		conds = append(conds, "t.status = ?")
		args = append(args, v)
	}
	if v := r.URL.Query().Get("q"); v != "" {
		conds = append(conds, "t.name LIKE ?")
		args = append(args, "%"+v+"%")
	}
	if len(conds) > 0 {
		q += " WHERE " + strings.Join(conds, " AND ")
	}
	q += " ORDER BY t.due_date IS NULL, t.due_date, t.id"
	rows, err := httpx.QueryMaps(s.DB, q, args...)
	if err != nil {
		httpx.Error(w, err)
		return
	}
	httpx.JSON(w, 200, map[string]any{"tasks": rows})
}

func (s *Server) CreateTask(w http.ResponseWriter, r *http.Request) {
	u := middleware.RequireUser(w, r)
	if u == nil || !middleware.NotViewer(w, u, "You do not have permission to create tasks.") {
		return
	}
	var body map[string]any
	if err := httpx.Decode(w, r, &body); err != nil {
		httpx.Error(w, err)
		return
	}
	name, err := httpx.AsString(body["name"], "Task name", true, 200)
	if err != nil {
		httpx.Error(w, err)
		return
	}
	description, _ := httpx.AsString(body["description"], "Description", false, 2000)
	projectID, err := httpx.AsNumber(body["projectId"], "Project", true, 0, true)
	if err != nil {
		httpx.Error(w, err)
		return
	}
	var assigneeID any
	if body["assigneeId"] != nil && body["assigneeId"] != "" {
		v, err := httpx.AsNumber(body["assigneeId"], "Assignee", false, 0, true)
		if err != nil {
			httpx.Error(w, err)
			return
		}
		assigneeID = v
	}
	startDate, err := httpx.AsDate(body["startDate"], "Start date", false)
	if err != nil {
		httpx.Error(w, err)
		return
	}
	dueDate, err := httpx.AsDate(body["dueDate"], "Due date", true)
	if err != nil {
		httpx.Error(w, err)
		return
	}
	status := "not-started"
	if body["status"] != nil {
		status, err = httpx.AsOneOf(body["status"], "Status", taskStatuses)
		if err != nil {
			httpx.Error(w, err)
			return
		}
	}
	priority := "medium"
	if body["priority"] != nil {
		priority, err = httpx.AsOneOf(body["priority"], "Priority", taskPriorities)
		if err != nil {
			httpx.Error(w, err)
			return
		}
	}
	var projectExists int
	if s.DB.QueryRow("SELECT COUNT(*) FROM projects WHERE id = ?", int64(projectID)).Scan(&projectExists); projectExists == 0 {
		httpx.Error(w, httpx.Err(400, "Select a valid project."))
		return
	}
	if assigneeID != nil {
		var c int
		s.DB.QueryRow("SELECT COUNT(*) FROM users WHERE id = ?", int64(assigneeID.(float64))).Scan(&c)
		if c == 0 {
			httpx.Error(w, httpx.Err(400, "Select a valid assignee."))
			return
		}
	}
	var sd, dd any
	if startDate != nil {
		sd = *startDate
	}
	dd = *dueDate
	res, err := s.DB.Exec(`INSERT INTO tasks (project_id, assignee_id, name, description, start_date, due_date, status, priority) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		int64(projectID), assigneeID, name, description, sd, dd, status, priority)
	if err != nil {
		httpx.Error(w, err)
		return
	}
	id, _ := res.LastInsertId()
	httpx.LogActivity(s.DB, int64(projectID), &httpx.Actor{ID: u.ID, Name: u.Name}, "task", "Created task “"+name+"”")
	httpx.TouchProject(s.DB, int64(projectID))
	row, err := httpx.QueryMap(s.DB, taskSelect+" WHERE t.id = ?", id)
	if err != nil {
		httpx.Error(w, err)
		return
	}
	httpx.JSON(w, 201, map[string]any{"task": row})
}

func (s *Server) PatchTask(w http.ResponseWriter, r *http.Request) {
	u := middleware.RequireUser(w, r)
	if u == nil || !middleware.NotViewer(w, u, "You do not have permission to update tasks.") {
		return
	}
	id := r.PathValue("id")
	var body map[string]any
	if err := httpx.Decode(w, r, &body); err != nil {
		httpx.Error(w, err)
		return
	}
	existing, err := httpx.QueryMap(s.DB, "SELECT * FROM tasks WHERE id = ?", id)
	if errors.Is(err, sql.ErrNoRows) {
		httpx.Error(w, httpx.Err(404, "Task not found."))
		return
	} else if err != nil {
		httpx.Error(w, err)
		return
	}
	var sets []string
	var args []any
	if v, ok := body["status"]; ok {
		s2, err := httpx.AsOneOf(v, "Status", taskStatuses)
		if err != nil {
			httpx.Error(w, err)
			return
		}
		sets = append(sets, "status = ?")
		args = append(args, s2)
	}
	if v, ok := body["priority"]; ok {
		p2, err := httpx.AsOneOf(v, "Priority", taskPriorities)
		if err != nil {
			httpx.Error(w, err)
			return
		}
		sets = append(sets, "priority = ?")
		args = append(args, p2)
	}
	if v, ok := body["name"]; ok {
		n, err := httpx.AsString(v, "Task name", true, 200)
		if err != nil {
			httpx.Error(w, err)
			return
		}
		sets = append(sets, "name = ?")
		args = append(args, n)
	}
	if v, ok := body["description"]; ok {
		d, _ := httpx.AsString(v, "Description", false, 2000)
		sets = append(sets, "description = ?")
		args = append(args, d)
	}
	if v, ok := body["assigneeId"]; ok {
		var assigneeID any
		if v != nil && v != "" {
			n, err := httpx.AsNumber(v, "Assignee", false, 0, true)
			if err != nil {
				httpx.Error(w, err)
				return
			}
			var c int
			s.DB.QueryRow("SELECT COUNT(*) FROM users WHERE id = ?", int64(n)).Scan(&c)
			if c == 0 {
				httpx.Error(w, httpx.Err(400, "Select a valid assignee."))
				return
			}
			assigneeID = n
		}
		sets = append(sets, "assignee_id = ?")
		args = append(args, assigneeID)
	}
	if v, ok := body["dueDate"]; ok {
		d, err := httpx.AsDate(v, "Due date", true)
		if err != nil {
			httpx.Error(w, err)
			return
		}
		sets = append(sets, "due_date = ?")
		args = append(args, *d)
	}
	if len(sets) == 0 {
		httpx.Error(w, httpx.Err(400, "No fields to update."))
		return
	}
	args = append(args, id)
	if _, err := s.DB.Exec("UPDATE tasks SET "+strings.Join(sets, ", ")+", updated_at = datetime('now') WHERE id = ?", args...); err != nil {
		httpx.Error(w, err)
		return
	}
	if v, ok := body["status"]; ok && v != existing["status"] {
		labels := map[string]string{"not-started": "Not started", "in-progress": "In progress", "blocked": "Blocked", "completed": "Completed"}
		httpx.LogActivity(s.DB, existing["project_id"], &httpx.Actor{ID: u.ID, Name: u.Name}, "task", "Marked “"+existing["name"].(string)+"” as "+labels[v.(string)])
	}
	httpx.TouchProject(s.DB, existing["project_id"].(int64))
	row, err := httpx.QueryMap(s.DB, taskSelect+" WHERE t.id = ?", id)
	if err != nil {
		httpx.Error(w, err)
		return
	}
	httpx.JSON(w, 200, map[string]any{"task": row})
}
