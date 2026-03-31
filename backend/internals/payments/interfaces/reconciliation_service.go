// Package interfaces defines the canonical service and repository contracts for the payments domain.
package interfaces

import (
	"context"
	"time"
)

// ReconciliationLinkType classifies how a transaction was linked to a bill.
type ReconciliationLinkType string

const (
	// ReconciliationLinkTypeAuto indicates the link was created automatically by the matching engine.
	ReconciliationLinkTypeAuto ReconciliationLinkType = "auto"
	// ReconciliationLinkTypeManual indicates the link was created by a user.
	ReconciliationLinkTypeManual ReconciliationLinkType = "manual"
)

// TransactionReconciliationStatus is the settlement state of a transaction line.
type TransactionReconciliationStatus string

const (
	TransactionUnmatched     TransactionReconciliationStatus = "unmatched"
	TransactionMatchedAuto   TransactionReconciliationStatus = "matched_auto"
	TransactionMatchedManual TransactionReconciliationStatus = "matched_manual"
	TransactionAmbiguous     TransactionReconciliationStatus = "ambiguous"
)

// ReconciliationLink is the materialized relationship between a transaction line and a bill record.
type ReconciliationLink struct {
	ID                string
	ProjectID         string
	TransactionLineID string
	BillRecordID      string
	LinkType          ReconciliationLinkType
	LinkedBy          *string // nil for auto links
	CreatedAt         time.Time
}

// ReconciliationSummaryEntry is a projection combining a transaction line with its linked bill (if any).
type ReconciliationSummaryEntry struct {
	TransactionLineID    string
	TransactionDate      string
	Description          string
	Amount               string
	Direction            string
	ReconciliationStatus TransactionReconciliationStatus
	LinkedBillID         *string
	LinkedBillDueDate    *string
	LinkedBillAmount     *string
	LinkType             *ReconciliationLinkType
}

// ReconciliationSummary is the project-scoped view of all reconciliation results for a period.
type ReconciliationSummary struct {
	ProjectID  string
	PeriodStart string
	PeriodEnd   string
	Entries    []ReconciliationSummaryEntry
}

// ReconciliationService defines the contract for auto and manual reconciliation operations.
// It is implemented by services.ReconciliationService.
type ReconciliationService interface {
	// AutoReconcile attempts to match all unmatched transaction lines in the given statement
	// against the project's bill records for the same period. Returns the updated summary.
	AutoReconcile(ctx context.Context, projectID, statementID string) (*ReconciliationSummary, error)

	// GetSummary returns the reconciliation summary for the project and optional date range.
	GetSummary(ctx context.Context, projectID, periodStart, periodEnd string) (*ReconciliationSummary, error)

	// CreateManualLink links a transaction line to a bill record as a user-confirmed match.
	// Returns ErrConflict if the pair is already linked.
	CreateManualLink(ctx context.Context, projectID, transactionLineID, billRecordID, linkedBy string) (*ReconciliationLink, error)
}

// ReconciliationRepository defines the persistence contract for reconciliation links and queries.
// It is implemented by repositories.PostgresReconciliationRepository.
type ReconciliationRepository interface {
	// GetUnmatchedTransactionLines returns all unmatched transaction lines for a statement.
	GetUnmatchedTransactionLines(ctx context.Context, projectID, statementID string) ([]ReconciliationSummaryEntry, error)

	// GetBillsForPeriod returns bill records within the given date range for matching.
	GetBillsForPeriod(ctx context.Context, projectID, periodStart, periodEnd string) ([]ReconciliationSummaryEntry, error)

	// CreateLink inserts a reconciliation link and updates the transaction line status.
	// Returns ErrConflict if the (transactionLineID, billRecordID) pair already exists.
	CreateLink(ctx context.Context, link ReconciliationLink) (*ReconciliationLink, error)

	// UpdateTransactionStatus updates the reconciliation_status on a transaction line.
	UpdateTransactionStatus(ctx context.Context, projectID, transactionLineID string, status TransactionReconciliationStatus) error

	// GetSummary returns all transaction lines with their linked bill data for the given period.
	GetSummary(ctx context.Context, projectID, periodStart, periodEnd string) (*ReconciliationSummary, error)
}
