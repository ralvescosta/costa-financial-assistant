// Package services implements use-case logic for the payments domain.
package services

import (
	"context"
	"errors"
	"strconv"

	"go.uber.org/zap"

	"github.com/ralvescosta/costa-financial-assistant/backend/internals/payments/interfaces"
	apperrors "github.com/ralvescosta/costa-financial-assistant/backend/pkgs/errors"
)

// ErrReconciliationConflict is returned when a (transaction, bill) pair is already linked.
var ErrReconciliationConflict = errors.New("reconciliation link already exists")

// reconciliationRepository is a local alias so callers only import the interfaces package.
type reconciliationRepository = interfaces.ReconciliationRepository

// ReconciliationService implements interfaces.ReconciliationService.
type ReconciliationService struct {
	repo   reconciliationRepository
	logger *zap.Logger
}

// NewReconciliationService constructs a ReconciliationService.
func NewReconciliationService(repo reconciliationRepository, logger *zap.Logger) interfaces.ReconciliationService {
	return &ReconciliationService{repo: repo, logger: logger}
}

// AutoReconcile attempts amount-based matching of unmatched transaction lines against
// unpaid bill records for the same project. Ambiguous matches (multiple bills with
// the same amount) are flagged accordingly rather than auto-linked.
func (s *ReconciliationService) AutoReconcile(ctx context.Context, projectID, statementID string) (*interfaces.ReconciliationSummary, error) {
	lines, err := s.repo.GetUnmatchedTransactionLines(ctx, projectID, statementID)
	if err != nil {
		s.logger.Error("reconciliation_service: get unmatched lines failed",
			zap.String("project_id", projectID),
			zap.String("statement_id", statementID),
			zap.Error(err))
		if appErr := apperrors.AsAppError(err); appErr != nil {
			return nil, appErr
		}
		return nil, apperrors.TranslateError(err, "service")
	}

	bills, err := s.repo.GetBillsForPeriod(ctx, projectID, "", "")
	if err != nil {
		s.logger.Error("reconciliation_service: get bills for period failed",
			zap.String("project_id", projectID),
			zap.Error(err))
		if appErr := apperrors.AsAppError(err); appErr != nil {
			return nil, appErr
		}
		return nil, apperrors.TranslateError(err, "service")
	}

	// Build an amount → bills index for O(1) lookup.
	billsByAmount := make(map[string][]interfaces.ReconciliationSummaryEntry, len(bills))
	for _, b := range bills {
		billsByAmount[b.Amount] = append(billsByAmount[b.Amount], b)
	}

	for _, line := range lines {
		if line.Direction != "debit" {
			continue // only debit transactions can pay bills
		}

		candidates, ok := billsByAmount[line.Amount]
		if !ok || len(candidates) == 0 {
			continue // no match
		}

		if len(candidates) > 1 {
			// Ambiguous — flag and skip auto-link
			if updateErr := s.repo.UpdateTransactionStatus(ctx, projectID, line.TransactionLineID, interfaces.TransactionAmbiguous); updateErr != nil {
				s.logger.Warn("reconciliation_service: failed to mark ambiguous",
					zap.String("transaction_line_id", line.TransactionLineID),
					zap.Error(updateErr))
			}
			continue
		}

		// Exactly one match — create auto link.
		// GetBillsForPeriod returns entries where TransactionLineID holds the bill_record.id.
		billRecordID := candidates[0].TransactionLineID
		link := interfaces.ReconciliationLink{
			ProjectID:         projectID,
			TransactionLineID: line.TransactionLineID,
			BillRecordID:      billRecordID,
			LinkType:          interfaces.ReconciliationLinkTypeAuto,
		}

		if _, createErr := s.repo.CreateLink(ctx, link); createErr != nil {
			s.logger.Warn("reconciliation_service: create auto link failed",
				zap.String("transaction_line_id", line.TransactionLineID),
				zap.String("bill_record_id", billRecordID),
				zap.Error(createErr))
		}
	}

	summary, err := s.repo.GetSummary(ctx, projectID, "", "")
	if err != nil {
		s.logger.Error("reconciliation_service: get summary after auto reconcile failed",
			zap.String("project_id", projectID),
			zap.Error(err))
		if appErr := apperrors.AsAppError(err); appErr != nil {
			return nil, appErr
		}
		return nil, apperrors.TranslateError(err, "service")
	}
	return summary, nil
}

// GetSummary returns the reconciliation view for the project and period.
func (s *ReconciliationService) GetSummary(ctx context.Context, projectID, periodStart, periodEnd string) (*interfaces.ReconciliationSummary, error) {
	summary, err := s.repo.GetSummary(ctx, projectID, periodStart, periodEnd)
	if err != nil {
		s.logger.Error("reconciliation_service: get summary failed",
			zap.String("project_id", projectID),
			zap.Error(err))
		if appErr := apperrors.AsAppError(err); appErr != nil {
			return nil, appErr
		}
		return nil, apperrors.TranslateError(err, "service")
	}
	return summary, nil
}

// CreateManualLink links a transaction line to a bill record as a user-confirmed match.
func (s *ReconciliationService) CreateManualLink(ctx context.Context, projectID, transactionLineID, billRecordID, linkedBy string) (*interfaces.ReconciliationLink, error) {
	link, err := s.repo.CreateLink(ctx, interfaces.ReconciliationLink{
		ProjectID:         projectID,
		TransactionLineID: transactionLineID,
		BillRecordID:      billRecordID,
		LinkType:          interfaces.ReconciliationLinkTypeManual,
		LinkedBy:          &linkedBy,
	})
	if err != nil {
		if errors.Is(err, ErrReconciliationConflict) {
			return nil, apperrors.NewCatalogError(apperrors.ErrConflict).WithError(err)
		}
		if appErr := apperrors.AsAppError(err); appErr != nil {
			return nil, appErr
		}
		s.logger.Error("reconciliation_service: create manual link failed",
			zap.String("project_id", projectID),
			zap.String("transaction_line_id", transactionLineID),
			zap.String("bill_record_id", billRecordID),
			zap.Error(err))
		return nil, apperrors.TranslateError(err, "service")
	}

	if updateErr := s.repo.UpdateTransactionStatus(ctx, projectID, transactionLineID, interfaces.TransactionMatchedManual); updateErr != nil {
		s.logger.Warn("reconciliation_service: update transaction status failed",
			zap.String("transaction_line_id", transactionLineID),
			zap.Error(updateErr))
	}

	return link, nil
}

// amountEqual compares two numeric string amounts for equality (ignoring trailing zeros).
func amountEqual(a, b string) bool {
	af, err1 := strconv.ParseFloat(a, 64)
	bf, err2 := strconv.ParseFloat(b, 64)
	if err1 != nil || err2 != nil {
		return a == b
	}
	return af == bf
}

var _ = amountEqual // suppress unused warning — helper available for future use
