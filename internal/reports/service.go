package reports

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"vecai/internal/api"
	"vecai/internal/auth"
)

// Service contains the dashboard, activity, daily report and
// intelligence business logic.
type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service { return &Service{repo: repo} }

// Dashboard assembles the workspace overview payload.
func (s *Service) Dashboard() (map[string]any, error) {
	projects, err := s.repo.Query("SELECT * FROM projects ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	if projects == nil {
		projects = []map[string]any{}
	}
	statsRow, _ := s.repo.Query(`SELECT
		(SELECT COUNT(*) FROM projects WHERE status != 'Completed') AS active_projects,
		(SELECT COUNT(*) FROM tasks WHERE status != 'completed' AND due_date IS NOT NULL AND due_date <= to_char((now() AT TIME ZONE 'UTC')::date + INTERVAL '7 days', 'YYYY-MM-DD')) AS tasks_due,
		(SELECT COUNT(*) FROM tasks WHERE status = 'blocked') AS blocked_tasks,
		(SELECT COUNT(*) FROM materials WHERE status IN ('critical','low')) AS material_alerts`)
	stats := map[string]any{}
	if len(statsRow) > 0 {
		stats = statsRow[0]
	}
	materialAlerts, _ := s.repo.Query(`SELECT m.id, m.name, m.available_qty, m.required_qty, m.unit, m.status, p.name AS project_name
		FROM materials m JOIN projects p ON p.id = m.project_id WHERE m.status IN ('critical','low')
		ORDER BY CASE m.status WHEN 'critical' THEN 0 ELSE 1 END, m.name LIMIT 8`)
	if materialAlerts == nil {
		materialAlerts = []map[string]any{}
	}
	deadlines, _ := s.repo.Query(`SELECT t.id, t.name, t.due_date, t.status, p.name AS project_name
		FROM tasks t JOIN projects p ON p.id = t.project_id
		WHERE t.status != 'completed' AND t.due_date IS NOT NULL AND t.due_date <= to_char((now() AT TIME ZONE 'UTC')::date + INTERVAL '14 days', 'YYYY-MM-DD')
		ORDER BY t.due_date LIMIT 6`)
	if deadlines == nil {
		deadlines = []map[string]any{}
	}
	activity, _ := s.repo.Query("SELECT * FROM activity_log ORDER BY created_at DESC, id DESC LIMIT 8")
	if activity == nil {
		activity = []map[string]any{}
	}
	return map[string]any{"stats": stats, "projects": projects, "materialAlerts": materialAlerts, "deadlines": deadlines, "activity": activity}, nil
}

// Activity returns the recent audit feed.
func (s *Service) Activity(projectID any) ([]map[string]any, error) {
	var rows []map[string]any
	var err error
	if projectID != nil {
		rows, err = s.repo.Query("SELECT * FROM activity_log WHERE project_id = $1 ORDER BY created_at DESC, id DESC LIMIT 50", projectID)
	} else {
		rows, err = s.repo.Query("SELECT * FROM activity_log ORDER BY created_at DESC, id DESC LIMIT 50")
	}
	if err != nil {
		return nil, err
	}
	if rows == nil {
		rows = []map[string]any{}
	}
	return rows, nil
}

// ListDailyReports returns site reports, newest first.
func (s *Service) ListDailyReports(projectID any) ([]map[string]any, error) {
	q := `SELECT su.*, p.name AS project_name, u.name AS submitted_by_name
		FROM daily_reports su JOIN projects p ON p.id = su.project_id LEFT JOIN users u ON u.id = su.submitted_by`
	args := []any{}
	if projectID != nil {
		q += " WHERE su.project_id = $1"
		args = append(args, projectID)
	}
	q += " ORDER BY su.date DESC, su.id DESC LIMIT 50"
	rows, err := s.repo.Query(q, args...)
	if err != nil {
		return nil, err
	}
	out := []map[string]any{}
	for _, row := range rows {
		mats, _ := s.repo.Query("SELECT name, quantity, unit FROM daily_report_materials WHERE daily_report_id = $1", row["id"])
		if mats == nil {
			mats = []map[string]any{}
		}
		row["materialsUsed"] = mats
		out = append(out, row)
	}
	return out, nil
}

