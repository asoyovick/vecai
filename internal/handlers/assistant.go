package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"vecai/internal/httpx"
	"vecai/internal/middleware"
)

type assistantLine struct {
	Text string `json:"text"`
	Tone string `json:"tone,omitempty"`
}

type assistantSection struct {
	Heading string          `json:"heading,omitempty"`
	Lines   []assistantLine `json:"lines"`
}

type assistantAnswer struct {
	Title    string             `json:"title"`
	Intro    string             `json:"intro"`
	Sections []assistantSection `json:"sections"`
	Footnote string             `json:"footnote"`
}

func daysBetween(from string, to time.Time) int {
	t, err := time.Parse("2006-01-02", from)
	if err != nil {
		return 1
	}
	d := int(to.Sub(t).Hours() / 24)
	if d < 1 {
		return 1
	}
	return d
}

func materialDaysLeft(available, used float64, startDate any, now time.Time) *int {
	sd, ok := startDate.(string)
	if !ok || sd == "" || used <= 0 {
		return nil
	}
	rate := used / float64(daysBetween(sd, now))
	if rate <= 0 {
		return nil
	}
	d := int(available / rate)
	return &d
}

func fmtQty(n float64) string {
	if n == float64(int64(n)) {
		return withCommas(int64(n))
	}
	if n < 1 {
		return strconv.FormatFloat(n, 'f', 2, 64)
	}
	return strconv.FormatFloat(n, 'f', 1, 64)
}

func withCommas(n int64) string {
	s := strconv.FormatInt(n, 10)
	if len(s) <= 3 {
		return s
	}
	var b strings.Builder
	for i, r := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			b.WriteByte(',')
		}
		b.WriteRune(r)
	}
	return b.String()
}

func KES(n float64) string { return withCommas(int64(n)) }

