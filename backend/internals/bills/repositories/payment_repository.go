// Package repositories implements the persistence layer for the bills domain.
package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/ralvescosta/costa-financial-assistant/backend/internals/bills/interfaces"
	billsv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/bills/v1"
)

// ErrBillNotFound is returned when a bill record is not found.
var ErrBillNotFound = errors.New("bill record not found")

// PostgresBillPaymentRepository implements interfaces.BillPaymentRepository.
type PostgresBillPaymentRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewBillPaymentRepository constructs a PostgresBillPaymentRepository.
func NewBillPaymentRepository(db *sql.DB, logger *zap.Logger) interfaces.BillPaymentRepository {
	return &PostgresBillPaymentRepository{db: db, logger: logger}
}

// GetBill fetches a single bill record by ID within the project.
func (r *PostgresBillPaymentRepository) GetBill(ctx context.Context, projectID, billID string) (*billsv1.BillRecord, error) {
	const q = `
		SELECT br.id, br.project_id, br.document_id, br.bill_type_id,
		       br.due_date, br.amount_due, br.pix_payload, br.pix_qr_image_ref,
		       br.barcode, br.payment_status::text, br.paid_at, br.marked_paid_by,
		       br.created_at, br.updated_at
		FROM bill_records br
		WHERE br.id = $1 AND br.project_id = $2`

	return r.scanBillRecord(ctx, r.db.QueryRowContext(ctx, q, billID, projectID))
}

