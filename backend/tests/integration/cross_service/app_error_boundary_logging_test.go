package cross_service

import (
	"context"
	nativeerrors "errors"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"

	bffservices "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/services"
	billsservices "github.com/ralvescosta/costa-financial-assistant/backend/internals/bills/services"
	identitygrpc "github.com/ralvescosta/costa-financial-assistant/backend/internals/identity/transport/grpc"
	onboardinggrpc "github.com/ralvescosta/costa-financial-assistant/backend/internals/onboarding/transport/grpc"
	billsv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/bills/v1"
	commonv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/common/v1"
	identityv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/identity/v1"
	onboardingv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/onboarding/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type billsRepoBoundaryErrStub struct{}

func (billsRepoBoundaryErrStub) GetDashboardEntries(context.Context, string, string, string, int32, string) ([]*billsv1.PaymentDashboardEntry, string, error) {
	return nil, "", nativeerrors.New("dashboard query failed")
}
func (billsRepoBoundaryErrStub) MarkPaid(context.Context, string, string, string) (*billsv1.BillRecord, error) {
	return nil, nil
}
func (billsRepoBoundaryErrStub) FindIdempotencyKey(context.Context, string) (string, error) {
	return "", nil
}
func (billsRepoBoundaryErrStub) StoreIdempotencyKey(context.Context, string, string, string) error {
	return nil
}
func (billsRepoBoundaryErrStub) GetBill(context.Context, string, string) (*billsv1.BillRecord, error) {
	return nil, nil
}
func (billsRepoBoundaryErrStub) ListBills(context.Context, string, billsv1.PaymentStatus, int32, string) ([]*billsv1.BillRecord, string, error) {
	return nil, "", nil
}

type identitySvcBoundaryErrStub struct{}

func (identitySvcBoundaryErrStub) IssueBootstrapToken(context.Context, string, string, string) (string, int64, error) {
	return "", 0, nativeerrors.New("token signer unavailable")
}
func (identitySvcBoundaryErrStub) ValidateToken(context.Context, string) (bool, *identityv1.JwtClaims, error) {
	return false, nil, nil
}
func (identitySvcBoundaryErrStub) GetJwksMetadata(context.Context) (*identityv1.JwksMetadata, error) {
	return nil, nil
}

type onboardingSvcBoundaryErrStub struct{}

func (onboardingSvcBoundaryErrStub) CreateProject(context.Context, string, string, onboardingv1.ProjectType) (*onboardingv1.Project, error) {
	return nil, nativeerrors.New("onboarding db unavailable")
}
func (onboardingSvcBoundaryErrStub) GetProject(context.Context, string) (*onboardingv1.Project, error) {
	return nil, nil
}
func (onboardingSvcBoundaryErrStub) InviteProjectMember(context.Context, string, string, onboardingv1.ProjectMemberRole, string) (*onboardingv1.ProjectMember, error) {
	return nil, nil
}
func (onboardingSvcBoundaryErrStub) UpdateProjectMemberRole(context.Context, string, string, onboardingv1.ProjectMemberRole) (*onboardingv1.ProjectMember, error) {
	return nil, nil
}
func (onboardingSvcBoundaryErrStub) ListProjectMembers(context.Context, string, int32, string) ([]*onboardingv1.ProjectMember, string, error) {
	return nil, "", nil
}

func TestBoundaryLogging_TableDriven_T064(t *testing.T) {
	tests := []struct {
		name          string
		run           func(context.Context, *zap.Logger) error
		expectedMsg   string
		expectedField string
		expectedValue string
	}{
		{
			name: "Given BFF boundary failure When service call fails Then one structured boundary log is emitted",
			run: func(ctx context.Context, logger *zap.Logger) error {
				svc := bffservices.NewHistoryService(logger)
				_, err := svc.GetTimeline(ctx, "project-1", 1)
				return err
			},
			expectedMsg:   "history_svc: get timeline failed",
			expectedField: "project_id",
			expectedValue: "project-1",
		},
		{
			name: "Given bills boundary failure When dashboard call fails Then one structured boundary log is emitted",
			run: func(ctx context.Context, logger *zap.Logger) error {
				svc := billsservices.NewBillPaymentService(billsRepoBoundaryErrStub{}, logger)
				_, _, err := svc.GetPaymentDashboard(ctx, "project-1", "2025-01-01", "2025-01-31", 10, "")
				return err
			},
			expectedMsg:   "bill_payment_service: get dashboard failed",
			expectedField: "project_id",
			expectedValue: "project-1",
		},
		{
			name: "Given identity boundary failure When token issue fails Then one structured boundary log is emitted",
			run: func(ctx context.Context, logger *zap.Logger) error {
				server := identitygrpc.NewServer(identitySvcBoundaryErrStub{}, logger)
				_, err := server.IssueBootstrapToken(ctx, &identityv1.IssueBootstrapTokenRequest{
					UserId:    "user-1",
					ProjectId: "project-1",
					Role:      "owner",
				})
				return err
			},
			expectedMsg: "IssueBootstrapToken failed",
		},
		{
			name: "Given onboarding boundary failure When create project fails Then one structured boundary log is emitted",
			run: func(ctx context.Context, logger *zap.Logger) error {
				_, server := onboardinggrpc.NewServer(onboardingSvcBoundaryErrStub{}, logger)
				_, err := server.CreateProject(ctx, &onboardingv1.CreateProjectRequest{
					Ctx:  &commonv1.ProjectContext{UserId: "user-1"},
					Name: "Test project",
					Type: onboardingv1.ProjectType_PROJECT_TYPE_PERSONAL,
				})
				return err
			},
			expectedMsg:   "grpc.CreateProject failed",
			expectedField: "owner_id",
			expectedValue: "user-1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			core, observed := observer.New(zap.ErrorLevel)
			logger := zap.New(core)
			ctx := context.Background()

			// Act
			err := tt.run(ctx, logger)

			// Assert
			require.Error(t, err)
			require.Equal(t, 1, observed.Len())
			entry := observed.All()[0]
			assert.Equal(t, tt.expectedMsg, entry.Message)
			_, hasError := entry.ContextMap()["error"]
			assert.True(t, hasError)
			if tt.expectedField != "" {
				assert.Equal(t, tt.expectedValue, entry.ContextMap()[tt.expectedField])
			}
		})
	}
}
