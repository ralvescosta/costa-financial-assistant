package services

import (
	"os"
	"strings"
	"testing"
)

func TestFilesAndBillsBoundaryPointerPolicy(t *testing.T) {
	t.Parallel()

	t.Run("GivenFilesServiceWhenAuditedThenStructBoundariesUsePointers", func(t *testing.T) {
		// Given
		filesPath := "document_service.go"

		// Arrange
		content, err := os.ReadFile(filesPath)
		if err != nil {
			t.Fatalf("failed to read files service: %v", err)
		}
		text := string(content)

		// When
		hasPointerInput := strings.Contains(text, "UploadDocument(ctx context.Context, req *UploadDocumentInput)")

		// Then
		if !hasPointerInput {
			t.Fatalf("expected pointer input for upload document boundary")
		}
	})

	t.Run("GivenBillsServiceWhenAuditedThenStructBoundariesUsePointers", func(t *testing.T) {
		// Given
		billsPath := "../../bills/services/payment_service.go"

		// Arrange
		content, err := os.ReadFile(billsPath)
		if err != nil {
			t.Fatalf("failed to read bills service: %v", err)
		}
		text := string(content)

		// When
		hasPointerReturn := strings.Contains(text, "MarkBillPaid(ctx context.Context, projectID, billID, markedBy string) (*billsv1.BillRecord, error)")

		// Then
		if !hasPointerReturn {
			t.Fatalf("expected pointer return for mark bill paid boundary")
		}
	})
}
