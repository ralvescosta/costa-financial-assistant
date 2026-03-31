package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"

	filesv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/files/v1"
)

// ErrStatementRecordNotFound is returned when a statement record does not exist for the given document.
var ErrStatementRecordNotFound = errors.New("statement record not found")

// StatementRecordRepository defines the persistence contract for extracted statement data.
type StatementRecordRepository interface {
	Create(ctx context.Context, tx *sql.Tx, record *filesv1.StatementRecord) (*filesv1.StatementRecord, error)
	FindByProjectAndDocumentID(ctx context.Context, projectID, documentID string) (*filesv1.StatementRecord, error)
}

// PostgresStatementRecordRepository implements StatementRecordRepository using PostgreSQL.
type PostgresStatementRecordRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewStatementRecordRepository constructs a PostgresStatementRecordRepository.
func NewStatementRecordRepository(db *sql.DB, logger *zap.Logger) StatementRecordRepository {
	return &PostgresStatementRecordRepository{db: db, logger: logger}
}

// Create inserts a new statement record and its transaction lines inside the provided transaction.
func (r *PostgresStatementRecordRepository) Create(ctx context.Context, tx *sql.Tx, record *filesv1.StatementRecord) (*filesv1.StatementRecord, error) {
	ctx, span := tracer.Start(ctx, "statement_record.create")
	defer span.End()

	span.SetAttributes(attribute.String("document_id", record.DocumentId))

	const stmtQuery = `
		INSERT INTO statement_records
			(project_id, document_id, bank_account_id, period_start, period_end)
		VALUES ($1, $2, NULLIF($3, '')::uuid, $4, $5)
		RETURNING id, created_at, updated_at`

	var id, createdAt, updatedAt string
	err := tx.QueryRowContext(ctx, stmtQuery,
		record.ProjectId,
		record.DocumentId,
		record.BankAccountId,
		record.PeriodStart,
		record.PeriodEnd,
	).Scan(&id, &createdAt, &updatedAt)
	if err != nil {
		span.RecordError(err)
		r.logger.Error("statement_record.create: insert statement failed",
			zap.String("document_id", record.DocumentId),
			zap.Error(err))
		return nil, fmt.Errorf("statement record repository: create: %w", err)
	}

	record.Id = id
	record.CreatedAt = createdAt
	record.UpdatedAt = updatedAt

	const lineQuery = `
		INSERT INTO transaction_lines
			(project_id, statement_id, transaction_date, description, amount, direction)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, reconciliation_status, created_at`

	for _, line := range record.Lines {
		line.ProjectId = record.ProjectId
		line.StatementId = id
		var lineID, reconcStatus, lineCreatedAt string
		err := tx.QueryRowContext(ctx, lineQuery,
			line.ProjectId,
			line.StatementId,
			line.TransactionDate,
			line.Description,
			line.Amount,
			line.Direction,
		).Scan(&lineID, &reconcStatus, &lineCreatedAt)
		if err != nil {
			span.RecordError(err)
			r.logger.Error("statement_record.create: insert line failed",
				zap.String("statement_id", id),
				zap.Error(err))
			return nil, fmt.Errorf("statement record repository: create line: %w", err)
		}
		line.Id = lineID
		line.ReconciliationStatus = reconcStatus
		line.CreatedAt = lineCreatedAt
	}

	return record, nil
}

// FindByProjectAndDocumentID returns the statement record and lines for a given project-scoped document.
func (r *PostgresStatementRecordRepository) FindByProjectAndDocumentID(ctx context.Context, projectID, documentID string) (*filesv1.StatementRecord, error) {
	ctx, span := tracer.Start(ctx, "statement_record.findByDocumentID")
	defer span.End()

	span.SetAttributes(attribute.String("document_id", documentID))

	const stmtQuery = `
		SELECT id, project_id, document_id,
		       COALESCE(bank_account_id::text,''), period_start, period_end,
		       created_at, updated_at
		FROM statement_records
		WHERE project_id = $1 AND document_id = $2`

	record := &filesv1.StatementRecord{}
	err := r.db.QueryRowContext(ctx, stmtQuery, projectID, documentID).Scan(
		&record.Id, &record.ProjectId, &record.DocumentId,
		&record.BankAccountId, &record.PeriodStart, &record.PeriodEnd,
		&record.CreatedAt, &record.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrStatementRecordNotFound
		}
		span.RecordError(err)
		r.logger.Error("statement_record.findByDocumentID: query failed",
			zap.String("document_id", documentID),
			zap.Error(err))
		return nil, fmt.Errorf("statement record repository: find by document: %w", err)
	}

	const lineQuery = `
		SELECT id, project_id, statement_id, transaction_date,
		       description, amount::text, direction, reconciliation_status, created_at
		FROM transaction_lines
		WHERE statement_id = $1
		ORDER BY transaction_date ASC, id ASC`

	rows, err := r.db.QueryContext(ctx, lineQuery, record.Id)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("statement record repository: list lines: %w", err)
	}
	defer rows.Close() //nolint:errcheck

	for rows.Next() {
		line := &filesv1.TransactionLine{}
		if err := rows.Scan(
			&line.Id, &line.ProjectId, &line.StatementId,
			&line.TransactionDate, &line.Description,
			&line.Amount, &line.Direction,
			&line.ReconciliationStatus, &line.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("statement record repository: scan line: %w", err)
		}
		record.Lines = append(record.Lines, line)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("statement record repository: rows error: %w", err)
	}
	return record, nil
}
