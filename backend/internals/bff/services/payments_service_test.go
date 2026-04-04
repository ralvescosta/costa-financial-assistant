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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	bffinterfaces "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/interfaces"
	"github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/services"
	bffmiddleware "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/transport/http/middleware"
	apperrors "github.com/ralvescosta/costa-financial-assistant/backend/pkgs/errors"
	billsv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/bills/v1"
	identityv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/identity/v1"
	paymentsv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/payments/v1"
)

// ─── mock: BillsServiceClient ─────────────────────────────────────────────────

type mockBillsClient struct{ mock.Mock }

func (m *mockBillsClient) GetPaymentDashboard(ctx context.Context, in *billsv1.GetPaymentDashboardRequest, opts ...grpc.CallOption) (*billsv1.GetPaymentDashboardResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*billsv1.GetPaymentDashboardResponse), args.Error(1)
}

func (m *mockBillsClient) MarkBillPaid(ctx context.Context, in *billsv1.MarkBillPaidRequest, opts ...grpc.CallOption) (*billsv1.MarkBillPaidResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*billsv1.MarkBillPaidResponse), args.Error(1)
}

func (m *mockBillsClient) GetBill(ctx context.Context, in *billsv1.GetBillRequest, opts ...grpc.CallOption) (*billsv1.GetBillResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*billsv1.GetBillResponse), args.Error(1)
}

func (m *mockBillsClient) ListBills(ctx context.Context, in *billsv1.ListBillsRequest, opts ...grpc.CallOption) (*billsv1.ListBillsResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*billsv1.ListBillsResponse), args.Error(1)
}

// ─── mock: PaymentsServiceClient ──────────────────────────────────────────────

type mockPaymentsClient struct{ mock.Mock }

func (m *mockPaymentsClient) GetCyclePreference(ctx context.Context, in *paymentsv1.GetCyclePreferenceRequest, opts ...grpc.CallOption) (*paymentsv1.GetCyclePreferenceResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*paymentsv1.GetCyclePreferenceResponse), args.Error(1)
}

func (m *mockPaymentsClient) SetCyclePreference(ctx context.Context, in *paymentsv1.SetCyclePreferenceRequest, opts ...grpc.CallOption) (*paymentsv1.SetCyclePreferenceResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*paymentsv1.SetCyclePreferenceResponse), args.Error(1)
}

func (m *mockPaymentsClient) GetHistoryTimeline(ctx context.Context, in *paymentsv1.GetHistoryTimelineRequest, opts ...grpc.CallOption) (*paymentsv1.GetHistoryTimelineResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*paymentsv1.GetHistoryTimelineResponse), args.Error(1)
}

func (m *mockPaymentsClient) GetHistoryCategoryBreakdown(ctx context.Context, in *paymentsv1.GetHistoryCategoryBreakdownRequest, opts ...grpc.CallOption) (*paymentsv1.GetHistoryCategoryBreakdownResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*paymentsv1.GetHistoryCategoryBreakdownResponse), args.Error(1)
}

func (m *mockPaymentsClient) GetHistoryCompliance(ctx context.Context, in *paymentsv1.GetHistoryComplianceRequest, opts ...grpc.CallOption) (*paymentsv1.GetHistoryComplianceResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*paymentsv1.GetHistoryComplianceResponse), args.Error(1)
}

func (m *mockPaymentsClient) GetReconciliationSummary(ctx context.Context, in *paymentsv1.GetReconciliationSummaryRequest, opts ...grpc.CallOption) (*paymentsv1.GetReconciliationSummaryResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*paymentsv1.GetReconciliationSummaryResponse), args.Error(1)
}

func (m *mockPaymentsClient) CreateManualLink(ctx context.Context, in *paymentsv1.CreateManualLinkRequest, opts ...grpc.CallOption) (*paymentsv1.CreateManualLinkResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*paymentsv1.CreateManualLinkResponse), args.Error(1)
}

// ─── helpers ──────────────────────────────────────────────────────────────────

func newPaymentsService(t *testing.T, bills billsv1.BillsServiceClient, payments paymentsv1.PaymentsServiceClient) bffinterfaces.PaymentsService {
	t.Helper()
	return services.NewPaymentsService(zaptest.NewLogger(t), bills, payments)
}

