// Package services implements use-case logic for the payments domain.
package services

import (
	"context"

	"go.uber.org/zap"

	"github.com/ralvescosta/costa-financial-assistant/backend/internals/payments/interfaces"
	apperrors "github.com/ralvescosta/costa-financial-assistant/backend/pkgs/errors"
)

// historyRepository is a local alias so callers only import the interfaces package.
type historyRepository = interfaces.HistoryRepository

// HistoryService implements interfaces.HistoryService.
type HistoryService struct {
	repo   historyRepository
	logger *zap.Logger
}

// NewHistoryService constructs a HistoryService.
func NewHistoryService(repo historyRepository, logger *zap.Logger) interfaces.HistoryService {
	return &HistoryService{repo: repo, logger: logger}
}

// GetTimeline returns monthly expenditure totals for the given project.
func (s *HistoryService) GetTimeline(ctx context.Context, projectID string, months int) ([]interfaces.MonthlyTimelineEntry, error) {
	entries, err := s.repo.GetTimeline(ctx, projectID, months)
	if err != nil {
		s.logger.Error("history_service: get timeline failed",
			zap.String("project_id", projectID),
			zap.Int("months", months),
			zap.Error(err))
		if appErr := apperrors.AsAppError(err); appErr != nil {
			return nil, appErr
		}
		return nil, apperrors.TranslateError(err, "service")
	}
	return entries, nil
}

// GetCategoryBreakdown returns monthly spend totals grouped by bill type.
func (s *HistoryService) GetCategoryBreakdown(ctx context.Context, projectID string, months int) ([]interfaces.CategoryBreakdownEntry, error) {
	entries, err := s.repo.GetCategoryBreakdown(ctx, projectID, months)
	if err != nil {
		s.logger.Error("history_service: get category breakdown failed",
			zap.String("project_id", projectID),
			zap.Int("months", months),
			zap.Error(err))
		if appErr := apperrors.AsAppError(err); appErr != nil {
			return nil, appErr
		}
		return nil, apperrors.TranslateError(err, "service")
	}
	return entries, nil
}

// GetComplianceMetrics returns the monthly on-time payment metrics for the project.
func (s *HistoryService) GetComplianceMetrics(ctx context.Context, projectID string, months int) ([]interfaces.MonthlyComplianceEntry, error) {
	entries, err := s.repo.GetComplianceMetrics(ctx, projectID, months)
	if err != nil {
		s.logger.Error("history_service: get compliance metrics failed",
			zap.String("project_id", projectID),
			zap.Int("months", months),
			zap.Error(err))
		if appErr := apperrors.AsAppError(err); appErr != nil {
			return nil, appErr
		}
		return nil, apperrors.TranslateError(err, "service")
	}
	return entries, nil
}
