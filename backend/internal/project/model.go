package project

import (
	"errors"
	"time"
)

var (
	ErrProjectNotFound  = errors.New("project not found")
	ErrMemberNotFound   = errors.New("team member not found")
	ErrInvalidInput     = errors.New("invalid project input")
	ErrDuplicateProject = errors.New("project with this ID already exists")
	ErrUnauthorized     = errors.New("unauthorized: insufficient permissions")
	ErrAccessDenied     = errors.New("access denied")
)

// Project mirrors the projects table from docs/database.md.
type Project struct {
	ID          string    `json:"id"`
	OwnerID     string    `json:"owner_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Location    string    `json:"location"`
	ClientName  string    `json:"client_name"`
	Status      string    `json:"status"`
	StartDate   string    `json:"start_date,omitempty"`
	EndDate     string    `json:"expected_end_date,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Member mirrors the project_members table: many users work on many projects.
type Member struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	UserID    string    `json:"user_id"`
	Name      string    `json:"name,omitempty"`
	Email     string    `json:"email,omitempty"`
	Role      string    `json:"role"`
	JoinedAt  time.Time `json:"joined_at"`
}

// ValidProjectStatuses are the lifecycle states a project can move through.
var ValidProjectStatuses = map[string]bool{
	"Planning":  true,
	"Active":    true,
	"Paused":    true,
	"Completed": true,
	"Archived":  true,
}

func IsValidProjectStatus(status string) bool {
	return ValidProjectStatuses[status]
}

// Blueprint mirrors the blueprints table (drawing files attached to a project).
type Blueprint struct {
	ID          string    `json:"id"`
	ProjectID   string    `json:"project_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	FilePath    string    `json:"file_path"`
	Version     int       `json:"version"`
	UploaderID  string    `json:"uploader_id"`
	CreatedAt   time.Time `json:"created_at"`
}

// Spec mirrors the specs table (technical specification sections).
type Spec struct {
	ID          string    `json:"id"`
	ProjectID   string    `json:"project_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	SectionCode string    `json:"section_code"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
}
