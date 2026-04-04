package services_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	bffinterfaces "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/interfaces"
	"github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/services"
	apperrors "github.com/ralvescosta/costa-financial-assistant/backend/pkgs/errors"
)

func newHistoryService(t *testing.T) bffinterfaces.HistoryService {
	t.Helper()
	return services.NewHistoryService(zaptest.NewLogger(t))
}

func TestHistoryService_GetTimeline_ReturnsDependencyError(t *testing.T) {
	svc := newHistoryService(t)

	result, err := svc.GetTimeline(context.Background(), "proj-1", 6)

	assert.Nil(t, result)
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperrors.CategoryDependencyGRPC, appErr.Category)
}

func TestHistoryService_GetCategoryBreakdown_ReturnsDependencyError(t *testing.T) {
	svc := newHistoryService(t)

	result, err := svc.GetCategoryBreakdown(context.Background(), "proj-1", 6)

	assert.Nil(t, result)
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperrors.CategoryDependencyGRPC, appErr.Category)
}

func TestHistoryService_GetComplianceMetrics_ReturnsDependencyError(t *testing.T) {
	svc := newHistoryService(t)

	result, err := svc.GetComplianceMetrics(context.Background(), "proj-1", 6)

	assert.Nil(t, result)
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperrors.CategoryDependencyGRPC, appErr.Category)
}
