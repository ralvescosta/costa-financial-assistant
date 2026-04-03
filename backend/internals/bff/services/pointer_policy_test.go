package services_test

import (
	"os"
	"strings"
	"testing"
)

func TestBFFServiceContractsPointerPolicy(t *testing.T) {
	t.Parallel()

	t.Run("GivenBFFServiceContractsWhenAuditedThenResponsesUsePointerSemantics", func(t *testing.T) {
		// Given
		filePath := "../interfaces/services.go"

		// Arrange
		content, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("failed to read BFF interfaces file: %v", err)
		}
		text := string(content)

		// When
		hasPointerResponses := strings.Contains(text, ") (*bffcontracts.")
		hasValueResponses := strings.Contains(text, ") (bffcontracts.")

		// Then
		if !hasPointerResponses {
			t.Fatalf("expected pointer-based contract responses in BFF interfaces")
		}
		if hasValueResponses {
			t.Fatalf("unexpected value-based contract responses in BFF interfaces")
		}
	})
}
