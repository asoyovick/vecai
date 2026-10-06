package reports

// Repository is the persistence contract for dashboard, activity and
// daily-report data. Query runs read-only SQL authored by the service;
// it exists because the intelligence pages already compose SQL across
// domains (projects, tasks, materials, budgets, daily reports).
type Repository interface {
	Query(q string, args ...any) ([]map[string]any, error)
	QueryMap(q string, args ...any) (map[string]any, error)
	CreateDailyReport(projectID int64, submittedBy int64, date, weather, temperature string, workers int64, workCompleted, issues, notes string) (int64, error)
	CreateDailyReportMaterial(reportID int64, name string, quantity float64, unit string) error
	ProjectName(projectID int64) (string, error)
	LogActivity(projectID any, actorID any, name, typ, text string)
}