func (s *Server) materialsForecast(projectID *int64, now time.Time) assistantAnswer {
	q := `SELECT m.*, p.name AS project_name, p.start_date FROM materials m JOIN projects p ON p.id = m.project_id`
	args := []any{}
	if projectID != nil {
		q += ` WHERE m.project_id = ?`
		args = append(args, *projectID)
	}
	q += ` ORDER BY CASE m.status WHEN 'critical' THEN 0 WHEN 'low' THEN 1 WHEN 'ordered' THEN 2 ELSE 3 END, m.name`
	rows, _ := httpx.QueryMaps(s.DB, q, args...)
	var critical, low, ordered []map[string]any
	for _, m := range rows {
		switch m["status"] {
		case "critical":
			critical = append(critical, m)
		case "low":
			low = append(low, m)
		case "ordered":
			ordered = append(ordered, m)
		}
	}
	var sections []assistantSection
	if len(critical) > 0 {
		var lines []assistantLine
		for _, m := range critical {
			days := materialDaysLeft(m["available_qty"].(float64), toF(m["used_qty"]), m["start_date"], now)
			text := fmt.Sprintf("%v (%v) — %s %v remaining", m["name"], m["project_name"], fmtQty(m["available_qty"].(float64)), m["unit"])
			if days != nil {
				text += fmt.Sprintf(", roughly %d day%s of supply at the current usage rate", *days, plural(*days))
			}
			lines = append(lines, assistantLine{Text: text + ".", Tone: "danger"})
		}
		sections = append(sections, assistantSection{Heading: "Critical — reorder immediately", Lines: lines})
	}
	if len(low) > 0 {
		var lines []assistantLine
		for _, m := range low {
			days := materialDaysLeft(m["available_qty"].(float64), toF(m["used_qty"]), m["start_date"], now)
			text := fmt.Sprintf("%v (%v) — %s of %s %v", m["name"], m["project_name"], fmtQty(m["available_qty"].(float64)), fmtQty(m["required_qty"].(float64)), m["unit"])
			if days != nil {
				text += fmt.Sprintf(", roughly %d days of supply", *days)
			}
			lines = append(lines, assistantLine{Text: text + ".", Tone: "warn"})
		}
		sections = append(sections, assistantSection{Heading: "Low stock — order within the week", Lines: lines})
	}
	if len(ordered) > 0 {
		var lines []assistantLine
		for _, m := range ordered {
			lines = append(lines, assistantLine{Text: fmt.Sprintf("%v (%v) — %s %v expected; not yet consumed on site.", m["name"], m["project_name"], fmtQty(m["available_qty"].(float64)), m["unit"])})
		}
		sections = append(sections, assistantSection{Heading: "On order", Lines: lines})
	}
	if len(sections) == 0 {
		sections = append(sections, assistantSection{Lines: []assistantLine{{Text: "No materials are at critical or low stock levels right now.", Tone: "success"}}})
	}
	intro := "Computed across all projects from tracked stock levels."
	if projectID != nil {
		intro = "Computed from tracked stock levels and average consumption since each project started."
	}
	return assistantAnswer{Title: "Material Stock Forecast", Intro: intro, Sections: sections, Footnote: "Computed from live project data."}
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

func toF(v any) float64 {
	switch x := v.(type) {
	case float64:
		return x
	case int64:
		return float64(x)
	case json.Number:
		f, _ := x.Float64()
		return f
	}
	return 0
}

func (s *Server) tasksAtRisk(projectID *int64, now time.Time) assistantAnswer {
	today := now.Format("2006-01-02")
	soon := now.AddDate(0, 0, 7).Format("2006-01-02")
	where := ""
	args := []any{}
	if projectID != nil {
		where = " AND t.project_id = ?"
		args = append(args, *projectID)
	}
	blocked, _ := httpx.QueryMaps(s.DB, `SELECT t.name, t.due_date, p.name AS project_name, u.name AS assignee FROM tasks t JOIN projects p ON p.id = t.project_id LEFT JOIN users u ON u.id = t.assignee_id WHERE t.status = 'blocked'`+where+` ORDER BY t.due_date`, args...)
	overdue, _ := httpx.QueryMaps(s.DB, `SELECT t.name, t.due_date, p.name AS project_name FROM tasks t JOIN projects p ON p.id = t.project_id WHERE t.status != 'completed' AND t.status != 'blocked' AND t.due_date IS NOT NULL AND t.due_date < ?`+where+` ORDER BY t.due_date`, append([]any{today}, args...)...)
	dueSoon, _ := httpx.QueryMaps(s.DB, `SELECT t.name, t.due_date, p.name AS project_name FROM tasks t JOIN projects p ON p.id = t.project_id WHERE t.status = 'in-progress' AND t.due_date BETWEEN ? AND ?`+where+` ORDER BY t.due_date LIMIT 10`, append([]any{today, soon}, args...)...)
	var sections []assistantSection
	if len(blocked) > 0 {
		var lines []assistantLine
		for _, t := range blocked {
			text := fmt.Sprintf("%v (%v)", t["name"], t["project_name"])
			if t["assignee"] != nil {
				text += " — assigned to " + t["assignee"].(string)
			}
			if t["due_date"] != nil {
				text += " · was due " + t["due_date"].(string)
			}
			lines = append(lines, assistantLine{Text: text + ".", Tone: "danger"})
		}
		sections = append(sections, assistantSection{Heading: "Blocked — needs intervention", Lines: lines})
	}
	if len(overdue) > 0 {
		var lines []assistantLine
		for _, t := range overdue {
			lines = append(lines, assistantLine{Text: fmt.Sprintf("%v (%v) — due %v.", t["name"], t["project_name"], t["due_date"]), Tone: "warn"})
		}
		sections = append(sections, assistantSection{Heading: "Overdue", Lines: lines})
	}
	if len(dueSoon) > 0 {
		var lines []assistantLine
		for _, t := range dueSoon {
			lines = append(lines, assistantLine{Text: fmt.Sprintf("%v (%v) — due %v.", t["name"], t["project_name"], t["due_date"])})
		}
		sections = append(sections, assistantSection{Heading: "In progress — due within 7 days", Lines: lines})
	}
	if len(sections) == 0 {
		sections = append(sections, assistantSection{Lines: []assistantLine{{Text: "No blocked, overdue, or soon-due tasks found.", Tone: "success"}}})
	}
	return assistantAnswer{Title: "Tasks Behind Schedule or Blocked", Intro: "Live task status across the selected scope.", Sections: sections, Footnote: "Computed from live project data."}
}

func (s *Server) budgetVariance(projectID *int64) assistantAnswer {
	scope := ""
	args := []any{}
	if projectID != nil {
		scope = " WHERE bm.project_id = ?"
		args = append(args, *projectID)
	}
	months, _ := httpx.QueryMaps(s.DB, `SELECT bm.ym, SUM(bm.planned) AS planned, SUM(bm.actual) AS actual FROM budget_months bm`+scope+` GROUP BY bm.ym ORDER BY bm.ym`, args...)
	var totalPlanned, totalActual float64
	for _, m := range months {
		totalPlanned += toF(m["planned"])
		totalActual += toF(m["actual"])
	}
	variance := totalActual - totalPlanned
	pct := 0.0
	if totalPlanned != 0 {
		pct = variance / totalPlanned * 100
	}
	position := fmt.Sprintf("Under budget by KES %s (%d%%)", KES(abs(variance)), int64(abs(pct)))
	tone := "success"
	if variance > 0 {
		position = fmt.Sprintf("Over budget by KES %s (%d%%)", KES(variance), int64(pct))
		tone = "danger"
	}
	sections := []assistantSection{{
		Heading: "Position to date",
		Lines: []assistantLine{
			{Text: "Planned: KES " + KES(totalPlanned)},
			{Text: "Actual: KES " + KES(totalActual)},
			{Text: position, Tone: tone},
		},
	}}
	var variances []monthVariance
	for _, m := range months {
		variances = append(variances, monthVariance{fmt.Sprint(m["ym"]), toF(m["actual"]) - toF(m["planned"])})
	}
	sortByAbsDesc(variances)
	if len(variances) > 3 {
		variances = variances[:3]
	}
	if len(variances) > 0 {
		var lines []assistantLine
		for _, v := range variances {
			sign := ""
			t := "success"
			if v.variance > 0 {
				sign = "+"
				t = "warn"
			}
			lines = append(lines, assistantLine{Text: fmt.Sprintf("%s: %sKES %s against plan.", v.ym, sign, KES(v.variance)), Tone: t})
		}
		sections = append(sections, assistantSection{Heading: "Largest monthly variances", Lines: lines})
	}
	return assistantAnswer{Title: "Budget Variance Analysis", Intro: "Planned vs. actual monthly expenditure from the budget ledger.", Sections: sections, Footnote: "Computed from live project data."}
}

type monthVariance struct {
	ym       string
	variance float64
}

func abs(f float64) float64 {
	if f < 0 {
		return -f
	}
	return f
}

func sortByAbsDesc(vs []monthVariance) {
	for i := 0; i < len(vs); i++ {
		for j := i + 1; j < len(vs); j++ {
			if abs(vs[j].variance) > abs(vs[i].variance) {
				vs[i], vs[j] = vs[j], vs[i]
			}
		}
	}
}

func (s *Server) siteSummary(projectID *int64) assistantAnswer {
	where := ""
	args := []any{}
	if projectID != nil {
		where = " WHERE su.project_id = ?"
		args = append(args, *projectID)
	}
	rows, _ := httpx.QueryMaps(s.DB, `SELECT su.*, p.name AS project_name, u.name AS submitted_by FROM site_updates su JOIN projects p ON p.id = su.project_id LEFT JOIN users u ON u.id = su.submitted_by`+where+` ORDER BY su.date DESC, su.id DESC LIMIT 2`, args...)
	if len(rows) == 0 {
		return assistantAnswer{
			Title: "Site Report Summary", Intro: "No site updates have been submitted for this scope yet.",
			Sections: []assistantSection{{Lines: []assistantLine{{Text: "Submit a site update from the Site Update page to start building the daily record."}}}},
			Footnote: "Computed from live project data.",
		}
	}
	var totalWorkers float64
	for _, r := range rows {
		totalWorkers += toF(r["workers_on_site"])
	}
	avgWorkers := int64(totalWorkers / float64(len(rows)))
	var sections []assistantSection
	for _, su := range rows {
		weather := strOr(su["weather"], "not recorded")
		text := "Weather: " + weather
		if t, ok := su["temperature"].(string); ok && t != "" {
			text += " · " + t + "°C"
		}
		text += fmt.Sprintf(" · %v workers on site", su["workers_on_site"])
		lines := []assistantLine{{Text: text}}
		var segments []string
		for _, part := range strings.Split(fmt.Sprint(su["work_completed"]), ". ") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			if !strings.HasSuffix(part, ".") {
				part += "."
			}
			segments = append(segments, part)
			if len(segments) >= 4 {
				break
			}
		}
		for _, sgm := range segments {
			lines = append(lines, assistantLine{Text: sgm})
		}
		if issues, ok := su["issues"].(string); ok && issues != "" {
			lines = append(lines, assistantLine{Text: "Issues: " + issues, Tone: "warn"})
		}
		heading := fmt.Sprintf("%v — %v (submitted by %v)", su["project_name"], su["date"], strOr(su["submitted_by"], "unknown"))
		sections = append(sections, assistantSection{Heading: heading, Lines: lines})
	}
	return assistantAnswer{
		Title:    "Site Report Summary",
		Intro:    fmt.Sprintf("Latest %d update%s · average %d workers on site.", len(rows), plural(len(rows)), avgWorkers),
		Sections: sections,
		Footnote: "Computed from live project data.",
	}
}

