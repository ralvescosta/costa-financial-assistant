package services_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/services"
	apperrors "github.com/ralvescosta/costa-financial-assistant/backend/pkgs/errors"
	identityv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/identity/v1"
)

type mockIdentityClient struct{ mock.Mock }

func (m *mockIdentityClient) AuthenticateUser(ctx context.Context, in *identityv1.AuthenticateUserRequest, opts ...grpc.CallOption) (*identityv1.AuthenticateUserResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*identityv1.AuthenticateUserResponse), args.Error(1)
}

func (m *mockIdentityClient) RefreshSession(ctx context.Context, in *identityv1.RefreshSessionRequest, opts ...grpc.CallOption) (*identityv1.RefreshSessionResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*identityv1.RefreshSessionResponse), args.Error(1)
}

func (m *mockIdentityClient) ValidateToken(ctx context.Context, in *identityv1.ValidateTokenRequest, opts ...grpc.CallOption) (*identityv1.ValidateTokenResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*identityv1.ValidateTokenResponse), args.Error(1)
}

func (m *mockIdentityClient) GetJwksMetadata(ctx context.Context, in *identityv1.GetJwksMetadataRequest, opts ...grpc.CallOption) (*identityv1.GetJwksMetadataResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*identityv1.GetJwksMetadataResponse), args.Error(1)
}

func TestAuthService_Login_ReturnsSessionEnvelope(t *testing.T) {
	identity := &mockIdentityClient{}
	svc := services.NewAuthService(zaptest.NewLogger(t), identity)
	ctx := context.Background()

	identity.On("AuthenticateUser", ctx, mock.AnythingOfType("*identityv1.AuthenticateUserRequest")).Return(
		&identityv1.AuthenticateUserResponse{
			Token:       "jwt-token",
			ExpiresAt:   4102444800,
			ProjectName: "Costa Financial Assistant",
			Claims: &identityv1.JwtClaims{
				Subject:   "00000000-0000-0000-0000-000000000001",
				Username:  "ralvescosta",
				Email:     "ralvescosta@local.dev",
				ProjectId: "00000000-0000-0000-0000-000000000010",
				Role:      "write",
			},
		}, nil,
	)

	resp, err := svc.Login(ctx, "ralvescosta", "mudar@1234")

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "jwt-token", resp.AccessToken)
	assert.Equal(t, "ralvescosta", resp.User.Username)
	assert.Equal(t, "Costa Financial Assistant", resp.ActiveProject.Name)
}

func TestAuthService_Login_MapsUnauthorizedError(t *testing.T) {
	identity := &mockIdentityClient{}
	svc := services.NewAuthService(zaptest.NewLogger(t), identity)
	ctx := context.Background()

	identity.On("AuthenticateUser", ctx, mock.AnythingOfType("*identityv1.AuthenticateUserRequest")).Return(nil, status.Error(codes.Unauthenticated, "invalid credentials"))

	resp, err := svc.Login(ctx, "ralvescosta", "wrong")

	assert.Nil(t, resp)
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperrors.CategoryAuth, appErr.Category)
}
