package grpc

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/ralvescosta/costa-financial-assistant/backend/internals/identity/services"
	apperrors "github.com/ralvescosta/costa-financial-assistant/backend/pkgs/errors"
	identityv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/identity/v1"
)

// Server implements identityv1.IdentityServiceServer.
type Server struct {
	identityv1.UnimplementedIdentityServiceServer
	svc    services.TokenServiceIface
	logger *zap.Logger
}

// NewServer constructs an identity gRPC server.
func NewServer(svc services.TokenServiceIface, logger *zap.Logger) *Server {
	return &Server{svc: svc, logger: logger}
}

// AuthenticateUser validates the seeded bootstrap credentials and returns a signed session.
func (s *Server) AuthenticateUser(ctx context.Context, req *identityv1.AuthenticateUserRequest) (*identityv1.AuthenticateUserResponse, error) {
	if req.GetUsername() == "" || req.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "username and password are required")
	}

	token, expiresAt, claims, projectName, err := s.svc.AuthenticateUser(ctx, req.GetUsername(), req.GetPassword())
	if err != nil {
		if appErr := apperrors.AsAppError(err); appErr != nil {
			return nil, toGRPCStatusError(appErr)
		}
		s.logger.Error("AuthenticateUser failed", zap.String("username", req.GetUsername()), zap.Error(err))
		return nil, status.Error(codes.Internal, "authentication failed")
	}

	return &identityv1.AuthenticateUserResponse{
		Token:       token,
		ExpiresAt:   expiresAt,
		Claims:      claims,
		ProjectName: projectName,
	}, nil
}

// RefreshSession validates the current session token and reissues a fresh one.
func (s *Server) RefreshSession(ctx context.Context, req *identityv1.RefreshSessionRequest) (*identityv1.RefreshSessionResponse, error) {
	if req.GetToken() == "" {
		return nil, status.Error(codes.InvalidArgument, "token is required")
	}

	token, expiresAt, claims, projectName, err := s.svc.RefreshSession(ctx, req.GetToken())
	if err != nil {
		if appErr := apperrors.AsAppError(err); appErr != nil {
			return nil, toGRPCStatusError(appErr)
		}
		s.logger.Error("RefreshSession failed", zap.Error(err))
		return nil, status.Error(codes.Internal, "session refresh failed")
	}

	return &identityv1.RefreshSessionResponse{
		Token:       token,
		ExpiresAt:   expiresAt,
		Claims:      claims,
		ProjectName: projectName,
	}, nil
}

// IssueBootstrapToken signs and returns a JWT for the bootstrap user/project.
func (s *Server) IssueBootstrapToken(ctx context.Context, req *identityv1.IssueBootstrapTokenRequest) (*identityv1.IssueBootstrapTokenResponse, error) {
	if req.UserId == "" || req.ProjectId == "" || req.Role == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id, project_id, and role are required")
	}

	token, expiresAt, err := s.svc.IssueBootstrapToken(ctx, req.UserId, req.ProjectId, req.Role)
	if err != nil {
		if appErr := apperrors.AsAppError(err); appErr != nil {
			return nil, toGRPCStatusError(appErr)
		}
		s.logger.Error("IssueBootstrapToken failed", zap.Error(err))
		return nil, status.Error(codes.Internal, "token issuance failed")
	}

	s.logger.Info("bootstrap token issued",
		zap.String("user_id", req.UserId),
		zap.String("project_id", req.ProjectId),
	)
	return &identityv1.IssueBootstrapTokenResponse{Token: token, ExpiresAt: expiresAt}, nil
}

// ValidateToken verifies a JWT and returns the decoded claims.
func (s *Server) ValidateToken(ctx context.Context, req *identityv1.ValidateTokenRequest) (*identityv1.ValidateTokenResponse, error) {
	if req.Token == "" {
		return nil, status.Error(codes.InvalidArgument, "token is required")
	}

	valid, claims, err := s.svc.ValidateToken(ctx, req.Token)
	if err != nil {
		s.logger.Debug("token validation error", zap.Error(err))
		return &identityv1.ValidateTokenResponse{Valid: false}, nil
	}

	return &identityv1.ValidateTokenResponse{Valid: valid, Claims: claims}, nil
}

// GetJwksMetadata returns the public JWKS for external validators.
func (s *Server) GetJwksMetadata(ctx context.Context, _ *identityv1.GetJwksMetadataRequest) (*identityv1.GetJwksMetadataResponse, error) {
	jwks, err := s.svc.GetJwksMetadata(ctx)
	if err != nil {
		if appErr := apperrors.AsAppError(err); appErr != nil {
			return nil, toGRPCStatusError(appErr)
		}
		s.logger.Error("GetJwksMetadata failed", zap.Error(err))
		return nil, status.Error(codes.Internal, "jwks fetch failed")
	}
	return &identityv1.GetJwksMetadataResponse{Jwks: jwks}, nil
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
		return status.Error(codes.Unauthenticated, message)
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
