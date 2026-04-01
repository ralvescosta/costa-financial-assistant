package controllers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"

	"github.com/danielgtaylor/huma/v2"
	"go.uber.org/zap"

	bffmiddleware "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/transport/http/middleware"
	commonv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/common/v1"
	filesv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/files/v1"
)

// ─── Input / Output types ─────────────────────────────────────────────────────

// UploadDocumentInput carries the PDF bytes and metadata for document upload.
type UploadDocumentInput struct {
	// FileName is the original filename, passed as a query parameter.
	FileName string `query:"fileName" required:"true" doc:"Original filename of the uploaded PDF"`
	// RawBody holds the raw PDF bytes sent as the request body.
	RawBody []byte
}

// ClassifyDocumentInput provides the document kind for classification.
type ClassifyDocumentInput struct {
	DocumentID string `path:"documentId" doc:"Document UUID"`
	Body       struct {
		Kind string `json:"kind" enum:"bill,statement" doc:"Document kind: bill or statement"`
	}
}

// ListDocumentsInput carries optional filters and pagination for document listing.
type ListDocumentsInput struct {
	PageSize  int32  `query:"pageSize"  minimum:"1" maximum:"100" doc:"Page size (default 25)"`
	PageToken string `query:"pageToken" doc:"Opaque cursor from a previous list response"`
}

// GetDocumentInput carries the document ID path parameter.
type GetDocumentInput struct {
	DocumentID string `path:"documentId" doc:"Document UUID"`
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
	ID            string                    `json:"id"`
	BankAccountID string                    `json:"bankAccountId,omitempty"`
	PeriodStart   string                    `json:"periodStart"`
	PeriodEnd     string                    `json:"periodEnd"`
	Lines         []TransactionLineResponse `json:"lines"`
}

// DocumentDetailResponse extends DocumentResponse with optional extraction data.
type DocumentDetailResponse struct {
	DocumentResponse
	BillRecord      *BillRecordResponse      `json:"billRecord,omitempty"`
	StatementRecord *StatementRecordResponse `json:"statementRecord,omitempty"`
}

// ListDocumentsResponse is the JSON body for the list endpoint.
type ListDocumentsResponse struct {
	Items         []DocumentResponse `json:"items"`
	NextPageToken string             `json:"nextPageToken,omitempty"`
}

// ─── Controller ───────────────────────────────────────────────────────────────

// DocumentsController handles BFF document HTTP endpoints.
type DocumentsController struct {
	BaseController
	filesClient filesv1.FilesServiceClient
}

// NewDocumentsController constructs a DocumentsController.
func NewDocumentsController(logger *zap.Logger, filesClient filesv1.FilesServiceClient) *DocumentsController {
	return &DocumentsController{BaseController: BaseController{logger: logger}, filesClient: filesClient}
}

// ─── Handlers ─────────────────────────────────────────────────────────────────

// HandleUpload processes a raw PDF upload and registers the document record.
func (c *DocumentsController) HandleUpload(ctx context.Context, input *UploadDocumentInput) (*struct{ Body DocumentResponse }, error) {
	claims := bffmiddleware.ClaimsFromContext(ctx)
	if claims == nil {
		return nil, huma.Error403Forbidden("missing project context")
	}

	if len(input.RawBody) == 0 {
		return nil, huma.Error400BadRequest("request body must be the PDF file bytes")
	}

	// Compute SHA-256 hash of the incoming bytes for project-scoped deduplication.
	hash := sha256.New()
	if _, err := io.Copy(hash, io.NopCloser(newByteReader(input.RawBody))); err != nil {
		c.logger.Error("upload: hash computation failed", zap.Error(err))
		return nil, huma.Error500InternalServerError("upload failed")
	}
	fileHash := hex.EncodeToString(hash.Sum(nil))

	// For Phase 3, persist with a local storage reference.
	// Phase 4 will replace this with a real S3 upload before calling the gRPC service.
	storageKey := fmt.Sprintf("local/%s", fileHash)

	resp, err := c.filesClient.UploadDocument(ctx, &filesv1.UploadDocumentRequest{
		Ctx: &commonv1.ProjectContext{
			ProjectId: claims.GetProjectId(),
		},
		FileName:        input.FileName,
		FileHash:        fileHash,
		StorageProvider: "local",
		StorageKey:      storageKey,
		Audit: &commonv1.AuditMetadata{
			PerformedBy: claims.GetSubject(),
		},
	})
	if err != nil {
		return nil, c.grpcToHumaError(err, "upload failed")
	}

	c.logger.Info("upload: document registered",
		zap.String("document_id", resp.Document.Id),
		zap.String("project_id", claims.GetProjectId()))
	return &struct{ Body DocumentResponse }{Body: protoToResponse(resp.Document)}, nil
}

