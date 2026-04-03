package rmq

import (
	"errors"
	"testing"

	apperrors "github.com/ralvescosta/costa-financial-assistant/backend/pkgs/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTranslatePublisherError(t *testing.T) {
	nativeErr := errors.New("connection reset by peer")
	appErr := TranslatePublisherError(nativeErr)

	require.NotNil(t, appErr)
	assert.NotEmpty(t, appErr.Code)
	assert.NotEmpty(t, appErr.Message)
}

func TestTranslatePublisherErrorPassthroughAppError(t *testing.T) {
	input := apperrors.NewCatalogError(apperrors.ErrDatabaseTimeout)
	out := TranslatePublisherError(input)

	require.NotNil(t, out)
	assert.Equal(t, input.Code, out.Code)
	assert.Equal(t, input.Category, out.Category)
}

func TestShouldRequeuePublisherError(t *testing.T) {
	retryable := apperrors.NewCatalogError(apperrors.ErrDatabaseTimeout)
	nonRetryable := apperrors.NewCatalogError(apperrors.ErrValidationError)

	assert.True(t, ShouldRequeuePublisherError(retryable))
	assert.False(t, ShouldRequeuePublisherError(nonRetryable))
	assert.False(t, ShouldRequeuePublisherError(nil))
}
