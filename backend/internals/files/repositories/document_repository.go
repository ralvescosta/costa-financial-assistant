package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"

	"github.com/ralvescosta/costa-financial-assistant/backend/internals/files/interfaces"
	filesv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/files/v1"
)

var tracer = otel.Tracer("files/repositories")

// ErrDocumentNotFound is returned when a queried document does not exist in the project scope.
var ErrDocumentNotFound = errors.New("document not found")

// ErrDuplicateDocument is returned when a file with the same hash already exists in the project.
var ErrDuplicateDocument = errors.New("document already uploaded in this project")

// PostgresDocumentRepository implements DocumentRepository using PostgreSQL.
type PostgresDocumentRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewDocumentRepository constructs a PostgresDocumentRepository.
func NewDocumentRepository(db *sql.DB, logger *zap.Logger) interfaces.DocumentRepository {
	return &PostgresDocumentRepository{db: db, logger: logger}
}

// Create inserts a new document row inside the provided transaction.
func (r *PostgresDocumentRepository) Create(ctx context.Context, tx *sql.Tx, doc *filesv1.Document) (*filesv1.Document, error) {
	ctx, span := tracer.Start(ctx, "document.create")
	defer span.End()

	const query = `
		INSERT INTO documents
			(project_id, uploaded_by, kind, storage_provider, storage_key, file_name, file_hash, analysis_status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, uploaded_at, updated_at`

	var id, uploadedAt, updatedAt string
	err := tx.QueryRowContext(ctx, query,
		doc.ProjectId,
		doc.UploadedBy,
		documentKindToSQL(doc.Kind),
		doc.StorageProvider,
		doc.StorageKey,
		doc.FileName,
		doc.FileHash,
		analysisStatusToSQL(doc.AnalysisStatus),
	).Scan(&id, &uploadedAt, &updatedAt)
	if err != nil {
		span.RecordError(err)
		if isDuplicateConstraint(err) {
			return nil, ErrDuplicateDocument
		}
		r.logger.Error("document.create: insert failed",
			zap.String("project_id", doc.ProjectId),
			zap.Error(err))
		return nil, fmt.Errorf("document repository: create: %w", err)
	}

	doc.Id = id
	doc.UploadedAt = uploadedAt
	doc.UpdatedAt = updatedAt
	return doc, nil
}

// FindByProjectAndHash returns the document matching (projectID, fileHash), or ErrDocumentNotFound.
func (r *PostgresDocumentRepository) FindByProjectAndHash(ctx context.Context, projectID, hash string) (*filesv1.Document, error) {
	ctx, span := tracer.Start(ctx, "document.findByProjectAndHash")
	defer span.End()
	span.SetAttributes(attribute.String("project_id", projectID))

	const query = `
		SELECT id, project_id, uploaded_by, kind, storage_provider, storage_key,
		       file_name, file_hash, analysis_status, COALESCE(failure_reason,''), uploaded_at, updated_at
		FROM documents
		WHERE project_id = $1 AND file_hash = $2`

	row := r.db.QueryRowContext(ctx, query, projectID, hash)
	doc, err := scanDocument(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrDocumentNotFound
		}
		span.RecordError(err)
		r.logger.Error("document.findByProjectAndHash: query failed",
			zap.String("project_id", projectID),
			zap.Error(err))
		return nil, fmt.Errorf("document repository: findByProjectAndHash: %w", err)
	}
	return doc, nil
}

