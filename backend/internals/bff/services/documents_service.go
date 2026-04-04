// Package services provides BFF service implementations that own all downstream
// gRPC orchestration on behalf of HTTP controllers.
package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"

	"go.uber.org/zap"

	bffinterfaces "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/interfaces"
	bffcontracts "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/services/contracts"
	apperrors "github.com/ralvescosta/costa-financial-assistant/backend/pkgs/errors"
	commonv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/common/v1"
	filesv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/files/v1"
)

// DocumentsServiceImpl implements bffinterfaces.DocumentsService using the Files gRPC client.
type DocumentsServiceImpl struct {
	logger      *zap.Logger
	filesClient bffinterfaces.FilesClient
}

// NewDocumentsService constructs a DocumentsServiceImpl.
func NewDocumentsService(logger *zap.Logger, filesClient bffinterfaces.FilesClient) bffinterfaces.DocumentsService {
	return &DocumentsServiceImpl{logger: logger, filesClient: filesClient}
}

// UploadDocument registers a document with the downstream files service.
func (s *DocumentsServiceImpl) UploadDocument(ctx context.Context, projectID, uploadedBy, fileName string, fileBytes []byte) (*bffcontracts.DocumentResponse, error) {
	hash := sha256.New()
	if _, err := io.Copy(hash, byteReader(fileBytes)); err != nil {
		s.logger.Error("documents_svc: hash computation failed", zap.Error(err))
		return nil, apperrors.TranslateError(err, "service")
	}
	fileHash := hex.EncodeToString(hash.Sum(nil))
	storageKey := fmt.Sprintf("local/%s", fileHash)

	resp, err := s.filesClient.UploadDocument(ctx, &filesv1.UploadDocumentRequest{
		Ctx:             projectContextFromContext(ctx, projectID, uploadedBy),
		Session:         sessionFromContext(ctx),
		FileName:        fileName,
		FileHash:        fileHash,
		StorageProvider: "local",
		StorageKey:      storageKey,
		Audit:           &commonv1.AuditMetadata{PerformedBy: uploadedBy},
	})
	if err != nil {
		s.logger.Error("documents_svc: upload downstream call failed",
			zap.String("project_id", projectID),
			zap.Error(err))
		if appErr := apperrors.AsAppError(err); appErr != nil {
			return nil, appErr
		}
		return nil, apperrors.TranslateError(err, "service")
	}
	s.logger.Info("documents_svc: document uploaded",
		zap.String("document_id", resp.Document.Id),
		zap.String("project_id", projectID))
	result := protoDocToView(resp.Document)
	return &result, nil
}

// ClassifyDocument updates a document's kind.
func (s *DocumentsServiceImpl) ClassifyDocument(ctx context.Context, projectID, documentID, kind string) (*bffcontracts.DocumentResponse, error) {
	projectCtx := projectContextFromContext(ctx, projectID, "")
	performedBy := projectCtx.GetUserId()
	if performedBy == "" {
		performedBy = projectID
	}

	resp, err := s.filesClient.ClassifyDocument(ctx, &filesv1.ClassifyDocumentRequest{
		Ctx:        projectCtx,
		Session:    sessionFromContext(ctx),
		DocumentId: documentID,
		Kind:       kindFromString(kind),
		Audit:      &commonv1.AuditMetadata{PerformedBy: performedBy},
	})
	if err != nil {
		s.logger.Error("documents_svc: classify downstream call failed",
			zap.String("project_id", projectID),
			zap.String("document_id", documentID),
			zap.Error(err))
		if appErr := apperrors.AsAppError(err); appErr != nil {
			return nil, appErr
		}
		return nil, apperrors.TranslateError(err, "service")
	}
	s.logger.Info("documents_svc: document classified",
		zap.String("document_id", documentID),
		zap.String("kind", kind))
	result := protoDocToView(resp.Document)
	return &result, nil
}

// ListDocuments returns a project-scoped page of documents.
func (s *DocumentsServiceImpl) ListDocuments(ctx context.Context, projectID string, pageSize int32, pageToken string) (*bffcontracts.ListDocumentsResponse, error) {
	resp, err := s.filesClient.ListDocuments(ctx, &filesv1.ListDocumentsRequest{
		Ctx:        projectContextFromContext(ctx, projectID, ""),
		Session:    sessionFromContext(ctx),
		Pagination: defaultPagination(pageSize, pageToken, 25),
	})
	if err != nil {
		s.logger.Error("documents_svc: list downstream call failed",
			zap.String("project_id", projectID),
			zap.Error(err))
		if appErr := apperrors.AsAppError(err); appErr != nil {
			return nil, appErr
		}
		return nil, apperrors.TranslateError(err, "service")
	}

	items := make([]*bffcontracts.DocumentResponse, 0, len(resp.Documents))
	for _, d := range resp.Documents {
		v := protoDocToView(d)
		items = append(items, &v)
	}
	result := bffcontracts.ListDocumentsResponse{Items: items}
	if resp.Pagination != nil {
		result.NextPageToken = resp.Pagination.NextPageToken
	}
	return &result, nil
}

// GetDocument returns full document metadata including extraction fields.
func (s *DocumentsServiceImpl) GetDocument(ctx context.Context, projectID, documentID string) (*bffcontracts.DocumentDetailResponse, error) {
	resp, err := s.filesClient.GetDocument(ctx, &filesv1.GetDocumentRequest{
		Ctx:        projectContextFromContext(ctx, projectID, ""),
		Session:    sessionFromContext(ctx),
		DocumentId: documentID,
	})
	if err != nil {
		s.logger.Error("documents_svc: get downstream call failed",
			zap.String("project_id", projectID),
			zap.String("document_id", documentID),
			zap.Error(err))
		if appErr := apperrors.AsAppError(err); appErr != nil {
			return nil, appErr
		}
		return nil, apperrors.TranslateError(err, "service")
	}

	result := bffcontracts.DocumentDetailResponse{DocumentResponse: protoDocToView(resp.Document)}
	if resp.BillRecord != nil {
		br := protoBillToView(resp.BillRecord)
		result.BillRecord = &br
	}
	if resp.StatementRecord != nil {
		sr := protoStatementToView(resp.StatementRecord)
		result.StatementRecord = &sr
	}
	return &result, nil
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func protoDocToView(d *filesv1.Document) bffcontracts.DocumentResponse {
	if d == nil {
		return bffcontracts.DocumentResponse{}
	}
	return bffcontracts.DocumentResponse{
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

func protoBillToView(b *filesv1.BillRecord) bffcontracts.BillRecordResponse {
	if b == nil {
		return bffcontracts.BillRecordResponse{}
	}
	return bffcontracts.BillRecordResponse{
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

func protoStatementToView(s *filesv1.StatementRecord) bffcontracts.StatementRecordResponse {
	if s == nil {
		return bffcontracts.StatementRecordResponse{}
	}
	lines := make([]*bffcontracts.TransactionLineResponse, 0, len(s.Lines))
	for _, l := range s.Lines {
		lines = append(lines, &bffcontracts.TransactionLineResponse{
			ID:                   l.Id,
			TransactionDate:      l.TransactionDate,
			Description:          l.Description,
			Amount:               l.Amount,
			Direction:            l.Direction,
			ReconciliationStatus: l.ReconciliationStatus,
		})
	}
	return bffcontracts.StatementRecordResponse{
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

// byteReader wraps a byte slice as an io.Reader for hashing.
func byteReader(b []byte) io.Reader {
	return &byteSliceReader{data: b}
}

type byteSliceReader struct {
	data []byte
	pos  int
}

func (r *byteSliceReader) Read(p []byte) (int, error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n := copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}
