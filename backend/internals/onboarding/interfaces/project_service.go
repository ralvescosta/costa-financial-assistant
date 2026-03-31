// Package interfaces defines the canonical service and repository contracts for the onboarding domain.
// These interfaces consolidate the key contracts used by the gRPC server and are used as mock targets in tests.
package interfaces

import (
	"context"

	onboardingv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/onboarding/v1"
)

// ProjectMembersService defines the contract for project membership lifecycle management.
// It is implemented by services.ProjectMembersService and consumed by the onboarding gRPC server.
type ProjectMembersService interface {
	// CreateProject creates a new project tenant owned by the given user.
	CreateProject(ctx context.Context, ownerID, name string, projectType onboardingv1.ProjectType) (*onboardingv1.Project, error)

	// GetProject returns a project by its ID.
	GetProject(ctx context.Context, projectID string) (*onboardingv1.Project, error)

	// InviteProjectMember resolves the invitee by email and creates a membership with the given role.
	InviteProjectMember(ctx context.Context, projectID, inviteeEmail string, role onboardingv1.ProjectMemberRole, invitedBy string) (*onboardingv1.ProjectMember, error)

	// UpdateProjectMemberRole changes the role of a project member identified by memberID.
	UpdateProjectMemberRole(ctx context.Context, projectID, memberID string, newRole onboardingv1.ProjectMemberRole) (*onboardingv1.ProjectMember, error)

	// ListProjectMembers returns all members of the given project with optional cursor pagination.
	ListProjectMembers(ctx context.Context, projectID string, pageSize int32, pageToken string) ([]*onboardingv1.ProjectMember, string, error)
}

// ProjectMembersRepository defines the persistence contract for project membership data.
// It is implemented by repositories.PostgresProjectMembersRepository.
type ProjectMembersRepository interface {
	// CreateProject inserts a new project record and returns the created entity.
	CreateProject(ctx context.Context, ownerID, name, projectType string) (*onboardingv1.Project, error)

	// GetProject fetches a project by its ID.
	GetProject(ctx context.Context, projectID string) (*onboardingv1.Project, error)

	// FindUserByEmail looks up a user ID by their email address.
	FindUserByEmail(ctx context.Context, email string) (string, error)

	// CreateMember inserts a project membership record.
	CreateMember(ctx context.Context, projectID, userID, invitedBy, role string) (*onboardingv1.ProjectMember, error)

	// FindMemberByID fetches a single project member by memberID within a project.
	FindMemberByID(ctx context.Context, projectID, memberID string) (*onboardingv1.ProjectMember, error)

	// UpdateMemberRole updates the role of an existing member and returns the updated record.
	UpdateMemberRole(ctx context.Context, projectID, memberID, newRole string) (*onboardingv1.ProjectMember, error)

	// ListMembers returns all members for the project, with cursor-based pagination.
	ListMembers(ctx context.Context, projectID string, pageSize int32, pageToken string) ([]*onboardingv1.ProjectMember, string, error)
}
