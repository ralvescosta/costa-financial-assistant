package services

import (
	"context"
	"errors"

	"go.uber.org/zap"

	"github.com/ralvescosta/costa-financial-assistant/backend/internals/files/interfaces"
	"github.com/ralvescosta/costa-financial-assistant/backend/internals/files/repositories"
	apperrors "github.com/ralvescosta/costa-financial-assistant/backend/pkgs/errors"
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
	docRepo       interfaces.DocumentRepository
	jobRepo       interfaces.AnalysisJobRepository
	billRepo      interfaces.BillRecordRepository
	statementRepo interfaces.StatementRecordRepository
	uow           interfaces.UnitOfWork
	logger        *zap.Logger
	// extractor is the PDF extraction backend (swappable for testing).
	extractor PDFExtractorIface
}

// NewExtractionService constructs an ExtractionService.
func NewExtractionService(
	docRepo interfaces.DocumentRepository,
	jobRepo interfaces.AnalysisJobRepository,
	billRepo interfaces.BillRecordRepository,
	statementRepo interfaces.StatementRecordRepository,
	uow interfaces.UnitOfWork,
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
		s.logger.Error("extraction: begin tx failed",
			zap.String("project_id", projectID),
			zap.String("document_id", documentID),
			zap.Error(err))
		return apperrors.TranslateError(err, "service")
	}
	defer s.uow.Rollback(tx) //nolint:errcheck

	if err := s.jobRepo.UpdateDocumentAnalysisStatus(ctx, tx, projectID, documentID, "processing", ""); err != nil {
		if appErr := apperrors.AsAppError(err); appErr != nil {
			return appErr
		}
		s.logger.Error("extraction: set processing status failed",
			zap.String("document_id", documentID),
			zap.Error(err))
		return apperrors.TranslateError(err, "service")
	}
	if err := s.jobRepo.UpdateStatus(ctx, tx, jobID, "running", "", 1); err != nil {
		if appErr := apperrors.AsAppError(err); appErr != nil {
			return appErr
		}
		s.logger.Error("extraction: set job running status failed",
			zap.String("job_id", jobID),
			zap.String("document_id", documentID),
			zap.Error(err))
		return apperrors.TranslateError(err, "service")
	}
	if err := s.uow.Commit(tx); err != nil {
		s.logger.Error("extraction: commit processing flag failed",
			zap.String("job_id", jobID),
			zap.String("document_id", documentID),
			zap.Error(err))
		return apperrors.TranslateError(err, "service")
	}

	// ── 2. Perform PDF extraction (outside the DB transaction) ──────────────
	var extractErr error
	switch kind {
	case filesv1.DocumentKind_DOCUMENT_KIND_BILL:
		extractErr = s.processBill(ctx, projectID, documentID)
	case filesv1.DocumentKind_DOCUMENT_KIND_STATEMENT:
		extractErr = s.processStatement(ctx, projectID, documentID)
	default:
		extractErr = apperrors.NewCatalogError(apperrors.ErrValidationError)
	}

	// ── 3. Persist final status ──────────────────────────────────────────────
	finalTx, err := s.uow.Begin(ctx)
	if err != nil {
		s.logger.Error("extraction: begin final tx failed",
			zap.String("job_id", jobID),
			zap.String("document_id", documentID),
			zap.Error(err))
		return apperrors.TranslateError(err, "service")
	}
	defer s.uow.Rollback(finalTx) //nolint:errcheck

	if extractErr != nil {
		if appErr := apperrors.AsAppError(extractErr); appErr != nil {
			extractErr = appErr
		} else {
			extractErr = apperrors.TranslateError(extractErr, "service")
		}
		s.logger.Warn("extraction: analysis failed",
			zap.String("document_id", documentID),
			zap.Error(extractErr))
		_ = s.jobRepo.UpdateDocumentAnalysisStatus(ctx, finalTx, projectID, documentID, "analysis_failed", extractErr.Error())
		_ = s.jobRepo.UpdateStatus(ctx, finalTx, jobID, "failed", extractErr.Error(), 1)
		_ = s.uow.Commit(finalTx)
		return extractErr
	}

	if err := s.jobRepo.UpdateDocumentAnalysisStatus(ctx, finalTx, projectID, documentID, "analysed", ""); err != nil {
		if appErr := apperrors.AsAppError(err); appErr != nil {
			return appErr
		}
		s.logger.Error("extraction: set analysed status failed",
			zap.String("document_id", documentID),
			zap.Error(err))
		return apperrors.TranslateError(err, "service")
	}
	if err := s.jobRepo.UpdateStatus(ctx, finalTx, jobID, "succeeded", "", 1); err != nil {
		if appErr := apperrors.AsAppError(err); appErr != nil {
			return appErr
		}
		s.logger.Error("extraction: set job succeeded status failed",
			zap.String("job_id", jobID),
			zap.Error(err))
		return apperrors.TranslateError(err, "service")
	}
	if err := s.uow.Commit(finalTx); err != nil {
		s.logger.Error("extraction: commit final status failed",
			zap.String("job_id", jobID),
			zap.String("document_id", documentID),
			zap.Error(err))
		return apperrors.TranslateError(err, "service")
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
			if appErr := apperrors.AsAppError(err); appErr != nil {
				return doc, nil, nil, appErr
			}
			s.logger.Error("extraction: get bill record failed",
				zap.String("document_id", documentID),
				zap.Error(err))
			return doc, nil, nil, apperrors.TranslateError(err, "service")
		}
		return doc, bill, nil, nil

	case filesv1.DocumentKind_DOCUMENT_KIND_STATEMENT:
		stmt, err := s.statementRepo.FindByProjectAndDocumentID(ctx, projectID, documentID)
		if err != nil && !errors.Is(err, repositories.ErrStatementRecordNotFound) {
			if appErr := apperrors.AsAppError(err); appErr != nil {
				return doc, nil, nil, appErr
			}
			s.logger.Error("extraction: get statement record failed",
				zap.String("document_id", documentID),
				zap.Error(err))
			return doc, nil, nil, apperrors.TranslateError(err, "service")
		}
		return doc, nil, stmt, nil
	}

	return doc, nil, nil, nil
}

