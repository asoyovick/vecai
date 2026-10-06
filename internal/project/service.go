package project

import (
	"errors"
	"fmt"

	"vecai/internal/api"
	"vecai/internal/auth"
)

// Service holds the project and task business rules.
type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service { return &Service{repo: repo} }

// MapProject enriches a raw project row like the API has always returned it.
func (s *Service) MapProject(row map[string]any) map[string]any {
	id := row["id"].(int64)
	stats, _ := s.projectStats(id)
	team, _ := s.repo.ProjectTeam(id)
	if team == nil {
		team = []map[string]any{}
	}
	return map[string]any{
		"id": row["id"], "name": row["name"], "location": row["location"], "status": row["status"],
		"progress": row["progress"], "startDate": row["start_date"], "expectedCompletion": row["expected_completion"],
		"client": row["client"], "contractor": row["contractor"], "budgetTotal": row["budget_total"],
		"budgetSpent": row["budget_spent"], "description": row["description"], "imageUrl": row["image_url"],
		"createdAt": row["created_at"], "updatedAt": row["updated_at"],
		"stats": stats, "team": team,
	}
}

func (s *Service) projectStats(id int64) (map[string]any, error) {
	total, completed, dueWeek, blocked, err := s.repo.ProjectStats(id)
	if err != nil {
		return nil, err
	}
	alerts, _ := s.repo.MaterialAlerts(id)
	recentDocs, _ := s.repo.RecentDocuments(id)
	return map[string]any{
		"tasksTotal": total, "tasksCompleted": completed, "tasksDueThisWeek": dueWeek,
		"tasksBlocked": blocked, "materialsAlert": alerts, "documentsRecent": recentDocs,
	}, nil
}

func (s *Service) ListProjects() ([]map[string]any, error) {
	rows, err := s.repo.ListProjects()
	if err != nil {
		return nil, err
	}
	out := []map[string]any{}
	for _, row := range rows {
		out = append(out, s.MapProject(row))
	}
	return out, nil
}

// CreateProject validates and creates a project; the creating admin joins its team.
func (s *Service) CreateProject(admin *auth.User, input map[string]any) (map[string]any, error) {
	name, err := api.AsString(input["name"], "Project name", true, 160)
	if err != nil {
		return nil, err
	}
	location, _ := api.AsString(input["location"], "Location", false, 160)
	status := "Tendering"
	if input["status"] != nil {
		status, err = api.AsOneOf(input["status"], "Status", Statuses)
		if err != nil {
			return nil, err
		}
	}
	startDate, err := api.AsDate(input["startDate"], "Start date", false)
	if err != nil {
		return nil, err
	}
	completion, err := api.AsDate(input["expectedCompletion"], "Expected completion", false)
	if err != nil {
		return nil, err
	}
	client, _ := api.AsString(input["client"], "Client", false, 160)
	contractor, _ := api.AsString(input["contractor"], "Contractor", false, 160)
	budgetTotal, err := api.AsNumber(input["budgetTotal"], "Total budget", false, 0, true)
	if err != nil {
		return nil, err
	}
	description, _ := api.AsString(input["description"], "Description", false, 4000)
	id, err := s.repo.CreateProject(name, location, status, strptr(startDate), strptr(completion), client, contractor, int64(budgetTotal), description)
	if err != nil {
		return nil, err
	}
	s.repo.AddProjectMember(id, admin.ID)
	s.repo.LogActivity(id, admin.ID, admin.Name, "project", "Created project “"+name+"”")
	row, err := s.repo.GetProject(id)
	if err != nil {
		return nil, err
	}
	return s.MapProject(row), nil
}

// ProjectDetail returns the full project view.
func (s *Service) ProjectDetail(id int64) (map[string]any, error) {
	row, err := s.repo.GetProject(id)
	if err != nil {
		return nil, ErrProjectNotFound
	}
	tasks, _ := s.repo.TasksByProject(id)
	materials, _ := s.repo.MaterialsByProject(id)
	documents, _ := s.repo.DocumentsByProject(id)
	activity, _ := s.repo.ActivityByProject(id)
	budget, _ := s.repo.BudgetMonths(id)
	for _, arr := range []*[]map[string]any{&tasks, &materials, &documents, &activity, &budget} {
		if *arr == nil {
			*arr = []map[string]any{}
		}
	}
	return map[string]any{
		"project": s.MapProject(row), "tasks": tasks, "materials": materials,
		"documents": documents, "activity": activity, "budget": budget,
	}, nil
}

