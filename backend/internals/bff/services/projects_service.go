package services

import (
	"context"

	"go.uber.org/zap"

	bffinterfaces "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/interfaces"
	bffcontracts "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/services/contracts"
	apperrors "github.com/ralvescosta/costa-financial-assistant/backend/pkgs/errors"
	commonv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/common/v1"
	onboardingv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/onboarding/v1"
)

// ProjectsServiceImpl implements bffinterfaces.ProjectsService using the Onboarding gRPC client.
type ProjectsServiceImpl struct {
	logger           *zap.Logger
	onboardingClient bffinterfaces.OnboardingClient
}

// NewProjectsService constructs a ProjectsServiceImpl.
func NewProjectsService(logger *zap.Logger, onboardingClient bffinterfaces.OnboardingClient) bffinterfaces.ProjectsService {
	return &ProjectsServiceImpl{logger: logger, onboardingClient: onboardingClient}
}

// GetCurrentProject returns the project identified by the caller's JWT project_id.
func (s *ProjectsServiceImpl) GetCurrentProject(ctx context.Context, projectID, userID, role string) (*bffcontracts.ProjectResponse, error) {
	resp, err := s.onboardingClient.GetProject(ctx, &onboardingv1.GetProjectRequest{
		Ctx: &commonv1.ProjectContext{
			ProjectId: projectID,
			UserId:    userID,
			Role:      role,
		},
		Session: sessionFromContext(ctx),
	})
	if err != nil {
		s.logger.Error("projects_svc: get project downstream call failed",
			zap.String("project_id", projectID),
			zap.Error(err))
		if appErr := apperrors.AsAppError(err); appErr != nil {
			return nil, appErr
		}
		return nil, apperrors.TranslateError(err, "service")
	}
	p := resp.GetProject()
	return &bffcontracts.ProjectResponse{
		ID:        p.GetId(),
		OwnerID:   p.GetOwnerId(),
		Name:      p.GetName(),
		CreatedAt: p.GetCreatedAt(),
		UpdatedAt: p.GetUpdatedAt(),
	}, nil
}

// ListMembers returns all members for the caller's project.
func (s *ProjectsServiceImpl) ListMembers(ctx context.Context, projectID, userID, role string, pageSize int32, pageToken string) (*bffcontracts.ListMembersResponse, error) {
	resp, err := s.onboardingClient.ListProjectMembers(ctx, &onboardingv1.ListProjectMembersRequest{
		Ctx: &commonv1.ProjectContext{
			ProjectId: projectID,
			UserId:    userID,
			Role:      role,
		},
		Session:    sessionFromContext(ctx),
		Pagination: defaultPagination(pageSize, pageToken, 25),
	})
	if err != nil {
		s.logger.Error("projects_svc: list members downstream call failed",
			zap.String("project_id", projectID),
			zap.Error(err))
		if appErr := apperrors.AsAppError(err); appErr != nil {
			return nil, appErr
		}
		return nil, apperrors.TranslateError(err, "service")
	}

	items := make([]*bffcontracts.ProjectMemberResponse, 0, len(resp.GetMembers()))
	for _, m := range resp.GetMembers() {
		items = append(items, &bffcontracts.ProjectMemberResponse{
			ID:        m.GetId(),
			ProjectID: m.GetProjectId(),
			UserID:    m.GetUserId(),
			Role:      protoRoleToString(m.GetRole()),
			InvitedBy: m.GetInvitedBy(),
			CreatedAt: m.GetCreatedAt(),
			UpdatedAt: m.GetUpdatedAt(),
		})
	}
	var nextToken string
	if resp.GetPagination() != nil {
		nextToken = resp.GetPagination().GetNextPageToken()
	}
	return &bffcontracts.ListMembersResponse{Items: items, NextPageToken: nextToken}, nil
}

// InviteMember sends an invitation to the given email with the specified role.
func (s *ProjectsServiceImpl) InviteMember(ctx context.Context, projectID, inviterID, inviterRole, email, role string) (*bffcontracts.ProjectMemberResponse, error) {
	resp, err := s.onboardingClient.InviteProjectMember(ctx, &onboardingv1.InviteProjectMemberRequest{
		Ctx: &commonv1.ProjectContext{
			ProjectId: projectID,
			UserId:    inviterID,
			Role:      inviterRole,
		},
		Session:      sessionFromContext(ctx),
		InviteeEmail: email,
		Role:         roleStringToProto(role),
		Audit:        &commonv1.AuditMetadata{PerformedBy: inviterID},
	})
	if err != nil {
		s.logger.Error("projects_svc: invite member downstream call failed",
			zap.String("project_id", projectID),
			zap.String("invitee_email", email),
			zap.Error(err))
		if appErr := apperrors.AsAppError(err); appErr != nil {
			return nil, appErr
		}
		return nil, apperrors.TranslateError(err, "service")
	}
	m := resp.GetMember()
	return &bffcontracts.ProjectMemberResponse{
		ID:        m.GetId(),
		ProjectID: m.GetProjectId(),
		UserID:    m.GetUserId(),
		Role:      protoRoleToString(m.GetRole()),
		InvitedBy: m.GetInvitedBy(),
		CreatedAt: m.GetCreatedAt(),
		UpdatedAt: m.GetUpdatedAt(),
	}, nil
}

// UpdateMemberRole changes the role of an existing project member.
func (s *ProjectsServiceImpl) UpdateMemberRole(ctx context.Context, projectID, callerID, callerRole, memberID, newRole string) (*bffcontracts.ProjectMemberResponse, error) {
	resp, err := s.onboardingClient.UpdateProjectMemberRole(ctx, &onboardingv1.UpdateProjectMemberRoleRequest{
		Ctx: &commonv1.ProjectContext{
			ProjectId: projectID,
			UserId:    callerID,
			Role:      callerRole,
		},
		Session:  sessionFromContext(ctx),
		MemberId: memberID,
		NewRole:  roleStringToProto(newRole),
		Audit:    &commonv1.AuditMetadata{PerformedBy: callerID},
	})
	if err != nil {
		s.logger.Error("projects_svc: update member role downstream call failed",
			zap.String("project_id", projectID),
			zap.String("member_id", memberID),
			zap.Error(err))
		if appErr := apperrors.AsAppError(err); appErr != nil {
			return nil, appErr
		}
		return nil, apperrors.TranslateError(err, "service")
	}
	m := resp.GetMember()
	return &bffcontracts.ProjectMemberResponse{
		ID:        m.GetId(),
		ProjectID: m.GetProjectId(),
		UserID:    m.GetUserId(),
		Role:      protoRoleToString(m.GetRole()),
		InvitedBy: m.GetInvitedBy(),
		CreatedAt: m.GetCreatedAt(),
		UpdatedAt: m.GetUpdatedAt(),
	}, nil
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func roleStringToProto(role string) onboardingv1.ProjectMemberRole {
	switch role {
	case "update":
		return onboardingv1.ProjectMemberRole_PROJECT_MEMBER_ROLE_UPDATE
	case "write":
		return onboardingv1.ProjectMemberRole_PROJECT_MEMBER_ROLE_WRITE
	default:
		return onboardingv1.ProjectMemberRole_PROJECT_MEMBER_ROLE_READ_ONLY
	}
}

func protoRoleToString(r onboardingv1.ProjectMemberRole) string {
	switch r {
	case onboardingv1.ProjectMemberRole_PROJECT_MEMBER_ROLE_UPDATE:
		return "update"
	case onboardingv1.ProjectMemberRole_PROJECT_MEMBER_ROLE_WRITE:
		return "write"
	default:
		return "read_only"
	}
}
