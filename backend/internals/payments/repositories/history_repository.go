// Package repositories implements the persistence layer for the payments domain.
package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"go.uber.org/zap"

	"github.com/ralvescosta/costa-financial-assistant/backend/internals/payments/interfaces"
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
		return nil, fmt.Errorf("history repo: get timeline: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var result []interfaces.MonthlyTimelineEntry
	for rows.Next() {
		var e interfaces.MonthlyTimelineEntry
		if err := rows.Scan(&e.Month, &e.TotalAmount, &e.BillCount); err != nil {
			return nil, fmt.Errorf("history repo: scan timeline: %w", err)
		}
		result = append(result, e)
	}
	return result, rows.Err()
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
		return nil, fmt.Errorf("history repo: get category breakdown: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var result []interfaces.CategoryBreakdownEntry
	for rows.Next() {
		var e interfaces.CategoryBreakdownEntry
		if err := rows.Scan(&e.Month, &e.BillTypeName, &e.TotalAmount, &e.BillCount); err != nil {
			return nil, fmt.Errorf("history repo: scan category breakdown: %w", err)
		}
		result = append(result, e)
	}
	return result, rows.Err()
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
		return nil, fmt.Errorf("history repo: get compliance: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var result []interfaces.MonthlyComplianceEntry
	for rows.Next() {
		var e interfaces.MonthlyComplianceEntry
		var complianceRate sql.NullString
		if err := rows.Scan(&e.Month, &e.TotalBills, &e.PaidOnTime, &e.Overdue, &complianceRate); err != nil {
			return nil, fmt.Errorf("history repo: scan compliance: %w", err)
		}
		if complianceRate.Valid {
			e.ComplianceRate = complianceRate.String
		} else {
			e.ComplianceRate = "0.00"
		}
		result = append(result, e)
	}
	return result, rows.Err()
}