// PatchProject applies a validated partial update.
func (s *Service) PatchProject(admin *auth.User, id int64, input map[string]any) (map[string]any, error) {
	_, err := s.repo.GetProject(id)
	if err != nil {
		return nil, ErrProjectNotFound
	}
	updates := []FieldUpdate{}
	if v, ok := input["name"]; ok {
		n, err := api.AsString(v, "Project name", true, 160)
		if err != nil {
			return nil, err
		}
		updates = append(updates, FieldUpdate{"name", n})
	}
	if v, ok := input["location"]; ok {
		l, err := api.AsString(v, "Location", false, 160)
		if err != nil {
			return nil, err
		}
		updates = append(updates, FieldUpdate{"location", l})
	}
	if v, ok := input["status"]; ok {
		st, err := api.AsOneOf(v, "Status", Statuses)
		if err != nil {
			return nil, err
		}
		updates = append(updates, FieldUpdate{"status", st})
	}
	if v, ok := input["progress"]; ok {
		n, err := api.AsNumber(v, "Progress", false, 0, true)
		if err != nil {
			return nil, err
		}
		if n > 100 {
			return nil, errors.New("Progress must be at most 100.")
		}
		updates = append(updates, FieldUpdate{"progress", int64(n)})
	}
	if v, ok := input["budgetSpent"]; ok {
		n, err := api.AsNumber(v, "Budget spent", false, 0, true)
		if err != nil {
			return nil, err
		}
		updates = append(updates, FieldUpdate{"budget_spent", int64(n)})
	}
	if v, ok := input["description"]; ok {
		d, err := api.AsString(v, "Description", false, 4000)
		if err != nil {
			return nil, err
		}
		updates = append(updates, FieldUpdate{"description", d})
	}
	if len(updates) == 0 {
		return nil, ErrNoFields
	}
	if err := s.repo.UpdateProject(id, updates); err != nil {
		return nil, err
	}
	if v, ok := input["progress"]; ok {
		s.repo.LogActivity(id, admin.ID, admin.Name, "project", "Updated progress to "+fmt.Sprint(v)+"%")
	} else {
		s.repo.LogActivity(id, admin.ID, admin.Name, "project", "Updated project details")
	}
	row, err := s.repo.GetProject(id)
	if err != nil {
		return nil, ErrProjectNotFound
	}
	return s.MapProject(row), nil
}

func (s *Service) Budget(id int64) ([]map[string]any, error) {
	rows, err := s.repo.BudgetMonths(id)
	if err != nil {
		return nil, err
	}
	if rows == nil {
		rows = []map[string]any{}
	}
	return rows, nil
}

// ListTasks filters the task register.
func (s *Service) ListTasks(projectID, status, q any) ([]map[string]any, error) {
	rows, err := s.repo.ListTasks(projectID, status, q)
	if err != nil {
		return nil, err
	}
	if rows == nil {
		rows = []map[string]any{}
	}
	return rows, nil
}

