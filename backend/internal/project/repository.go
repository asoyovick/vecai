package project

import "context"

type Repository interface {
	CreateProject(ctx context.Context, project *Project) error
	GetProjectByID(ctx context.Context, id string) (*Project, error)
	ListProjectsByOwner(ctx context.Context, ownerID string) ([]*Project, error)
	ListProjectsByUser(ctx context.Context, userID string) ([]*Project, error)
	ListAllProjects(ctx context.Context) ([]*Project, error)
	UpdateProject(ctx context.Context, project *Project) error
	DeleteProject(ctx context.Context, id string) error

	AddBlueprint(ctx context.Context, blueprint *Blueprint) error
	GetBlueprintsByProject(ctx context.Context, projectID string) ([]*Blueprint, error)

	AddSpec(ctx context.Context, spec *Spec) error
	GetSpecsByProject(ctx context.Context, projectID string) ([]*Spec, error)

	AddMember(ctx context.Context, member *Member) error
	GetMembersByProject(ctx context.Context, projectID string) ([]*Member, error)
	GetMemberByID(ctx context.Context, projectID, memberID string) (*Member, error)
	DeleteMember(ctx context.Context, projectID, memberID string) error
}
