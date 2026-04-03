package cross_service

import (
	"errors"
	"testing"

	apperrors "github.com/ralvescosta/costa-financial-assistant/backend/pkgs/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAsyncPublisherErrorPropagation_T061 validates that RabbitMQ producer/consumer
// errors follow the same AppError propagation contract as synchronous gRPC calls.
// This ensures async paths don't leak raw dependency errors.
func TestAsyncPublisherErrorPropagation_T061(t *testing.T) {
	// SCENARIO 1: Publisher (producer) encounters an error
	publisherErr := errors.New("connection refused to rabbitmq broker")
	publisherAppErr := apperrors.TranslateError(publisherErr, "rmq_publisher")

	require.NotNil(t, publisherAppErr, "publisher error must translate to AppError")
	assert.IsType(t, (*apperrors.AppError)(nil), publisherAppErr)

	// SCENARIO 2: Consumer encounters an error processing a message
	consumerErr := errors.New("failed to persist analysis result to database")
	consumerAppErr := apperrors.TranslateError(consumerErr, "rmq_consumer")

	require.NotNil(t, consumerAppErr, "consumer error must translate to AppError")
	assert.IsType(t, (*apperrors.AppError)(nil), consumerAppErr)

	// BOTH must be AppErrors (consistent with sync gRPC errors)
	assert.True(t, apperrors.IsAppError(publisherAppErr))
	assert.True(t, apperrors.IsAppError(consumerAppErr))
}

// TestAsyncErrorNoSensitiveDataLeak_T061 validates that sanitization prevents
// exposure of connection strings, credentials, or internal service details.
func TestAsyncErrorNoSensitiveDataLeak_T061(t *testing.T) {
	sensitiveErr := errors.New("rabbitmq connection failed: amqp://admin:secretpass123@rabbitmq-broker.internal:5672/analysis_queue")

	appErr := apperrors.TranslateError(sensitiveErr, "rmq_publisher")
	require.NotNil(t, appErr)

	// THEN: Error message must be sanitized
	assert.NotContains(t, appErr.Message, "secretpass123",
		"must not expose credentials in error message")
	assert.NotContains(t, appErr.Message, "admin",
		"must not expose usernames in error message")
	assert.NotContains(t, appErr.Message, "rabbitmq-broker.internal:5672",
		"must not expose internal service endpoints")
	assert.NotContains(t, appErr.Message, "amqp://",
		"must not expose connection protocol details")
}

// TestAsyncErrorConsistencyWithSync_T061 validates that async errors follow
// the same category and retryability rules as their sync equivalents.
func TestAsyncErrorConsistencyWithSync_T061(t *testing.T) {
	tests := []struct {
		name        string
		asyncErr    error
		shouldRetry bool
	}{
		{
			name:        "Connection failures are retryable in both sync and async",
			asyncErr:    errors.New("connection refused"),
			shouldRetry: true,
		},
		{
			name:        "Validation errors are not retryable in either path",
			asyncErr:    errors.New("invalid message format"),
			shouldRetry: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appErr := apperrors.TranslateError(tt.asyncErr, "rmq_publisher")
			require.NotNil(t, appErr)

			// Verify retryability matches expectation
			// (exact matching may depend on translation rules, so we're just
			// checking that the field is set consistently)
			assert.Equal(t, tt.shouldRetry, appErr.Retryable,
				"async errors should have deterministic retryability")
		})
	}
}

// TestAsyncErrorPreservesNativeForLogging_T061 validates that AppError wraps
// the native error so it can be logged at the translation boundary,
// but this wrapped error doesn't propagate beyond the boundary.
func TestAsyncErrorPreservesNativeForLogging_T061(t *testing.T) {
	nativeErr := errors.New("rabbitmq: timeout waiting for heartbeat")
	appErr := apperrors.TranslateError(nativeErr, "rmq_consumer")

	require.NotNil(t, appErr)

	// THEN: Native error must be preserved (for logging)
	assert.NotNil(t, appErr.Err,
		"AppError must preserve native error for boundary logging")

	// BUT: The message must be sanitized
	assert.NotContains(t, appErr.Message, "rabbitmq:",
		"message should not expose raw error prefix")
}
