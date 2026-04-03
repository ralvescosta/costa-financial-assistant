package services

import (
	"context"

	"go.uber.org/zap"

	bffinterfaces "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/interfaces"
	views "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/transport/http/views"
	paymentsinterfaces "github.com/ralvescosta/costa-financial-assistant/backend/internals/payments/interfaces"
	apperrors "github.com/ralvescosta/costa-financial-assistant/backend/pkgs/errors"
)

// ReconciliationServiceImpl implements bffinterfaces.ReconciliationService.
type ReconciliationServiceImpl struct {
	logger   *zap.Logger
	reconSvc paymentsinterfaces.ReconciliationService
}

// NewReconciliationService constructs a ReconciliationServiceImpl.
func NewReconciliationService(logger *zap.Logger, reconSvc paymentsinterfaces.ReconciliationService) bffinterfaces.ReconciliationService {
	return &ReconciliationServiceImpl{logger: logger, reconSvc: reconSvc}
}

// GetSummary returns the reconciliation summary for the project and period.
func (s *ReconciliationServiceImpl) GetSummary(ctx context.Context, projectID, periodStart, periodEnd string) (*views.ReconciliationSummaryResponse, error) {
	summary, err := s.reconSvc.GetSummary(ctx, projectID, periodStart, periodEnd)
	if err != nil {
		s.logger.Error("reconciliation_svc: get summary failed",
			zap.String("project_id", projectID),
			zap.Error(err))
		if appErr := apperrors.AsAppError(err); appErr != nil {
			return nil, appErr
		}
		return nil, apperrors.TranslateError(err, "service")
	}

	entries := make([]*views.ReconciliationEntryResponse, 0, len(summary.Entries))
	for _, e := range summary.Entries {
		entry := views.ReconciliationEntryResponse{
			TransactionLineID:    e.TransactionLineID,
			TransactionDate:      e.TransactionDate,
			Description:          e.Description,
			Amount:               e.Amount,
			Direction:            e.Direction,
			ReconciliationStatus: string(e.ReconciliationStatus),
			LinkedBillID:         e.LinkedBillID,
			LinkedBillDueDate:    e.LinkedBillDueDate,
			LinkedBillAmount:     e.LinkedBillAmount,
		}
		if e.LinkType != nil {
			lt := string(*e.LinkType)
			entry.LinkType = &lt
		}
		entries = append(entries, &entry)
	}
	return &views.ReconciliationSummaryResponse{
		ProjectID:   summary.ProjectID,
		PeriodStart: summary.PeriodStart,
		PeriodEnd:   summary.PeriodEnd,
		Entries:     entries,
	}, nil
}

// CreateManualLink manually links a statement transaction to a bill record.
func (s *ReconciliationServiceImpl) CreateManualLink(ctx context.Context, projectID, transactionLineID, billRecordID, linkedBy string) (*views.ReconciliationLinkResponse, error) {
	link, err := s.reconSvc.CreateManualLink(ctx, projectID, transactionLineID, billRecordID, linkedBy)
	if err != nil {
		s.logger.Error("reconciliation_svc: create manual link failed",
			zap.String("project_id", projectID),
			zap.String("transaction_line_id", transactionLineID),
			zap.String("bill_record_id", billRecordID),
			zap.Error(err))
		if appErr := apperrors.AsAppError(err); appErr != nil {
			return nil, appErr
		}
		return nil, apperrors.TranslateError(err, "service")
	}
	return &views.ReconciliationLinkResponse{
		ID:                link.ID,
		ProjectID:         link.ProjectID,
		TransactionLineID: link.TransactionLineID,
		BillRecordID:      link.BillRecordID,
		LinkType:          string(link.LinkType),
		LinkedBy:          link.LinkedBy,
		CreatedAt:         link.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}, nil
}
