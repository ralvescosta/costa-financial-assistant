// Package repositories implements the persistence layer for the payments domain.
package repositories

import (
	"context"
	"database/sql"

	"go.uber.org/zap"

	"github.com/ralvescosta/costa-financial-assistant/backend/internals/payments/interfaces"
	apperrors "github.com/ralvescosta/costa-financial-assistant/backend/pkgs/errors"
)

// HistoryTimelineEntry is a value type returned by GetTimeline for use in integration tests.
type HistoryTimelineEntry = interfaces.MonthlyTimelineEntry

// PostgresHistoryRepository implements interfaces.HistoryRepository using direct SQL aggregations.
type PostgresHistoryRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewHistoryRepository constructs a PostgresHistoryRepository.
func NewHistoryRepository(db *sql.DB, logger *zap.Logger) interfaces.HistoryRepository {
	return &PostgresHistoryRepository{db: db, logger: logger}
}

// GetTimeline returns one row per calendar month containing the total bill amounts.
// When months > 0 only the last N calendar months (relative to NOW()) are returned.
func (r *PostgresHistoryRepository) GetTimeline(
	ctx context.Context,
	projectID string,
	months int,
) ([]interfaces.MonthlyTimelineEntry, error) {
	const baseQuery = `
		SELECT
			date_trunc('month', due_date)::date AS month,
			SUM(amount_due)::numeric(14,2)       AS total_amount,
			COUNT(*)                              AS bill_count
		FROM bill_records
		WHERE project_id = $1
		  AND ($2 = 0 OR due_date >= date_trunc('month', NOW() - ($2 || ' months')::interval))
		GROUP BY 1
		ORDER BY 1`

	rows, err := r.db.QueryContext(ctx, baseQuery, projectID, months)
	if err != nil {
		r.logger.Error("history repo: get timeline query failed",
			zap.String("project_id", projectID),
			zap.Int("months", months),
			zap.Error(err))
		return nil, apperrors.TranslateError(err, "repository")
	}
	defer func() { _ = rows.Close() }()

	var result []interfaces.MonthlyTimelineEntry
	for rows.Next() {
		var e interfaces.MonthlyTimelineEntry
		if err := rows.Scan(&e.Month, &e.TotalAmount, &e.BillCount); err != nil {
			r.logger.Error("history repo: scan timeline failed",
				zap.String("project_id", projectID),
				zap.Error(err))
			return nil, apperrors.TranslateError(err, "repository")
		}
		result = append(result, e)
	}
	if err := rows.Err(); err != nil {
		r.logger.Error("history repo: timeline rows iteration failed",
			zap.String("project_id", projectID),
			zap.Error(err))
		return nil, apperrors.TranslateError(err, "repository")
	}
	return result, nil
}

// GetCategoryBreakdown returns per-bill-type totals grouped by month.
// Entries without a bill_type_id are labelled "Uncategorised".
func (r *PostgresHistoryRepository) GetCategoryBreakdown(
	ctx context.Context,
	projectID string,
	months int,
) ([]interfaces.CategoryBreakdownEntry, error) {
	const query = `
		SELECT
			date_trunc('month', br.due_date)::date               AS month,
			COALESCE(bt.name, 'Uncategorised')                   AS bill_type_name,
			SUM(br.amount_due)::numeric(14,2)                    AS total_amount,
			COUNT(*)                                              AS bill_count
		FROM bill_records br
		LEFT JOIN bill_types bt ON bt.id = br.bill_type_id
		WHERE br.project_id = $1
		  AND ($2 = 0 OR br.due_date >= date_trunc('month', NOW() - ($2 || ' months')::interval))
		GROUP BY 1, 2
		ORDER BY 1, 2`

	rows, err := r.db.QueryContext(ctx, query, projectID, months)
	if err != nil {
		r.logger.Error("history repo: get category breakdown query failed",
			zap.String("project_id", projectID),
			zap.Int("months", months),
			zap.Error(err))
		return nil, apperrors.TranslateError(err, "repository")
	}
	defer func() { _ = rows.Close() }()

	var result []interfaces.CategoryBreakdownEntry
	for rows.Next() {
		var e interfaces.CategoryBreakdownEntry
		if err := rows.Scan(&e.Month, &e.BillTypeName, &e.TotalAmount, &e.BillCount); err != nil {
			r.logger.Error("history repo: scan category breakdown failed",
				zap.String("project_id", projectID),
				zap.Error(err))
			return nil, apperrors.TranslateError(err, "repository")
		}
		result = append(result, e)
	}
	if err := rows.Err(); err != nil {
		r.logger.Error("history repo: category breakdown rows iteration failed",
			zap.String("project_id", projectID),
			zap.Error(err))
		return nil, apperrors.TranslateError(err, "repository")
	}
	return result, nil
}

// GetComplianceMetrics returns on-time vs overdue payment counts per calendar month.
// A bill is considered on-time when paid_at <= due_date. Unpaid bills whose due_date < NOW() are overdue.
func (r *PostgresHistoryRepository) GetComplianceMetrics(
	ctx context.Context,
	projectID string,
	months int,
) ([]interfaces.MonthlyComplianceEntry, error) {
	const query = `
		SELECT
			date_trunc('month', due_date)::date                         AS month,
			COUNT(*)                                                     AS total_bills,
			COUNT(*) FILTER (WHERE payment_status = 'paid'
			                   AND paid_at::date <= due_date)           AS paid_on_time,
			COUNT(*) FILTER (WHERE (payment_status = 'unpaid' AND due_date < NOW()::date)
			                    OR (payment_status = 'paid'   AND paid_at::date > due_date)) AS overdue,
			ROUND(
				COUNT(*) FILTER (WHERE payment_status = 'paid' AND paid_at::date <= due_date)::numeric
				* 100.0
				/ NULLIF(COUNT(*), 0),
			2)::text                                                     AS compliance_rate
		FROM bill_records
		WHERE project_id = $1
		  AND ($2 = 0 OR due_date >= date_trunc('month', NOW() - ($2 || ' months')::interval))
		GROUP BY 1
		ORDER BY 1`

	rows, err := r.db.QueryContext(ctx, query, projectID, months)
	if err != nil {
		r.logger.Error("history repo: get compliance query failed",
			zap.String("project_id", projectID),
			zap.Int("months", months),
			zap.Error(err))
		return nil, apperrors.TranslateError(err, "repository")
	}
	defer func() { _ = rows.Close() }()

	var result []interfaces.MonthlyComplianceEntry
	for rows.Next() {
		var e interfaces.MonthlyComplianceEntry
		var complianceRate sql.NullString
		if err := rows.Scan(&e.Month, &e.TotalBills, &e.PaidOnTime, &e.Overdue, &complianceRate); err != nil {
			r.logger.Error("history repo: scan compliance failed",
				zap.String("project_id", projectID),
				zap.Error(err))
			return nil, apperrors.TranslateError(err, "repository")
		}
		if complianceRate.Valid {
			e.ComplianceRate = complianceRate.String
		} else {
			e.ComplianceRate = "0.00"
		}
		result = append(result, e)
	}
	if err := rows.Err(); err != nil {
		r.logger.Error("history repo: compliance rows iteration failed",
			zap.String("project_id", projectID),
			zap.Error(err))
		return nil, apperrors.TranslateError(err, "repository")
	}
	return result, nil
}
