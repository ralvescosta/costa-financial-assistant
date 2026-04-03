package cross_service

import (
	nativeerrors "errors"
	"testing"

	apperrors "github.com/ralvescosta/costa-financial-assistant/backend/pkgs/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnknownFallbackTranslation_T038(t *testing.T) {
	tests := []struct {
		name      string
		layer     string
		nativeErr error
	}{
		{
			name:      "Given unmapped layer When translating error Then unknown fallback is returned",
			layer:     "unknown_boundary",
			nativeErr: nativeerrors.New("opaque dependency failure"),
		},
		{
			name:      "Given future layer When translating error Then fallback remains deterministic",
			layer:     "future_layer",
			nativeErr: nativeerrors.New("dependency exploded"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			nativeErr := tt.nativeErr

			// Act
			first := apperrors.TranslateError(nativeErr, tt.layer)
			second := apperrors.TranslateError(nativeErr, tt.layer)

			// Assert
			require.NotNil(t, first)
			require.NotNil(t, second)
			assert.Equal(t, apperrors.ErrUnknown.Code, first.Code)
			assert.Equal(t, apperrors.ErrUnknown.Category, first.Category)
			assert.Equal(t, apperrors.ErrUnknown.Message, first.Message)
			assert.Equal(t, nativeErr, first.Err)
			assert.Equal(t, first.Code, second.Code)
			assert.Equal(t, first.Category, second.Category)
			assert.Equal(t, first.Message, second.Message)
			assert.Equal(t, first.Retryable, second.Retryable)
		})
	}
}