// ─── GetPaymentDashboard ──────────────────────────────────────────────────────

func TestPaymentsService_GetPaymentDashboard_ReturnsDashboard(t *testing.T) {
	bills := &mockBillsClient{}
	payments := &mockPaymentsClient{}
	svc := newPaymentsService(t, bills, payments)
	ctx := context.Background()

	bills.On("GetPaymentDashboard", ctx, mock.AnythingOfType("*billsv1.GetPaymentDashboardRequest")).Return(
		&billsv1.GetPaymentDashboardResponse{
			Entries: []*billsv1.PaymentDashboardEntry{
				{Bill: &billsv1.BillRecord{Id: "b1"}, IsOverdue: false, DaysUntilDue: 5},
			},
		}, nil)

	result, err := svc.GetPaymentDashboard(ctx, "proj-1", "user-1", "", "", 10, "")

	require.NoError(t, err)
	assert.Len(t, result.Entries, 1)
	assert.Equal(t, "b1", result.Entries[0].Bill.ID)
	bills.AssertExpectations(t)
}

func TestPaymentsService_GetPaymentDashboard_ClientError(t *testing.T) {
	bills := &mockBillsClient{}
	payments := &mockPaymentsClient{}
	svc := newPaymentsService(t, bills, payments)
	ctx := context.Background()

	bills.On("GetPaymentDashboard", ctx, mock.Anything).Return(nil, errors.New("bills unavailable"))

	result, err := svc.GetPaymentDashboard(ctx, "proj-1", "user-1", "", "", 0, "")

	assert.Nil(t, result)
	assert.Error(t, err)
}

func TestPaymentsService_GetPaymentDashboard_ForwardsSessionAndDefaultPagination(t *testing.T) {
	bills := &mockBillsClient{}
	payments := &mockPaymentsClient{}
	svc := newPaymentsService(t, bills, payments)
	ctx := context.WithValue(context.Background(), bffmiddleware.ProjectContextKey, &identityv1.JwtClaims{
		Subject:   "user-1",
		ProjectId: "proj-1",
		Role:      "write",
		Email:     "ralvescosta@local.dev",
		Username:  "ralvescosta",
	})

	var capturedReq *billsv1.GetPaymentDashboardRequest
	bills.On("GetPaymentDashboard", ctx, mock.MatchedBy(func(req *billsv1.GetPaymentDashboardRequest) bool {
		capturedReq = req
		return true
	})).Return(&billsv1.GetPaymentDashboardResponse{}, nil)

	_, err := svc.GetPaymentDashboard(ctx, "proj-1", "user-1", "", "", 0, "")

	require.NoError(t, err)
	require.NotNil(t, capturedReq)
	require.NotNil(t, capturedReq.GetSession())
	assert.Equal(t, "user-1", capturedReq.GetSession().GetId())
	assert.EqualValues(t, 20, capturedReq.GetPagination().GetPageSize())
}

// ─── MarkBillPaid ─────────────────────────────────────────────────────────────

func TestPaymentsService_MarkBillPaid_Success(t *testing.T) {
	bills := &mockBillsClient{}
	payments := &mockPaymentsClient{}
	svc := newPaymentsService(t, bills, payments)
	ctx := context.Background()

	bills.On("MarkBillPaid", ctx, mock.AnythingOfType("*billsv1.MarkBillPaidRequest")).Return(
		&billsv1.MarkBillPaidResponse{
			Bill: &billsv1.BillRecord{Id: "b1", PaymentStatus: billsv1.PaymentStatus_PAYMENT_STATUS_PAID},
		}, nil)

	result, err := svc.MarkBillPaid(ctx, "proj-1", "b1", "user-1")

	require.NoError(t, err)
	assert.Equal(t, "b1", result.Bill.ID)
}

// ─── GetCyclePreference ───────────────────────────────────────────────────────