// ── internal helpers ──────────────────────────────────────────────────────────

func (s *ExtractionService) processBill(ctx context.Context, projectID, documentID string) error {
	doc, err := s.docRepo.FindByProjectAndID(ctx, projectID, documentID)
	if err != nil {
		if appErr := apperrors.AsAppError(err); appErr != nil {
			return appErr
		}
		s.logger.Error("extraction: load document for bill failed",
			zap.String("project_id", projectID),
			zap.String("document_id", documentID),
			zap.Error(err))
		return apperrors.TranslateError(err, "service")
	}

	result, err := s.extractor.ExtractBill(ctx, doc.StorageKey)
	if err != nil {
		s.logger.Error("extraction: bill PDF extraction failed",
			zap.String("project_id", projectID),
			zap.String("document_id", documentID),
			zap.Error(err))
		return apperrors.TranslateError(err, "service")
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
		s.logger.Error("extraction: begin tx for bill persistence failed",
			zap.String("project_id", projectID),
			zap.String("document_id", documentID),
			zap.Error(err))
		return apperrors.TranslateError(err, "service")
	}
	defer s.uow.Rollback(tx) //nolint:errcheck

	if _, err := s.billRepo.Create(ctx, tx, record); err != nil {
		if appErr := apperrors.AsAppError(err); appErr != nil {
			return appErr
		}
		s.logger.Error("extraction: persist bill record failed",
			zap.String("project_id", projectID),
			zap.String("document_id", documentID),
			zap.Error(err))
		return apperrors.TranslateError(err, "service")
	}
	if err := s.uow.Commit(tx); err != nil {
		s.logger.Error("extraction: commit bill record failed",
			zap.String("project_id", projectID),
			zap.String("document_id", documentID),
			zap.Error(err))
		return apperrors.TranslateError(err, "service")
	}
	return nil
}

func (s *ExtractionService) processStatement(ctx context.Context, projectID, documentID string) error {
	doc, err := s.docRepo.FindByProjectAndID(ctx, projectID, documentID)
	if err != nil {
		if appErr := apperrors.AsAppError(err); appErr != nil {
			return appErr
		}
		s.logger.Error("extraction: load document for statement failed",
			zap.String("project_id", projectID),
			zap.String("document_id", documentID),
			zap.Error(err))
		return apperrors.TranslateError(err, "service")
	}

	result, err := s.extractor.ExtractStatement(ctx, doc.StorageKey)
	if err != nil {
		s.logger.Error("extraction: statement PDF extraction failed",
			zap.String("project_id", projectID),
			zap.String("document_id", documentID),
			zap.Error(err))
		return apperrors.TranslateError(err, "service")
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
		s.logger.Error("extraction: begin tx for statement persistence failed",
			zap.String("project_id", projectID),
			zap.String("document_id", documentID),
			zap.Error(err))
		return apperrors.TranslateError(err, "service")
	}
	defer s.uow.Rollback(tx) //nolint:errcheck

	if _, err := s.statementRepo.Create(ctx, tx, record); err != nil {
		if appErr := apperrors.AsAppError(err); appErr != nil {
			return appErr
		}
		s.logger.Error("extraction: persist statement record failed",
			zap.String("project_id", projectID),
			zap.String("document_id", documentID),
			zap.Error(err))
		return apperrors.TranslateError(err, "service")
	}
	if err := s.uow.Commit(tx); err != nil {
		s.logger.Error("extraction: commit statement record failed",
			zap.String("project_id", projectID),
			zap.String("document_id", documentID),
			zap.Error(err))
		return apperrors.TranslateError(err, "service")
	}
	return nil
}
