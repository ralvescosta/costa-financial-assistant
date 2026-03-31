package grpc

import (
	"context"
	"errors"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/ralvescosta/costa-financial-assistant/backend/internals/onboarding/repositories"
	"github.com/ralvescosta/costa-financial-assistant/backend/internals/onboarding/services"
	commonv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/common/v1"
	onboardingv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/onboarding/v1"
)

// Server implements onboardingv1.OnboardingServiceServer.
type Server struct {
	onboardingv1.UnimplementedOnboardingServiceServer
	svc    services.ProjectMembersServiceIface
	logger *zap.Logger
}

// NewServer constructs an onboarding gRPC server with OTel interceptors.
func NewServer(svc services.ProjectMembersServiceIface, logger *zap.Logger) (*grpc.Server, *Server) {
	srv := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
	)
	handler := &Server{svc: svc, logger: logger}
	onboardingv1.RegisterOnboardingServiceServer(srv, handler)
	return srv, handler
}

// CreateProject creates a new project tenant owned by the caller.
func (s *Server) CreateProject(ctx context.Context, req *onboardingv1.CreateProjectRequest) (*onboardingv1.CreateProjectResponse, error) {
	if req.GetCtx() == nil || req.GetCtx().GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "project name is required")
	}

	project, err := s.svc.CreateProject(ctx, req.GetCtx().GetUserId(), req.GetName(), req.GetType())
	if err != nil {
		s.logger.Error("grpc.CreateProject failed",
			zap.String("owner_id", req.GetCtx().GetUserId()),
			zap.Error(err))
		return nil, status.Error(codes.Internal, "create project failed")
	}
	return &onboardingv1.CreateProjectResponse{Project: project}, nil
}

// GetProject returns a single project by its context project_id.
func (s *Server) GetProject(ctx context.Context, req *onboardingv1.GetProjectRequest) (*onboardingv1.GetProjectResponse, error) {
	if req.GetCtx() == nil || req.GetCtx().GetProjectId() == "" {
		return nil, status.Error(codes.InvalidArgument, "project_id is required")
	}

	project, err := s.svc.GetProject(ctx, req.GetCtx().GetProjectId())
	if err != nil {
		if errors.Is(err, repositories.ErrProjectNotFound) {
			return nil, status.Error(codes.NotFound, "project not found")
		}
		s.logger.Error("grpc.GetProject failed",
			zap.String("project_id", req.GetCtx().GetProjectId()),
			zap.Error(err))
		return nil, status.Error(codes.Internal, "get project failed")
	}
	return &onboardingv1.GetProjectResponse{Project: project}, nil
}

// InviteProjectMember adds a user to a project with a given role.
func (s *Server) InviteProjectMember(ctx context.Context, req *onboardingv1.InviteProjectMemberRequest) (*onboardingv1.InviteProjectMemberResponse, error) {
	if req.GetCtx() == nil || req.GetCtx().GetProjectId() == "" {
		return nil, status.Error(codes.InvalidArgument, "project_id is required")
	}
	if req.GetInviteeEmail() == "" {
		return nil, status.Error(codes.InvalidArgument, "invitee_email is required")
	}

	invitedBy := ""
	if req.GetAudit() != nil {
		invitedBy = req.GetAudit().GetPerformedBy()
	}

	member, err := s.svc.InviteProjectMember(ctx, req.GetCtx().GetProjectId(), req.GetInviteeEmail(), req.GetRole(), invitedBy)
	if err != nil {
		if errors.Is(err, repositories.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found by email")
		}
		if errors.Is(err, repositories.ErrMemberAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, "user is already a member of this project")
		}
		s.logger.Error("grpc.InviteProjectMember failed",
			zap.String("project_id", req.GetCtx().GetProjectId()),
			zap.String("invitee_email", req.GetInviteeEmail()),
			zap.Error(err))
		return nil, status.Error(codes.Internal, "invite member failed")
	}
	return &onboardingv1.InviteProjectMemberResponse{Member: member}, nil
}

// UpdateProjectMemberRole changes the role of an existing project member.
func (s *Server) UpdateProjectMemberRole(ctx context.Context, req *onboardingv1.UpdateProjectMemberRoleRequest) (*onboardingv1.UpdateProjectMemberRoleResponse, error) {
	if req.GetCtx() == nil || req.GetCtx().GetProjectId() == "" {
		return nil, status.Error(codes.InvalidArgument, "project_id is required")
	}
	if req.GetMemberId() == "" {
		return nil, status.Error(codes.InvalidArgument, "member_id is required")
	}

	member, err := s.svc.UpdateProjectMemberRole(ctx, req.GetCtx().GetProjectId(), req.GetMemberId(), req.GetNewRole())
	if err != nil {
		if errors.Is(err, repositories.ErrMemberNotFound) {
			return nil, status.Error(codes.NotFound, "project member not found")
		}
		s.logger.Error("grpc.UpdateProjectMemberRole failed",
			zap.String("project_id", req.GetCtx().GetProjectId()),
			zap.String("member_id", req.GetMemberId()),
			zap.Error(err))
		return nil, status.Error(codes.Internal, "update member role failed")
	}
	return &onboardingv1.UpdateProjectMemberRoleResponse{Member: member}, nil
}

// ListProjectMembers returns all members of a given project.
func (s *Server) ListProjectMembers(ctx context.Context, req *onboardingv1.ListProjectMembersRequest) (*onboardingv1.ListProjectMembersResponse, error) {
	if req.GetCtx() == nil || req.GetCtx().GetProjectId() == "" {
		return nil, status.Error(codes.InvalidArgument, "project_id is required")
	}

	pageSize := int32(25)
	pageToken := ""
	if req.GetPagination() != nil {
		if req.GetPagination().GetPageSize() > 0 {
			pageSize = req.GetPagination().GetPageSize()
		}
		pageToken = req.GetPagination().GetPageToken()
	}

	members, nextToken, err := s.svc.ListProjectMembers(ctx, req.GetCtx().GetProjectId(), pageSize, pageToken)
	if err != nil {
		s.logger.Error("grpc.ListProjectMembers failed",
			zap.String("project_id", req.GetCtx().GetProjectId()),
			zap.Error(err))
		return nil, status.Error(codes.Internal, "list members failed")
	}

	return &onboardingv1.ListProjectMembersResponse{
		Members: members,
		Pagination: &commonv1.PaginationResult{
			NextPageToken: nextToken,
		},
	}, nil
}
