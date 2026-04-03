package repositories_test

import (
	"database/sql"
	"testing"

	apperrors "github.com/ralvescosta/costa-financial-assistant/backend/pkgs/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDocumentRepositoryErrorTranslation_T011 validates that repository-layer database
// errors are translated to AppError before crossing the repository boundary.
// This is the core requirement for the MVP error propagation contract.
func TestDocumentRepositoryErrorTranslation_T011(t *testing.T) {
	tests := []struct {
		name         string
		givenErr     error
		expectAppErr bool
		expectCode   string
	}{
		{
			name:         "sql.ErrNoRows translates to AppError",
			givenErr:     sql.ErrNoRows,
			expectAppErr: true,
			expectCode:   "not_found",
		},
		{
			name:         "sql.ErrConnDone translates to AppError",
			givenErr:     sql.ErrConnDone,
			expectAppErr: true,
			expectCode:   "database_connection_failed",
		},
		{
			name:         "unknown error translates to AppError with fallback",
			givenErr:     sql.ErrTxDone,
			expectAppErr: true,
			expectCode:   "database_error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// WHEN: Repository layer receives a native database error
			translated := apperrors.TranslateError(tt.givenErr, "repository")

			// THEN: It MUST translate to AppError (non-nil)
			require.NotNil(t, translated, "native error must be translated to AppError")

			// AND: The AppError must have appropriate code
			if tt.expectCode != "" {
				assert.Equal(t, tt.expectCode, translated.Code,
					"translated error code should match expected")
			}

			// AND: The AppError must preserve the native error reference for logging
			assert.NotNil(t, translated.Err,
				"AppError should preserve native error for logging")

			// AND: The message must be sanitized
			assert.NotContains(t, translated.Message, "sql",
				"error message should be sanitized, not expose SQL details")
		})
	}
}

// BenchmarkErrorTranslation_T011 measures translation performance.
func BenchmarkErrorTranslation_T011(b *testing.B) {
	nativeErr := sql.ErrNoRows
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = apperrors.TranslateError(nativeErr, "repository")
	}
}
