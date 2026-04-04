// Package views defines all HTTP request and response contracts for BFF routes.
// No Huma or Echo framework types are imported; only JSON/doc/validation tags.
package views

// UploadDocumentInput carries the PDF bytes and metadata for document upload.
type UploadDocumentInput struct {
	// FileName is the original filename, passed as a query parameter.
	FileName string `query:"fileName" required:"true" doc:"Original filename of the uploaded PDF" validate:"required"`
	// RawBody holds the raw PDF bytes sent as the request body.
	RawBody []byte
}

// ClassifyDocumentInput provides the document kind for classification.
type ClassifyDocumentInput struct {
	DocumentID string `path:"documentId" doc:"Document UUID" validate:"required,uuid4"`
	Body       struct {
		Kind string `json:"kind" enum:"bill,statement" doc:"Document kind: bill or statement" validate:"required,oneof=bill statement"`
	}
}

// ListDocumentsInput carries optional filters and pagination for document listing.
type ListDocumentsInput struct {
	PageSize  int32  `query:"pageSize"  minimum:"1" maximum:"100" doc:"Page size (default 25 for document lists)"`
	PageToken string `query:"pageToken" doc:"Opaque cursor from a previous list response"`
}

// GetDocumentInput carries the document ID path parameter.
type GetDocumentInput struct {
	DocumentID string `path:"documentId" doc:"Document UUID" validate:"required,uuid4"`
}

// DocumentResponse is the JSON shape returned for a single document.
type DocumentResponse struct {
	ID              string `json:"id"`
	ProjectID       string `json:"projectId"`
	UploadedBy      string `json:"uploadedBy"`
	Kind            string `json:"kind"`
	FileName        string `json:"fileName"`
	AnalysisStatus  string `json:"analysisStatus"`
	StorageProvider string `json:"storageProvider,omitempty"`
	UploadedAt      string `json:"uploadedAt"`
	UpdatedAt       string `json:"updatedAt"`
}

// BillRecordResponse is the JSON shape returned for an extracted bill.
type BillRecordResponse struct {
	ID            string `json:"id"`
	DueDate       string `json:"dueDate"`
	AmountDue     string `json:"amountDue"`
	PixPayload    string `json:"pixPayload,omitempty"`
	PixQRImageRef string `json:"pixQrImageRef,omitempty"`
	Barcode       string `json:"barcode,omitempty"`
	PaymentStatus string `json:"paymentStatus"`
	PaidAt        string `json:"paidAt,omitempty"`
}

// TransactionLineResponse is a single line from a bank statement.
type TransactionLineResponse struct {
	ID                   string `json:"id"`
	TransactionDate      string `json:"transactionDate"`
	Description          string `json:"description"`
	Amount               string `json:"amount"`
	Direction            string `json:"direction"`
	ReconciliationStatus string `json:"reconciliationStatus"`
}

// StatementRecordResponse is the JSON shape returned for an extracted statement.
type StatementRecordResponse struct {
	ID            string                     `json:"id"`
	BankAccountID string                     `json:"bankAccountId,omitempty"`
	PeriodStart   string                     `json:"periodStart"`
	PeriodEnd     string                     `json:"periodEnd"`
	Lines         []*TransactionLineResponse `json:"lines"`
}

// DocumentDetailResponse extends DocumentResponse with optional extraction data.
type DocumentDetailResponse struct {
	DocumentResponse
	BillRecord      *BillRecordResponse      `json:"billRecord,omitempty"`
	StatementRecord *StatementRecordResponse `json:"statementRecord,omitempty"`
}

// ListDocumentsResponse is the JSON body for the list endpoint.
type ListDocumentsResponse struct {
	Items         []*DocumentResponse `json:"items"`
	NextPageToken string              `json:"nextPageToken,omitempty"`
}
