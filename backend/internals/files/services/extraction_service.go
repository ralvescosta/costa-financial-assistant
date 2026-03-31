package services

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"

	"github.com/ralvescosta/costa-financial-assistant/backend/internals/files/repositories"
	filesv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/files/v1"
)

// ExtractionServiceIface orchestrates async analysis for a classified document.
type ExtractionServiceIface interface {
	// ProcessDocument transitions the document to "processing", runs extraction
	// for the given kind, persists the result, and transitions to "analysed" or
	// "analysis_failed". It also updates the corresponding analysis job status.
	ProcessDocument(ctx context.Context, jobID, projectID, documentID string, kind filesv1.DocumentKind) error

	// GetDocumentDetail returns the document together with its extracted
	// BillRecord or StatementRecord if the analysis has completed.
	GetDocumentDetail(ctx context.Context, projectID, documentID string) (*filesv1.Document, *filesv1.BillRecord, *filesv1.StatementRecord, error)
}

// ExtractionService implements ExtractionServiceIface.
type ExtractionService struct {
	docRepo       repositories.DocumentRepository
	jobRepo       repositories.AnalysisJobRepository
	billRepo      repositories.BillRecordRepository
	statementRepo repositories.StatementRecordRepository
	uow           repositories.UnitOfWork
	logger        *zap.Logger
	// extractor is the PDF extraction backend (swappable for testing).
	extractor PDFExtractorIface
}

// NewExtractionService constructs an ExtractionService.
func NewExtractionService(
	docRepo repositories.DocumentRepository,
	jobRepo repositories.AnalysisJobRepository,
	billRepo repositories.BillRecordRepository,
	statementRepo repositories.StatementRecordRepository,
	uow repositories.UnitOfWork,
	extractor PDFExtractorIface,
	logger *zap.Logger,
) ExtractionServiceIface {
	return &ExtractionService{
		docRepo:       docRepo,
		jobRepo:       jobRepo,
		billRepo:      billRepo,
		statementRepo: statementRepo,
		uow:           uow,
		extractor:     extractor,
		logger:        logger,
	}
}

// ProcessDocument runs the full extraction pipeline for a single document.
func (s *ExtractionService) ProcessDocument(ctx context.Context, jobID, projectID, documentID string, kind filesv1.DocumentKind) error {
	// ── 1. Transition document → processing ────────────────────────────────
	tx, err := s.uow.Begin(ctx)
	if err != nil {
		return fmt.Errorf("extraction service: begin tx: %w", err)
	}
	defer s.uow.Rollback(tx) //nolint:errcheck

	if err := s.jobRepo.UpdateDocumentAnalysisStatus(ctx, tx, projectID, documentID, "processing", ""); err != nil {
		s.logger.Error("extraction: set processing status failed",
			zap.String("document_id", documentID),
			zap.Error(err))
		return fmt.Errorf("extraction service: set processing: %w", err)
	}
	if err := s.jobRepo.UpdateStatus(ctx, tx, jobID, "running", "", 1); err != nil {
		return fmt.Errorf("extraction service: set job running: %w", err)
	}
	if err := s.uow.Commit(tx); err != nil {
		return fmt.Errorf("extraction service: commit processing flag: %w", err)
	}

	// ── 2. Perform PDF extraction (outside the DB transaction) ──────────────
	var extractErr error
	switch kind {
	case filesv1.DocumentKind_DOCUMENT_KIND_BILL:
		extractErr = s.processBill(ctx, projectID, documentID)
	case filesv1.DocumentKind_DOCUMENT_KIND_STATEMENT:
		extractErr = s.processStatement(ctx, projectID, documentID)
	default:
		extractErr = fmt.Errorf("unsupported document kind: %v", kind)
	}

	// ── 3. Persist final status ──────────────────────────────────────────────
	finalTx, err := s.uow.Begin(ctx)
	if err != nil {
		return fmt.Errorf("extraction service: begin final tx: %w", err)
	}
	defer s.uow.Rollback(finalTx) //nolint:errcheck

	if extractErr != nil {
		s.logger.Warn("extraction: analysis failed",
			zap.String("document_id", documentID),
			zap.Error(extractErr))
		_ = s.jobRepo.UpdateDocumentAnalysisStatus(ctx, finalTx, projectID, documentID, "analysis_failed", extractErr.Error())
		_ = s.jobRepo.UpdateStatus(ctx, finalTx, jobID, "failed", extractErr.Error(), 1)
		_ = s.uow.Commit(finalTx)
		return fmt.Errorf("extraction service: extraction failed: %w", extractErr)
	}

	if err := s.jobRepo.UpdateDocumentAnalysisStatus(ctx, finalTx, projectID, documentID, "analysed", ""); err != nil {
		return fmt.Errorf("extraction service: set analysed: %w", err)
	}
	if err := s.jobRepo.UpdateStatus(ctx, finalTx, jobID, "succeeded", "", 1); err != nil {
		return fmt.Errorf("extraction service: set job succeeded: %w", err)
	}
	if err := s.uow.Commit(finalTx); err != nil {
		return fmt.Errorf("extraction service: commit final status: %w", err)
	}

	s.logger.Info("extraction: document analysed",
		zap.String("document_id", documentID),
		zap.String("project_id", projectID))
	return nil
}

