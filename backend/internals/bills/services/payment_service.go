// Package services implements use-case logic for the bills domain.
package services

import (
	"context"

	"go.uber.org/zap"

	"github.com/ralvescosta/costa-financial-assistant/backend/internals/bills/interfaces"
	apperrors "github.com/ralvescosta/costa-financial-assistant/backend/pkgs/errors"
	billsv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/bills/v1"
)

// BillPaymentService implements interfaces.BillPaymentService.
type BillPaymentService struct {
	repo   interfaces.BillPaymentRepository
	logger *zap.Logger
}

// NewBillPaymentService constructs a BillPaymentService.
func NewBillPaymentService(repo interfaces.BillPaymentRepository, logger *zap.Logger) interfaces.BillPaymentService {
	return &BillPaymentService{repo: repo, logger: logger}
}

// GetPaymentDashboard returns outstanding and overdue bills for the given cycle date range.
func (s *BillPaymentService) GetPaymentDashboard(
	ctx context.Context,
	projectID, cycleStart, cycleEnd string,
	pageSize int32,
	pageToken string,
) ([]*billsv1.PaymentDashboardEntry, string, error) {
	entries, nextToken, err := s.repo.GetDashboardEntries(ctx, projectID, cycleStart, cycleEnd, pageSize, pageToken)
	if err != nil {
		s.logger.Error("bill_payment_service: get dashboard failed",
			zap.String("project_id", projectID),
			zap.String("cycle_start", cycleStart),
			zap.String("cycle_end", cycleEnd),
			zap.Error(err))
		if appErr := apperrors.AsAppError(err); appErr != nil {
			return nil, "", appErr
		}
		return nil, "", apperrors.TranslateError(err, "service")
	}
	return entries, nextToken, nil
}

// MarkBillPaid idempotently marks a bill as paid.
// The idempotency key is scoped to the project + bill ID pair to prevent double-marking.
func (s *BillPaymentService) MarkBillPaid(ctx context.Context, projectID, billID, markedBy string) (*billsv1.BillRecord, error) {
	idempotencyKey := projectID + ":mark-paid:" + billID

	payload, err := s.repo.FindIdempotencyKey(ctx, idempotencyKey)
	if err != nil {
		s.logger.Error("bill_payment_service: idempotency check failed",
			zap.String("project_id", projectID),
			zap.String("bill_id", billID),
			zap.Error(err))
		if appErr := apperrors.AsAppError(err); appErr != nil {
			return nil, appErr
		}
		return nil, apperrors.TranslateError(err, "service")
	}
	if payload != "" {
		s.logger.Info("bill_payment_service: mark-paid idempotent hit",
			zap.String("project_id", projectID),
			zap.String("bill_id", billID))
		return s.repo.GetBill(ctx, projectID, billID)
	}

	bill, err := s.repo.MarkPaid(ctx, projectID, billID, markedBy)
	if err != nil {
		s.logger.Error("bill_payment_service: mark paid failed",
			zap.String("project_id", projectID),
			zap.String("bill_id", billID),
			zap.Error(err))
		if appErr := apperrors.AsAppError(err); appErr != nil {
			return nil, appErr
		}
		return nil, apperrors.TranslateError(err, "service")
	}

	if storeErr := s.repo.StoreIdempotencyKey(ctx, idempotencyKey, "bills", bill.GetId()); storeErr != nil {
		s.logger.Warn("bill_payment_service: failed to store idempotency key",
			zap.String("project_id", projectID),
			zap.String("bill_id", billID),
			zap.Error(storeErr))
	}

	s.logger.Info("bill_payment_service: bill marked paid",
		zap.String("project_id", projectID),
		zap.String("bill_id", billID),
		zap.String("marked_by", markedBy))

	return bill, nil
}

// GetBill returns a single bill record by ID, scoped to the project.
func (s *BillPaymentService) GetBill(ctx context.Context, projectID, billID string) (*billsv1.BillRecord, error) {
	bill, err := s.repo.GetBill(ctx, projectID, billID)
	if err != nil {
		s.logger.Error("bill_payment_service: get bill failed",
			zap.String("project_id", projectID),
			zap.String("bill_id", billID),
			zap.Error(err))
		if appErr := apperrors.AsAppError(err); appErr != nil {
			return nil, appErr
		}
		return nil, apperrors.TranslateError(err, "service")
	}
	return bill, nil
}

// ListBills returns project-scoped bill records filtered by optional payment status.
func (s *BillPaymentService) ListBills(
	ctx context.Context,
	projectID string,
	status billsv1.PaymentStatus,
	pageSize int32,
	pageToken string,
) ([]*billsv1.BillRecord, string, error) {
	bills, nextToken, err := s.repo.ListBills(ctx, projectID, status, pageSize, pageToken)
	if err != nil {
		s.logger.Error("bill_payment_service: list bills failed",
			zap.String("project_id", projectID),
			zap.Error(err))
		if appErr := apperrors.AsAppError(err); appErr != nil {
			return nil, "", appErr
		}
		return nil, "", apperrors.TranslateError(err, "service")
	}
	return bills, nextToken, nil
}
