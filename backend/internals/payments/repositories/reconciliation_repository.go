// Package repositories implements the persistence layer for the payments domain.
package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/lib/pq"

	"github.com/ralvescosta/costa-financial-assistant/backend/internals/payments/interfaces"
	apperrors "github.com/ralvescosta/costa-financial-assistant/backend/pkgs/errors"
)

// PostgresReconciliationRepository implements interfaces.ReconciliationRepository.
type PostgresReconciliationRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewReconciliationRepository constructs a PostgresReconciliationRepository.
func NewReconciliationRepository(db *sql.DB, logger *zap.Logger) interfaces.ReconciliationRepository {
	return &PostgresReconciliationRepository{db: db, logger: logger}
}

// GetUnmatchedTransactionLines returns all unmatched transaction lines for the given statement.
func (r *PostgresReconciliationRepository) GetUnmatchedTransactionLines(
	ctx context.Context,
	projectID, statementID string,
) ([]interfaces.ReconciliationSummaryEntry, error) {
	const q = `
		SELECT id, transaction_date, description, amount::text, direction, reconciliation_status
		FROM transaction_lines
		WHERE project_id = $1
		  AND statement_id = $2
		  AND reconciliation_status = 'unmatched'
		ORDER BY transaction_date, id`

	rows, err := r.db.QueryContext(ctx, q, projectID, statementID)
	if err != nil {
		r.logger.Error("reconciliation repo: get unmatched lines query failed",
			zap.String("project_id", projectID),
			zap.String("statement_id", statementID),
			zap.Error(err))
		return nil, apperrors.TranslateError(err, "repository")
	}
	defer func() { _ = rows.Close() }()

	var result []interfaces.ReconciliationSummaryEntry
	for rows.Next() {
		var e interfaces.ReconciliationSummaryEntry
		var status string
		if err := rows.Scan(&e.TransactionLineID, &e.TransactionDate, &e.Description, &e.Amount, &e.Direction, &status); err != nil {
			r.logger.Error("reconciliation repo: scan unmatched line failed",
				zap.String("project_id", projectID),
				zap.Error(err))
			return nil, apperrors.TranslateError(err, "repository")
		}
		e.ReconciliationStatus = interfaces.TransactionReconciliationStatus(status)
		result = append(result, e)
	}
	if err := rows.Err(); err != nil {
		r.logger.Error("reconciliation repo: unmatched lines rows iteration failed",
			zap.String("project_id", projectID),
			zap.Error(err))
		return nil, apperrors.TranslateError(err, "repository")
	}

	return result, nil
}

// GetBillsForPeriod returns unpaid bill records within the optional date range.
// Returns entries where TransactionLineID holds the bill_record.id (projection convention).
func (r *PostgresReconciliationRepository) GetBillsForPeriod(
	ctx context.Context,
	projectID, periodStart, periodEnd string,
) ([]interfaces.ReconciliationSummaryEntry, error) {
	args := []any{projectID}
	conditions := []string{"project_id = $1", "payment_status = 'unpaid'"}

	if periodStart != "" {
		args = append(args, periodStart)
		conditions = append(conditions, fmt.Sprintf("due_date >= $%d", len(args)))
	}
	if periodEnd != "" {
		args = append(args, periodEnd)
		conditions = append(conditions, fmt.Sprintf("due_date <= $%d", len(args)))
	}

	q := fmt.Sprintf(
		`SELECT id, due_date, amount_due::text FROM bill_records WHERE %s ORDER BY due_date, id`,
		strings.Join(conditions, " AND "),
	)

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		r.logger.Error("reconciliation repo: get bills for period query failed",
			zap.String("project_id", projectID),
			zap.Error(err))
		return nil, apperrors.TranslateError(err, "repository")
	}
	defer func() { _ = rows.Close() }()

	var result []interfaces.ReconciliationSummaryEntry
	for rows.Next() {
		var e interfaces.ReconciliationSummaryEntry
		// TransactionLineID is reused as the bill_record.id for the matching index.
		if err := rows.Scan(&e.TransactionLineID, &e.TransactionDate, &e.Amount); err != nil {
			r.logger.Error("reconciliation repo: scan bill row failed",
				zap.String("project_id", projectID),
				zap.Error(err))
			return nil, apperrors.TranslateError(err, "repository")
		}
		result = append(result, e)
	}
	if err := rows.Err(); err != nil {
		r.logger.Error("reconciliation repo: bills rows iteration failed",
			zap.String("project_id", projectID),
			zap.Error(err))
		return nil, apperrors.TranslateError(err, "repository")
	}

	return result, nil
}