func strOr(v any, fallback string) string {
	if s2, ok := v.(string); ok && s2 != "" {
		return s2
	}
	return fallback
}

func (s *Server) weeklyReport(now time.Time) assistantAnswer {
	projects, _ := httpx.QueryMaps(s.DB, `SELECT * FROM projects ORDER BY status != 'Completed' DESC, name`)
	today := now.Format("2006-01-02")
	var sections []assistantSection
	for _, p := range projects {
		var blocked, overdue, alerts int
		s.DB.QueryRow("SELECT COUNT(*) FROM tasks WHERE project_id = ? AND status = 'blocked'", p["id"]).Scan(&blocked)
		s.DB.QueryRow("SELECT COUNT(*) FROM tasks WHERE project_id = ? AND status != 'completed' AND due_date < ?", p["id"], today).Scan(&overdue)
		s.DB.QueryRow("SELECT COUNT(*) FROM materials WHERE project_id = ? AND status IN ('critical','low')", p["id"]).Scan(&alerts)
		spentPct := 0.0
		if toF(p["budget_total"]) != 0 {
			spentPct = toF(p["budget_spent"]) / toF(p["budget_total"]) * 100
		}
		lines := []assistantLine{{Text: fmt.Sprintf("%v · %v%% complete · budget %d%% spent (KES %.1fM of %.1fM)",
			p["status"], p["progress"], int64(spentPct), toF(p["budget_spent"])/1e6, toF(p["budget_total"])/1e6)}}
		if blocked > 0 {
			verb := "tasks need"
			if blocked == 1 {
				verb = "task needs"
			}
			lines = append(lines, assistantLine{Text: fmt.Sprintf("%d blocked %s attention.", blocked, verb), Tone: "danger"})
		}
		if overdue > 0 {
			lines = append(lines, assistantLine{Text: fmt.Sprintf("%d overdue task%s.", overdue, plural(overdue)), Tone: "warn"})
		}
		if alerts > 0 {
			lines = append(lines, assistantLine{Text: fmt.Sprintf("%d material alert%s (critical or low stock).", alerts, plural(alerts)), Tone: "warn"})
		}
		if blocked == 0 && overdue == 0 && alerts == 0 {
			lines = append(lines, assistantLine{Text: "No blocked tasks, overdue items, or stock alerts.", Tone: "success"})
		}
		sections = append(sections, assistantSection{Heading: fmt.Sprintf("%v — %v", p["name"], p["location"]), Lines: lines})
	}
	return assistantAnswer{
		Title:    "Weekly Project Status Report",
		Intro:    "All projects · week ending " + today + ".",
		Sections: sections,
		Footnote: "Computed from live project data.",
	}
}

