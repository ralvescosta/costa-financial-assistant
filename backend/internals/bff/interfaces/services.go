// Package interfaces defines the BFF-facing service contracts that decouple controllers
// from downstream gRPC clients and repository-backed services.
// Controllers depend only on these narrow interfaces; concrete implementations live in
// backend/internals/bff/services/.
package interfaces

import (
	"context"

	bffcontracts "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/services/contracts"
)

// AuthService defines the transport-agnostic authentication operations consumed by the BFF.
type AuthService interface {
	// Login validates the seeded owner credentials and returns the authenticated session payload.
	Login(ctx context.Context, username, password string) (*bffcontracts.AuthSessionResponse, error)

	// Refresh validates the current session token and returns a refreshed session payload.
	Refresh(ctx context.Context, token string) (*bffcontracts.RefreshSessionResponse, error)
}

// DocumentsService defines the transport-agnostic document operations consumed by the BFF.
// Pointer policy: struct payloads crossing service boundaries use pointer semantics.
type DocumentsService interface {
	// UploadDocument registers a document with the downstream files service.
	// fileBytes must be non-empty; the service computes the SHA-256 hash internally.
	UploadDocument(ctx context.Context, projectID, uploadedBy, fileName string, fileBytes []byte) (*bffcontracts.DocumentResponse, error)

	// ClassifyDocument updates the kind (bill or statement) of an existing document.
	ClassifyDocument(ctx context.Context, projectID, documentID, kind string) (*bffcontracts.DocumentResponse, error)

	// ListDocuments returns a project-scoped page of documents.
	ListDocuments(ctx context.Context, projectID string, pageSize int32, pageToken string) (*bffcontracts.ListDocumentsResponse, error)

	// GetDocument returns full document metadata including extraction fields.
	GetDocument(ctx context.Context, projectID, documentID string) (*bffcontracts.DocumentDetailResponse, error)
}

// ProjectsService defines the transport-agnostic project collaboration operations consumed by the BFF.
type ProjectsService interface {
	// GetCurrentProject returns the project identified by the caller's JWT claim.
	GetCurrentProject(ctx context.Context, projectID, userID, role string) (*bffcontracts.ProjectResponse, error)

	// ListMembers returns all members for the caller's project.
	ListMembers(ctx context.Context, projectID, userID, role string, pageSize int32, pageToken string) (*bffcontracts.ListMembersResponse, error)

	// InviteMember sends an invitation to the given email with the specified role.
	InviteMember(ctx context.Context, projectID, inviterID, inviterRole, email, role string) (*bffcontracts.ProjectMemberResponse, error)

	// UpdateMemberRole changes the role of an existing project member.
	UpdateMemberRole(ctx context.Context, projectID, callerID, callerRole, memberID, newRole string) (*bffcontracts.ProjectMemberResponse, error)
}

// SettingsService defines the transport-agnostic bank account operations consumed by the BFF.
type SettingsService interface {
	// ListBankAccounts returns all bank accounts for the project.
	ListBankAccounts(ctx context.Context, projectID string) (*bffcontracts.ListBankAccountsResponse, error)

	// CreateBankAccount registers a new bank account label for the project.
	CreateBankAccount(ctx context.Context, projectID, createdBy, label string) (*bffcontracts.BankAccountResponse, error)

	// DeleteBankAccount removes a bank account from the project.
	DeleteBankAccount(ctx context.Context, projectID, bankAccountID string) error
}

// PaymentsService defines the transport-agnostic payment operations consumed by the BFF.
type PaymentsService interface {
	// GetPaymentDashboard returns outstanding bills for the project's payment cycle.
	GetPaymentDashboard(ctx context.Context, projectID, userID, cycleStart, cycleEnd string, pageSize int32, pageToken string) (*bffcontracts.PaymentDashboardResponse, error)

	// MarkBillPaid idempotently marks a bill as paid.
	MarkBillPaid(ctx context.Context, projectID, billID, paidBy string) (*bffcontracts.MarkBillPaidResponse, error)

	// GetCyclePreference returns the project's preferred payment day.
	GetCyclePreference(ctx context.Context, projectID string) (*bffcontracts.CyclePreferenceResponse, error)

	// SetCyclePreference creates or updates the project's preferred payment day.
	SetCyclePreference(ctx context.Context, projectID string, dayOfMonth int, updatedBy string) (*bffcontracts.CyclePreferenceResponse, error)
}

// ReconciliationService defines the transport-agnostic reconciliation operations consumed by the BFF.
type ReconciliationService interface {
	// GetSummary returns the reconciliation summary for the project and period.
	GetSummary(ctx context.Context, projectID, periodStart, periodEnd string) (*bffcontracts.ReconciliationSummaryResponse, error)

	// CreateManualLink manually links a statement transaction to a bill record.
	CreateManualLink(ctx context.Context, projectID, transactionLineID, billRecordID, linkedBy string) (*bffcontracts.ReconciliationLinkResponse, error)
}

// HistoryService defines the transport-agnostic financial history analytics consumed by the BFF.
type HistoryService interface {
	// GetTimeline returns aggregated bill amounts per calendar month.
	GetTimeline(ctx context.Context, projectID string, months int) (*bffcontracts.TimelineResponse, error)

	// GetCategoryBreakdown returns bill amounts grouped by bill type and month.
	GetCategoryBreakdown(ctx context.Context, projectID string, months int) (*bffcontracts.CategoryBreakdownResponse, error)

	// GetComplianceMetrics returns on-time vs overdue bill counts and compliance rate.
	GetComplianceMetrics(ctx context.Context, projectID string, months int) (*bffcontracts.ComplianceResponse, error)
}
