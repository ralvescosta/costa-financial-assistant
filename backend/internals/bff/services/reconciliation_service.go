package services

import (
	"context"

	"go.uber.org/zap"

	bffinterfaces "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/interfaces"
	bffcontracts "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/services/contracts"
	apperrors "github.com/ralvescosta/costa-financial-assistant/backend/pkgs/errors"
)

const reconciliationGRPCUnavailableMessage = "reconciliation actions require a downstream payments gRPC contract before the BFF can serve this flow"

// ReconciliationServiceImpl implements bffinterfaces.ReconciliationService.
type ReconciliationServiceImpl struct {
	logger *zap.Logger
}

// NewReconciliationService constructs a ReconciliationServiceImpl.
func NewReconciliationService(logger *zap.Logger) bffinterfaces.ReconciliationService {
	return &ReconciliationServiceImpl{logger: logger}
}

func (s *ReconciliationServiceImpl) downstreamUnavailable(logMessage, projectID string, extraFields ...zap.Field) error {
	appErr := apperrors.NewWithCategory(reconciliationGRPCUnavailableMessage, apperrors.CategoryDependencyGRPC)
	fields := []zap.Field{zap.String("project_id", projectID), zap.Error(appErr)}
	fields = append(fields, extraFields...)
	s.logger.Error(logMessage, fields...)
	return appErr
}

// GetSummary returns the reconciliation summary for the project and period.
func (s *ReconciliationServiceImpl) GetSummary(ctx context.Context, projectID, periodStart, periodEnd string) (*bffcontracts.ReconciliationSummaryResponse, error) {
	_ = ctx
	_ = periodStart
	_ = periodEnd
	return nil, s.downstreamUnavailable("reconciliation_svc: get summary failed", projectID)
}

// CreateManualLink manually links a statement transaction to a bill record.
func (s *ReconciliationServiceImpl) CreateManualLink(ctx context.Context, projectID, transactionLineID, billRecordID, linkedBy string) (*bffcontracts.ReconciliationLinkResponse, error) {
	_ = ctx
	_ = linkedBy
	return nil, s.downstreamUnavailable(
		"reconciliation_svc: create manual link failed",
		projectID,
		zap.String("transaction_line_id", transactionLineID),
		zap.String("bill_record_id", billRecordID),
	)
}
