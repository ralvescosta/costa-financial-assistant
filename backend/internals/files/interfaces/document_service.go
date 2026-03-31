// Package interfaces defines the canonical service and repository contracts for the files domain.
// These interfaces consolidate the key contracts used by the gRPC server and are used as mock targets in tests.
package interfaces

import (
	"context"
	"database/sql"

	filesv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/files/v1"
)

// DocumentService defines the contract for document upload, classification, and retrieval.
// It is implemented by services.DocumentService and consumed by the files gRPC server.
type DocumentService interface {
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

// ExtractionService defines the contract for async PDF analysis orchestration.
// It is implemented by services.ExtractionService and consumed by the files gRPC server and RMQ consumer.
type ExtractionService interface {
	// ProcessDocument transitions the document through analysis states and persists the result.
	ProcessDocument(ctx context.Context, jobID, projectID, documentID string, kind filesv1.DocumentKind) error

	// GetDocumentDetail returns the document with its extracted BillRecord or StatementRecord.
	GetDocumentDetail(ctx context.Context, projectID, documentID string) (*filesv1.Document, *filesv1.BillRecord, *filesv1.StatementRecord, error)
}

// DocumentRepository defines the project-scoped persistence contract for documents.
// It is implemented by repositories.PostgresDocumentRepository.
type DocumentRepository interface {
	Create(ctx context.Context, tx *sql.Tx, doc *filesv1.Document) (*filesv1.Document, error)
	FindByProjectAndHash(ctx context.Context, projectID, hash string) (*filesv1.Document, error)
	FindByProjectAndID(ctx context.Context, projectID, id string) (*filesv1.Document, error)
	UpdateKind(ctx context.Context, tx *sql.Tx, projectID, id string, kind filesv1.DocumentKind) (*filesv1.Document, error)
	ListByProject(ctx context.Context, projectID string, pageSize int32, offsetToken string) ([]*filesv1.Document, error)
}

// AnalysisJobRepository defines the persistence contract for async analysis jobs.
// It is implemented by repositories.PostgresAnalysisJobRepository.
type AnalysisJobRepository interface {
	Create(ctx context.Context, tx *sql.Tx, job *filesv1.AnalysisJob) (*filesv1.AnalysisJob, error)
	FindByDocumentID(ctx context.Context, projectID, documentID string) (*filesv1.AnalysisJob, error)
	UpdateStatus(ctx context.Context, tx *sql.Tx, jobID, status, lastError string, attemptCount int32) error
	UpdateDocumentAnalysisStatus(ctx context.Context, tx *sql.Tx, projectID, documentID, analysisStatus, failureReason string) error
}
