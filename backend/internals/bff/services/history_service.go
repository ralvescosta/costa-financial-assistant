package services

import (
	"context"

	"go.uber.org/zap"

	bffinterfaces "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/interfaces"
	bffcontracts "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/services/contracts"
	apperrors "github.com/ralvescosta/costa-financial-assistant/backend/pkgs/errors"
	paymentsv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/payments/v1"
)

// HistoryServiceImpl implements bffinterfaces.HistoryService.
type HistoryServiceImpl struct {
	logger         *zap.Logger
	paymentsClient paymentsv1.PaymentsServiceClient
}

// NewHistoryService constructs a HistoryServiceImpl.
func NewHistoryService(logger *zap.Logger, paymentsClient paymentsv1.PaymentsServiceClient) bffinterfaces.HistoryService {
	return &HistoryServiceImpl{logger: logger, paymentsClient: paymentsClient}
}

// defaultMonths returns 12 when months is negative (0 is "all history").
func defaultMonths(m int) int {
	if m < 0 {
		return 12
	}
	return m
}

// GetTimeline returns aggregated bill amounts per calendar month.
func (s *HistoryServiceImpl) GetTimeline(ctx context.Context, projectID string, months int) (*bffcontracts.TimelineResponse, error) {
	resolvedMonths := defaultMonths(months)
	resp, err := s.paymentsClient.GetHistoryTimeline(ctx, &paymentsv1.GetHistoryTimelineRequest{
		Ctx:        projectContextFromContext(ctx, projectID, ""),
		Session:    sessionFromContext(ctx),
		Pagination: defaultPagination(0, "", 20),
		Months:     int32(resolvedMonths),
	})
	if err != nil {
		s.logger.Error("history_svc: get timeline failed",
			zap.String("project_id", projectID),
			zap.Int("months", resolvedMonths),
			zap.Error(err))
		if appErr := apperrors.AsAppError(err); appErr != nil {
			return nil, appErr
		}
		return nil, apperrors.TranslateError(err, "service")
	}

	result := &bffcontracts.TimelineResponse{
		ProjectID: resp.GetProjectId(),
		Months:    int(resp.GetMonths()),
		Timeline:  make([]*bffcontracts.MonthlyTimelineEntryResponse, 0, len(resp.GetEntries())),
	}
	for _, entry := range resp.GetEntries() {
		result.Timeline = append(result.Timeline, &bffcontracts.MonthlyTimelineEntryResponse{
			Month:       entry.GetMonth(),
			TotalAmount: entry.GetTotalAmount(),
			BillCount:   int(entry.GetBillCount()),
		})
	}
	return result, nil
}

// GetCategoryBreakdown returns bill amounts grouped by bill type and month.
func (s *HistoryServiceImpl) GetCategoryBreakdown(ctx context.Context, projectID string, months int) (*bffcontracts.CategoryBreakdownResponse, error) {
	resolvedMonths := defaultMonths(months)
	resp, err := s.paymentsClient.GetHistoryCategoryBreakdown(ctx, &paymentsv1.GetHistoryCategoryBreakdownRequest{
		Ctx:        projectContextFromContext(ctx, projectID, ""),
		Session:    sessionFromContext(ctx),
		Pagination: defaultPagination(0, "", 20),
		Months:     int32(resolvedMonths),
	})
	if err != nil {
		s.logger.Error("history_svc: get category breakdown failed",
			zap.String("project_id", projectID),
			zap.Int("months", resolvedMonths),
			zap.Error(err))
		if appErr := apperrors.AsAppError(err); appErr != nil {
			return nil, appErr
		}
		return nil, apperrors.TranslateError(err, "service")
	}

	result := &bffcontracts.CategoryBreakdownResponse{
		ProjectID:  resp.GetProjectId(),
		Months:     int(resp.GetMonths()),
		Categories: make([]*bffcontracts.CategoryBreakdownEntryResponse, 0, len(resp.GetEntries())),
	}
	for _, entry := range resp.GetEntries() {
		result.Categories = append(result.Categories, &bffcontracts.CategoryBreakdownEntryResponse{
			Month:        entry.GetMonth(),
			BillTypeName: entry.GetBillTypeName(),
			TotalAmount:  entry.GetTotalAmount(),
			BillCount:    int(entry.GetBillCount()),
		})
	}
	return result, nil
}

// GetComplianceMetrics returns on-time vs overdue bill counts and compliance rate.
func (s *HistoryServiceImpl) GetComplianceMetrics(ctx context.Context, projectID string, months int) (*bffcontracts.ComplianceResponse, error) {
	resolvedMonths := defaultMonths(months)
	resp, err := s.paymentsClient.GetHistoryCompliance(ctx, &paymentsv1.GetHistoryComplianceRequest{
		Ctx:        projectContextFromContext(ctx, projectID, ""),
		Session:    sessionFromContext(ctx),
		Pagination: defaultPagination(0, "", 20),
		Months:     int32(resolvedMonths),
	})
	if err != nil {
		s.logger.Error("history_svc: get compliance metrics failed",
			zap.String("project_id", projectID),
			zap.Int("months", resolvedMonths),
			zap.Error(err))
		if appErr := apperrors.AsAppError(err); appErr != nil {
			return nil, appErr
		}
		return nil, apperrors.TranslateError(err, "service")
	}

	result := &bffcontracts.ComplianceResponse{
		ProjectID:  resp.GetProjectId(),
		Months:     int(resp.GetMonths()),
		Compliance: make([]*bffcontracts.MonthlyComplianceEntryResponse, 0, len(resp.GetEntries())),
	}
	for _, entry := range resp.GetEntries() {
		result.Compliance = append(result.Compliance, &bffcontracts.MonthlyComplianceEntryResponse{
			Month:          entry.GetMonth(),
			TotalBills:     int(entry.GetTotalBills()),
			PaidOnTime:     int(entry.GetPaidOnTime()),
			Overdue:        int(entry.GetOverdue()),
			ComplianceRate: entry.GetComplianceRate(),
		})
	}
	return result, nil
}
