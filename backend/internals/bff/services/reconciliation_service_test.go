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

func newReconciliationService(t *testing.T) bffinterfaces.ReconciliationService {
	t.Helper()
	return services.NewReconciliationService(zaptest.NewLogger(t))
}

func TestReconciliationService_GetSummary_ReturnsDependencyError(t *testing.T) {
	svc := newReconciliationService(t)

	result, err := svc.GetSummary(context.Background(), "proj-1", "2024-01-01", "2024-01-31")

	assert.Nil(t, result)
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperrors.CategoryDependencyGRPC, appErr.Category)
}

func TestReconciliationService_CreateManualLink_ReturnsDependencyError(t *testing.T) {
	svc := newReconciliationService(t)

	result, err := svc.CreateManualLink(context.Background(), "proj-1", "txn-1", "bill-1", "user-1")

	assert.Nil(t, result)
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperrors.CategoryDependencyGRPC, appErr.Category)
}
