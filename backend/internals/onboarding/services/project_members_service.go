package services

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"

	"github.com/ralvescosta/costa-financial-assistant/backend/internals/onboarding/repositories"
	onboardingv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/onboarding/v1"
)

// projectTypeStrings maps proto enum values to their PostgreSQL string representations.
var projectTypeStrings = map[onboardingv1.ProjectType]string{
	onboardingv1.ProjectType_PROJECT_TYPE_PERSONAL: "personal",
	onboardingv1.ProjectType_PROJECT_TYPE_CONJUGAL: "conjugal",
	onboardingv1.ProjectType_PROJECT_TYPE_SHARED:   "shared",
}

// projectMemberRoleStrings maps proto role enum values to PostgreSQL string representations.
var projectMemberRoleStrings = map[onboardingv1.ProjectMemberRole]string{
	onboardingv1.ProjectMemberRole_PROJECT_MEMBER_ROLE_READ_ONLY: "read_only",
	onboardingv1.ProjectMemberRole_PROJECT_MEMBER_ROLE_UPDATE:    "update",
	onboardingv1.ProjectMemberRole_PROJECT_MEMBER_ROLE_WRITE:     "write",
}

// ProjectMembersServiceIface is the narrow interface consumed by the gRPC server.
type ProjectMembersServiceIface interface {
	CreateProject(ctx context.Context, ownerID, name string, projectType onboardingv1.ProjectType) (*onboardingv1.Project, error)
	GetProject(ctx context.Context, projectID string) (*onboardingv1.Project, error)
	InviteProjectMember(ctx context.Context, projectID, inviteeEmail string, role onboardingv1.ProjectMemberRole, invitedBy string) (*onboardingv1.ProjectMember, error)
	UpdateProjectMemberRole(ctx context.Context, projectID, memberID string, newRole onboardingv1.ProjectMemberRole) (*onboardingv1.ProjectMember, error)
	ListProjectMembers(ctx context.Context, projectID string, pageSize int32, pageToken string) ([]*onboardingv1.ProjectMember, string, error)
}

// ProjectMembersService implements ProjectMembersServiceIface.
type ProjectMembersService struct {
	repo   *repositories.PostgresProjectMembersRepository
	logger *zap.Logger
}

// NewProjectMembersService constructs a ProjectMembersService.
func NewProjectMembersService(repo *repositories.PostgresProjectMembersRepository, logger *zap.Logger) ProjectMembersServiceIface {
	return &ProjectMembersService{repo: repo, logger: logger}
}

// CreateProject creates a new project tenant owned by the given user.
func (s *ProjectMembersService) CreateProject(ctx context.Context, ownerID, name string, projectType onboardingv1.ProjectType) (*onboardingv1.Project, error) {
	if name == "" {
		return nil, fmt.Errorf("project_members_service: project name is required")
	}

	typeStr, ok := projectTypeStrings[projectType]
	if !ok {
		typeStr = "personal"
	}

	project, err := s.repo.CreateProject(ctx, ownerID, name, typeStr)
	if err != nil {
		s.logger.Error("project_members_service.create_project: failed",
			zap.String("owner_id", ownerID),
			zap.Error(err))
		return nil, fmt.Errorf("project_members_service: create project: %w", err)
	}
	return project, nil
}

// GetProject returns a project by its ID.
func (s *ProjectMembersService) GetProject(ctx context.Context, projectID string) (*onboardingv1.Project, error) {
	project, err := s.repo.GetProject(ctx, projectID)
	if err != nil {
		if errors.Is(err, repositories.ErrProjectNotFound) {
			return nil, repositories.ErrProjectNotFound
		}
		s.logger.Error("project_members_service.get_project: failed",
			zap.String("project_id", projectID),
			zap.Error(err))
		return nil, fmt.Errorf("project_members_service: get project: %w", err)
	}
	return project, nil
}

// InviteProjectMember resolves the invitee by email and creates a membership with the given role.
func (s *ProjectMembersService) InviteProjectMember(ctx context.Context, projectID, inviteeEmail string, role onboardingv1.ProjectMemberRole, invitedBy string) (*onboardingv1.ProjectMember, error) {
	if inviteeEmail == "" {
		return nil, fmt.Errorf("project_members_service: invitee email is required")
	}

	userID, err := s.repo.FindUserByEmail(ctx, inviteeEmail)
	if err != nil {
		if errors.Is(err, repositories.ErrUserNotFound) {
			return nil, repositories.ErrUserNotFound
		}
		s.logger.Error("project_members_service.invite: find user failed",
			zap.String("email", inviteeEmail),
			zap.Error(err))
		return nil, fmt.Errorf("project_members_service: invite member: user lookup: %w", err)
	}

	roleStr, ok := projectMemberRoleStrings[role]
	if !ok {
		roleStr = "read_only"
	}

	member, err := s.repo.CreateMember(ctx, projectID, userID, invitedBy, roleStr)
	if err != nil {
		if errors.Is(err, repositories.ErrMemberAlreadyExists) {
			return nil, repositories.ErrMemberAlreadyExists
		}
		s.logger.Error("project_members_service.invite: create member failed",
			zap.String("project_id", projectID),
			zap.String("user_id", userID),
			zap.Error(err))
		return nil, fmt.Errorf("project_members_service: invite member: create: %w", err)
	}
	return member, nil
}

// UpdateProjectMemberRole changes the role of a project member.
func (s *ProjectMembersService) UpdateProjectMemberRole(ctx context.Context, projectID, memberID string, newRole onboardingv1.ProjectMemberRole) (*onboardingv1.ProjectMember, error) {
	if memberID == "" {
		return nil, fmt.Errorf("project_members_service: member_id is required")
	}

	roleStr, ok := projectMemberRoleStrings[newRole]
	if !ok {
		roleStr = "read_only"
	}

	member, err := s.repo.UpdateMemberRole(ctx, projectID, memberID, roleStr)
	if err != nil {
		if errors.Is(err, repositories.ErrMemberNotFound) {
			return nil, repositories.ErrMemberNotFound
		}
		s.logger.Error("project_members_service.update_role: failed",
			zap.String("project_id", projectID),
			zap.String("member_id", memberID),
			zap.Error(err))
		return nil, fmt.Errorf("project_members_service: update member role: %w", err)
	}
	return member, nil
}

// ListProjectMembers returns all members of the given project with pagination.
func (s *ProjectMembersService) ListProjectMembers(ctx context.Context, projectID string, pageSize int32, pageToken string) ([]*onboardingv1.ProjectMember, string, error) {
	members, nextToken, err := s.repo.ListMembers(ctx, projectID, pageSize, pageToken)
	if err != nil {
		s.logger.Error("project_members_service.list_members: failed",
			zap.String("project_id", projectID),
			zap.Error(err))
		return nil, "", fmt.Errorf("project_members_service: list members: %w", err)
	}
	return members, nextToken, nil
}