func (s *Server) riskAnalysis(projectID *int64, now time.Time) assistantAnswer {
	if projectID == nil {
		return riskMissing()
	}
	project, err := httpx.QueryMap(s.DB, "SELECT * FROM projects WHERE id = ?", *projectID)
	if err != nil || project["start_date"] == nil || project["expected_completion"] == nil {
		return riskMissing()
	}
	startDate, _ := project["start_date"].(string)
	expected, _ := project["expected_completion"].(string)
	expEnd, _ := time.Parse("2006-01-02", expected)
	totalDays := daysBetween(startDate, expEnd)
	elapsed := daysBetween(startDate, now)
	elapsedPct := elapsed * 100 / totalDays
	if elapsedPct > 100 {
		elapsedPct = 100
	}
	progress := int64(toF(project["progress"]))
	var criticalMaterials, blocked, overdue int
	s.DB.QueryRow("SELECT COUNT(*) FROM materials WHERE project_id = ? AND status = 'critical'", *projectID).Scan(&criticalMaterials)
	s.DB.QueryRow("SELECT COUNT(*) FROM tasks WHERE project_id = ? AND status = 'blocked'", *projectID).Scan(&blocked)
	s.DB.QueryRow("SELECT COUNT(*) FROM tasks WHERE project_id = ? AND status != 'completed' AND due_date IS NOT NULL AND due_date < date('now')", *projectID).Scan(&overdue)
	var a, p *float64
	s.DB.QueryRow("SELECT SUM(actual), SUM(planned) FROM budget_months WHERE project_id = ?", *projectID).Scan(&a, &p)
	overBudget := a != nil && p != nil && *a > *p
	lines := []assistantLine{{Text: fmt.Sprintf("Programme: %d%% of the schedule elapsed, %d%% physical progress reported.", elapsedPct, progress)}}
	if progress < int64(elapsedPct)-10 {
		lines = append(lines, assistantLine{Text: fmt.Sprintf("Progress is trailing the programme by %d points — the completion date is at risk if the trend continues.", int64(elapsedPct)-progress), Tone: "warn"})
	} else {
		lines = append(lines, assistantLine{Text: "Physical progress is broadly in line with the programme.", Tone: "success"})
	}
	if criticalMaterials > 0 {
		lines = append(lines, assistantLine{Text: fmt.Sprintf("%d critical material%s could halt site work — reorder immediately.", criticalMaterials, plural(criticalMaterials)), Tone: "danger"})
	}
	if blocked > 0 {
		lines = append(lines, assistantLine{Text: fmt.Sprintf("%d blocked task%s need management intervention.", blocked, plural(blocked)), Tone: "danger"})
	}
	if overdue > 0 {
		taskWord := "tasks are"
		if overdue == 1 {
			taskWord = "task is"
		}
		lines = append(lines, assistantLine{Text: fmt.Sprintf("%d %s already past the due date.", overdue, taskWord), Tone: "warn"})
	}
	if overBudget {
		lines = append(lines, assistantLine{Text: "Actual expenditure is running ahead of the plan to date.", Tone: "warn"})
	}
	if criticalMaterials == 0 && blocked == 0 && overdue == 0 && !overBudget && progress >= int64(elapsedPct)-10 {
		lines = append(lines, assistantLine{Text: "No material schedule or cost risks detected in the current data.", Tone: "success"})
	}
	remaining := totalDays - elapsed
	if remaining < 0 {
		remaining = 0
	}
	return assistantAnswer{
		Title:    "Risk Analysis — " + fmt.Sprint(project["name"]),
		Intro:    fmt.Sprintf("Completion date %s · %d days remaining.", expected, remaining),
		Sections: []assistantSection{{Lines: lines}},
		Footnote: "Computed from live project data.",
	}
}