func TestPaymentsService_GetCyclePreference_ReturnsPreference(t *testing.T) {
	bills := &mockBillsClient{}
	payments := &mockPaymentsClient{}
	svc := newPaymentsService(t, bills, payments)
	ctx := context.Background()

	payments.On("GetCyclePreference", ctx, mock.AnythingOfType("*paymentsv1.GetCyclePreferenceRequest")).Return(
		&paymentsv1.GetCyclePreferenceResponse{
			Preference: &paymentsv1.CyclePreference{
				ProjectId:           "proj-1",
				PreferredDayOfMonth: 15,
				UpdatedAt:           "2026-04-04T10:00:00Z",
			},
		}, nil,
	)

	result, err := svc.GetCyclePreference(ctx, "proj-1")

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "proj-1", result.ProjectID)
	assert.Equal(t, 15, result.PreferredDayOfMonth)
	assert.Equal(t, "2026-04-04T10:00:00Z", result.UpdatedAt)
}

func TestPaymentsService_GetCyclePreference_MapsGRPCError(t *testing.T) {
	bills := &mockBillsClient{}
	payments := &mockPaymentsClient{}
	svc := newPaymentsService(t, bills, payments)
	ctx := context.Background()

	payments.On("GetCyclePreference", ctx, mock.AnythingOfType("*paymentsv1.GetCyclePreferenceRequest")).Return(nil, status.Error(codes.Unavailable, "payments unavailable"))

	result, err := svc.GetCyclePreference(ctx, "proj-1")

	assert.Nil(t, result)
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperrors.CategoryDependencyGRPC, appErr.Category)
}

// ─── SetCyclePreference ───────────────────────────────────────────────────────

func TestPaymentsService_SetCyclePreference_InvalidDayReturnsValidationError(t *testing.T) {
	bills := &mockBillsClient{}
	payments := &mockPaymentsClient{}
	svc := newPaymentsService(t, bills, payments)
	ctx := context.Background()

	result, err := svc.SetCyclePreference(ctx, "proj-1", 0, "user-1")

	assert.Nil(t, result)
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperrors.CategoryValidation, appErr.Category)
}

func TestPaymentsService_SetCyclePreference_ReturnsPreference(t *testing.T) {
	bills := &mockBillsClient{}
	payments := &mockPaymentsClient{}
	svc := newPaymentsService(t, bills, payments)
	ctx := context.Background()

	payments.On("SetCyclePreference", ctx, mock.AnythingOfType("*paymentsv1.SetCyclePreferenceRequest")).Return(
		&paymentsv1.SetCyclePreferenceResponse{
			Preference: &paymentsv1.CyclePreference{
				ProjectId:           "proj-1",
				PreferredDayOfMonth: 12,
				UpdatedAt:           "2026-04-04T11:00:00Z",
			},
		}, nil,
	)

	result, err := svc.SetCyclePreference(ctx, "proj-1", 12, "user-1")

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "proj-1", result.ProjectID)
	assert.Equal(t, 12, result.PreferredDayOfMonth)
}

func TestPaymentsReconciliationServiceBoundaryContracts(t *testing.T) {
	t.Parallel()

	t.Run("GivenPaymentsServiceWhenBoundaryImportsAreCheckedThenTransportViewsAreNotImported", func(t *testing.T) {
		servicePath := "payments_service.go"

		content, err := os.ReadFile(servicePath)
		require.NoError(t, err)
		text := string(content)

		hasViewsImport := strings.Contains(text, "transport/http/views")
		hasContractsImport := strings.Contains(text, "services/contracts")
		hasPaymentsDomainImport := strings.Contains(text, "internals/payments/interfaces")

		assert.False(t, hasViewsImport)
		assert.True(t, hasContractsImport)
		assert.False(t, hasPaymentsDomainImport)
	})

	t.Run("GivenReconciliationServiceWhenBoundaryImportsAreCheckedThenTransportViewsAreNotImported", func(t *testing.T) {
		servicePath := "reconciliation_service.go"

		content, err := os.ReadFile(servicePath)
		require.NoError(t, err)
		text := string(content)

		hasViewsImport := strings.Contains(text, "transport/http/views")
		hasContractsImport := strings.Contains(text, "services/contracts")

		assert.False(t, hasViewsImport)
		assert.True(t, hasContractsImport)
	})
}
