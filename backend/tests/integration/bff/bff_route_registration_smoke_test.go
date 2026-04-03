//go:build integration

package integration

import (
	"context"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	bfftransportroutes "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/transport/http/routes"
	views "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/transport/http/views"
)

// ── Stub capability implementations ──────────────────────────────────────────
// These stubs satisfy the capacity interfaces without requiring real gRPC clients
// or database connections. They are used exclusively to verify route registration.

type stubDocuments struct{}

func (stubDocuments) HandleUpload(_ context.Context, _ *views.UploadDocumentInput) (*struct{ Body views.DocumentResponse }, error) {
	return nil, nil
}
func (stubDocuments) HandleClassify(_ context.Context, _ *views.ClassifyDocumentInput) (*struct{ Body views.DocumentResponse }, error) {
	return nil, nil
}
func (stubDocuments) HandleList(_ context.Context, _ *views.ListDocumentsInput) (*struct {
	Body views.ListDocumentsResponse
}, error) {
	return nil, nil
}
func (stubDocuments) HandleGet(_ context.Context, _ *views.GetDocumentInput) (*struct {
	Body views.DocumentDetailResponse
}, error) {
	return nil, nil
}

type stubProjects struct{}

func (stubProjects) HandleGetCurrent(_ context.Context, _ *struct{}) (*struct{ Body views.ProjectResponse }, error) {
	return nil, nil
}
func (stubProjects) HandleListMembers(_ context.Context, _ *views.ListMembersInput) (*struct {
	Body views.ListMembersResponse
}, error) {
	return nil, nil
}
func (stubProjects) HandleInvite(_ context.Context, _ *views.InviteMemberInput) (*struct {
	Body views.ProjectMemberResponse
}, error) {
	return nil, nil
}
func (stubProjects) HandleUpdateRole(_ context.Context, _ *views.UpdateMemberRoleInput) (*struct {
	Body views.ProjectMemberResponse
}, error) {
	return nil, nil
}

type stubSettings struct{}

func (stubSettings) HandleList(_ context.Context, _ *struct{}) (*struct {
	Body views.ListBankAccountsResponse
}, error) {
	return nil, nil
}
func (stubSettings) HandleCreate(_ context.Context, _ *views.CreateBankAccountInput) (*struct {
	Body views.BankAccountResponse
}, error) {
	return nil, nil
}
func (stubSettings) HandleDelete(_ context.Context, _ *views.DeleteBankAccountInput) (*struct{}, error) {
	return nil, nil
}

type stubPayments struct{}

func (stubPayments) HandleGetDashboard(_ context.Context, _ *views.GetPaymentDashboardInput) (*struct {
	Body views.PaymentDashboardResponse
}, error) {
	return nil, nil
}
func (stubPayments) HandleMarkPaid(_ context.Context, _ *views.MarkBillPaidInput) (*struct {
	Body views.MarkBillPaidResponse
}, error) {
	return nil, nil
}
func (stubPayments) HandleGetPreferredDay(_ context.Context, _ *struct{}) (*struct {
	Body views.CyclePreferenceResponse
}, error) {
	return nil, nil
}
func (stubPayments) HandleSetPreferredDay(_ context.Context, _ *views.SetPreferredDayInput) (*struct {
	Body views.CyclePreferenceResponse
}, error) {
	return nil, nil
}

type stubReconciliation struct{}

func (stubReconciliation) HandleGetSummary(_ context.Context, _ *views.ReconciliationSummaryInput) (*struct {
	Body views.ReconciliationSummaryResponse
}, error) {
	return nil, nil
}
func (stubReconciliation) HandleCreateLink(_ context.Context, _ *views.CreateReconciliationLinkInput) (*struct {
	Body views.ReconciliationLinkResponse
}, error) {
	return nil, nil
}

type stubHistory struct{}

func (stubHistory) HandleGetTimeline(_ context.Context, _ *views.HistoryQueryInput) (*struct{ Body views.TimelineResponse }, error) {
	return nil, nil
}
func (stubHistory) HandleGetCategories(_ context.Context, _ *views.HistoryQueryInput) (*struct {
	Body views.CategoryBreakdownResponse
}, error) {
	return nil, nil
}
func (stubHistory) HandleGetCompliance(_ context.Context, _ *views.HistoryQueryInput) (*struct {
	Body views.ComplianceResponse
}, error) {
	return nil, nil
}

// ── Smoke test ────────────────────────────────────────────────────────────────

// TestBFFRouteRegistrationSmoke verifies that every expected OpenAPI operation ID
// is registered when all six route modules are wired together using the
// buildBFFTestServer helper. This test does NOT exercise handler logic — it only
// validates that the registration wiring is complete and correct.
func TestBFFRouteRegistrationSmoke(t *testing.T) {
	// Given all active BFF route modules are wired into the in-process test server.
	// Arrange
	logger := zap.NewNop()

	routeModules := []bfftransportroutes.Route{
		bfftransportroutes.NewDocumentsRoute(stubDocuments{}, logger),
		bfftransportroutes.NewProjectsRoute(stubProjects{}, logger),
		bfftransportroutes.NewSettingsRoute(stubSettings{}, logger),
		bfftransportroutes.NewPaymentsRoute(stubPayments{}, logger),
		bfftransportroutes.NewReconciliationRoute(stubReconciliation{}, logger),
		bfftransportroutes.NewHistoryRoute(stubHistory{}, logger),
	}

	// When the OpenAPI document is produced from registered routes.
	// Act
	_, api := buildBFFTestServer(t, routeModules...)

	expectedOperationIDs := []string{
		// Documents (4)
		"upload-document",
		"classify-document",
		"list-documents",
		"get-document",
		// Projects (4)
		"get-current-project",
		"list-project-members",
		"invite-project-member",
		"update-project-member-role",
		// Settings (3)
		"list-bank-accounts",
		"create-bank-account",
		"delete-bank-account",
		// Payments (4)
		"get-payment-dashboard",
		"mark-bill-paid",
		"get-preferred-payment-day",
		"set-preferred-payment-day",
		// Reconciliation (2)
		"get-reconciliation-summary",
		"create-reconciliation-link",
		// History (3)
		"get-history-timeline",
		"get-history-categories",
		"get-history-compliance",
	}

	// Collect all registered operation IDs from the OpenAPI spec.
	registered := map[string]bool{}
	for _, pathItem := range api.OpenAPI().Paths {
		for _, op := range []*huma.Operation{
			pathItem.Get,
			pathItem.Post,
			pathItem.Put,
			pathItem.Patch,
			pathItem.Delete,
		} {
			if op != nil && op.OperationID != "" {
				registered[op.OperationID] = true
			}
		}
	}

	// Then every expected operation is registered exactly once.
	// Assert
	assert.Len(t, registered, len(expectedOperationIDs),
		"expected exactly %d registered operations", len(expectedOperationIDs))

	for _, id := range expectedOperationIDs {
		assert.True(t, registered[id], "operation %q not registered", id)
	}
}

