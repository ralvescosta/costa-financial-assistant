package services

import (
	"context"
	"fmt"
	"strconv"

	"go.uber.org/zap"

	bffinterfaces "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/interfaces"
	views "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/transport/http/views"
	paymentsinterfaces "github.com/ralvescosta/costa-financial-assistant/backend/internals/payments/interfaces"
	apperrors "github.com/ralvescosta/costa-financial-assistant/backend/pkgs/errors"
	billsv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/bills/v1"
	commonv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/common/v1"
)

// PaymentsServiceImpl implements bffinterfaces.PaymentsService.
type PaymentsServiceImpl struct {
	logger       *zap.Logger
	billsClient  billsv1.BillsServiceClient
	cycleService paymentsinterfaces.PaymentCycleService
}

// NewPaymentsService constructs a PaymentsServiceImpl.
func NewPaymentsService(
	logger *zap.Logger,
	billsClient billsv1.BillsServiceClient,
	cycleService paymentsinterfaces.PaymentCycleService,
) bffinterfaces.PaymentsService {
	return &PaymentsServiceImpl{
		logger:       logger,
		billsClient:  billsClient,
		cycleService: cycleService,
	}
}

// GetPaymentDashboard returns outstanding bills for the project's active payment cycle.
func (s *PaymentsServiceImpl) GetPaymentDashboard(ctx context.Context, projectID, userID, cycleStart, cycleEnd string, pageSize int32, pageToken string) (*views.PaymentDashboardResponse, error) {
	if pageSize == 0 {
		pageSize = 20
	}
	resp, err := s.billsClient.GetPaymentDashboard(ctx, &billsv1.GetPaymentDashboardRequest{
		Ctx:        &commonv1.ProjectContext{ProjectId: projectID, UserId: userID},
		CycleStart: cycleStart,
		CycleEnd:   cycleEnd,
		Pagination: &commonv1.Pagination{PageSize: pageSize, PageToken: pageToken},
	})
	if err != nil {
		s.logger.Error("payments_svc: dashboard downstream call failed",
			zap.String("project_id", projectID),
			zap.Error(err))
		if appErr := apperrors.AsAppError(err); appErr != nil {
			return nil, appErr
		}
		return nil, apperrors.TranslateError(err, "service")
	}

	entries := make([]*views.PaymentDashboardEntryResponse, 0, len(resp.GetEntries()))
	for _, e := range resp.GetEntries() {
		entry := views.PaymentDashboardEntryResponse{
			Bill:         protoBillRecordToView(e.GetBill()),
			IsOverdue:    e.GetIsOverdue(),
			DaysUntilDue: e.GetDaysUntilDue(),
		}
		if bt := e.GetBillType(); bt != nil {
			entry.BillType = &views.PaymentBillTypeResponse{
				ID:        bt.GetId(),
				ProjectID: bt.GetProjectId(),
				Name:      bt.GetName(),
			}
		}
		entries = append(entries, &entry)
	}

	var nextToken string
	if resp.GetPagination() != nil {
		nextToken = resp.GetPagination().GetNextPageToken()
	}
	return &views.PaymentDashboardResponse{Entries: entries, NextPageToken: nextToken}, nil
}

// MarkBillPaid idempotently marks a bill as paid.
func (s *PaymentsServiceImpl) MarkBillPaid(ctx context.Context, projectID, billID, paidBy string) (*views.MarkBillPaidResponse, error) {
	resp, err := s.billsClient.MarkBillPaid(ctx, &billsv1.MarkBillPaidRequest{
		Ctx:    &commonv1.ProjectContext{ProjectId: projectID, UserId: paidBy},
		BillId: billID,
		Audit:  &commonv1.AuditMetadata{PerformedBy: paidBy},
	})
	if err != nil {
		s.logger.Error("payments_svc: mark paid downstream call failed",
			zap.String("project_id", projectID),
			zap.String("bill_id", billID),
			zap.Error(err))
		if appErr := apperrors.AsAppError(err); appErr != nil {
			return nil, appErr
		}
		return nil, apperrors.TranslateError(err, "service")
	}
	s.logger.Info("payments_svc: bill marked paid",
		zap.String("bill_id", billID),
		zap.String("project_id", projectID))
	return &views.MarkBillPaidResponse{Bill: protoBillRecordToView(resp.GetBill())}, nil
}

// GetCyclePreference returns the project's preferred payment day.
func (s *PaymentsServiceImpl) GetCyclePreference(ctx context.Context, projectID string) (*views.CyclePreferenceResponse, error) {
	pref, err := s.cycleService.GetCyclePreference(ctx, projectID)
	if err != nil {
		s.logger.Error("payments_svc: get cycle preference failed",
			zap.String("project_id", projectID),
			zap.Error(err))
		if appErr := apperrors.AsAppError(err); appErr != nil {
			return nil, appErr
		}
		return nil, apperrors.TranslateError(err, "service")
	}
	if pref == nil {
		return nil, nil
	}
	return &views.CyclePreferenceResponse{
		ProjectID:           pref.ProjectID,
		PreferredDayOfMonth: pref.PreferredDayOfMonth,
		UpdatedAt:           pref.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

// SetCyclePreference creates or updates the project's preferred payment day.
func (s *PaymentsServiceImpl) SetCyclePreference(ctx context.Context, projectID string, dayOfMonth int, updatedBy string) (*views.CyclePreferenceResponse, error) {
	if dayOfMonth < 1 || dayOfMonth > 28 {
		return nil, fmt.Errorf("preferredDayOfMonth must be between 1 and 28, got %s", strconv.Itoa(dayOfMonth))
	}
	pref, err := s.cycleService.UpsertCyclePreference(ctx, projectID, dayOfMonth, updatedBy)
	if err != nil {
		s.logger.Error("payments_svc: set cycle preference failed",
			zap.String("project_id", projectID),
			zap.Int("day_of_month", dayOfMonth),
			zap.Error(err))
		if appErr := apperrors.AsAppError(err); appErr != nil {
			return nil, appErr
		}
		return nil, apperrors.TranslateError(err, "service")
	}
	s.logger.Info("payments_svc: preferred day set",
		zap.String("project_id", projectID),
		zap.Int("day", dayOfMonth))
	return &views.CyclePreferenceResponse{
		ProjectID:           pref.ProjectID,
		PreferredDayOfMonth: pref.PreferredDayOfMonth,
		UpdatedAt:           pref.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func protoBillRecordToView(b *billsv1.BillRecord) views.PaymentBillRecordResponse {
	if b == nil {
		return views.PaymentBillRecordResponse{}
	}
	return views.PaymentBillRecordResponse{
		ID:            b.GetId(),
		ProjectID:     b.GetProjectId(),
		DocumentID:    b.GetDocumentId(),
		BillTypeID:    b.GetBillTypeId(),
		DueDate:       b.GetDueDate(),
		AmountDue:     b.GetAmountDue(),
		PixPayload:    b.GetPixPayload(),
		PixQRImageRef: b.GetPixQrImageRef(),
		Barcode:       b.GetBarcode(),
		PaymentStatus: b.GetPaymentStatus().String(),
		PaidAt:        b.GetPaidAt(),
		MarkedPaidBy:  b.GetMarkedPaidBy(),
		CreatedAt:     b.GetCreatedAt(),
		UpdatedAt:     b.GetUpdatedAt(),
	}
}
