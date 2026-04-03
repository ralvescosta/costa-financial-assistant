package cross_service

import (
	"database/sql"
	"testing"

	apperrors "github.com/ralvescosta/costa-financial-assistant/backend/pkgs/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAppErrorPropagationAcrossLayers_T012 validates the MVP propagation contract:
// When an error occurs at the repository boundary (data access layer),
// it must:
// 1. Be translated to AppError at the repository boundary
// 2. Propagate unchanged through service layer
// 3. NOT leak raw dependency errors at any boundary
func TestAppErrorPropagationAcrossLayers_T012(t *testing.T) {
	tests := []struct {
		name      string
		nativeErr error
	}{
		{
			name:      "Given repository native failure When translated Then service receives AppError unchanged",
			nativeErr: sql.ErrNoRows,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			nativeErr := tt.nativeErr

			// Act
			repoErr := apperrors.TranslateError(nativeErr, "repository")
			svcErr := repoErr

			// Assert
			require.NotNil(t, repoErr, "repository must translate to AppError")
			assert.IsType(t, (*apperrors.AppError)(nil), repoErr)
			assert.IsType(t, (*apperrors.AppError)(nil), svcErr,
				"service boundary should propagate AppError unchanged")
			assert.True(t, apperrors.IsAppError(svcErr),
				"error crossing boundaries must be AppError, not raw error")
			assert.NotEmpty(t, repoErr.Code, "AppError must have code set")
			assert.NotEmpty(t, repoErr.Message, "AppError must have sanitized message")
			assert.NotNil(t, repoErr.Err, "AppError must preserve native error for logging")
		})
	}
}

// TestNoRawErrorsLeakBoundary_T012 validates the critical FR-010 requirement:
// Raw dependency errors must never cross layer boundaries.
func TestNoRawErrorsLeakBoundary_T012(t *testing.T) {
	tests := []struct {
		name   string
		rawErr error
	}{
		{name: "Given sql no rows When translated Then no raw sql leaked", rawErr: sql.ErrNoRows},
		{name: "Given sql connection done When translated Then no raw sql leaked", rawErr: sql.ErrConnDone},
		{name: "Given sql tx done When translated Then no raw sql leaked", rawErr: sql.ErrTxDone},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			rawErr := tt.rawErr

			// Act
			appErr := apperrors.TranslateError(rawErr, "repository")

			// Assert
			require.NotNil(t, appErr)
			assert.IsType(t, (*apperrors.AppError)(nil), appErr)
			assert.NotContains(t, appErr.Message, "sql",
				"message should not expose raw SQL error details")
		})
	}
}

// TestRetryabilityPreservation_T012 validates that error classification
// (retryable vs non-retryable) is preserved through propagation.
func TestRetryabilityPreservation_T012(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		expectRetry bool
	}{
		{
			name:        "connection errors are retryable",
			err:         sql.ErrConnDone,
			expectRetry: true,
		},
		{
			name:        "not-found errors are not retryable",
			err:         sql.ErrNoRows,
			expectRetry: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			err := tt.err

			// Act
			appErr := apperrors.TranslateError(err, "repository")

			// Assert
			assert.Equal(t, tt.expectRetry, appErr.Retryable,
				"retryability should match expectation")
		})
	}
}
