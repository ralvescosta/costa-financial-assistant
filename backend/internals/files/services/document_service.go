package services

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"

	"github.com/ralvescosta/costa-financial-assistant/backend/internals/files/repositories"
	filesv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/files/v1"
)

// DocumentServiceIface is the narrow interface consumed by the gRPC server.
type DocumentServiceIface interface {
	UploadDocument(ctx context.Context, req *UploadDocumentInput) (*filesv1.Document, error)
	ClassifyDocument(ctx context.Context, projectID, documentID string, kind filesv1.DocumentKind) (*filesv1.Document, error)
	GetDocument(ctx context.Context, projectID, documentID string) (*filesv1.Document, error)
	ListDocuments(ctx context.Context, projectID string, pageSize int32, pageToken string) ([]*filesv1.Document, error)
}

// UploadDocumentInput carries all fields needed to register a new uploaded PDF.
type UploadDocumentInput struct {
	ProjectID       string
	UploadedBy      string
	FileName        string
	FileHash        string
	StorageProvider string
	StorageKey      string
}

// DocumentService implements DocumentServiceIface.
type DocumentService struct {
	repo   repositories.DocumentRepository
	uow    repositories.UnitOfWork
	logger *zap.Logger
}

// NewDocumentService constructs a DocumentService.
func NewDocumentService(repo repositories.DocumentRepository, uow repositories.UnitOfWork, logger *zap.Logger) DocumentServiceIface {
	return &DocumentService{repo: repo, uow: uow, logger: logger}
}

// UploadDocument registers a newly uploaded PDF. Returns ErrDuplicateDocument when the
// same file hash already exists in the project (project-scoped deduplication).
func (s *DocumentService) UploadDocument(ctx context.Context, req *UploadDocumentInput) (*filesv1.Document, error) {
	// Project-scoped duplicate detection — same hash already uploaded.
	existing, err := s.repo.FindByProjectAndHash(ctx, req.ProjectID, req.FileHash)
	if err != nil && !errors.Is(err, repositories.ErrDocumentNotFound) {
		s.logger.Error("upload: duplicate check failed",
			zap.String("project_id", req.ProjectID),
			zap.Error(err))
		return nil, fmt.Errorf("document service: upload: %w", err)
	}
	if existing != nil {
		s.logger.Info("upload: duplicate document detected",
			zap.String("project_id", req.ProjectID),
			zap.String("file_hash", req.FileHash))
		return nil, repositories.ErrDuplicateDocument
	}

	tx, err := s.uow.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("document service: upload: begin tx: %w", err)
	}
	defer s.uow.Rollback(tx) //nolint:errcheck

	doc := &filesv1.Document{
		ProjectId:       req.ProjectID,
		UploadedBy:      req.UploadedBy,
		Kind:            filesv1.DocumentKind_DOCUMENT_KIND_UNSPECIFIED,
		StorageProvider: req.StorageProvider,
		StorageKey:      req.StorageKey,
		FileName:        req.FileName,
		FileHash:        req.FileHash,
		AnalysisStatus:  filesv1.AnalysisStatus_ANALYSIS_STATUS_PENDING,
	}

	created, err := s.repo.Create(ctx, tx, doc)
	if err != nil {
		s.logger.Error("upload: create document failed",
			zap.String("project_id", req.ProjectID),
			zap.Error(err))
		return nil, fmt.Errorf("document service: upload: %w", err)
	}

	if err := s.uow.Commit(tx); err != nil {
		s.logger.Error("upload: commit failed",
			zap.String("document_id", created.Id),
			zap.Error(err))
		return nil, fmt.Errorf("document service: upload: commit: %w", err)
	}

	s.logger.Info("upload: document created",
		zap.String("document_id", created.Id),
		zap.String("project_id", req.ProjectID))
	return created, nil
}

// ClassifyDocument updates the document kind for an existing project-scoped document.
func (s *DocumentService) ClassifyDocument(ctx context.Context, projectID, documentID string, kind filesv1.DocumentKind) (*filesv1.Document, error) {
	tx, err := s.uow.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("document service: classify: begin tx: %w", err)
	}
	defer s.uow.Rollback(tx) //nolint:errcheck

	updated, err := s.repo.UpdateKind(ctx, tx, projectID, documentID, kind)
	if err != nil {
		if errors.Is(err, repositories.ErrDocumentNotFound) {
			return nil, repositories.ErrDocumentNotFound
		}
		s.logger.Error("classify: update kind failed",
			zap.String("project_id", projectID),
			zap.String("document_id", documentID),
			zap.Error(err))
		return nil, fmt.Errorf("document service: classify: %w", err)
	}

	if err := s.uow.Commit(tx); err != nil {
		s.logger.Error("classify: commit failed",
			zap.String("document_id", documentID),
			zap.Error(err))
		return nil, fmt.Errorf("document service: classify: commit: %w", err)
	}

	s.logger.Info("classify: document classified",
		zap.String("document_id", documentID),
		zap.String("project_id", projectID))
	return updated, nil
}

// GetDocument retrieves a single project-scoped document by ID.
func (s *DocumentService) GetDocument(ctx context.Context, projectID, documentID string) (*filesv1.Document, error) {
	doc, err := s.repo.FindByProjectAndID(ctx, projectID, documentID)
	if err != nil {
		if errors.Is(err, repositories.ErrDocumentNotFound) {
			return nil, repositories.ErrDocumentNotFound
		}
		s.logger.Error("get: find document failed",
			zap.String("project_id", projectID),
			zap.String("document_id", documentID),
			zap.Error(err))
		return nil, fmt.Errorf("document service: get: %w", err)
	}
	return doc, nil
}

// ListDocuments returns project-scoped documents in reverse-chronological order.
func (s *DocumentService) ListDocuments(ctx context.Context, projectID string, pageSize int32, pageToken string) ([]*filesv1.Document, error) {
	docs, err := s.repo.ListByProject(ctx, projectID, pageSize, pageToken)
	if err != nil {
		s.logger.Error("list: query failed",
			zap.String("project_id", projectID),
			zap.Error(err))
		return nil, fmt.Errorf("document service: list: %w", err)
	}
	return docs, nil
}
