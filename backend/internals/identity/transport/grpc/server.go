package grpc

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/ralvescosta/costa-financial-assistant/backend/internals/identity/services"
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

// IssueBootstrapToken signs and returns a JWT for the bootstrap user/project.
func (s *Server) IssueBootstrapToken(ctx context.Context, req *identityv1.IssueBootstrapTokenRequest) (*identityv1.IssueBootstrapTokenResponse, error) {
	if req.UserId == "" || req.ProjectId == "" || req.Role == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id, project_id, and role are required")
	}

	token, expiresAt, err := s.svc.IssueBootstrapToken(ctx, req.UserId, req.ProjectId, req.Role)
	if err != nil {
		s.logger.Error("IssueBootstrapToken failed", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "token issuance failed: %v", err)
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
		s.logger.Error("GetJwksMetadata failed", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "jwks fetch failed: %v", err)
	}
	return &identityv1.GetJwksMetadataResponse{Jwks: jwks}, nil
}
