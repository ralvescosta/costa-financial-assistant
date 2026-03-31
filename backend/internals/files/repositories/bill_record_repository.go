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

// ErrBillRecordNotFound is returned when a bill record does not exist for the given document.
var ErrBillRecordNotFound = errors.New("bill record not found")

// BillRecordRepository defines the persistence contract for extracted bill data.
type BillRecordRepository interface {
	Create(ctx context.Context, tx *sql.Tx, record *filesv1.BillRecord) (*filesv1.BillRecord, error)
	FindByProjectAndDocumentID(ctx context.Context, projectID, documentID string) (*filesv1.BillRecord, error)
}

// PostgresBillRecordRepository implements BillRecordRepository using PostgreSQL.
type PostgresBillRecordRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewBillRecordRepository constructs a PostgresBillRecordRepository.
func NewBillRecordRepository(db *sql.DB, logger *zap.Logger) BillRecordRepository {
	return &PostgresBillRecordRepository{db: db, logger: logger}
}

// Create inserts a new bill record inside the provided transaction.
func (r *PostgresBillRecordRepository) Create(ctx context.Context, tx *sql.Tx, record *filesv1.BillRecord) (*filesv1.BillRecord, error) {
	ctx, span := tracer.Start(ctx, "bill_record.create")
	defer span.End()

	span.SetAttributes(attribute.String("document_id", record.DocumentId))

	const query = `
		INSERT INTO bill_records
			(project_id, document_id, due_date, amount_due,
			 pix_payload, pix_qr_image_ref, barcode)
		VALUES ($1, $2, $3, $4, NULLIF($5,''), NULLIF($6,''), NULLIF($7,''))
		RETURNING id, payment_status, created_at, updated_at`

	var id, paymentStatus, createdAt, updatedAt string
	err := tx.QueryRowContext(ctx, query,
		record.ProjectId,
		record.DocumentId,
		record.DueDate,
		record.AmountDue,
		record.PixPayload,
		record.PixQrImageRef,
		record.Barcode,
	).Scan(&id, &paymentStatus, &createdAt, &updatedAt)
	if err != nil {
		span.RecordError(err)
		r.logger.Error("bill_record.create: insert failed",
			zap.String("document_id", record.DocumentId),
			zap.Error(err))
		return nil, fmt.Errorf("bill record repository: create: %w", err)
	}

	record.Id = id
	record.PaymentStatus = paymentStatus
	record.CreatedAt = createdAt
	record.UpdatedAt = updatedAt
	return record, nil
}

// FindByProjectAndDocumentID returns the bill record for a given project-scoped document.
func (r *PostgresBillRecordRepository) FindByProjectAndDocumentID(ctx context.Context, projectID, documentID string) (*filesv1.BillRecord, error) {
	ctx, span := tracer.Start(ctx, "bill_record.findByDocumentID")
	defer span.End()

	span.SetAttributes(attribute.String("document_id", documentID))

	const query = `
		SELECT id, project_id, document_id,
		       due_date, amount_due::text,
		       COALESCE(pix_payload,''), COALESCE(pix_qr_image_ref,''), COALESCE(barcode,''),
		       payment_status, COALESCE(paid_at::text,''),
		       created_at, updated_at
		FROM bill_records
		WHERE project_id = $1 AND document_id = $2`

	record := &filesv1.BillRecord{}
	var dueDate, paidAt string
	err := r.db.QueryRowContext(ctx, query, projectID, documentID).Scan(
		&record.Id, &record.ProjectId, &record.DocumentId,
		&dueDate, &record.AmountDue,
		&record.PixPayload, &record.PixQrImageRef, &record.Barcode,
		&record.PaymentStatus, &paidAt,
		&record.CreatedAt, &record.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrBillRecordNotFound
		}
		span.RecordError(err)
		r.logger.Error("bill_record.findByDocumentID: query failed",
			zap.String("document_id", documentID),
			zap.Error(err))
		return nil, fmt.Errorf("bill record repository: find by document: %w", err)
	}
	record.DueDate = dueDate
	record.PaidAt = paidAt
	return record, nil
}
