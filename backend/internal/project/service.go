package project

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// CreateProject validates and persists a new project owned by ownerID.
func (s *Service) CreateProject(ctx context.Context, ownerID string, req CreateRequest) (*Project, error) {
	if strings.TrimSpace(ownerID) == "" {
		return nil, fmt.Errorf("%w: authenticated user required", ErrInvalidInput)
	}
	if err := validateCreate(req); err != nil {
		return nil, err
	}

	// The status was validated in validateCreate, so honor the caller's
	// choice instead of silently defaulting every project to Planning.
	status := "Planning"
	if req.Status != "" {
		status = req.Status
	}

	p := &Project{
		ID:          uuid.NewString(),
		OwnerID:     ownerID,
		Name:        strings.TrimSpace(req.Name),
		Description: strings.TrimSpace(req.Description),
		Location:    strings.TrimSpace(req.Location),
		ClientName:  strings.TrimSpace(req.ClientName),
		Status:      status,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		CreatedAt:   time.Now().UTC(),
	}

	if err := s.repo.CreateProject(ctx, p); err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}
	return p, nil
}

func (s *Service) GetProjectByID(ctx context.Context, id string) (*Project, error) {
	return s.repo.GetProjectByID(ctx, id)
}

// AuthorizeProjectAccess verifies the user is the project owner or a project
// member. It returns the project when access is allowed.
func (s *Service) AuthorizeProjectAccess(ctx context.Context, projectID, userID string) (*Project, error) {
	p, err := s.repo.GetProjectByID(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if p.OwnerID == userID {
		return p, nil
	}
	if _, err := s.repo.GetMemberByUserID(ctx, projectID, userID); err == nil {
		return p, nil
	} else if !errors.Is(err, ErrMemberNotFound) {
		return nil, err
	}
	return nil, ErrAccessDenied
}

func (s *Service) ListProjectsByOwner(ctx context.Context, ownerID string) ([]*Project, error) {
	return s.repo.ListProjectsByOwner(ctx, ownerID)
}

// ListProjects returns all projects visible to the given user: projects they
// own plus projects they are a team member of.
func (s *Service) ListProjects(ctx context.Context, userID string) ([]*Project, error) {
	return s.repo.ListProjectsByOwner(ctx, userID)
}

// UpdateProject applies a partial update to an existing project.
func (s *Service) UpdateProject(ctx context.Context, id string, req UpdateRequest) (*Project, error) {
	existing, err := s.repo.GetProjectByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		if strings.TrimSpace(*req.Name) == "" {
			return nil, fmt.Errorf("%w: name cannot be empty", ErrInvalidInput)
		}
		existing.Name = strings.TrimSpace(*req.Name)
	}
	if req.Description != nil {
		existing.Description = strings.TrimSpace(*req.Description)
	}
	if req.Location != nil {
		existing.Location = strings.TrimSpace(*req.Location)
	}
	if req.ClientName != nil {
		existing.ClientName = strings.TrimSpace(*req.ClientName)
	}
	if req.Status != nil {
		if !IsValidProjectStatus(*req.Status) {
			return nil, fmt.Errorf("%w: status must be one of Planning, Active, Paused, Completed, Archived", ErrInvalidInput)
		}
		existing.Status = *req.Status
	}
	if req.StartDate != nil {
		existing.StartDate = *req.StartDate
	}
	if req.EndDate != nil {
		existing.EndDate = *req.EndDate
	}

	if err := s.repo.UpdateProject(ctx, existing); err != nil {
		return nil, fmt.Errorf("failed to update project: %w", err)
	}
	return existing, nil
}

func (s *Service) DeleteProject(ctx context.Context, id string) error {
	if err := s.repo.DeleteProject(ctx, id); err != nil {
		return err
	}
	return nil
}

