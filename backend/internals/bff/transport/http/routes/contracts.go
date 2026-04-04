// Package routes contains route module implementations that own all Huma
// operation registrations for the BFF HTTP transport layer.
// Controllers remain pure behaviour structs; route modules own the HTTP contract.
package routes

import (
	"context"

	"github.com/danielgtaylor/huma/v2"

	views "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/transport/http/views"
)

// Route is the shared registration contract for all BFF route modules.
// Every route module's Register method is called exactly once from container.go.
type Route interface {
	Register(api huma.API, auth func(huma.Context, func(huma.Context)))
}

// DocumentsCapability is the narrow handler interface consumed by DocumentsRoute.
type DocumentsCapability interface {
	HandleUpload(ctx context.Context, input *views.UploadDocumentInput) (*struct{ Body views.DocumentResponse }, error)
	HandleClassify(ctx context.Context, input *views.ClassifyDocumentInput) (*struct{ Body views.DocumentResponse }, error)
	HandleList(ctx context.Context, input *views.ListDocumentsInput) (*struct{ Body views.ListDocumentsResponse }, error)
	HandleGet(ctx context.Context, input *views.GetDocumentInput) (*struct{ Body views.DocumentDetailResponse }, error)
}

// ProjectsCapability is the narrow handler interface consumed by ProjectsRoute.
type ProjectsCapability interface {
	HandleGetCurrent(ctx context.Context, _ *struct{}) (*struct{ Body views.ProjectResponse }, error)
	HandleListMembers(ctx context.Context, input *views.ListMembersInput) (*struct{ Body views.ListMembersResponse }, error)
	HandleInvite(ctx context.Context, input *views.InviteMemberInput) (*struct{ Body views.ProjectMemberResponse }, error)
	HandleUpdateRole(ctx context.Context, input *views.UpdateMemberRoleInput) (*struct{ Body views.ProjectMemberResponse }, error)
}

// SettingsCapability is the narrow handler interface consumed by SettingsRoute.
type SettingsCapability interface {
	HandleList(ctx context.Context, _ *struct{}) (*struct {
		Body views.ListBankAccountsResponse
	}, error)
	HandleCreate(ctx context.Context, input *views.CreateBankAccountInput) (*struct{ Body views.BankAccountResponse }, error)
	HandleDelete(ctx context.Context, input *views.DeleteBankAccountInput) (*struct{}, error)
}

// AuthCapability is the narrow handler interface consumed by AuthRoute.
type AuthCapability interface {
	HandleLogin(ctx context.Context, input *views.LoginInput) (*views.LoginOutput, error)
	HandleRefresh(ctx context.Context, input *views.RefreshInput) (*views.RefreshOutput, error)
}

// PaymentsCapability is the narrow handler interface consumed by PaymentsRoute.
type PaymentsCapability interface {
	HandleGetDashboard(ctx context.Context, input *views.GetPaymentDashboardInput) (*struct {
		Body views.PaymentDashboardResponse
	}, error)
	HandleMarkPaid(ctx context.Context, input *views.MarkBillPaidInput) (*struct{ Body views.MarkBillPaidResponse }, error)
	HandleGetPreferredDay(ctx context.Context, _ *struct{}) (*struct{ Body views.CyclePreferenceResponse }, error)
	HandleSetPreferredDay(ctx context.Context, input *views.SetPreferredDayInput) (*struct{ Body views.CyclePreferenceResponse }, error)
}

// ReconciliationCapability is the narrow handler interface consumed by ReconciliationRoute.
type ReconciliationCapability interface {
	HandleGetSummary(ctx context.Context, input *views.ReconciliationSummaryInput) (*struct {
		Body views.ReconciliationSummaryResponse
	}, error)
	HandleCreateLink(ctx context.Context, input *views.CreateReconciliationLinkInput) (*struct {
		Body views.ReconciliationLinkResponse
	}, error)
}

// HistoryCapability is the narrow handler interface consumed by HistoryRoute.
type HistoryCapability interface {
	HandleGetTimeline(ctx context.Context, input *views.HistoryQueryInput) (*struct{ Body views.TimelineResponse }, error)
	HandleGetCategories(ctx context.Context, input *views.HistoryQueryInput) (*struct {
		Body views.CategoryBreakdownResponse
	}, error)
	HandleGetCompliance(ctx context.Context, input *views.HistoryQueryInput) (*struct{ Body views.ComplianceResponse }, error)
}

// Compile-time assertions: every route module satisfies the Route interface.
var (
	_ Route = (*DocumentsRoute)(nil)
	_ Route = (*ProjectsRoute)(nil)
	_ Route = (*SettingsRoute)(nil)
	_ Route = (*AuthRoute)(nil)
	_ Route = (*PaymentsRoute)(nil)
	_ Route = (*ReconciliationRoute)(nil)
	_ Route = (*HistoryRoute)(nil)
)
