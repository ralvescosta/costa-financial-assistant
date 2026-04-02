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

	bffinterfaces "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/interfaces"
	"github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/services"
	paymentsinterfaces "github.com/ralvescosta/costa-financial-assistant/backend/internals/payments/interfaces"
)

// ─── mock: ReconciliationService ─────────────────────────────────────────────

type mockReconciliationSvc struct{ mock.Mock }

func (m *mockReconciliationSvc) AutoReconcile(ctx context.Context, projectID, statementID string) (*paymentsinterfaces.ReconciliationSummary, error) {
	args := m.Called(ctx, projectID, statementID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*paymentsinterfaces.ReconciliationSummary), args.Error(1)
}

func (m *mockReconciliationSvc) GetSummary(ctx context.Context, projectID, periodStart, periodEnd string) (*paymentsinterfaces.ReconciliationSummary, error) {
	args := m.Called(ctx, projectID, periodStart, periodEnd)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*paymentsinterfaces.ReconciliationSummary), args.Error(1)
}

func (m *mockReconciliationSvc) CreateManualLink(ctx context.Context, projectID, transactionLineID, billRecordID, linkedBy string) (*paymentsinterfaces.ReconciliationLink, error) {
	args := m.Called(ctx, projectID, transactionLineID, billRecordID, linkedBy)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*paymentsinterfaces.ReconciliationLink), args.Error(1)
}

// ─── helper ───────────────────────────────────────────────────────────────────

func newReconciliationService(t *testing.T, svc paymentsinterfaces.ReconciliationService) bffinterfaces.ReconciliationService {
	t.Helper()
	return services.NewReconciliationService(zaptest.NewLogger(t), svc)
}

// ─── GetSummary ───────────────────────────────────────────────────────────────

func TestReconciliationService_GetSummary_ReturnsMappedEntries(t *testing.T) {
	// Arrange
	reconSvc := &mockReconciliationSvc{}
	svc := newReconciliationService(t, reconSvc)
	ctx := context.Background()

	linkType := paymentsinterfaces.ReconciliationLinkTypeAuto
	reconSvc.On("GetSummary", ctx, "proj-1", "2024-01-01", "2024-01-31").Return(
		&paymentsinterfaces.ReconciliationSummary{
			ProjectID:   "proj-1",
			PeriodStart: "2024-01-01",
			PeriodEnd:   "2024-01-31",
			Entries: []paymentsinterfaces.ReconciliationSummaryEntry{
				{
					TransactionLineID:    "txn-1",
					TransactionDate:      "2024-01-10",
					Description:          "Energy bill",
					Amount:               "150.00",
					Direction:            "debit",
					ReconciliationStatus: paymentsinterfaces.TransactionMatchedAuto,
					LinkType:             &linkType,
				},
			},
		}, nil)

	// Act
	result, err := svc.GetSummary(ctx, "proj-1", "2024-01-01", "2024-01-31")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "proj-1", result.ProjectID)
	assert.Len(t, result.Entries, 1)
	assert.Equal(t, "txn-1", result.Entries[0].TransactionLineID)
	assert.Equal(t, "matched_auto", result.Entries[0].ReconciliationStatus)
	reconSvc.AssertExpectations(t)
}

func TestReconciliationService_GetSummary_ClientError(t *testing.T) {
	// Arrange
	reconSvc := &mockReconciliationSvc{}
	svc := newReconciliationService(t, reconSvc)
	ctx := context.Background()

	reconSvc.On("GetSummary", ctx, "proj-1", "", "").Return(nil, errors.New("db unavailable"))

	// Act
	result, err := svc.GetSummary(ctx, "proj-1", "", "")

	// Assert
	assert.Nil(t, result)
	assert.Error(t, err)
}

// ─── CreateManualLink ─────────────────────────────────────────────────────────

func TestReconciliationService_CreateManualLink_Success(t *testing.T) {
	// Arrange
	reconSvc := &mockReconciliationSvc{}
	svc := newReconciliationService(t, reconSvc)
	ctx := context.Background()

	linkedBy := "user-1"
	reconSvc.On("CreateManualLink", ctx, "proj-1", "txn-1", "bill-1", "user-1").Return(
		&paymentsinterfaces.ReconciliationLink{
			ID:                "link-1",
			ProjectID:         "proj-1",
			TransactionLineID: "txn-1",
			BillRecordID:      "bill-1",
			LinkType:          paymentsinterfaces.ReconciliationLinkTypeManual,
			LinkedBy:          &linkedBy,
			CreatedAt:         time.Now(),
		}, nil)

	// Act
	result, err := svc.CreateManualLink(ctx, "proj-1", "txn-1", "bill-1", "user-1")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "link-1", result.ID)
	assert.Equal(t, "manual", result.LinkType)
	reconSvc.AssertExpectations(t)
}

func TestReconciliationService_CreateManualLink_ClientError(t *testing.T) {
	// Arrange
	reconSvc := &mockReconciliationSvc{}
	svc := newReconciliationService(t, reconSvc)
	ctx := context.Background()

	reconSvc.On("CreateManualLink", ctx, "proj-1", "txn-1", "bill-1", "user-1").Return(nil, errors.New("conflict"))

	// Act
	result, err := svc.CreateManualLink(ctx, "proj-1", "txn-1", "bill-1", "user-1")

	// Assert
	assert.Nil(t, result)
	assert.Error(t, err)
}
