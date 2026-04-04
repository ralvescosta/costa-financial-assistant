package grpc

import (
	"context"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/ralvescosta/costa-financial-assistant/backend/internals/onboarding/services"
	apperrors "github.com/ralvescosta/costa-financial-assistant/backend/pkgs/errors"
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
	pc := req.GetCtx()
	if pc == nil {
		return nil, status.Error(codes.InvalidArgument, "project context is required")
	}
	sessionUserID, err := requireSessionUserID(req.GetSession())
	if err != nil {
		return nil, err
	}
	ownerUserID := pc.GetUserId()
	if ownerUserID == "" {
		ownerUserID = sessionUserID
	}
	if ownerUserID == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "project name is required")
	}

	project, err := s.svc.CreateProject(ctx, ownerUserID, req.GetName(), req.GetType())
	if err != nil {
		if appErr := apperrors.AsAppError(err); appErr != nil {
			return nil, toGRPCStatusError(appErr)
		}
		s.logger.Error("grpc.CreateProject failed",
			zap.String("owner_id", req.GetCtx().GetUserId()),
			zap.Error(err))
		return nil, status.Error(codes.Internal, "internal service error")
	}
	return &onboardingv1.CreateProjectResponse{Project: project}, nil
}

// GetProject returns a single project by its context project_id.
func (s *Server) GetProject(ctx context.Context, req *onboardingv1.GetProjectRequest) (*onboardingv1.GetProjectResponse, error) {
	if req.GetCtx() == nil || req.GetCtx().GetProjectId() == "" {
		return nil, status.Error(codes.InvalidArgument, "project_id is required")
	}
	if _, err := requireSessionUserID(req.GetSession()); err != nil {
		return nil, err
	}

	project, err := s.svc.GetProject(ctx, req.GetCtx().GetProjectId())
	if err != nil {
		if appErr := apperrors.AsAppError(err); appErr != nil {
			return nil, toGRPCStatusError(appErr)
		}
		s.logger.Error("grpc.GetProject failed",
			zap.String("project_id", req.GetCtx().GetProjectId()),
			zap.Error(err))
		return nil, status.Error(codes.Internal, "internal service error")
	}
	return &onboardingv1.GetProjectResponse{Project: project}, nil
}

// InviteProjectMember adds a user to a project with a given role.
func (s *Server) InviteProjectMember(ctx context.Context, req *onboardingv1.InviteProjectMemberRequest) (*onboardingv1.InviteProjectMemberResponse, error) {
	if req.GetCtx() == nil || req.GetCtx().GetProjectId() == "" {
		return nil, status.Error(codes.InvalidArgument, "project_id is required")
	}
	sessionUserID, err := requireSessionUserID(req.GetSession())
	if err != nil {
		return nil, err
	}
	if req.GetInviteeEmail() == "" {
		return nil, status.Error(codes.InvalidArgument, "invitee_email is required")
	}

	invitedBy := sessionUserID
	if req.GetAudit() != nil && req.GetAudit().GetPerformedBy() != "" {
		invitedBy = req.GetAudit().GetPerformedBy()
	}

	member, err := s.svc.InviteProjectMember(ctx, req.GetCtx().GetProjectId(), req.GetInviteeEmail(), req.GetRole(), invitedBy)
	if err != nil {
		if appErr := apperrors.AsAppError(err); appErr != nil {
			return nil, toGRPCStatusError(appErr)
		}
		s.logger.Error("grpc.InviteProjectMember failed",
			zap.String("project_id", req.GetCtx().GetProjectId()),
			zap.String("invitee_email", req.GetInviteeEmail()),
			zap.Error(err))
		return nil, status.Error(codes.Internal, "internal service error")
	}
	return &onboardingv1.InviteProjectMemberResponse{Member: member}, nil
}

// UpdateProjectMemberRole changes the role of an existing project member.
func (s *Server) UpdateProjectMemberRole(ctx context.Context, req *onboardingv1.UpdateProjectMemberRoleRequest) (*onboardingv1.UpdateProjectMemberRoleResponse, error) {
	if req.GetCtx() == nil || req.GetCtx().GetProjectId() == "" {
		return nil, status.Error(codes.InvalidArgument, "project_id is required")
	}
	if _, err := requireSessionUserID(req.GetSession()); err != nil {
		return nil, err
	}
	if req.GetMemberId() == "" {
		return nil, status.Error(codes.InvalidArgument, "member_id is required")
	}

	member, err := s.svc.UpdateProjectMemberRole(ctx, req.GetCtx().GetProjectId(), req.GetMemberId(), req.GetNewRole())
	if err != nil {
		if appErr := apperrors.AsAppError(err); appErr != nil {
			return nil, toGRPCStatusError(appErr)
		}
		s.logger.Error("grpc.UpdateProjectMemberRole failed",
			zap.String("project_id", req.GetCtx().GetProjectId()),
			zap.String("member_id", req.GetMemberId()),
			zap.Error(err))
		return nil, status.Error(codes.Internal, "internal service error")
	}
	return &onboardingv1.UpdateProjectMemberRoleResponse{Member: member}, nil
}

// ListProjectMembers returns all members of a given project.
func (s *Server) ListProjectMembers(ctx context.Context, req *onboardingv1.ListProjectMembersRequest) (*onboardingv1.ListProjectMembersResponse, error) {
	if req.GetCtx() == nil || req.GetCtx().GetProjectId() == "" {
		return nil, status.Error(codes.InvalidArgument, "project_id is required")
	}
	if _, err := requireSessionUserID(req.GetSession()); err != nil {
		return nil, err
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
		if appErr := apperrors.AsAppError(err); appErr != nil {
			return nil, toGRPCStatusError(appErr)
		}
		s.logger.Error("grpc.ListProjectMembers failed",
			zap.String("project_id", req.GetCtx().GetProjectId()),
			zap.Error(err))
		return nil, status.Error(codes.Internal, "internal service error")
	}

	return &onboardingv1.ListProjectMembersResponse{
		Members: members,
		Pagination: &commonv1.PaginationResult{
			NextPageToken: nextToken,
		},
	}, nil
}

func requireSessionUserID(session *commonv1.Session) (string, error) {
	if session == nil || session.GetId() == "" {
		return "", status.Error(codes.Unauthenticated, "session is required")
	}
	return session.GetId(), nil
}

func toGRPCStatusError(appErr *apperrors.AppError) error {
	if appErr == nil {
		return status.Error(codes.Internal, "internal service error")
	}

	message := appErr.Message
	if message == "" {
		message = "internal service error"
	}

	switch appErr.Category {
	case apperrors.CategoryValidation:
		return status.Error(codes.InvalidArgument, message)
	case apperrors.CategoryAuth:
		return status.Error(codes.PermissionDenied, message)
	case apperrors.CategoryNotFound:
		return status.Error(codes.NotFound, message)
	case apperrors.CategoryConflict:
		return status.Error(codes.AlreadyExists, message)
	case apperrors.CategoryDependencyDB, apperrors.CategoryDependencyGRPC, apperrors.CategoryDependencyNet:
		if appErr.Retryable {
			return status.Error(codes.Unavailable, message)
		}
		return status.Error(codes.FailedPrecondition, message)
	default:
		return status.Error(codes.Internal, message)
	}
}
