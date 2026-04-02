package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"

	"github.com/ralvescosta/costa-financial-assistant/backend/internals/files/interfaces"
	filesv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/files/v1"
)

// ErrAnalysisJobNotFound is returned when a queried analysis job does not exist.
var ErrAnalysisJobNotFound = errors.New("analysis job not found")

// PostgresAnalysisJobRepository implements AnalysisJobRepository using PostgreSQL.
type PostgresAnalysisJobRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewAnalysisJobRepository constructs a PostgresAnalysisJobRepository.
func NewAnalysisJobRepository(db *sql.DB, logger *zap.Logger) interfaces.AnalysisJobRepository {
	return &PostgresAnalysisJobRepository{db: db, logger: logger}
}

// Create inserts a new analysis job inside the provided transaction.
func (r *PostgresAnalysisJobRepository) Create(ctx context.Context, tx *sql.Tx, job *filesv1.AnalysisJob) (*filesv1.AnalysisJob, error) {
	ctx, span := tracer.Start(ctx, "analysis_job.create")
	defer span.End()

	span.SetAttributes(attribute.String("document_id", job.DocumentId))

	const query = `
		INSERT INTO analysis_jobs (project_id, document_id, job_type, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id, attempt_count, created_at, updated_at`

	var id, createdAt, updatedAt string
	var attemptCount int32
	err := tx.QueryRowContext(ctx, query,
		job.ProjectId,
		job.DocumentId,
		job.JobType,
		job.Status,
	).Scan(&id, &attemptCount, &createdAt, &updatedAt)
	if err != nil {
		span.RecordError(err)
		r.logger.Error("analysis_job.create: insert failed",
			zap.String("document_id", job.DocumentId),
			zap.Error(err))
		return nil, fmt.Errorf("analysis job repository: create: %w", err)
	}

	job.Id = id
	job.AttemptCount = attemptCount
	job.CreatedAt = createdAt
	job.UpdatedAt = updatedAt
	return job, nil
}

// FindByDocumentID returns the latest analysis job for the given document in the project.
func (r *PostgresAnalysisJobRepository) FindByDocumentID(ctx context.Context, projectID, documentID string) (*filesv1.AnalysisJob, error) {
	ctx, span := tracer.Start(ctx, "analysis_job.findByDocumentID")
	defer span.End()

	span.SetAttributes(attribute.String("document_id", documentID))

	const query = `
		SELECT id, project_id, document_id, job_type, status, attempt_count,
		       COALESCE(last_error, ''), created_at, updated_at
		FROM analysis_jobs
		WHERE project_id = $1 AND document_id = $2
		ORDER BY created_at DESC
		LIMIT 1`

	row := r.db.QueryRowContext(ctx, query, projectID, documentID)
	job := &filesv1.AnalysisJob{}
	err := row.Scan(
		&job.Id, &job.ProjectId, &job.DocumentId, &job.JobType,
		&job.Status, &job.AttemptCount, &job.LastError,
		&job.CreatedAt, &job.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrAnalysisJobNotFound
		}
		span.RecordError(err)
		r.logger.Error("analysis_job.findByDocumentID: query failed",
			zap.String("document_id", documentID),
			zap.Error(err))
		return nil, fmt.Errorf("analysis job repository: find by document: %w", err)
	}
	return job, nil
}

// UpdateStatus transitions an analysis job to a new status and increments attempt count.
func (r *PostgresAnalysisJobRepository) UpdateStatus(ctx context.Context, tx *sql.Tx, jobID, status, lastError string, attemptCount int32) error {
	ctx, span := tracer.Start(ctx, "analysis_job.updateStatus")
	defer span.End()

	const query = `
		UPDATE analysis_jobs
		SET status = $1, last_error = NULLIF($2, ''), attempt_count = $3, updated_at = NOW()
		WHERE id = $4`

	_, err := tx.ExecContext(ctx, query, status, lastError, attemptCount, jobID)
	if err != nil {
		span.RecordError(err)
		r.logger.Error("analysis_job.updateStatus: update failed",
			zap.String("job_id", jobID),
			zap.Error(err))
		return fmt.Errorf("analysis job repository: update status: %w", err)
	}
	return nil
}

// UpdateDocumentAnalysisStatus transitions a document's analysis_status column.
func (r *PostgresAnalysisJobRepository) UpdateDocumentAnalysisStatus(ctx context.Context, tx *sql.Tx, projectID, documentID, analysisStatus, failureReason string) error {
	ctx, span := tracer.Start(ctx, "analysis_job.updateDocumentStatus")
	defer span.End()

	const query = `
		UPDATE documents
		SET analysis_status = $1, failure_reason = NULLIF($2, ''), updated_at = NOW()
		WHERE project_id = $3 AND id = $4`

	_, err := tx.ExecContext(ctx, query, analysisStatus, failureReason, projectID, documentID)
	if err != nil {
		span.RecordError(err)
		r.logger.Error("analysis_job.updateDocumentStatus: update failed",
			zap.String("document_id", documentID),
			zap.Error(err))
		return fmt.Errorf("analysis job repository: update document status: %w", err)
	}
	return nil
}
