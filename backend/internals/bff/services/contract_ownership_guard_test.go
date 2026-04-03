package services_test

import (
	"os"
	"strings"
	"testing"
)

func TestContractOwnershipGuard(t *testing.T) {
	t.Parallel()

	t.Run("GivenBFFInterfacesWhenContractOwnershipCheckedThenInterfacesDoNotImportViewsDirectly", func(t *testing.T) {
		// Given
		filePath := "../interfaces/services.go"

		// Arrange
		content, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("failed to read interfaces file: %v", err)
		}
		text := string(content)

		// When
		hasDirectViewsImport := strings.Contains(text, "transport/http/views")
		hasContractsImport := strings.Contains(text, "services/contracts")

		// Then
		if hasDirectViewsImport {
			t.Fatalf("interfaces must not import transport views directly")
		}
		if !hasContractsImport {
			t.Fatalf("interfaces must import service contracts package")
		}
	})
}