// GetDocumentDetail returns the document with its extracted bill or statement record.
func (s *ExtractionService) GetDocumentDetail(ctx context.Context, projectID, documentID string) (*filesv1.Document, *filesv1.BillRecord, *filesv1.StatementRecord, error) {
	doc, err := s.docRepo.FindByProjectAndID(ctx, projectID, documentID)
	if err != nil {
		return nil, nil, nil, err
	}

	if doc.AnalysisStatus != filesv1.AnalysisStatus_ANALYSIS_STATUS_ANALYSED {
		return doc, nil, nil, nil
	}

	switch doc.Kind {
	case filesv1.DocumentKind_DOCUMENT_KIND_BILL:
		bill, err := s.billRepo.FindByProjectAndDocumentID(ctx, projectID, documentID)
		if err != nil && !errors.Is(err, repositories.ErrBillRecordNotFound) {
			s.logger.Error("extraction: get bill record failed",
				zap.String("document_id", documentID),
				zap.Error(err))
			return doc, nil, nil, fmt.Errorf("extraction service: get bill record: %w", err)
		}
		return doc, bill, nil, nil

	case filesv1.DocumentKind_DOCUMENT_KIND_STATEMENT:
		stmt, err := s.statementRepo.FindByProjectAndDocumentID(ctx, projectID, documentID)
		if err != nil && !errors.Is(err, repositories.ErrStatementRecordNotFound) {
			s.logger.Error("extraction: get statement record failed",
				zap.String("document_id", documentID),
				zap.Error(err))
			return doc, nil, nil, fmt.Errorf("extraction service: get statement record: %w", err)
		}
		return doc, nil, stmt, nil
	}

	return doc, nil, nil, nil
}

// ── internal helpers ──────────────────────────────────────────────────────────

func (s *ExtractionService) processBill(ctx context.Context, projectID, documentID string) error {
	doc, err := s.docRepo.FindByProjectAndID(ctx, projectID, documentID)
	if err != nil {
		return fmt.Errorf("load document: %w", err)
	}

	result, err := s.extractor.ExtractBill(ctx, doc.StorageKey)
	if err != nil {
		return fmt.Errorf("pdf extraction: %w", err)
	}

	record := &filesv1.BillRecord{
		ProjectId:     projectID,
		DocumentId:    documentID,
		DueDate:       result.DueDate,
		AmountDue:     result.AmountDue,
		PixPayload:    result.PixPayload,
		PixQrImageRef: result.PixQRImageRef,
		Barcode:       result.Barcode,
	}

	tx, err := s.uow.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer s.uow.Rollback(tx) //nolint:errcheck

	if _, err := s.billRepo.Create(ctx, tx, record); err != nil {
		return fmt.Errorf("persist bill record: %w", err)
	}
	return s.uow.Commit(tx)
}

func (s *ExtractionService) processStatement(ctx context.Context, projectID, documentID string) error {
	doc, err := s.docRepo.FindByProjectAndID(ctx, projectID, documentID)
	if err != nil {
		return fmt.Errorf("load document: %w", err)
	}

	result, err := s.extractor.ExtractStatement(ctx, doc.StorageKey)
	if err != nil {
		return fmt.Errorf("pdf extraction: %w", err)
	}

	lines := make([]*filesv1.TransactionLine, 0, len(result.Lines))
	for _, l := range result.Lines {
		lines = append(lines, &filesv1.TransactionLine{
			TransactionDate: l.Date,
			Description:     l.Description,
			Amount:          l.Amount,
			Direction:       l.Direction,
		})
	}

	record := &filesv1.StatementRecord{
		ProjectId:   projectID,
		DocumentId:  documentID,
		PeriodStart: result.PeriodStart,
		PeriodEnd:   result.PeriodEnd,
		Lines:       lines,
	}

	tx, err := s.uow.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer s.uow.Rollback(tx) //nolint:errcheck

	if _, err := s.statementRepo.Create(ctx, tx, record); err != nil {
		return fmt.Errorf("persist statement record: %w", err)
	}
	return s.uow.Commit(tx)
}