// CreateTask validates and inserts a task.
func (s *Service) CreateTask(u *auth.User, input map[string]any) (map[string]any, error) {
	if u.Role == "viewer" {
		return nil, errors.New("You do not have permission to create tasks.")
	}
	name, err := api.AsString(input["name"], "Task name", true, 200)
	if err != nil {
		return nil, err
	}
	description, _ := api.AsString(input["description"], "Description", false, 2000)
	projectID, err := api.AsNumber(input["projectId"], "Project", true, 0, true)
	if err != nil {
		return nil, err
	}
	var assigneeID any
	if input["assigneeId"] != nil && input["assigneeId"] != "" {
		v, err := api.AsNumber(input["assigneeId"], "Assignee", false, 0, true)
		if err != nil {
			return nil, err
		}
		if ok, _ := s.repo.UserExists(int64(v)); !ok {
			return nil, errors.New("Select a valid assignee.")
		}
		assigneeID = int64(v)
	}
	startDate, err := api.AsDate(input["startDate"], "Start date", false)
	if err != nil {
		return nil, err
	}
	dueDate, err := api.AsDate(input["dueDate"], "Due date", true)
	if err != nil {
		return nil, err
	}
	status := "not-started"
	if input["status"] != nil {
		status, err = api.AsOneOf(input["status"], "Status", TaskStatuses)
		if err != nil {
			return nil, err
		}
	}
	priority := "medium"
	if input["priority"] != nil {
		priority, err = api.AsOneOf(input["priority"], "Priority", TaskPriorities)
		if err != nil {
			return nil, err
		}
	}
	if ok, _ := s.repo.ProjectExists(int64(projectID)); !ok {
		return nil, errors.New("Select a valid project.")
	}
	id, err := s.repo.CreateTask(int64(projectID), assigneeID, name, description, strptr(startDate), strptr(dueDate), status, priority)
	if err != nil {
		return nil, err
	}
	s.repo.LogActivity(int64(projectID), u.ID, u.Name, "task", "Created task “"+name+"”")
	s.repo.TouchProject(int64(projectID))
	return s.repo.GetTask(id)
}

// PatchTask applies a partial update to a task.
func (s *Service) PatchTask(u *auth.User, id int64, input map[string]any) (map[string]any, error) {
	if u.Role == "viewer" {
		return nil, errors.New("You do not have permission to update tasks.")
	}
	existing, err := s.repo.GetTask(id)
	if err != nil {
		return nil, ErrTaskNotFound
	}
	updates := []FieldUpdate{}
	if v, ok := input["status"]; ok {
		s2, err := api.AsOneOf(v, "Status", TaskStatuses)
		if err != nil {
			return nil, err
		}
		updates = append(updates, FieldUpdate{"status", s2})
	}
	if v, ok := input["priority"]; ok {
		p2, err := api.AsOneOf(v, "Priority", TaskPriorities)
		if err != nil {
			return nil, err
		}
		updates = append(updates, FieldUpdate{"priority", p2})
	}
	if v, ok := input["name"]; ok {
		n, err := api.AsString(v, "Task name", true, 200)
		if err != nil {
			return nil, err
		}
		updates = append(updates, FieldUpdate{"name", n})
	}
	if v, ok := input["description"]; ok {
		d, err := api.AsString(v, "Description", false, 2000)
		if err != nil {
			return nil, err
		}
		updates = append(updates, FieldUpdate{"description", d})
	}
	if v, ok := input["assigneeId"]; ok {
		var assignee any
		if v != nil && v != "" {
			n, err := api.AsNumber(v, "Assignee", false, 0, true)
			if err != nil {
				return nil, err
			}
			if ok, _ := s.repo.UserExists(int64(n)); !ok {
				return nil, errors.New("Select a valid assignee.")
			}
			assignee = int64(n)
		}
		updates = append(updates, FieldUpdate{"assignee_id", assignee})
	}
	if v, ok := input["dueDate"]; ok {
		d, err := api.AsDate(v, "Due date", true)
		if err != nil {
			return nil, err
		}
		updates = append(updates, FieldUpdate{"due_date", *d})
	}
	if len(updates) == 0 {
		return nil, ErrNoFields
	}
	if err := s.repo.UpdateTask(id, updates); err != nil {
		return nil, err
	}
	if v, ok := input["status"]; ok && v != existing["status"] {
		labels := map[string]string{"not-started": "Not started", "in-progress": "In progress", "blocked": "Blocked", "completed": "Completed"}
		s.repo.LogActivity(existing["project_id"], u.ID, u.Name, "task", "Marked “"+existing["name"].(string)+"” as "+labels[v.(string)])
	}
	s.repo.TouchProject(existing["project_id"].(int64))
	return s.repo.GetTask(id)
}

func strptr(s *string) any {
	if s == nil {
		return nil
	}
	return *s
}
