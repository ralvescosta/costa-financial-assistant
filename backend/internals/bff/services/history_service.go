package services

import (
	"context"

	"go.uber.org/zap"

	bffinterfaces "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/interfaces"
	views "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/transport/http/views"
	paymentsinterfaces "github.com/ralvescosta/costa-financial-assistant/backend/internals/payments/interfaces"
	apperrors "github.com/ralvescosta/costa-financial-assistant/backend/pkgs/errors"
)

// HistoryServiceImpl implements bffinterfaces.HistoryService.
type HistoryServiceImpl struct {
	logger      *zap.Logger
	historyRepo paymentsinterfaces.HistoryRepository
}

// NewHistoryService constructs a HistoryServiceImpl.
func NewHistoryService(logger *zap.Logger, historyRepo paymentsinterfaces.HistoryRepository) bffinterfaces.HistoryService {
	return &HistoryServiceImpl{logger: logger, historyRepo: historyRepo}
}

// defaultMonths returns 12 when months is negative (0 is "all history").
func defaultMonths(m int) int {
	if m < 0 {
		return 12
	}
	return m
}

// GetTimeline returns aggregated bill amounts per calendar month.
func (s *HistoryServiceImpl) GetTimeline(ctx context.Context, projectID string, months int) (*views.TimelineResponse, error) {
	months = defaultMonths(months)
	entries, err := s.historyRepo.GetTimeline(ctx, projectID, months)
	if err != nil {
		s.logger.Error("history_svc: get timeline failed",
			zap.String("project_id", projectID),
			zap.Error(err))
		if appErr := apperrors.AsAppError(err); appErr != nil {
			return nil, appErr
		}
		return nil, apperrors.TranslateError(err, "service")
	}

	rows := make([]*views.MonthlyTimelineEntryResponse, 0, len(entries))
	for _, e := range entries {
		rows = append(rows, &views.MonthlyTimelineEntryResponse{
			Month:       e.Month,
			TotalAmount: e.TotalAmount,
			BillCount:   e.BillCount,
		})
	}
	return &views.TimelineResponse{ProjectID: projectID, Months: months, Timeline: rows}, nil
}

// GetCategoryBreakdown returns bill amounts grouped by bill type and month.
func (s *HistoryServiceImpl) GetCategoryBreakdown(ctx context.Context, projectID string, months int) (*views.CategoryBreakdownResponse, error) {
	months = defaultMonths(months)
	entries, err := s.historyRepo.GetCategoryBreakdown(ctx, projectID, months)
	if err != nil {
		s.logger.Error("history_svc: get category breakdown failed",
			zap.String("project_id", projectID),
			zap.Error(err))
		if appErr := apperrors.AsAppError(err); appErr != nil {
			return nil, appErr
		}
		return nil, apperrors.TranslateError(err, "service")
	}

	rows := make([]*views.CategoryBreakdownEntryResponse, 0, len(entries))
	for _, e := range entries {
		rows = append(rows, &views.CategoryBreakdownEntryResponse{
			Month:        e.Month,
			BillTypeName: e.BillTypeName,
			TotalAmount:  e.TotalAmount,
			BillCount:    e.BillCount,
		})
	}
	return &views.CategoryBreakdownResponse{ProjectID: projectID, Months: months, Categories: rows}, nil
}

// GetComplianceMetrics returns on-time vs overdue bill counts and compliance rate.
func (s *HistoryServiceImpl) GetComplianceMetrics(ctx context.Context, projectID string, months int) (*views.ComplianceResponse, error) {
	months = defaultMonths(months)
	entries, err := s.historyRepo.GetComplianceMetrics(ctx, projectID, months)
	if err != nil {
		s.logger.Error("history_svc: get compliance metrics failed",
			zap.String("project_id", projectID),
			zap.Error(err))
		if appErr := apperrors.AsAppError(err); appErr != nil {
			return nil, appErr
		}
		return nil, apperrors.TranslateError(err, "service")
	}

	rows := make([]*views.MonthlyComplianceEntryResponse, 0, len(entries))
	for _, e := range entries {
		rows = append(rows, &views.MonthlyComplianceEntryResponse{
			Month:          e.Month,
			TotalBills:     e.TotalBills,
			PaidOnTime:     e.PaidOnTime,
			Overdue:        e.Overdue,
			ComplianceRate: e.ComplianceRate,
		})
	}
	return &views.ComplianceResponse{ProjectID: projectID, Months: months, Compliance: rows}, nil
}