// HandleClassify updates a document's kind (bill or statement).
func (c *DocumentsController) HandleClassify(ctx context.Context, input *ClassifyDocumentInput) (*struct{ Body DocumentResponse }, error) {
	claims := bffmiddleware.ClaimsFromContext(ctx)
	if claims == nil {
		return nil, huma.Error403Forbidden("missing project context")
	}

	kind := kindFromString(input.Body.Kind)
	if kind == filesv1.DocumentKind_DOCUMENT_KIND_UNSPECIFIED {
		return nil, huma.Error400BadRequest("kind must be 'bill' or 'statement'")
	}

	resp, err := c.filesClient.ClassifyDocument(ctx, &filesv1.ClassifyDocumentRequest{
		Ctx: &commonv1.ProjectContext{
			ProjectId: claims.GetProjectId(),
		},
		DocumentId: input.DocumentID,
		Kind:       kind,
		Audit: &commonv1.AuditMetadata{
			PerformedBy: claims.GetSubject(),
		},
	})
	if err != nil {
		return nil, c.grpcToHumaError(err, "classify failed")
	}

	c.logger.Info("classify: document classified",
		zap.String("document_id", input.DocumentID),
		zap.String("kind", input.Body.Kind))
	return &struct{ Body DocumentResponse }{Body: protoToResponse(resp.Document)}, nil
}

// HandleList returns project-scoped documents with pagination.
func (c *DocumentsController) HandleList(ctx context.Context, input *ListDocumentsInput) (*struct{ Body ListDocumentsResponse }, error) {
	claims := bffmiddleware.ClaimsFromContext(ctx)
	if claims == nil {
		return nil, huma.Error403Forbidden("missing project context")
	}

	pageSize := input.PageSize
	if pageSize == 0 {
		pageSize = 25
	}

	resp, err := c.filesClient.ListDocuments(ctx, &filesv1.ListDocumentsRequest{
		Ctx: &commonv1.ProjectContext{
			ProjectId: claims.GetProjectId(),
		},
		Pagination: &commonv1.Pagination{
			PageSize:  pageSize,
			PageToken: input.PageToken,
		},
	})
	if err != nil {
		return nil, c.grpcToHumaError(err, "list documents failed")
	}

	items := make([]DocumentResponse, 0, len(resp.Documents))
	for _, d := range resp.Documents {
		items = append(items, protoToResponse(d))
	}

	body := ListDocumentsResponse{Items: items}
	if resp.Pagination != nil {
		body.NextPageToken = resp.Pagination.NextPageToken
	}
	return &struct{ Body ListDocumentsResponse }{Body: body}, nil
}

