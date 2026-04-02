//go:build integration

package integration

import (
	"context"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	controllers "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/transport/http/controllers"
	bfftransportroutes "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/transport/http/routes"
)

// ── Stub capability implementations ──────────────────────────────────────────
// These stubs satisfy the capacity interfaces without requiring real gRPC clients
// or database connections. They are used exclusively to verify route registration.

type stubDocuments struct{}

func (stubDocuments) HandleUpload(_ context.Context, _ *controllers.UploadDocumentInput) (*struct{ Body controllers.DocumentResponse }, error) {
	return nil, nil
}
func (stubDocuments) HandleClassify(_ context.Context, _ *controllers.ClassifyDocumentInput) (*struct{ Body controllers.DocumentResponse }, error) {
	return nil, nil
}
func (stubDocuments) HandleList(_ context.Context, _ *controllers.ListDocumentsInput) (*struct {
	Body controllers.ListDocumentsResponse
}, error) {
	return nil, nil
}
func (stubDocuments) HandleGet(_ context.Context, _ *controllers.GetDocumentInput) (*struct {
	Body controllers.DocumentDetailResponse
}, error) {
	return nil, nil
}

type stubProjects struct{}

func (stubProjects) HandleGetCurrent(_ context.Context, _ *struct{}) (*struct{ Body controllers.ProjectResponse }, error) {
	return nil, nil
}
func (stubProjects) HandleListMembers(_ context.Context, _ *controllers.ListMembersInput) (*struct {
	Body controllers.ListMembersResponse
}, error) {
	return nil, nil
}
func (stubProjects) HandleInvite(_ context.Context, _ *controllers.InviteMemberInput) (*struct {
	Body controllers.ProjectMemberResponse
}, error) {
	return nil, nil
}
func (stubProjects) HandleUpdateRole(_ context.Context, _ *controllers.UpdateMemberRoleInput) (*struct {
	Body controllers.ProjectMemberResponse
}, error) {
	return nil, nil
}

type stubSettings struct{}

func (stubSettings) HandleList(_ context.Context, _ *struct{}) (*struct {
	Body controllers.ListBankAccountsResponse
}, error) {
	return nil, nil
}
func (stubSettings) HandleCreate(_ context.Context, _ *controllers.CreateBankAccountInput) (*struct {
	Body controllers.BankAccountResponse
}, error) {
	return nil, nil
}
func (stubSettings) HandleDelete(_ context.Context, _ *controllers.DeleteBankAccountInput) (*struct{}, error) {
	return nil, nil
}

type stubPayments struct{}

func (stubPayments) HandleGetDashboard(_ context.Context, _ *controllers.GetPaymentDashboardInput) (*struct {
	Body controllers.PaymentDashboardResponse
}, error) {
	return nil, nil
}
func (stubPayments) HandleMarkPaid(_ context.Context, _ *controllers.MarkBillPaidInput) (*struct {
	Body controllers.MarkBillPaidResponse
}, error) {
	return nil, nil
}
func (stubPayments) HandleGetPreferredDay(_ context.Context, _ *struct{}) (*struct {
	Body controllers.CyclePreferenceResponse
}, error) {
	return nil, nil
}
func (stubPayments) HandleSetPreferredDay(_ context.Context, _ *controllers.SetPreferredDayInput) (*struct {
	Body controllers.CyclePreferenceResponse
}, error) {
	return nil, nil
}

type stubReconciliation struct{}

func (stubReconciliation) HandleGetSummary(_ context.Context, _ *controllers.ReconciliationSummaryInput) (*struct {
	Body controllers.ReconciliationSummaryResponse
}, error) {
	return nil, nil
}
func (stubReconciliation) HandleCreateLink(_ context.Context, _ *controllers.CreateReconciliationLinkInput) (*struct {
	Body controllers.ReconciliationLinkResponse
}, error) {
	return nil, nil
}

type stubHistory struct{}

func (stubHistory) HandleGetTimeline(_ context.Context, _ *controllers.HistoryQueryInput) (*struct{ Body controllers.TimelineResponse }, error) {
	return nil, nil
}
func (stubHistory) HandleGetCategories(_ context.Context, _ *controllers.HistoryQueryInput) (*struct {
	Body controllers.CategoryBreakdownResponse
}, error) {
	return nil, nil
}
func (stubHistory) HandleGetCompliance(_ context.Context, _ *controllers.HistoryQueryInput) (*struct {
	Body controllers.ComplianceResponse
}, error) {
	return nil, nil
}

// ── Smoke test ────────────────────────────────────────────────────────────────

// TestBFFRouteRegistrationSmoke verifies that every expected OpenAPI operation ID
// is registered when all six route modules are wired together using the
// buildBFFTestServer helper. This test does NOT exercise handler logic — it only
// validates that the registration wiring is complete and correct.
func TestBFFRouteRegistrationSmoke(t *testing.T) {
	logger := zap.NewNop()

	routeModules := []bfftransportroutes.Route{
		bfftransportroutes.NewDocumentsRoute(stubDocuments{}, logger),
		bfftransportroutes.NewProjectsRoute(stubProjects{}, logger),
		bfftransportroutes.NewSettingsRoute(stubSettings{}, logger),
		bfftransportroutes.NewPaymentsRoute(stubPayments{}, logger),
		bfftransportroutes.NewReconciliationRoute(stubReconciliation{}, logger),
		bfftransportroutes.NewHistoryRoute(stubHistory{}, logger),
	}

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

	assert.Len(t, registered, len(expectedOperationIDs),
		"expected exactly %d registered operations", len(expectedOperationIDs))

	for _, id := range expectedOperationIDs {
		assert.True(t, registered[id], "operation %q not registered", id)
	}
}
