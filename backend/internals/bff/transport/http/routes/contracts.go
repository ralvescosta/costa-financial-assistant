// Package routes contains route module implementations that own all Huma
// operation registrations for the BFF HTTP transport layer.
// Controllers remain pure behaviour structs; route modules own the HTTP contract.
package routes

import (
	"context"

	"github.com/danielgtaylor/huma/v2"

	controllers "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/transport/http/controllers"
)

// Route is the shared registration contract for all BFF route modules.
// Every route module's Register method is called exactly once from container.go.
type Route interface {
	Register(api huma.API, auth func(huma.Context, func(huma.Context)))
}

// DocumentsCapability is the narrow handler interface consumed by DocumentsRoute.
type DocumentsCapability interface {
	HandleUpload(ctx context.Context, input *controllers.UploadDocumentInput) (*struct{ Body controllers.DocumentResponse }, error)
	HandleClassify(ctx context.Context, input *controllers.ClassifyDocumentInput) (*struct{ Body controllers.DocumentResponse }, error)
	HandleList(ctx context.Context, input *controllers.ListDocumentsInput) (*struct{ Body controllers.ListDocumentsResponse }, error)
	HandleGet(ctx context.Context, input *controllers.GetDocumentInput) (*struct{ Body controllers.DocumentDetailResponse }, error)
}

// ProjectsCapability is the narrow handler interface consumed by ProjectsRoute.
type ProjectsCapability interface {
	HandleGetCurrent(ctx context.Context, _ *struct{}) (*struct{ Body controllers.ProjectResponse }, error)
	HandleListMembers(ctx context.Context, input *controllers.ListMembersInput) (*struct{ Body controllers.ListMembersResponse }, error)
	HandleInvite(ctx context.Context, input *controllers.InviteMemberInput) (*struct{ Body controllers.ProjectMemberResponse }, error)
	HandleUpdateRole(ctx context.Context, input *controllers.UpdateMemberRoleInput) (*struct{ Body controllers.ProjectMemberResponse }, error)
}

// SettingsCapability is the narrow handler interface consumed by SettingsRoute.
type SettingsCapability interface {
	HandleList(ctx context.Context, _ *struct{}) (*struct{ Body controllers.ListBankAccountsResponse }, error)
	HandleCreate(ctx context.Context, input *controllers.CreateBankAccountInput) (*struct{ Body controllers.BankAccountResponse }, error)
	HandleDelete(ctx context.Context, input *controllers.DeleteBankAccountInput) (*struct{}, error)
}

// PaymentsCapability is the narrow handler interface consumed by PaymentsRoute.
type PaymentsCapability interface {
	HandleGetDashboard(ctx context.Context, input *controllers.GetPaymentDashboardInput) (*struct{ Body controllers.PaymentDashboardResponse }, error)
	HandleMarkPaid(ctx context.Context, input *controllers.MarkBillPaidInput) (*struct{ Body controllers.MarkBillPaidResponse }, error)
	HandleGetPreferredDay(ctx context.Context, _ *struct{}) (*struct{ Body controllers.CyclePreferenceResponse }, error)
	HandleSetPreferredDay(ctx context.Context, input *controllers.SetPreferredDayInput) (*struct{ Body controllers.CyclePreferenceResponse }, error)
}

// ReconciliationCapability is the narrow handler interface consumed by ReconciliationRoute.
type ReconciliationCapability interface {
	HandleGetSummary(ctx context.Context, input *controllers.ReconciliationSummaryInput) (*struct{ Body controllers.ReconciliationSummaryResponse }, error)
	HandleCreateLink(ctx context.Context, input *controllers.CreateReconciliationLinkInput) (*struct{ Body controllers.ReconciliationLinkResponse }, error)
}

// HistoryCapability is the narrow handler interface consumed by HistoryRoute.
type HistoryCapability interface {
	HandleGetTimeline(ctx context.Context, input *controllers.HistoryQueryInput) (*struct{ Body controllers.TimelineResponse }, error)
	HandleGetCategories(ctx context.Context, input *controllers.HistoryQueryInput) (*struct{ Body controllers.CategoryBreakdownResponse }, error)
	HandleGetCompliance(ctx context.Context, input *controllers.HistoryQueryInput) (*struct{ Body controllers.ComplianceResponse }, error)
}

// Compile-time assertions: every route module satisfies the Route interface.
var (
	_ Route = (*DocumentsRoute)(nil)
	_ Route = (*ProjectsRoute)(nil)
	_ Route = (*SettingsRoute)(nil)
	_ Route = (*PaymentsRoute)(nil)
	_ Route = (*ReconciliationRoute)(nil)
	_ Route = (*HistoryRoute)(nil)
)