// CreateLink inserts a reconciliation link and updates the transaction line status.
func (r *PostgresReconciliationRepository) CreateLink(
	ctx context.Context,
	link interfaces.ReconciliationLink,
) (*interfaces.ReconciliationLink, error) {
	const q = `
		INSERT INTO reconciliation_links (project_id, transaction_line_id, bill_record_id, link_type, linked_by)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at`

	var linkedBy sql.NullString
	if link.LinkedBy != nil {
		linkedBy = sql.NullString{String: *link.LinkedBy, Valid: true}
	}

	var createdAt time.Time
	err := r.db.QueryRowContext(ctx, q,
		link.ProjectID,
		link.TransactionLineID,
		link.BillRecordID,
		string(link.LinkType),
		linkedBy,
	).Scan(&link.ID, &createdAt)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && string(pqErr.Code) == "23505" {
			return nil, apperrors.NewCatalogError(apperrors.ErrConflict).WithError(ErrReconciliationConflict)
		}
		r.logger.Error("reconciliation repo: create link failed",
			zap.String("project_id", link.ProjectID),
			zap.String("transaction_line_id", link.TransactionLineID),
			zap.String("bill_record_id", link.BillRecordID),
			zap.Error(err))
		return nil, apperrors.TranslateError(err, "repository")
	}

	link.CreatedAt = createdAt

	// Update the transaction line status
	statusUpdate := string(interfaces.TransactionMatchedAuto)
	if link.LinkType == interfaces.ReconciliationLinkTypeManual {
		statusUpdate = string(interfaces.TransactionMatchedManual)
	}
	if updateErr := r.UpdateTransactionStatus(ctx, link.ProjectID, link.TransactionLineID,
		interfaces.TransactionReconciliationStatus(statusUpdate)); updateErr != nil {
		r.logger.Warn("reconciliation repo: update transaction status after create link failed",
			zap.String("transaction_line_id", link.TransactionLineID),
			zap.Error(updateErr))
	}

	return &link, nil
}

// UpdateTransactionStatus updates the reconciliation_status on a transaction line.
func (r *PostgresReconciliationRepository) UpdateTransactionStatus(
	ctx context.Context,
	projectID, transactionLineID string,
	status interfaces.TransactionReconciliationStatus,
) error {
	const q = `
		UPDATE transaction_lines
		SET reconciliation_status = $1
		WHERE project_id = $2 AND id = $3`

	if _, err := r.db.ExecContext(ctx, q, string(status), projectID, transactionLineID); err != nil {
		r.logger.Error("reconciliation repo: update transaction status failed",
			zap.String("project_id", projectID),
			zap.String("transaction_line_id", transactionLineID),
			zap.Error(err))
		return apperrors.TranslateError(err, "repository")
	}
	return nil
}

// GetSummary returns all transaction lines with their linked bill data for the period.
func (r *PostgresReconciliationRepository) GetSummary(
	ctx context.Context,
	projectID, periodStart, periodEnd string,
) (*interfaces.ReconciliationSummary, error) {
	args := []any{projectID}
	conditions := []string{"tl.project_id = $1"}

	if periodStart != "" {
		args = append(args, periodStart)
		conditions = append(conditions, fmt.Sprintf("tl.transaction_date >= $%d", len(args)))
	}
	if periodEnd != "" {
		args = append(args, periodEnd)
		conditions = append(conditions, fmt.Sprintf("tl.transaction_date <= $%d", len(args)))
	}

	q := fmt.Sprintf(`
		SELECT
			tl.id,
			tl.transaction_date::text,
			tl.description,
			tl.amount::text,
			tl.direction,
			tl.reconciliation_status,
			rl.bill_record_id,
			br.due_date::text,
			br.amount_due::text,
			rl.link_type
		FROM transaction_lines tl
		LEFT JOIN reconciliation_links rl ON rl.transaction_line_id = tl.id AND rl.project_id = tl.project_id
		LEFT JOIN bill_records br ON br.id = rl.bill_record_id
		WHERE %s
		ORDER BY tl.transaction_date, tl.id`,
		strings.Join(conditions, " AND "),
	)

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		r.logger.Error("reconciliation repo: get summary query failed",
			zap.String("project_id", projectID),
			zap.Error(err))
		return nil, apperrors.TranslateError(err, "repository")
	}
	defer func() { _ = rows.Close() }()

	summary := &interfaces.ReconciliationSummary{
		ProjectID:   projectID,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
	}

	for rows.Next() {
		var e interfaces.ReconciliationSummaryEntry
		var status string
		var billID, billDueDate, billAmount, linkType sql.NullString

		if err := rows.Scan(
			&e.TransactionLineID,
			&e.TransactionDate,
			&e.Description,
			&e.Amount,
			&e.Direction,
			&status,
			&billID,
			&billDueDate,
			&billAmount,
			&linkType,
		); err != nil {
			r.logger.Error("reconciliation repo: scan summary row failed",
				zap.String("project_id", projectID),
				zap.Error(err))
			return nil, apperrors.TranslateError(err, "repository")
		}

		e.ReconciliationStatus = interfaces.TransactionReconciliationStatus(status)

		if billID.Valid {
			e.LinkedBillID = &billID.String
		}
		if billDueDate.Valid {
			e.LinkedBillDueDate = &billDueDate.String
		}
		if billAmount.Valid {
			e.LinkedBillAmount = &billAmount.String
		}
		if linkType.Valid {
			lt := interfaces.ReconciliationLinkType(linkType.String)
			e.LinkType = &lt
		}

		summary.Entries = append(summary.Entries, e)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("reconciliation repo: summary rows iteration failed",
			zap.String("project_id", projectID),
			zap.Error(err))
		return nil, apperrors.TranslateError(err, "repository")
	}

	return summary, nil
}

// ErrReconciliationConflict is returned when a (transaction, bill) link already exists.
var ErrReconciliationConflict = errors.New("reconciliation link already exists")