func riskMissing() assistantAnswer {
	return assistantAnswer{
		Title: "Project Risk Analysis", Intro: "Select a project with start and completion dates to run a schedule risk check.",
		Sections: []assistantSection{{Lines: []assistantLine{{Text: "No project selected, or the project has no programme dates recorded."}}}},
		Footnote: "Computed from live project data.",
	}
}

var (
	weeklyRe   = regexp.MustCompile(`weekly|status report|generate.*report|all projects`)
	siteSumRe  = regexp.MustCompile(`summar|site (report|update)|diary`)
	budgetRe   = regexp.MustCompile(`budget|spend|cost|variance|money|expenditure`)
	materialRe = regexp.MustCompile(`material|stock|run low|reorder|supply|deliver`)
	taskRe     = regexp.MustCompile(`task|blocked|behind|overdue|schedule`)
	riskRe     = regexp.MustCompile(`why|risk|delay|miss|completion date|on track`)
)

func (s *Server) AskAssistant(w http.ResponseWriter, r *http.Request) {
	if middleware.RequireUser(w, r) == nil {
		return
	}
	var body map[string]any
	if err := httpx.Decode(w, r, &body); err != nil {
		httpx.Error(w, err)
		return
	}
	question, err := httpx.AsString(body["question"], "Question", true, 500)
	if err != nil {
		httpx.Error(w, err)
		return
	}
	var projectID *int64
	if body["projectId"] != nil && body["projectId"] != "" {
		n, err := httpx.AsNumber(body["projectId"], "Project", false, 0, true)
		if err != nil {
			httpx.Error(w, err)
			return
		}
		v := int64(n)
		projectID = &v
	}
	now := time.Now()
	q := strings.ToLower(question)
	var answer assistantAnswer
	switch {
	case weeklyRe.MatchString(q):
		answer = s.weeklyReport(now)
	case siteSumRe.MatchString(q):
		answer = s.siteSummary(projectID)
	case budgetRe.MatchString(q):
		answer = s.budgetVariance(projectID)
	case materialRe.MatchString(q):
		answer = s.materialsForecast(projectID, now)
	case taskRe.MatchString(q):
		answer = s.tasksAtRisk(projectID, now)
	case riskRe.MatchString(q):
		answer = s.riskAnalysis(projectID, now)
	default:
		answer = s.weeklyReport(now)
	}
	httpx.JSON(w, 200, map[string]any{"answer": answer})
}