// AddBlueprint attaches a drawing file reference to a project.
func (s *Service) AddBlueprint(ctx context.Context, projectID, uploaderID string, req AddBlueprintRequest) (*Blueprint, error) {
	if strings.TrimSpace(req.Title) == "" || strings.TrimSpace(req.FilePath) == "" {
		return nil, fmt.Errorf("%w: title and file_path are required", ErrInvalidInput)
	}

	version := req.Version
	if version <= 0 {
		version = 1
	}

	b := &Blueprint{
		ID:          uuid.NewString(),
		ProjectID:   projectID,
		Title:       strings.TrimSpace(req.Title),
		Description: strings.TrimSpace(req.Description),
		FilePath:    strings.TrimSpace(req.FilePath),
		Version:     version,
		UploaderID:  uploaderID,
		CreatedAt:   time.Now().UTC(),
	}

	if err := s.repo.AddBlueprint(ctx, b); err != nil {
		return nil, fmt.Errorf("failed to add blueprint: %w", err)
	}
	return b, nil
}

func (s *Service) GetBlueprintsByProject(ctx context.Context, projectID string) ([]*Blueprint, error) {
	return s.repo.GetBlueprintsByProject(ctx, projectID)
}

// AddSpec records a technical specification section for a project.
func (s *Service) AddSpec(ctx context.Context, projectID, userID string, req AddSpecRequest) (*Spec, error) {
	if strings.TrimSpace(req.Title) == "" || strings.TrimSpace(req.SectionCode) == "" {
		return nil, fmt.Errorf("%w: title and section_code are required", ErrInvalidInput)
	}

	spec := &Spec{
		ID:          uuid.NewString(),
		ProjectID:   projectID,
		Title:       strings.TrimSpace(req.Title),
		Description: req.Description,
		SectionCode: strings.TrimSpace(req.SectionCode),
		CreatedBy:   userID,
		CreatedAt:   time.Now().UTC(),
	}

	if err := s.repo.AddSpec(ctx, spec); err != nil {
		return nil, fmt.Errorf("failed to add spec: %w", err)
	}
	return spec, nil
}

func (s *Service) GetSpecsByProject(ctx context.Context, projectID string) ([]*Spec, error) {
	return s.repo.GetSpecsByProject(ctx, projectID)
}

// InviteMember adds a user to a project team with the given role.
func (s *Service) InviteMember(ctx context.Context, projectID string, req InviteMemberRequest) (*Member, error) {
	if strings.TrimSpace(req.UserID) == "" {
		return nil, fmt.Errorf("%w: user_id is required", ErrInvalidInput)
	}
	if strings.TrimSpace(req.Role) == "" {
		return nil, fmt.Errorf("%w: role is required", ErrInvalidInput)
	}

	m := &Member{
		ID:        uuid.NewString(),
		ProjectID: projectID,
		UserID:    req.UserID,
		Role:      strings.TrimSpace(req.Role),
		JoinedAt:  time.Now().UTC(),
	}

	if err := s.repo.AddMember(ctx, m); err != nil {
		return nil, err
	}
	return m, nil
}

func (s *Service) GetMembersByProject(ctx context.Context, projectID string) ([]*Member, error) {
	return s.repo.GetMembersByProject(ctx, projectID)
}

func (s *Service) GetMemberByID(ctx context.Context, projectID, memberID string) (*Member, error) {
	return s.repo.GetMemberByID(ctx, projectID, memberID)
}

func (s *Service) RemoveMember(ctx context.Context, projectID, memberID string) error {
	return s.repo.DeleteMember(ctx, projectID, memberID)
}

func validateCreate(req CreateRequest) error {
	if strings.TrimSpace(req.Name) == "" {
		return fmt.Errorf("%w: name is required", ErrInvalidInput)
	}
	if req.StartDate != "" && !isValidDate(req.StartDate) {
		return fmt.Errorf("%w: start_date must be YYYY-MM-DD", ErrInvalidInput)
	}
	if req.EndDate != "" && !isValidDate(req.EndDate) {
		return fmt.Errorf("%w: expected_end_date must be YYYY-MM-DD", ErrInvalidInput)
	}
	if req.Status != "" && !IsValidProjectStatus(req.Status) {
		return fmt.Errorf("%w: status must be one of Planning, Active, Paused, Completed, Archived", ErrInvalidInput)
	}
	return nil
}

func isValidDate(value string) bool {
	_, err := time.Parse("2006-01-02", value)
	return err == nil
}