// HandleGet returns full document metadata including extraction fields.
func (c *DocumentsController) HandleGet(ctx context.Context, input *GetDocumentInput) (*struct{ Body DocumentDetailResponse }, error) {
	claims := bffmiddleware.ClaimsFromContext(ctx)
	if claims == nil {
		return nil, huma.Error403Forbidden("missing project context")
	}

	resp, err := c.filesClient.GetDocument(ctx, &filesv1.GetDocumentRequest{
		Ctx: &commonv1.ProjectContext{
			ProjectId: claims.GetProjectId(),
		},
		DocumentId: input.DocumentID,
	})
	if err != nil {
		return nil, c.grpcToHumaError(err, "get document failed")
	}

	body := DocumentDetailResponse{
		DocumentResponse: protoToResponse(resp.Document),
	}

	if resp.BillRecord != nil {
		body.BillRecord = protoBillToResponse(resp.BillRecord)
	}
	if resp.StatementRecord != nil {
		body.StatementRecord = protoStatementToResponse(resp.StatementRecord)
	}

	return &struct{ Body DocumentDetailResponse }{Body: body}, nil
}

// ─── helpers ──────────────────────────────────────────────────────────────────

func protoToResponse(d *filesv1.Document) DocumentResponse {
	if d == nil {
		return DocumentResponse{}
	}
	return DocumentResponse{
		ID:              d.Id,
		ProjectID:       d.ProjectId,
		UploadedBy:      d.UploadedBy,
		Kind:            kindToString(d.Kind),
		FileName:        d.FileName,
		AnalysisStatus:  analysisStatusToString(d.AnalysisStatus),
		StorageProvider: d.StorageProvider,
		UploadedAt:      d.UploadedAt,
		UpdatedAt:       d.UpdatedAt,
	}
}

func protoBillToResponse(b *filesv1.BillRecord) *BillRecordResponse {
	if b == nil {
		return nil
	}
	return &BillRecordResponse{
		ID:            b.Id,
		DueDate:       b.DueDate,
		AmountDue:     b.AmountDue,
		PixPayload:    b.PixPayload,
		PixQRImageRef: b.PixQrImageRef,
		Barcode:       b.Barcode,
		PaymentStatus: b.PaymentStatus,
		PaidAt:        b.PaidAt,
	}
}

func protoStatementToResponse(s *filesv1.StatementRecord) *StatementRecordResponse {
	if s == nil {
		return nil
	}
	lines := make([]TransactionLineResponse, 0, len(s.Lines))
	for _, l := range s.Lines {
		lines = append(lines, TransactionLineResponse{
			ID:                   l.Id,
			TransactionDate:      l.TransactionDate,
			Description:          l.Description,
			Amount:               l.Amount,
			Direction:            l.Direction,
			ReconciliationStatus: l.ReconciliationStatus,
		})
	}
	return &StatementRecordResponse{
		ID:            s.Id,
		BankAccountID: s.BankAccountId,
		PeriodStart:   s.PeriodStart,
		PeriodEnd:     s.PeriodEnd,
		Lines:         lines,
	}
}

func kindToString(k filesv1.DocumentKind) string {
	switch k {
	case filesv1.DocumentKind_DOCUMENT_KIND_BILL:
		return "bill"
	case filesv1.DocumentKind_DOCUMENT_KIND_STATEMENT:
		return "statement"
	default:
		return "unspecified"
	}
}

func kindFromString(s string) filesv1.DocumentKind {
	switch s {
	case "bill":
		return filesv1.DocumentKind_DOCUMENT_KIND_BILL
	case "statement":
		return filesv1.DocumentKind_DOCUMENT_KIND_STATEMENT
	default:
		return filesv1.DocumentKind_DOCUMENT_KIND_UNSPECIFIED
	}
}

func analysisStatusToString(s filesv1.AnalysisStatus) string {
	switch s {
	case filesv1.AnalysisStatus_ANALYSIS_STATUS_PROCESSING:
		return "processing"
	case filesv1.AnalysisStatus_ANALYSIS_STATUS_ANALYSED:
		return "analysed"
	case filesv1.AnalysisStatus_ANALYSIS_STATUS_ANALYSIS_FAILED:
		return "analysis_failed"
	default:
		return "pending"
	}
}

// newByteReader returns an io.Reader from a byte slice.
func newByteReader(b []byte) io.Reader {
	return &byteReader{data: b}
}

type byteReader struct {
	data []byte
	pos  int
}

func (r *byteReader) Read(p []byte) (int, error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n := copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}
