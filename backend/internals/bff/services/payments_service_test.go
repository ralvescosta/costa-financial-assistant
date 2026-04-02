package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"google.golang.org/grpc"

	bffinterfaces "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/interfaces"
	"github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/services"
	paymentsinterfaces "github.com/ralvescosta/costa-financial-assistant/backend/internals/payments/interfaces"
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

// ─── mock: PaymentCycleService ────────────────────────────────────────────────

type mockCycleService struct{ mock.Mock }

func (m *mockCycleService) GetCyclePreference(ctx context.Context, projectID string) (*paymentsinterfaces.CyclePreference, error) {
	args := m.Called(ctx, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*paymentsinterfaces.CyclePreference), args.Error(1)
}

func (m *mockCycleService) UpsertCyclePreference(ctx context.Context, projectID string, dayOfMonth int, updatedBy string) (*paymentsinterfaces.CyclePreference, error) {
	args := m.Called(ctx, projectID, dayOfMonth, updatedBy)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*paymentsinterfaces.CyclePreference), args.Error(1)
}

// ─── helpers ──────────────────────────────────────────────────────────────────

func newPaymentsService(t *testing.T, bills billsv1.BillsServiceClient, cycle paymentsinterfaces.PaymentCycleService) bffinterfaces.PaymentsService {
	t.Helper()
	return services.NewPaymentsService(zaptest.NewLogger(t), bills, cycle)
}

// ─── GetPaymentDashboard ──────────────────────────────────────────────────────

func TestPaymentsService_GetPaymentDashboard_ReturnsDashboard(t *testing.T) {
	// Arrange
	bills := &mockBillsClient{}
	cycle := &mockCycleService{}
	svc := newPaymentsService(t, bills, cycle)
	ctx := context.Background()

	bills.On("GetPaymentDashboard", ctx, mock.AnythingOfType("*billsv1.GetPaymentDashboardRequest")).Return(
		&billsv1.GetPaymentDashboardResponse{
			Entries: []*billsv1.PaymentDashboardEntry{
				{Bill: &billsv1.BillRecord{Id: "b1"}, IsOverdue: false, DaysUntilDue: 5},
			},
		}, nil)

	// Act
	result, err := svc.GetPaymentDashboard(ctx, "proj-1", "user-1", "", "", 10, "")

	// Assert
	require.NoError(t, err)
	assert.Len(t, result.Entries, 1)
	assert.Equal(t, "b1", result.Entries[0].Bill.ID)
	bills.AssertExpectations(t)
}

func TestPaymentsService_GetPaymentDashboard_ClientError(t *testing.T) {
	// Arrange
	bills := &mockBillsClient{}
	cycle := &mockCycleService{}
	svc := newPaymentsService(t, bills, cycle)
	ctx := context.Background()

	bills.On("GetPaymentDashboard", ctx, mock.Anything).Return(nil, errors.New("bills unavailable"))

	// Act
	result, err := svc.GetPaymentDashboard(ctx, "proj-1", "user-1", "", "", 0, "")

	// Assert
	assert.Nil(t, result)
	assert.Error(t, err)
}

// ─── MarkBillPaid ─────────────────────────────────────────────────────────────

func TestPaymentsService_MarkBillPaid_Success(t *testing.T) {
	// Arrange
	bills := &mockBillsClient{}
	cycle := &mockCycleService{}
	svc := newPaymentsService(t, bills, cycle)
	ctx := context.Background()

	bills.On("MarkBillPaid", ctx, mock.AnythingOfType("*billsv1.MarkBillPaidRequest")).Return(
		&billsv1.MarkBillPaidResponse{
			Bill: &billsv1.BillRecord{Id: "b1", PaymentStatus: billsv1.PaymentStatus_PAYMENT_STATUS_PAID},
		}, nil)

	// Act
	result, err := svc.MarkBillPaid(ctx, "proj-1", "b1", "user-1")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "b1", result.Bill.ID)
}

// ─── GetCyclePreference ───────────────────────────────────────────────────────

func TestPaymentsService_GetCyclePreference_ReturnsPreference(t *testing.T) {
	// Arrange
	bills := &mockBillsClient{}
	cycle := &mockCycleService{}
	svc := newPaymentsService(t, bills, cycle)
	ctx := context.Background()

	cycle.On("GetCyclePreference", ctx, "proj-1").Return(
		&paymentsinterfaces.CyclePreference{
			ProjectID:           "proj-1",
			PreferredDayOfMonth: 10,
			UpdatedAt:           time.Now(),
		}, nil)

	// Act
	result, err := svc.GetCyclePreference(ctx, "proj-1")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 10, result.PreferredDayOfMonth)
}

func TestPaymentsService_GetCyclePreference_NilWhenNotSet(t *testing.T) {
	// Arrange
	bills := &mockBillsClient{}
	cycle := &mockCycleService{}
	svc := newPaymentsService(t, bills, cycle)
	ctx := context.Background()

	cycle.On("GetCyclePreference", ctx, "proj-1").Return(nil, nil)

	// Act
	result, err := svc.GetCyclePreference(ctx, "proj-1")

	// Assert
	require.NoError(t, err)
	assert.Nil(t, result)
}

// ─── SetCyclePreference ───────────────────────────────────────────────────────

func TestPaymentsService_SetCyclePreference_InvalidDayReturnsError(t *testing.T) {
	// Arrange
	bills := &mockBillsClient{}
	cycle := &mockCycleService{}
	svc := newPaymentsService(t, bills, cycle)
	ctx := context.Background()

	// Act
	result, err := svc.SetCyclePreference(ctx, "proj-1", 0, "user-1")

	// Assert
	assert.Nil(t, result)
	assert.Error(t, err)
}

func TestPaymentsService_SetCyclePreference_Success(t *testing.T) {
	// Arrange
	bills := &mockBillsClient{}
	cycle := &mockCycleService{}
	svc := newPaymentsService(t, bills, cycle)
	ctx := context.Background()

	cycle.On("UpsertCyclePreference", ctx, "proj-1", 15, "user-1").Return(
		&paymentsinterfaces.CyclePreference{
			ProjectID:           "proj-1",
			PreferredDayOfMonth: 15,
			UpdatedAt:           time.Now(),
		}, nil)

	// Act
	result, err := svc.SetCyclePreference(ctx, "proj-1", 15, "user-1")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 15, result.PreferredDayOfMonth)
}