// ListBills returns bill records for the project with optional status filter and cursor pagination.
func (r *PostgresBillPaymentRepository) ListBills(
	ctx context.Context,
	projectID string,
	status billsv1.PaymentStatus,
	pageSize int32,
	pageToken string,
) ([]*billsv1.BillRecord, string, error) {
	if pageSize <= 0 {
		pageSize = 20
	}

	var (
		rows *sql.Rows
		err  error
	)

	if status == billsv1.PaymentStatus_PAYMENT_STATUS_UNSPECIFIED {
		const q = `
			SELECT br.id, br.project_id, br.document_id, br.bill_type_id,
			       br.due_date, br.amount_due, br.pix_payload, br.pix_qr_image_ref,
			       br.barcode, br.payment_status::text, br.paid_at, br.marked_paid_by,
			       br.created_at, br.updated_at
			FROM bill_records br
			WHERE br.project_id = $1
			  AND ($2 = '' OR br.id > $2::uuid)
			ORDER BY br.id
			LIMIT $3`
		rows, err = r.db.QueryContext(ctx, q, projectID, pageToken, pageSize+1)
	} else {
		statusStr := paymentStatusToString(status)
		const q = `
			SELECT br.id, br.project_id, br.document_id, br.bill_type_id,
			       br.due_date, br.amount_due, br.pix_payload, br.pix_qr_image_ref,
			       br.barcode, br.payment_status::text, br.paid_at, br.marked_paid_by,
			       br.created_at, br.updated_at
			FROM bill_records br
			WHERE br.project_id = $1 AND br.payment_status = $2::payment_status
			  AND ($3 = '' OR br.id > $3::uuid)
			ORDER BY br.id
			LIMIT $4`
		rows, err = r.db.QueryContext(ctx, q, projectID, statusStr, pageToken, pageSize+1)
	}
	if err != nil {
		return nil, "", fmt.Errorf("bill payment repo: list bills: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var bills []*billsv1.BillRecord
	for rows.Next() {
		bill, scanErr := r.scanRowBillRecord(rows)
		if scanErr != nil {
			return nil, "", fmt.Errorf("bill payment repo: list bills scan: %w", scanErr)
		}
		bills = append(bills, bill)
	}
	if err = rows.Err(); err != nil {
		return nil, "", fmt.Errorf("bill payment repo: list bills rows: %w", err)
	}

	var nextToken string
	if int32(len(bills)) > pageSize {
		nextToken = bills[pageSize].GetId()
		bills = bills[:pageSize]
	}

	return bills, nextToken, nil
}

// GetDashboardEntries returns bills for the project in the given cycle date range with overdue flags.
func (r *PostgresBillPaymentRepository) GetDashboardEntries(
	ctx context.Context,
	projectID, cycleStart, cycleEnd string,
	pageSize int32,
	pageToken string,
) ([]*billsv1.PaymentDashboardEntry, string, error) {
	if pageSize <= 0 {
		pageSize = 20
	}

	const q = `
		SELECT br.id, br.project_id, br.document_id, br.bill_type_id,
		       br.due_date, br.amount_due, br.pix_payload, br.pix_qr_image_ref,
		       br.barcode, br.payment_status::text, br.paid_at, br.marked_paid_by,
		       br.created_at, br.updated_at,
		       bt.id   AS bt_id,   bt.project_id AS bt_project_id,
		       bt.name AS bt_name, bt.created_at AS bt_created_at,
		       (br.payment_status = 'unpaid' AND br.due_date < CURRENT_DATE) AS is_overdue,
		       (br.due_date - CURRENT_DATE)::int                             AS days_until_due
		FROM bill_records br
		LEFT JOIN bill_types bt ON bt.id = br.bill_type_id
		WHERE br.project_id    = $1
		  AND br.payment_status != 'paid'
		  AND br.due_date BETWEEN $2::date AND $3::date
		  AND ($4 = '' OR br.id > $4::uuid)
		ORDER BY br.due_date, br.id
		LIMIT $5`

	rows, err := r.db.QueryContext(ctx, q, projectID, cycleStart, cycleEnd, pageToken, pageSize+1)
	if err != nil {
		return nil, "", fmt.Errorf("bill payment repo: dashboard entries: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var entries []*billsv1.PaymentDashboardEntry

	for rows.Next() {
		bill := &billsv1.BillRecord{}
		var (
			dueDate     string
			amountDue   float64
			billTypeID  sql.NullString
			pixPayload  sql.NullString
			pixQRRef    sql.NullString
			barcode     sql.NullString
			paidAt      sql.NullTime
			markedBy    sql.NullString
			createdAt   time.Time
			updatedAt   time.Time
			statusStr   string
			btID        sql.NullString
			btProjectID sql.NullString
			btName      sql.NullString
			btCreatedAt sql.NullTime
			isOverdue   bool
			daysUntil   int32
		)

		if scanErr := rows.Scan(
			&bill.Id, &bill.ProjectId, &bill.DocumentId, &billTypeID,
			&dueDate, &amountDue, &pixPayload, &pixQRRef,
			&barcode, &statusStr, &paidAt, &markedBy,
			&createdAt, &updatedAt,
			&btID, &btProjectID, &btName, &btCreatedAt,
			&isOverdue, &daysUntil,
		); scanErr != nil {
			return nil, "", fmt.Errorf("bill payment repo: dashboard scan: %w", scanErr)
		}

		bill.DueDate = dueDate
		bill.AmountDue = fmt.Sprintf("%.2f", amountDue)
		if billTypeID.Valid {
			bill.BillTypeId = billTypeID.String
		}
		if pixPayload.Valid {
			bill.PixPayload = pixPayload.String
		}
		if pixQRRef.Valid {
			bill.PixQrImageRef = pixQRRef.String
		}
		if barcode.Valid {
			bill.Barcode = barcode.String
		}
		bill.PaymentStatus = stringToPaymentStatus(statusStr)
		if paidAt.Valid {
			bill.PaidAt = paidAt.Time.Format(time.RFC3339)
		}
		if markedBy.Valid {
			bill.MarkedPaidBy = markedBy.String
		}
		bill.CreatedAt = createdAt.Format(time.RFC3339)
		bill.UpdatedAt = updatedAt.Format(time.RFC3339)

		entry := &billsv1.PaymentDashboardEntry{
			Bill:         bill,
			IsOverdue:    isOverdue,
			DaysUntilDue: daysUntil,
		}

		if btID.Valid {
			entry.BillType = &billsv1.BillType{
				Id:        btID.String,
				ProjectId: btProjectID.String,
				Name:      btName.String,
			}
			if btCreatedAt.Valid {
				entry.BillType.CreatedAt = btCreatedAt.Time.Format(time.RFC3339)
			}
		}

		entries = append(entries, entry)
	}
	if err = rows.Err(); err != nil {
		return nil, "", fmt.Errorf("bill payment repo: dashboard rows: %w", err)
	}

	var nextToken string
	if int32(len(entries)) > pageSize {
		nextToken = entries[pageSize].GetBill().GetId()
		entries = entries[:pageSize]
	}

	return entries, nextToken, nil
}

// MarkPaid updates the bill record as paid and returns the updated record.
func (r *PostgresBillPaymentRepository) MarkPaid(ctx context.Context, projectID, billID, markedBy string) (*billsv1.BillRecord, error) {
	const q = `
		UPDATE bill_records
		SET payment_status  = 'paid',
		    paid_at         = NOW(),
		    marked_paid_by  = $3::uuid,
		    updated_at      = NOW()
		WHERE id = $1 AND project_id = $2
		RETURNING id, project_id, document_id, bill_type_id,
		          due_date, amount_due, pix_payload, pix_qr_image_ref,
		          barcode, payment_status::text, paid_at, marked_paid_by,
		          created_at, updated_at`

	return r.scanBillRecord(ctx, r.db.QueryRowContext(ctx, q, billID, projectID, markedBy))
}

// FindIdempotencyKey looks up an idempotency key and returns its stored payload if present.
// Returns empty string if not found. Uses a compound key scoped to the bills service.
func (r *PostgresBillPaymentRepository) FindIdempotencyKey(ctx context.Context, key string) (string, error) {
	const q = `
		SELECT response_hash
		FROM idempotency_keys
		WHERE idempotency_key = $1 AND operation = 'bills'
		  AND expires_at > NOW()`

	var payload sql.NullString
	err := r.db.QueryRowContext(ctx, q, key).Scan(&payload)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("bill payment repo: find idempotency key: %w", err)
	}
	return payload.String, nil
}

// StoreIdempotencyKey persists an idempotency key. The key expires after 24 hours.
func (r *PostgresBillPaymentRepository) StoreIdempotencyKey(ctx context.Context, key, service, payload string) error {
	const q = `
		INSERT INTO idempotency_keys (project_id, operation, idempotency_key, response_hash, expires_at)
		VALUES ('00000000-0000-0000-0000-000000000000', $1, $2, $3, NOW() + INTERVAL '24 hours')
		ON CONFLICT (project_id, operation, idempotency_key) DO NOTHING`

	_, err := r.db.ExecContext(ctx, q, service, key, payload)
	if err != nil {
		return fmt.Errorf("bill payment repo: store idempotency key: %w", err)
	}
	return nil
}

// scanBillRecord scans a single BillRecord from a QueryRow.
func (r *PostgresBillPaymentRepository) scanBillRecord(_ context.Context, row *sql.Row) (*billsv1.BillRecord, error) {
	var (
		bill       billsv1.BillRecord
		dueDate    string
		amountDue  float64
		billTypeID sql.NullString
		pixPayload sql.NullString
		pixQRRef   sql.NullString
		barcode    sql.NullString
		paidAt     sql.NullTime
		markedBy   sql.NullString
		createdAt  time.Time
		updatedAt  time.Time
		statusStr  string
	)

	err := row.Scan(
		&bill.Id, &bill.ProjectId, &bill.DocumentId, &billTypeID,
		&dueDate, &amountDue, &pixPayload, &pixQRRef,
		&barcode, &statusStr, &paidAt, &markedBy,
		&createdAt, &updatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrBillNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("bill payment repo: scan: %w", err)
	}

	bill.DueDate = dueDate
	bill.AmountDue = fmt.Sprintf("%.2f", amountDue)
	if billTypeID.Valid {
		bill.BillTypeId = billTypeID.String
	}
	if pixPayload.Valid {
		bill.PixPayload = pixPayload.String
	}
	if pixQRRef.Valid {
		bill.PixQrImageRef = pixQRRef.String
	}
	if barcode.Valid {
		bill.Barcode = barcode.String
	}
	bill.PaymentStatus = stringToPaymentStatus(statusStr)
	if paidAt.Valid {
		bill.PaidAt = paidAt.Time.Format(time.RFC3339)
	}
	if markedBy.Valid {
		bill.MarkedPaidBy = markedBy.String
	}
	bill.CreatedAt = createdAt.Format(time.RFC3339)
	bill.UpdatedAt = updatedAt.Format(time.RFC3339)

	return &bill, nil
}

// scanRowBillRecord scans a single BillRecord from sql.Rows.
func (r *PostgresBillPaymentRepository) scanRowBillRecord(rows *sql.Rows) (*billsv1.BillRecord, error) {
	var (
		bill       billsv1.BillRecord
		dueDate    string
		amountDue  float64
		billTypeID sql.NullString
		pixPayload sql.NullString
		pixQRRef   sql.NullString
		barcode    sql.NullString
		paidAt     sql.NullTime
		markedBy   sql.NullString
		createdAt  time.Time
		updatedAt  time.Time
		statusStr  string
	)

	if err := rows.Scan(
		&bill.Id, &bill.ProjectId, &bill.DocumentId, &billTypeID,
		&dueDate, &amountDue, &pixPayload, &pixQRRef,
		&barcode, &statusStr, &paidAt, &markedBy,
		&createdAt, &updatedAt,
	); err != nil {
		return nil, fmt.Errorf("bill payment repo: scan row: %w", err)
	}

	bill.DueDate = dueDate
	bill.AmountDue = fmt.Sprintf("%.2f", amountDue)
	if billTypeID.Valid {
		bill.BillTypeId = billTypeID.String
	}
	if pixPayload.Valid {
		bill.PixPayload = pixPayload.String
	}
	if pixQRRef.Valid {
		bill.PixQrImageRef = pixQRRef.String
	}
	if barcode.Valid {
		bill.Barcode = barcode.String
	}
	bill.PaymentStatus = stringToPaymentStatus(statusStr)
	if paidAt.Valid {
		bill.PaidAt = paidAt.Time.Format(time.RFC3339)
	}
	if markedBy.Valid {
		bill.MarkedPaidBy = markedBy.String
	}
	bill.CreatedAt = createdAt.Format(time.RFC3339)
	bill.UpdatedAt = updatedAt.Format(time.RFC3339)

	return &bill, nil
}

// paymentStatusToString converts a PaymentStatus enum to its SQL text value.
func paymentStatusToString(s billsv1.PaymentStatus) string {
	switch s {
	case billsv1.PaymentStatus_PAYMENT_STATUS_PAID:
		return "paid"
	case billsv1.PaymentStatus_PAYMENT_STATUS_OVERDUE:
		return "overdue"
	case billsv1.PaymentStatus_PAYMENT_STATUS_UNPAID:
		return "unpaid"
	default:
		return "unpaid"
	}
}

// stringToPaymentStatus converts a SQL text payment_status value to the proto enum.
func stringToPaymentStatus(s string) billsv1.PaymentStatus {
	switch s {
	case "paid":
		return billsv1.PaymentStatus_PAYMENT_STATUS_PAID
	case "overdue":
		return billsv1.PaymentStatus_PAYMENT_STATUS_OVERDUE
	case "unpaid":
		return billsv1.PaymentStatus_PAYMENT_STATUS_UNPAID
	default:
		return billsv1.PaymentStatus_PAYMENT_STATUS_UNSPECIFIED
	}
}
