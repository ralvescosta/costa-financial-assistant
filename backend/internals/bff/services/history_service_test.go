package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	bffinterfaces "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/interfaces"
	"github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/services"
	paymentsinterfaces "github.com/ralvescosta/costa-financial-assistant/backend/internals/payments/interfaces"
)

// ─── mock: HistoryRepository ──────────────────────────────────────────────────

type mockHistoryRepo struct{ mock.Mock }

func (m *mockHistoryRepo) GetTimeline(ctx context.Context, projectID string, months int) ([]paymentsinterfaces.MonthlyTimelineEntry, error) {
	args := m.Called(ctx, projectID, months)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]paymentsinterfaces.MonthlyTimelineEntry), args.Error(1)
}

func (m *mockHistoryRepo) GetCategoryBreakdown(ctx context.Context, projectID string, months int) ([]paymentsinterfaces.CategoryBreakdownEntry, error) {
	args := m.Called(ctx, projectID, months)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]paymentsinterfaces.CategoryBreakdownEntry), args.Error(1)
}

func (m *mockHistoryRepo) GetComplianceMetrics(ctx context.Context, projectID string, months int) ([]paymentsinterfaces.MonthlyComplianceEntry, error) {
	args := m.Called(ctx, projectID, months)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]paymentsinterfaces.MonthlyComplianceEntry), args.Error(1)
}

// ─── helper ───────────────────────────────────────────────────────────────────

func newHistoryService(t *testing.T, repo paymentsinterfaces.HistoryRepository) bffinterfaces.HistoryService {
	t.Helper()
	return services.NewHistoryService(zaptest.NewLogger(t), repo)
}

// ─── GetTimeline ──────────────────────────────────────────────────────────────

func TestHistoryService_GetTimeline_ReturnsMappedRows(t *testing.T) {
	// Arrange
	repo := &mockHistoryRepo{}
	svc := newHistoryService(t, repo)
	ctx := context.Background()

	repo.On("GetTimeline", ctx, "proj-1", 6).Return(
[]paymentsinterfaces.MonthlyTimelineEntry{
{Month: "2024-01-01", TotalAmount: "1200.00", BillCount: 4},
{Month: "2024-02-01", TotalAmount: "950.00", BillCount: 3},
}, nil)

	// Act
	result, err := svc.GetTimeline(ctx, "proj-1", 6)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "proj-1", result.ProjectID)
	assert.Equal(t, 6, result.Months)
	assert.Len(t, result.Timeline, 2)
	assert.Equal(t, "2024-01-01", result.Timeline[0].Month)
	assert.Equal(t, "1200.00", result.Timeline[0].TotalAmount)
	repo.AssertExpectations(t)
}

func TestHistoryService_GetTimeline_NegativeMonthsDefaultsTo12(t *testing.T) {
	// Arrange - months=-1 is coerced to 12 before the repo call
	repo := &mockHistoryRepo{}
	svc := newHistoryService(t, repo)
	ctx := context.Background()

	repo.On("GetTimeline", ctx, "proj-1", 12).Return(
[]paymentsinterfaces.MonthlyTimelineEntry{}, nil)

	// Act
	result, err := svc.GetTimeline(ctx, "proj-1", -1)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 12, result.Months)
	repo.AssertExpectations(t)
}

func TestHistoryService_GetTimeline_ZeroMonthsMeansAllHistory(t *testing.T) {
	// Arrange - months=0 is "all history"; passes through to repo unchanged
	repo := &mockHistoryRepo{}
	svc := newHistoryService(t, repo)
	ctx := context.Background()

	repo.On("GetTimeline", ctx, "proj-1", 0).Return(
[]paymentsinterfaces.MonthlyTimelineEntry{}, nil)

	// Act
	result, err := svc.GetTimeline(ctx, "proj-1", 0)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 0, result.Months)
	repo.AssertExpectations(t)
}

func TestHistoryService_GetTimeline_RepoError(t *testing.T) {
	// Arrange
	repo := &mockHistoryRepo{}
	svc := newHistoryService(t, repo)
	ctx := context.Background()

	repo.On("GetTimeline", ctx, "proj-1", 3).Return(nil, errors.New("db timeout"))

	// Act
	result, err := svc.GetTimeline(ctx, "proj-1", 3)

	// Assert
	assert.Nil(t, result)
	assert.Error(t, err)
}

// ─── GetCategoryBreakdown ─────────────────────────────────────────────────────

func TestHistoryService_GetCategoryBreakdown_ReturnsMappedRows(t *testing.T) {
	// Arrange
	repo := &mockHistoryRepo{}
	svc := newHistoryService(t, repo)
	ctx := context.Background()

	repo.On("GetCategoryBreakdown", ctx, "proj-1", 6).Return(
[]paymentsinterfaces.CategoryBreakdownEntry{
{Month: "2024-01-01", BillTypeName: "Energy", TotalAmount: "300.00", BillCount: 2},
}, nil)

	// Act
	result, err := svc.GetCategoryBreakdown(ctx, "proj-1", 6)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "proj-1", result.ProjectID)
	assert.Len(t, result.Categories, 1)
	assert.Equal(t, "Energy", result.Categories[0].BillTypeName)
	repo.AssertExpectations(t)
}

// ─── GetComplianceMetrics ─────────────────────────────────────────────────────

func TestHistoryService_GetComplianceMetrics_ReturnsMappedRows(t *testing.T) {
	// Arrange
	repo := &mockHistoryRepo{}
	svc := newHistoryService(t, repo)
	ctx := context.Background()

	repo.On("GetComplianceMetrics", ctx, "proj-1", 6).Return(
[]paymentsinterfaces.MonthlyComplianceEntry{
{Month: "2024-01-01", TotalBills: 5, PaidOnTime: 4, Overdue: 1, ComplianceRate: "80.00"},
}, nil)

	// Act
	result, err := svc.GetComplianceMetrics(ctx, "proj-1", 6)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "proj-1", result.ProjectID)
	assert.Len(t, result.Compliance, 1)
	assert.Equal(t, "80.00", result.Compliance[0].ComplianceRate)
	repo.AssertExpectations(t)
}
