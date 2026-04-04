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
	apperrors "github.com/ralvescosta/costa-financial-assistant/backend/pkgs/errors"
	paymentsv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/payments/v1"
)

func newReconciliationService(t *testing.T, payments paymentsv1.PaymentsServiceClient) bffinterfaces.ReconciliationService {
	t.Helper()
	return services.NewReconciliationService(zaptest.NewLogger(t), payments)
}

func TestReconciliationService_GetSummary_ReturnsSummary(t *testing.T) {
	payments := &mockPaymentsClient{}
	svc := newReconciliationService(t, payments)
	ctx := context.Background()

	payments.On("GetReconciliationSummary", ctx, mock.AnythingOfType("*paymentsv1.GetReconciliationSummaryRequest")).Return(
		&paymentsv1.GetReconciliationSummaryResponse{
			Summary: &paymentsv1.ReconciliationSummary{
				ProjectId:   "proj-1",
				PeriodStart: "2024-01-01",
				PeriodEnd:   "2024-01-31",
				Entries: []*paymentsv1.ReconciliationSummaryEntry{{
					TransactionLineId:    "txn-1",
					TransactionDate:      "2024-01-10",
					Description:          "Energy bill",
					Amount:               "100.00",
					Direction:            "debit",
					ReconciliationStatus: paymentsv1.TransactionReconciliationStatus_TRANSACTION_RECONCILIATION_STATUS_MATCHED_MANUAL,
				}},
			},
		}, nil,
	)

	result, err := svc.GetSummary(ctx, "proj-1", "2024-01-01", "2024-01-31")

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "proj-1", result.ProjectID)
	assert.Len(t, result.Entries, 1)
}

func TestReconciliationService_CreateManualLink_ReturnsLink(t *testing.T) {
	payments := &mockPaymentsClient{}
	svc := newReconciliationService(t, payments)
	ctx := context.Background()
	linkedBy := "user-1"

	payments.On("CreateManualLink", ctx, mock.AnythingOfType("*paymentsv1.CreateManualLinkRequest")).Return(
		&paymentsv1.CreateManualLinkResponse{
			Link: &paymentsv1.ReconciliationLink{
				Id:                "link-1",
				ProjectId:         "proj-1",
				TransactionLineId: "txn-1",
				BillRecordId:      "bill-1",
				LinkType:          paymentsv1.ReconciliationLinkType_RECONCILIATION_LINK_TYPE_MANUAL,
				LinkedBy:          &linkedBy,
				CreatedAt:         "2026-04-04T12:00:00Z",
			},
		}, nil,
	)

	result, err := svc.CreateManualLink(ctx, "proj-1", "txn-1", "bill-1", "user-1")

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "link-1", result.ID)
	assert.Equal(t, "manual", result.LinkType)
	assert.NotNil(t, result.LinkedBy)
	assert.Equal(t, "user-1", *result.LinkedBy)
}

func TestReconciliationService_CreateManualLink_MapsGRPCConflict(t *testing.T) {
	payments := &mockPaymentsClient{}
	svc := newReconciliationService(t, payments)
	ctx := context.Background()

	payments.On("CreateManualLink", ctx, mock.AnythingOfType("*paymentsv1.CreateManualLinkRequest")).Return(nil, status.Error(codes.AlreadyExists, "already linked"))

	result, err := svc.CreateManualLink(ctx, "proj-1", "txn-1", "bill-1", "user-1")

	assert.Nil(t, result)
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperrors.CategoryConflict, appErr.Category)
}