// FindByProjectAndID returns the document matching (projectID, id), or ErrDocumentNotFound.
func (r *PostgresDocumentRepository) FindByProjectAndID(ctx context.Context, projectID, id string) (*filesv1.Document, error) {
	ctx, span := tracer.Start(ctx, "document.findByProjectAndID")
	defer span.End()
	span.SetAttributes(
		attribute.String("project_id", projectID),
		attribute.String("document_id", id),
	)

	const query = `
		SELECT id, project_id, uploaded_by, kind, storage_provider, storage_key,
		       file_name, file_hash, analysis_status, COALESCE(failure_reason,''), uploaded_at, updated_at
		FROM documents
		WHERE project_id = $1 AND id = $2`

	row := r.db.QueryRowContext(ctx, query, projectID, id)
	doc, err := scanDocument(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrDocumentNotFound
		}
		span.RecordError(err)
		r.logger.Error("document.findByProjectAndID: query failed",
			zap.String("project_id", projectID),
			zap.String("document_id", id),
			zap.Error(err))
		return nil, fmt.Errorf("document repository: findByProjectAndID: %w", err)
	}
	return doc, nil
}

// UpdateKind sets the document kind for the given project-scoped document.
func (r *PostgresDocumentRepository) UpdateKind(ctx context.Context, tx *sql.Tx, projectID, id string, kind filesv1.DocumentKind) (*filesv1.Document, error) {
	ctx, span := tracer.Start(ctx, "document.updateKind")
	defer span.End()
	span.SetAttributes(
		attribute.String("project_id", projectID),
		attribute.String("document_id", id),
	)

	const query = `
		UPDATE documents
		SET kind = $1, updated_at = $2
		WHERE project_id = $3 AND id = $4
		RETURNING id, project_id, uploaded_by, kind, storage_provider, storage_key,
		          file_name, file_hash, analysis_status, COALESCE(failure_reason,''), uploaded_at, updated_at`

	now := time.Now().UTC().Format(time.RFC3339)
	row := tx.QueryRowContext(ctx, query, documentKindToSQL(kind), now, projectID, id)
	doc, err := scanDocument(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrDocumentNotFound
		}
		span.RecordError(err)
		r.logger.Error("document.updateKind: update failed",
			zap.String("project_id", projectID),
			zap.String("document_id", id),
			zap.Error(err))
		return nil, fmt.Errorf("document repository: updateKind: %w", err)
	}
	return doc, nil
}

