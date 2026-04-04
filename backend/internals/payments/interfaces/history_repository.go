// Package interfaces defines the canonical service and repository contracts for the payments domain.
package interfaces

import "context"

// MonthlyTimelineEntry holds aggregated expenditure data for a single calendar month.
type MonthlyTimelineEntry struct {
	// Month is the first day of the calendar month in YYYY-MM-DD format.
	Month string
	// TotalAmount is the sum of all bill-record amounts for the month, formatted as a decimal string.
	TotalAmount string
	// BillCount is the number of bill records in the month.
	BillCount int
}

// CategoryBreakdownEntry holds the spend total for one bill type within a given month.
type CategoryBreakdownEntry struct {
	// Month is the first day of the relevant calendar month in YYYY-MM-DD format.
	Month string
	// BillTypeName is the human-readable label of the bill type (e.g., "Energy", "Credit Card").
	BillTypeName string
	// TotalAmount is the sum of bill amounts for this category and month, as a decimal string.
	TotalAmount string
	// BillCount is the number of bills in this category for the month.
	BillCount int
}

// MonthlyComplianceEntry holds payment-compliance metrics for a single calendar month.
type MonthlyComplianceEntry struct {
	// Month is the first day of the calendar month in YYYY-MM-DD format.
	Month string
	// TotalBills is the total number of bills in the month.
	TotalBills int
	// PaidOnTime is the count of bills paid on or before the due date.
	PaidOnTime int
	// Overdue is the count of bills paid after the due date or still unpaid past the due date.
	Overdue int
	// ComplianceRate is the percentage of bills paid on time, as a decimal string (e.g., "83.33").
	ComplianceRate string
}

// HistoryDashboard aggregates all three analytical views for the given project and period.
type HistoryDashboard struct {
	ProjectID  string
	Timeline   []MonthlyTimelineEntry
	Categories []CategoryBreakdownEntry
	Compliance []MonthlyComplianceEntry
}

// HistoryService defines the payments-owned use-case contract for history analytics.
// It is implemented by services.HistoryService.
type HistoryService interface {
	// GetTimeline returns the monthly expenditure totals for the project.
	GetTimeline(ctx context.Context, projectID string, months int) ([]MonthlyTimelineEntry, error)

	// GetCategoryBreakdown returns per-category totals for each month in the look-back window.
	GetCategoryBreakdown(ctx context.Context, projectID string, months int) ([]CategoryBreakdownEntry, error)

	// GetComplianceMetrics returns on-time vs overdue payment counts for each month.
	GetComplianceMetrics(ctx context.Context, projectID string, months int) ([]MonthlyComplianceEntry, error)
}

// HistoryRepository defines the read-only persistence contract for financial history analytics.
// It is implemented by repositories.PostgresHistoryRepository.
type HistoryRepository interface {
	// GetTimeline returns the monthly expenditure totals for the project.
	// months controls the look-back window; 0 means all available history.
	GetTimeline(ctx context.Context, projectID string, months int) ([]MonthlyTimelineEntry, error)

	// GetCategoryBreakdown returns per-category totals for each month in the look-back window.
	GetCategoryBreakdown(ctx context.Context, projectID string, months int) ([]CategoryBreakdownEntry, error)

	// GetComplianceMetrics returns on-time vs overdue payment counts for each month.
	GetComplianceMetrics(ctx context.Context, projectID string, months int) ([]MonthlyComplianceEntry, error)
}
