package services_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	bffinterfaces "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/interfaces"
	"github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/services"
	bffmiddleware "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/transport/http/middleware"
	apperrors "github.com/ralvescosta/costa-financial-assistant/backend/pkgs/errors"
	identityv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/identity/v1"
	paymentsv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/payments/v1"
)

func newHistoryService(t *testing.T, payments paymentsv1.PaymentsServiceClient) bffinterfaces.HistoryService {
	t.Helper()
	return services.NewHistoryService(zaptest.NewLogger(t), payments)
}

func TestHistoryService_GetTimeline_ReturnsTimeline(t *testing.T) {
	payments := &mockPaymentsClient{}
	svc := newHistoryService(t, payments)
	ctx := context.Background()

	payments.On("GetHistoryTimeline", ctx, mock.AnythingOfType("*paymentsv1.GetHistoryTimelineRequest")).Return(
		&paymentsv1.GetHistoryTimelineResponse{
			ProjectId: "proj-1",
			Months:    6,
			Entries: []*paymentsv1.MonthlyTimelineEntry{{
				Month:       "2026-03-01",
				TotalAmount: "120.50",
				BillCount:   2,
			}},
		}, nil,
	)

	result, err := svc.GetTimeline(ctx, "proj-1", 6)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "proj-1", result.ProjectID)
	assert.Equal(t, 6, result.Months)
	assert.Len(t, result.Timeline, 1)
}

func TestHistoryService_GetCategoryBreakdown_ReturnsBreakdown(t *testing.T) {
	payments := &mockPaymentsClient{}
	svc := newHistoryService(t, payments)
	ctx := context.Background()

	payments.On("GetHistoryCategoryBreakdown", ctx, mock.AnythingOfType("*paymentsv1.GetHistoryCategoryBreakdownRequest")).Return(
		&paymentsv1.GetHistoryCategoryBreakdownResponse{
			ProjectId: "proj-1",
			Months:    6,
			Entries: []*paymentsv1.CategoryBreakdownEntry{{
				Month:        "2026-03-01",
				BillTypeName: "Energy",
				TotalAmount:  "80.00",
				BillCount:    1,
			}},
		}, nil,
	)

	result, err := svc.GetCategoryBreakdown(ctx, "proj-1", 6)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "proj-1", result.ProjectID)
	assert.Len(t, result.Categories, 1)
}

func TestHistoryService_GetComplianceMetrics_ReturnsCompliance(t *testing.T) {
	payments := &mockPaymentsClient{}
	svc := newHistoryService(t, payments)
	ctx := context.Background()

	payments.On("GetHistoryCompliance", ctx, mock.AnythingOfType("*paymentsv1.GetHistoryComplianceRequest")).Return(
		&paymentsv1.GetHistoryComplianceResponse{
			ProjectId: "proj-1",
			Months:    6,
			Entries: []*paymentsv1.MonthlyComplianceEntry{{
				Month:          "2026-03-01",
				TotalBills:     3,
				PaidOnTime:     2,
				Overdue:        1,
				ComplianceRate: "66.67",
			}},
		}, nil,
	)

	result, err := svc.GetComplianceMetrics(ctx, "proj-1", 6)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "proj-1", result.ProjectID)
	assert.Len(t, result.Compliance, 1)
}

func TestHistoryService_GetTimeline_MapsGRPCError(t *testing.T) {
	payments := &mockPaymentsClient{}
	svc := newHistoryService(t, payments)
	ctx := context.Background()

	payments.On("GetHistoryTimeline", ctx, mock.AnythingOfType("*paymentsv1.GetHistoryTimelineRequest")).Return(nil, status.Error(codes.Unavailable, "payments unavailable"))

	result, err := svc.GetTimeline(ctx, "proj-1", 6)

	assert.Nil(t, result)
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperrors.CategoryDependencyGRPC, appErr.Category)
}

func TestHistoryService_GetTimeline_ForwardsSessionClaims(t *testing.T) {
	payments := &mockPaymentsClient{}
	svc := newHistoryService(t, payments)
	ctx := context.WithValue(context.Background(), bffmiddleware.ProjectContextKey, &identityv1.JwtClaims{
		Subject:   "user-1",
		ProjectId: "proj-1",
		Role:      "write",
		Email:     "ralvescosta@local.dev",
		Username:  "ralvescosta",
	})

	var capturedReq *paymentsv1.GetHistoryTimelineRequest
	payments.On("GetHistoryTimeline", ctx, mock.MatchedBy(func(req *paymentsv1.GetHistoryTimelineRequest) bool {
		capturedReq = req
		return true
	})).Return(&paymentsv1.GetHistoryTimelineResponse{}, nil)

	_, err := svc.GetTimeline(ctx, "proj-1", 6)

	require.NoError(t, err)
	require.NotNil(t, capturedReq)
	require.NotNil(t, capturedReq.GetSession())
	assert.Equal(t, "ralvescosta@local.dev", capturedReq.GetSession().GetEmail())
}