// ListByProject returns project-scoped documents ordered by upload time descending.
// offsetToken is an opaque cursor (RFC3339 timestamp) for keyset pagination.
func (r *PostgresDocumentRepository) ListByProject(ctx context.Context, projectID string, pageSize int32, offsetToken string) ([]*filesv1.Document, error) {
	ctx, span := tracer.Start(ctx, "document.listByProject")
	defer span.End()
	span.SetAttributes(attribute.String("project_id", projectID))

	if pageSize <= 0 || pageSize > 100 {
		pageSize = 25
	}

	var (
		rows *sql.Rows
		err  error
	)

	if offsetToken == "" {
		const query = `
			SELECT id, project_id, uploaded_by, kind, storage_provider, storage_key,
			       file_name, file_hash, analysis_status, COALESCE(failure_reason,''), uploaded_at, updated_at
			FROM documents
			WHERE project_id = $1
			ORDER BY uploaded_at DESC, id DESC
			LIMIT $2`
		rows, err = r.db.QueryContext(ctx, query, projectID, pageSize)
	} else {
		const query = `
			SELECT id, project_id, uploaded_by, kind, storage_provider, storage_key,
			       file_name, file_hash, analysis_status, COALESCE(failure_reason,''), uploaded_at, updated_at
			FROM documents
			WHERE project_id = $1 AND uploaded_at < $2::TIMESTAMPTZ
			ORDER BY uploaded_at DESC, id DESC
			LIMIT $3`
		rows, err = r.db.QueryContext(ctx, query, projectID, offsetToken, pageSize)
	}
	if err != nil {
		span.RecordError(err)
		r.logger.Error("document.listByProject: query failed",
			zap.String("project_id", projectID),
			zap.Error(err))
		return nil, fmt.Errorf("document repository: listByProject: %w", err)
	}
	defer rows.Close() //nolint:errcheck

	var docs []*filesv1.Document
	for rows.Next() {
		var (
			id, projectID, uploadedBy                            string
			kindStr, storageProvider, storageKey                 string
			fileName, fileHash, analysisStatusStr, failureReason string
			uploadedAt, updatedAt                                string
		)
		if err := rows.Scan(
			&id, &projectID, &uploadedBy,
			&kindStr, &storageProvider, &storageKey,
			&fileName, &fileHash, &analysisStatusStr, &failureReason,
			&uploadedAt, &updatedAt,
		); err != nil {
			span.RecordError(err)
			return nil, fmt.Errorf("document repository: listByProject scan: %w", err)
		}
		docs = append(docs, &filesv1.Document{
			Id:              id,
			ProjectId:       projectID,
			UploadedBy:      uploadedBy,
			Kind:            documentKindFromSQL(kindStr),
			StorageProvider: storageProvider,
			StorageKey:      storageKey,
			FileName:        fileName,
			FileHash:        fileHash,
			AnalysisStatus:  analysisStatusFromSQL(analysisStatusStr),
			FailureReason:   failureReason,
			UploadedAt:      uploadedAt,
			UpdatedAt:       updatedAt,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("document repository: listByProject rows: %w", err)
	}
	return docs, nil
}

// ─── helpers ──────────────────────────────────────────────────────────────────

type rowScanner interface {
	Scan(dest ...any) error
}

func scanDocument(row rowScanner) (*filesv1.Document, error) {
	var (
		id, projectID, uploadedBy                            string
		kindStr, storageProvider, storageKey                 string
		fileName, fileHash, analysisStatusStr, failureReason string
		uploadedAt, updatedAt                                string
	)
	if err := row.Scan(
		&id, &projectID, &uploadedBy,
		&kindStr, &storageProvider, &storageKey,
		&fileName, &fileHash, &analysisStatusStr, &failureReason,
		&uploadedAt, &updatedAt,
	); err != nil {
		return nil, err
	}
	return &filesv1.Document{
		Id:              id,
		ProjectId:       projectID,
		UploadedBy:      uploadedBy,
		Kind:            documentKindFromSQL(kindStr),
		StorageProvider: storageProvider,
		StorageKey:      storageKey,
		FileName:        fileName,
		FileHash:        fileHash,
		AnalysisStatus:  analysisStatusFromSQL(analysisStatusStr),
		FailureReason:   failureReason,
		UploadedAt:      uploadedAt,
		UpdatedAt:       updatedAt,
	}, nil
}

func documentKindToSQL(k filesv1.DocumentKind) string {
	switch k {
	case filesv1.DocumentKind_DOCUMENT_KIND_BILL:
		return "bill"
	case filesv1.DocumentKind_DOCUMENT_KIND_STATEMENT:
		return "statement"
	default:
		return "unspecified"
	}
}

func documentKindFromSQL(s string) filesv1.DocumentKind {
	switch s {
	case "bill":
		return filesv1.DocumentKind_DOCUMENT_KIND_BILL
	case "statement":
		return filesv1.DocumentKind_DOCUMENT_KIND_STATEMENT
	default:
		return filesv1.DocumentKind_DOCUMENT_KIND_UNSPECIFIED
	}
}

func analysisStatusToSQL(s filesv1.AnalysisStatus) string {
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

func analysisStatusFromSQL(s string) filesv1.AnalysisStatus {
	switch s {
	case "processing":
		return filesv1.AnalysisStatus_ANALYSIS_STATUS_PROCESSING
	case "analysed":
		return filesv1.AnalysisStatus_ANALYSIS_STATUS_ANALYSED
	case "analysis_failed":
		return filesv1.AnalysisStatus_ANALYSIS_STATUS_ANALYSIS_FAILED
	default:
		return filesv1.AnalysisStatus_ANALYSIS_STATUS_PENDING
	}
}

// isDuplicateConstraint detects the PostgreSQL unique-violation error code (23505).
func isDuplicateConstraint(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "23505") || strings.Contains(msg, "uq_documents_project_hash")
}
