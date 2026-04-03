package mappers_test

import (
	"testing"

	bffcontracts "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/services/contracts"
	controllermappers "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/transport/http/controllers/mappers"
	views "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/transport/http/views"
)

func TestDocumentsMapperNilSafety(t *testing.T) {
	t.Parallel()

	t.Run("GivenNilUploadInputWhenMappedThenReturnsEmptyRequest", func(t *testing.T) {
		// Given
		var input *views.UploadDocumentInput

		// Arrange
		// no additional setup

		// Act
		fileName, rawBody := controllermappers.ToUploadRequest(input)

		// Then
		if fileName != "" || rawBody != nil {
			t.Fatalf("expected empty upload request mapping")
		}
	})

	t.Run("GivenNilListDocumentsResponseWhenMappedThenReturnsEmptyItems", func(t *testing.T) {
		// Given
		var response *bffcontracts.ListDocumentsResponse

		// Arrange
		// no additional setup

		// Act
		mapped := controllermappers.ToListDocumentsResponse(response)

		// Then
		if mapped.Items == nil || len(mapped.Items) != 0 {
			t.Fatalf("expected non-nil empty list response items")
		}
	})
}

func TestHistoryMapperNilSafety(t *testing.T) {
	t.Parallel()

	t.Run("GivenNilHistoryInputWhenMappedThenMonthsDefaultToZero", func(t *testing.T) {
		// Given
		var input *views.HistoryQueryInput

		// Arrange
		// no additional setup

		// Act
		months := controllermappers.ToHistoryMonths(input)

		// Then
		if months != 0 {
			t.Fatalf("expected months default to zero")
		}
	})

	t.Run("GivenNilTimelineResponseWhenMappedThenReturnsEmptyTimeline", func(t *testing.T) {
		// Given
		var response *bffcontracts.TimelineResponse

		// Arrange
		// no additional setup

		// Act
		mapped := controllermappers.ToTimelineResponse(response)

		// Then
		if mapped.Timeline == nil || len(mapped.Timeline) != 0 {
			t.Fatalf("expected non-nil empty timeline")
		}
	})
}
