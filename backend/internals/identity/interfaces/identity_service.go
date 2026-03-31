// Package interfaces defines the canonical service and repository contracts for the identity domain.
// These interfaces are consumed by gRPC handlers and used as mock targets in tests.
package interfaces

import (
	"context"

	identityv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/identity/v1"
	onboardingv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/onboarding/v1"
)

// TokenService defines the contract for JWT issuance, validation, and JWKS exposition.
// It is implemented by services.TokenService and mocked in identity handler tests.
type TokenService interface {
	// IssueBootstrapToken signs and returns a short-lived JWT for the bootstrap tenant context.
	IssueBootstrapToken(ctx context.Context, userID, projectID, role string) (token string, expiresAt int64, err error)

	// ValidateToken parses and cryptographically verifies a JWT.
	// Returns (false, nil, nil) for invalid or expired tokens.
	ValidateToken(ctx context.Context, token string) (valid bool, claims *identityv1.JwtClaims, err error)

	// GetJwksMetadata returns the RSA public JWKS used by external validators (BFF middleware).
	GetJwksMetadata(ctx context.Context) (*identityv1.JwksMetadata, error)
}

// IdentityRepository defines the persistence contract for identity data.
// Phase 1 uses an in-process seeded bootstrap; Phase 2+ will bind a PostgreSQL implementation.
type IdentityRepository interface {
	// FindUserByID returns the bootstrap user record by its UUID.
	FindUserByID(ctx context.Context, userID string) (*onboardingv1.User, error)

	// FindProjectMember returns the project-member record for the given user and project.
	FindProjectMember(ctx context.Context, projectID, userID string) (*onboardingv1.ProjectMember, error)
}