// CreateDailyReport validates a new daily report and stores it.
func (s *Service) CreateDailyReport(u *auth.User, input map[string]any) (map[string]any, error) {
	if u.Role == "viewer" {
		return nil, errors.New("You do not have permission to submit site updates.")
	}
	projectID, err := api.AsNumber(input["projectId"], "Project", true, 0, true)
	if err != nil {
		return nil, err
	}
	date, err := api.AsDate(input["date"], "Date", true)
	if err != nil {
		return nil, err
	}
	weather, _ := api.AsString(input["weather"], "Weather", false, 60)
	temperature, _ := api.AsString(input["temperature"], "Temperature", false, 24)
	workers, err := api.AsNumber(input["workersOnSite"], "Workers on site", false, 0, true)
	if err != nil {
		return nil, err
	}
	workCompleted, err := api.AsString(input["workCompleted"], "Work completed", true, 5000)
	if err != nil {
		return nil, err
	}
	issues, _ := api.AsString(input["issues"], "Issues", false, 3000)
	notes, _ := api.AsString(input["notes"], "Notes", false, 3000)
	materials, err := parseMaterials(input["materialsUsed"])
	if err != nil {
		return nil, err
	}
	name, err := s.repo.ProjectName(int64(projectID))
	if err != nil {
		return nil, errors.New("Select a valid project.")
	}
	id, err := s.repo.CreateDailyReport(int64(projectID), u.ID, *date, weather, temperature, int64(workers), workCompleted, issues, notes)
	if err != nil {
		return nil, err
	}
	for _, m := range materials {
		s.repo.CreateDailyReportMaterial(id, fmt.Sprint(m["name"]), toF(m["quantity"]), fmt.Sprint(m["unit"]))
	}
	s.repo.LogActivity(int64(projectID), u.ID, u.Name, "site-update", "Submitted site update for "+name+" — "+*date)
	rows, err := s.repo.Query(`SELECT su.*, p.name AS project_name, u.name AS submitted_by_name
		FROM daily_reports su JOIN projects p ON p.id = su.project_id LEFT JOIN users u ON u.id = su.submitted_by WHERE su.id = $1`, id)
	if err != nil || len(rows) == 0 {
		return nil, errors.New("Could not load the report.")
	}
	if materials == nil {
		materials = []map[string]any{}
	}
	rows[0]["materialsUsed"] = materials
	return rows[0], nil
}

// Answer builds a computed answer to a plain-language question.
func (s *Service) Answer(question string, projectID *int64) assistantAnswer {
	now := time.Now()
	q := lower(question)
	switch {
	case weeklyRe.MatchString(q):
		return s.weeklyReport(now)
	case siteSumRe.MatchString(q):
		return s.siteSummary(projectID)
	case budgetRe.MatchString(q):
		return s.budgetVariance(projectID)
	case materialRe.MatchString(q):
		return s.materialsForecast(projectID, now)
	case taskRe.MatchString(q):
		return s.tasksAtRisk(projectID, now)
	case riskRe.MatchString(q):
		return s.riskAnalysis(projectID, now)
	default:
		return s.weeklyReport(now)
	}
}

func lower(s string) string { return strings.ToLower(s) }

var (
	weeklyRe   = regexp.MustCompile(`weekly|status report|generate.*report|all projects`)
	siteSumRe  = regexp.MustCompile(`summar|site (report|update)|diary`)
	budgetRe   = regexp.MustCompile(`budget|spend|cost|variance|money|expenditure`)
	materialRe = regexp.MustCompile(`material|stock|run low|reorder|supply|deliver`)
	taskRe     = regexp.MustCompile(`task|blocked|behind|overdue|schedule`)
	riskRe     = regexp.MustCompile(`why|risk|delay|miss|completion date|on track`)
)

// countScalar returns the first column of the first row of a COUNT query.
func (s *Service) countScalar(q string, args ...any) int {
	rows, _ := s.repo.Query(q, args...)
	if len(rows) == 0 {
		return 0
	}
	for _, v := range rows[0] {
		switch n := v.(type) {
		case int64:
			return int(n)
		case float64:
			return int(n)
		case string:
			i, _ := strconv.Atoi(n)
			return i
		}
		break
	}
	return 0
}

// budgetSum returns the planned/actual sum for a project across months.
func (s *Service) budgetSum(projectID int64) (*float64, *float64) {
	rows, err := s.repo.Query("SELECT SUM(actual) AS a, SUM(planned) AS p FROM budget_months WHERE project_id = $1", projectID)
	if err != nil || len(rows) == 0 {
		return nil, nil
	}
	var a, p *float64
	if rows[0]["a"] != nil {
		v := toF(rows[0]["a"])
		a = &v
	}
	if rows[0]["p"] != nil {
		v := toF(rows[0]["p"])
		p = &v
	}
	return a, p
}
