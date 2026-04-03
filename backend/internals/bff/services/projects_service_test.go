package services_test

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"google.golang.org/grpc"

	bffinterfaces "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/interfaces"
	"github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/services"
	onboardingv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/onboarding/v1"
)

// ─── mock: OnboardingClient ───────────────────────────────────────────────────

type mockOnboardingClient struct{ mock.Mock }

func (m *mockOnboardingClient) CreateProject(ctx context.Context, in *onboardingv1.CreateProjectRequest, opts ...grpc.CallOption) (*onboardingv1.CreateProjectResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*onboardingv1.CreateProjectResponse), args.Error(1)
}

func (m *mockOnboardingClient) InviteProjectMember(ctx context.Context, in *onboardingv1.InviteProjectMemberRequest, opts ...grpc.CallOption) (*onboardingv1.InviteProjectMemberResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*onboardingv1.InviteProjectMemberResponse), args.Error(1)
}

func (m *mockOnboardingClient) UpdateProjectMemberRole(ctx context.Context, in *onboardingv1.UpdateProjectMemberRoleRequest, opts ...grpc.CallOption) (*onboardingv1.UpdateProjectMemberRoleResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*onboardingv1.UpdateProjectMemberRoleResponse), args.Error(1)
}

func (m *mockOnboardingClient) ListProjectMembers(ctx context.Context, in *onboardingv1.ListProjectMembersRequest, opts ...grpc.CallOption) (*onboardingv1.ListProjectMembersResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*onboardingv1.ListProjectMembersResponse), args.Error(1)
}

func (m *mockOnboardingClient) GetProject(ctx context.Context, in *onboardingv1.GetProjectRequest, opts ...grpc.CallOption) (*onboardingv1.GetProjectResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*onboardingv1.GetProjectResponse), args.Error(1)
}

// ─── helpers ──────────────────────────────────────────────────────────────────

func newProjectsService(t *testing.T, client bffinterfaces.OnboardingClient) bffinterfaces.ProjectsService {
	t.Helper()
	return services.NewProjectsService(zaptest.NewLogger(t), client)
}

// ─── GetCurrentProject ────────────────────────────────────────────────────────

func TestProjectsService_GetCurrentProject_Success(t *testing.T) {
	// Arrange
	client := &mockOnboardingClient{}
	svc := newProjectsService(t, client)
	ctx := context.Background()

	client.On("GetProject", ctx, mock.AnythingOfType("*onboardingv1.GetProjectRequest")).Return(
		&onboardingv1.GetProjectResponse{
			Project: &onboardingv1.Project{
				Id:   "proj-1",
				Name: "My Project",
			},
		}, nil)

	// Act
	result, err := svc.GetCurrentProject(ctx, "proj-1", "user-1", "owner")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "proj-1", result.ID)
	assert.Equal(t, "My Project", result.Name)
	client.AssertExpectations(t)
}

func TestProjectsService_GetCurrentProject_ClientError(t *testing.T) {
	// Arrange
	client := &mockOnboardingClient{}
	svc := newProjectsService(t, client)
	ctx := context.Background()

	client.On("GetProject", ctx, mock.Anything).Return(nil, errors.New("not found"))

	// Act
	result, err := svc.GetCurrentProject(ctx, "proj-1", "user-1", "owner")

	// Assert
	assert.Nil(t, result)
	assert.Error(t, err)
}

// ─── ListMembers ──────────────────────────────────────────────────────────────

func TestProjectsService_ListMembers_ReturnsMembers(t *testing.T) {
	// Arrange
	client := &mockOnboardingClient{}
	svc := newProjectsService(t, client)
	ctx := context.Background()

	client.On("ListProjectMembers", ctx, mock.AnythingOfType("*onboardingv1.ListProjectMembersRequest")).Return(
		&onboardingv1.ListProjectMembersResponse{
			Members: []*onboardingv1.ProjectMember{
				{Id: "m1", ProjectId: "proj-1", UserId: "u1"},
				{Id: "m2", ProjectId: "proj-1", UserId: "u2"},
			},
		}, nil)

	// Act
	result, err := svc.ListMembers(ctx, "proj-1", "user-1", "owner", 10, "")

	// Assert
	require.NoError(t, err)
	assert.Len(t, result.Items, 2)
}

// ─── InviteMember ─────────────────────────────────────────────────────────────

func TestProjectsService_InviteMember_Success(t *testing.T) {
	// Arrange
	client := &mockOnboardingClient{}
	svc := newProjectsService(t, client)
	ctx := context.Background()

	client.On("InviteProjectMember", ctx, mock.AnythingOfType("*onboardingv1.InviteProjectMemberRequest")).Return(
		&onboardingv1.InviteProjectMemberResponse{
			Member: &onboardingv1.ProjectMember{Id: "m-new"},
		}, nil)

	// Act
	result, err := svc.InviteMember(ctx, "proj-1", "user-1", "owner", "new@example.com", "member")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "m-new", result.ID)
}

// ─── UpdateMemberRole ─────────────────────────────────────────────────────────

func TestProjectsService_UpdateMemberRole_Success(t *testing.T) {
	// Arrange
	client := &mockOnboardingClient{}
	svc := newProjectsService(t, client)
	ctx := context.Background()

	client.On("UpdateProjectMemberRole", ctx, mock.AnythingOfType("*onboardingv1.UpdateProjectMemberRoleRequest")).Return(
		&onboardingv1.UpdateProjectMemberRoleResponse{
			Member: &onboardingv1.ProjectMember{Id: "m1", UserId: "u1"},
		}, nil)

	// Act
	result, err := svc.UpdateMemberRole(ctx, "proj-1", "owner-id", "owner", "m1", "admin")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "m1", result.ID)
}

func TestProjectsSettingsServiceBoundaryContracts(t *testing.T) {
	t.Parallel()

	t.Run("GivenProjectsServiceWhenBoundaryImportsAreCheckedThenTransportViewsAreNotImported", func(t *testing.T) {
		// Given
		servicePath := "projects_service.go"

		// Arrange
		content, err := os.ReadFile(servicePath)
		require.NoError(t, err)
		text := string(content)

		// Act
		hasViewsImport := strings.Contains(text, "transport/http/views")
		hasContractsImport := strings.Contains(text, "services/contracts")

		// Then
		assert.False(t, hasViewsImport)
		assert.True(t, hasContractsImport)
	})

	t.Run("GivenSettingsServiceWhenBoundaryImportsAreCheckedThenTransportViewsAreNotImported", func(t *testing.T) {
		// Given
		servicePath := "settings_service.go"

		// Arrange
		content, err := os.ReadFile(servicePath)
		require.NoError(t, err)
		text := string(content)

		// Act
		hasViewsImport := strings.Contains(text, "transport/http/views")
		hasContractsImport := strings.Contains(text, "services/contracts")

		// Then
		assert.False(t, hasViewsImport)
		assert.True(t, hasContractsImport)
	})
}
