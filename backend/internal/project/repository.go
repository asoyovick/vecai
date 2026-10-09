package project

// Repository is the persistence contract for projects and tasks.
type Repository interface {
	ListProjects() ([]map[string]any, error)
	GetProject(id int64) (map[string]any, error)
	CreateProject(name, location, status string, startDate, expectedCompletion any, client, contractor string, budgetTotal int64, description string) (int64, error)
	UpdateProject(id int64, updates []FieldUpdate) error
	AddProjectMember(projectID int64, userID int64) error
	BudgetMonths(projectID int64) ([]map[string]any, error)
	TasksByProject(projectID int64) ([]map[string]any, error)
	MaterialsByProject(projectID int64) ([]map[string]any, error)
	DocumentsByProject(projectID int64) ([]map[string]any, error)
	ActivityByProject(projectID int64) ([]map[string]any, error)
	ProjectTeam(projectID int64) ([]map[string]any, error)
	ProjectStats(projectID int64) (total int, completed int64, dueWeek int64, blocked int64, err error)
	MaterialAlerts(projectID int64) (int, error)
	RecentDocuments(projectID int64) (int, error)
	TouchProject(projectID int64) error
	ListTasks(projectID any, status any, q any) ([]map[string]any, error)
	GetTask(id int64) (map[string]any, error)
	CreateTask(projectID int64, assigneeID any, name string, description string, startDate, dueDate any, status, priority string) (int64, error)
	UpdateTask(id int64, updates []FieldUpdate) error
	ProjectExists(id int64) (bool, error)
	UserExists(id int64) (bool, error)
	LogActivity(projectID any, actorID any, name, typ, text string)
}
