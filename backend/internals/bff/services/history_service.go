package services

import (
	"context"

	"go.uber.org/zap"

	bffinterfaces "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/interfaces"
	bffcontracts "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/services/contracts"
	apperrors "github.com/ralvescosta/costa-financial-assistant/backend/pkgs/errors"
)

const historyGRPCUnavailableMessage = "history analytics require a downstream payments gRPC contract before the BFF can serve this flow"

// HistoryServiceImpl implements bffinterfaces.HistoryService.
type HistoryServiceImpl struct {
	logger *zap.Logger
}

// NewHistoryService constructs a HistoryServiceImpl.
func NewHistoryService(logger *zap.Logger) bffinterfaces.HistoryService {
	return &HistoryServiceImpl{logger: logger}
}

// defaultMonths returns 12 when months is negative (0 is "all history").
func defaultMonths(m int) int {
	if m < 0 {
		return 12
	}
	return m
}

func (s *HistoryServiceImpl) downstreamUnavailable(logMessage, projectID string) error {
	appErr := apperrors.NewWithCategory(historyGRPCUnavailableMessage, apperrors.CategoryDependencyGRPC)
	s.logger.Error(logMessage,
		zap.String("project_id", projectID),
		zap.Error(appErr))
	return appErr
}

// GetTimeline returns aggregated bill amounts per calendar month.
func (s *HistoryServiceImpl) GetTimeline(ctx context.Context, projectID string, months int) (*bffcontracts.TimelineResponse, error) {
	_ = ctx
	_ = defaultMonths(months)
	return nil, s.downstreamUnavailable("history_svc: get timeline failed", projectID)
}

// GetCategoryBreakdown returns bill amounts grouped by bill type and month.
func (s *HistoryServiceImpl) GetCategoryBreakdown(ctx context.Context, projectID string, months int) (*bffcontracts.CategoryBreakdownResponse, error) {
	_ = ctx
	_ = defaultMonths(months)
	return nil, s.downstreamUnavailable("history_svc: get category breakdown failed", projectID)
}

// GetComplianceMetrics returns on-time vs overdue bill counts and compliance rate.
func (s *HistoryServiceImpl) GetComplianceMetrics(ctx context.Context, projectID string, months int) (*bffcontracts.ComplianceResponse, error) {
	_ = ctx
	_ = defaultMonths(months)
	return nil, s.downstreamUnavailable("history_svc: get compliance metrics failed", projectID)
}
