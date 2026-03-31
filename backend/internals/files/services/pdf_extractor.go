package services

import "context"

// BillExtractionResult holds structured data extracted from a bill PDF.
type BillExtractionResult struct {
	// DueDate is the bill due date in ISO format (YYYY-MM-DD).
	// Empty string when not found in the document.
	DueDate string
	// AmountDue is the total amount due as a decimal string (e.g. "1234.56").
	// Empty string when not found.
	AmountDue string
	// PixPayload is the Pix copy-and-paste string. Empty when not found.
	PixPayload string
	// PixQRImageRef is a storage reference to the extracted QR image. Empty when not found.
	PixQRImageRef string
	// Barcode is the numeric barcode string. Empty when not found.
	Barcode string
}

// StatementLineResult holds a single transaction entry from a statement PDF.
type StatementLineResult struct {
	Date        string // ISO date (YYYY-MM-DD)
	Description string
	Amount      string // decimal string
	Direction   string // "credit" or "debit"
}

// StatementExtractionResult holds structured data extracted from a bank statement PDF.
type StatementExtractionResult struct {
	PeriodStart string // ISO date
	PeriodEnd   string // ISO date
	Lines       []StatementLineResult
}

// PDFExtractorIface is the narrow interface for PDF parsing backends.
// Concrete implementations may use pdfcpu, Tika, or an external OCR service.
type PDFExtractorIface interface {
	// ExtractBill parses the PDF at storageKey and returns structured bill fields.
	// Fields not found in the PDF are returned as empty strings.
	ExtractBill(ctx context.Context, storageKey string) (*BillExtractionResult, error)

	// ExtractStatement parses the PDF at storageKey and returns all transaction lines.
	ExtractStatement(ctx context.Context, storageKey string) (*StatementExtractionResult, error)
}

// StubPDFExtractor is a no-op implementation used in development and integration tests
// when actual PDF parsing is not available.
type StubPDFExtractor struct{}

// NewStubPDFExtractor returns a StubPDFExtractor.
func NewStubPDFExtractor() PDFExtractorIface {
	return &StubPDFExtractor{}
}

// ExtractBill returns placeholder values, marking all optional fields as not found.
func (e *StubPDFExtractor) ExtractBill(_ context.Context, _ string) (*BillExtractionResult, error) {
	return &BillExtractionResult{
		DueDate:   "1900-01-01",
		AmountDue: "0.00",
	}, nil
}

// ExtractStatement returns an empty transaction list.
func (e *StubPDFExtractor) ExtractStatement(_ context.Context, _ string) (*StatementExtractionResult, error) {
	return &StatementExtractionResult{
		PeriodStart: "1900-01-01",
		PeriodEnd:   "1900-01-31",
	}, nil
}
