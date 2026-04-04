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
	apperrors "github.com/ralvescosta/costa-financial-assistant/backend/pkgs/errors"
	billsv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/bills/v1"
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

// ─── helpers ──────────────────────────────────────────────────────────────────

func newPaymentsService(t *testing.T, bills billsv1.BillsServiceClient) bffinterfaces.PaymentsService {
	t.Helper()
	return services.NewPaymentsService(zaptest.NewLogger(t), bills)
}

// ─── GetPaymentDashboard ──────────────────────────────────────────────────────

func TestPaymentsService_GetPaymentDashboard_ReturnsDashboard(t *testing.T) {
	bills := &mockBillsClient{}
	svc := newPaymentsService(t, bills)
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
	svc := newPaymentsService(t, bills)
	ctx := context.Background()

	bills.On("GetPaymentDashboard", ctx, mock.Anything).Return(nil, errors.New("bills unavailable"))

	result, err := svc.GetPaymentDashboard(ctx, "proj-1", "user-1", "", "", 0, "")

	assert.Nil(t, result)
	assert.Error(t, err)
}

// ─── MarkBillPaid ─────────────────────────────────────────────────────────────

func TestPaymentsService_MarkBillPaid_Success(t *testing.T) {
	bills := &mockBillsClient{}
	svc := newPaymentsService(t, bills)
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

func TestPaymentsService_GetCyclePreference_ReturnsDependencyError(t *testing.T) {
	bills := &mockBillsClient{}
	svc := newPaymentsService(t, bills)
	ctx := context.Background()

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
	svc := newPaymentsService(t, bills)
	ctx := context.Background()

	result, err := svc.SetCyclePreference(ctx, "proj-1", 0, "user-1")

	assert.Nil(t, result)
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperrors.CategoryValidation, appErr.Category)
}

func TestPaymentsService_SetCyclePreference_ReturnsDependencyError(t *testing.T) {
	bills := &mockBillsClient{}
	svc := newPaymentsService(t, bills)
	ctx := context.Background()

	result, err := svc.SetCyclePreference(ctx, "proj-1", 15, "user-1")

	assert.Nil(t, result)
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperrors.CategoryDependencyGRPC, appErr.Category)
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
