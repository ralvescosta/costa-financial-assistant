package rmq

import (
	apperrors "github.com/ralvescosta/costa-financial-assistant/backend/pkgs/errors"
)

// TranslatePublisherError sanitizes native broker/publisher failures
// into the canonical AppError contract used across async boundaries.
//
// Use this helper at RabbitMQ producer boundaries before propagating
// errors to services/controllers so raw dependency errors do not leak.
func TranslatePublisherError(err error) *apperrors.AppError {
	if err == nil {
		return nil
	}

	if appErr := apperrors.AsAppError(err); appErr != nil {
		return appErr
	}

	// Publisher path currently reuses async-consumer translation policy.
	// This keeps deterministic categorization/retryability until a dedicated
	// async_producer layer policy is introduced.
	return apperrors.TranslateError(err, "async_consumer")
}

// ShouldRequeuePublisherError decides whether a failed publish operation
// should be retried based on the translated AppError retryability semantics.
func ShouldRequeuePublisherError(err error) bool {
	appErr := TranslatePublisherError(err)
	if appErr == nil {
		return false
	}
	return appErr.Retryable
}
