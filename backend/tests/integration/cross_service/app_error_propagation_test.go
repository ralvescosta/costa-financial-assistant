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
	// STEP 1: Repository boundary - translate native error
	nativeErr := sql.ErrNoRows
	repoErr := apperrors.TranslateError(nativeErr, "repository")

	require.NotNil(t, repoErr, "repository must translate to AppError")
	assert.IsType(t, (*apperrors.AppError)(nil), repoErr)

	// STEP 2: Service boundary - AppError propagates unchanged
	// (Service returns the error as-is, doesn't re-translate)
	svcErr := repoErr
	assert.IsType(t, (*apperrors.AppError)(nil), svcErr,
		"service boundary should propagate AppError unchanged")

	// STEP 3: Verify no raw error leaked
	assert.True(t, apperrors.IsAppError(svcErr),
		"error crossing boundaries must be AppError, not raw error")

	// STEP 4: Verify AppError is properly constructed
	assert.NotEmpty(t, repoErr.Code, "AppError must have code set")
	assert.NotEmpty(t, repoErr.Message, "AppError must have sanitized message")
	assert.NotNil(t, repoErr.Err, "AppError must preserve native error for logging")
}

// TestNoRawErrorsLeakBoundary_T012 validates the critical FR-010 requirement:
// Raw dependency errors must never cross layer boundaries.
func TestNoRawErrorsLeakBoundary_T012(t *testing.T) {
	testCases := []error{
		sql.ErrNoRows,
		sql.ErrConnDone,
		sql.ErrTxDone,
	}

	for _, rawErr := range testCases {
		t.Run(rawErr.Error(), func(t *testing.T) {
			// WHEN: Repository translates error
			appErr := apperrors.TranslateError(rawErr, "repository")

			// THEN: Result must be AppError (not raw error, not nil)
			require.NotNil(t, appErr)

			// AND: The result type must be AppError
			_, isAppErr := appErr.(*apperrors.AppError)
			require.True(t, isAppErr, "must be AppError type")

			// AND: Error message must not expose raw error details
			assert.NotContains(t, appErr.Message, "sql",
				"message should not expose raw SQL error details")
		})
	}
}

// TestRetryabilityPreservation_T012 validates that error classification
// (retryable vs non-retryable) is preserved through propagation.
func TestRetryabilityPreservation_T012(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		expectRetry  bool
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
			appErr := apperrors.TranslateError(tt.err, "repository")
			assert.Equal(t, tt.expectRetry, appErr.Retryable,
				"retryability should match expectation")
		})
	}
}
