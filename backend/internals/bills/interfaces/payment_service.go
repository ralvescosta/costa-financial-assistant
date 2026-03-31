// Package interfaces defines the canonical service and repository contracts for the bills domain.
// These interfaces consolidate the key contracts used by the bills gRPC server and are used as mock targets in tests.
package interfaces

import (
	"context"

	billsv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/bills/v1"
)

// BillPaymentService defines the contract for bill payment dashboard and mark-as-paid operations.
// It is implemented by services.BillPaymentService and consumed by the bills gRPC server.
type BillPaymentService interface {
	// GetPaymentDashboard returns outstanding and overdue bills for the given cycle date range.
	GetPaymentDashboard(ctx context.Context, projectID, cycleStart, cycleEnd string, pageSize int32, pageToken string) ([]*billsv1.PaymentDashboardEntry, string, error)

	// MarkBillPaid idempotently marks a bill as paid.
	// Returns the updated BillRecord whether it is newly marked or was already paid.
	MarkBillPaid(ctx context.Context, projectID, billID, markedBy string) (*billsv1.BillRecord, error)

	// GetBill returns a single bill record by ID, scoped to the project.
	GetBill(ctx context.Context, projectID, billID string) (*billsv1.BillRecord, error)

	// ListBills returns project-scoped bill records filtered by optional payment status.
	// Pass billsv1.PaymentStatus_PAYMENT_STATUS_UNSPECIFIED to retrieve all statuses.
	ListBills(ctx context.Context, projectID string, status billsv1.PaymentStatus, pageSize int32, pageToken string) ([]*billsv1.BillRecord, string, error)
}

// BillPaymentRepository defines the persistence contract for bill payment operations.
// It is implemented by repositories.PostgresBillPaymentRepository.
type BillPaymentRepository interface {
	// GetBill fetches a single bill record by ID within the project.
	GetBill(ctx context.Context, projectID, billID string) (*billsv1.BillRecord, error)

	// ListBills returns bill records for the project with optional status filter and cursor pagination.
	ListBills(ctx context.Context, projectID string, status billsv1.PaymentStatus, pageSize int32, pageToken string) ([]*billsv1.BillRecord, string, error)

	// GetDashboardEntries returns bills for the project in the given cycle date range with overdue flags.
	GetDashboardEntries(ctx context.Context, projectID, cycleStart, cycleEnd string, pageSize int32, pageToken string) ([]*billsv1.PaymentDashboardEntry, string, error)

	// MarkPaid updates the bill record as paid and returns the updated record.
	MarkPaid(ctx context.Context, projectID, billID, markedBy string) (*billsv1.BillRecord, error)

	// FindIdempotencyKey looks up an idempotency key and returns its stored payload if present.
	FindIdempotencyKey(ctx context.Context, key string) (string, error)

	// StoreIdempotencyKey persists an idempotency key with its payload and service label.
	StoreIdempotencyKey(ctx context.Context, key, service, payload string) error
}
