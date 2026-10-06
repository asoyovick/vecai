// Package handlers implements the VECAI REST API.
package handlers

import (
	"database/sql"

	"vecai/internal/config"
	"vecai/internal/httpx"
)

type Server struct {
	DB  *sql.DB
	Cfg *config.Config
}

func New(db *sql.DB, cfg *config.Config) *Server {
	return &Server{DB: db, Cfg: cfg}
}

var projectStatuses = []string{"Tendering", "Design", "Foundation", "Construction", "Roofing", "Finishing", "Handover", "Completed"}

func (s *Server) projectStats(projectID int64) map[string]any {
	var total int
	var completed, dueWeek, blocked sql.NullInt64
	s.DB.QueryRow(`SELECT COUNT(*),
		SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END),
		SUM(CASE WHEN status != 'completed' AND due_date IS NOT NULL AND due_date <= date('now', '+7 days') THEN 1 ELSE 0 END),
		SUM(CASE WHEN status = 'blocked' THEN 1 ELSE 0 END)
		FROM tasks WHERE project_id = ?`, projectID).Scan(&total, &completed, &dueWeek, &blocked)
	var alerts int
	s.DB.QueryRow("SELECT COUNT(*) FROM materials WHERE project_id = ? AND status IN ('critical','low')", projectID).Scan(&alerts)
	var recentDocs int
	s.DB.QueryRow("SELECT COUNT(*) FROM documents WHERE project_id = ? AND created_at >= datetime('now', '-14 days')", projectID).Scan(&recentDocs)
	return map[string]any{
		"tasksTotal": total, "tasksCompleted": completed.Int64, "tasksDueThisWeek": dueWeek.Int64,
		"tasksBlocked": blocked.Int64, "materialsAlert": alerts, "documentsRecent": recentDocs,
	}
}

func (s *Server) mapProject(row map[string]any) map[string]any {
	id := row["id"].(int64)
	team, _ := httpx.QueryMaps(s.DB, `SELECT u.id, u.name, u.role, u.job_title AS roleLabel, u.initials, u.color, u.email, u.phone
		FROM project_members pm JOIN users u ON u.id = pm.user_id WHERE pm.project_id = ? ORDER BY u.name`, id)
	if team == nil {
		team = []map[string]any{}
	}
	return map[string]any{
		"id": row["id"], "name": row["name"], "location": row["location"], "status": row["status"],
		"progress": row["progress"], "startDate": row["start_date"], "expectedCompletion": row["expected_completion"],
		"client": row["client"], "contractor": row["contractor"], "budgetTotal": row["budget_total"],
		"budgetSpent": row["budget_spent"], "description": row["description"], "imageUrl": row["image_url"],
		"createdAt": row["created_at"], "updatedAt": row["updated_at"],
		"stats": s.projectStats(id), "team": team,
	}
}
