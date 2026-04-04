//go:build integration

package cross_service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"google.golang.org/grpc"

	bffinterfaces "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/interfaces"
	bffservices "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/services"
	bffmiddleware "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/transport/http/middleware"
	billsv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/bills/v1"
	filesv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/files/v1"
	identityv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/identity/v1"
	onboardingv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/onboarding/v1"
	paymentsv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/payments/v1"
)

const (
	seededOwnerUserID    = "00000000-0000-0000-0000-000000000001"
	seededOwnerProjectID = "00000000-0000-0000-0000-000000000010"
)

func newAuthenticatedClaimsContext() context.Context {
	return context.WithValue(context.Background(), bffmiddleware.ProjectContextKey, &identityv1.JwtClaims{
		Subject:   seededOwnerUserID,
		ProjectId: seededOwnerProjectID,
		Role:      "write",
		Email:     "ralvescosta@local.dev",
		Username:  "ralvescosta",
	})
}

type recordingFilesClient struct {
	listDocumentsReq *filesv1.ListDocumentsRequest
}

func (c *recordingFilesClient) UploadDocument(context.Context, *filesv1.UploadDocumentRequest, ...grpc.CallOption) (*filesv1.UploadDocumentResponse, error) {
	return &filesv1.UploadDocumentResponse{}, nil
}
func (c *recordingFilesClient) ClassifyDocument(context.Context, *filesv1.ClassifyDocumentRequest, ...grpc.CallOption) (*filesv1.ClassifyDocumentResponse, error) {
	return &filesv1.ClassifyDocumentResponse{}, nil
}
func (c *recordingFilesClient) GetDocument(context.Context, *filesv1.GetDocumentRequest, ...grpc.CallOption) (*filesv1.GetDocumentResponse, error) {
	return &filesv1.GetDocumentResponse{}, nil
}
func (c *recordingFilesClient) ListDocuments(_ context.Context, in *filesv1.ListDocumentsRequest, _ ...grpc.CallOption) (*filesv1.ListDocumentsResponse, error) {
	c.listDocumentsReq = in
	return &filesv1.ListDocumentsResponse{Documents: []*filesv1.Document{}}, nil
}
func (c *recordingFilesClient) CreateBankAccount(context.Context, *filesv1.CreateBankAccountRequest, ...grpc.CallOption) (*filesv1.CreateBankAccountResponse, error) {
	return &filesv1.CreateBankAccountResponse{}, nil
}
func (c *recordingFilesClient) ListBankAccounts(context.Context, *filesv1.ListBankAccountsRequest, ...grpc.CallOption) (*filesv1.ListBankAccountsResponse, error) {
	return &filesv1.ListBankAccountsResponse{}, nil
}
func (c *recordingFilesClient) DeleteBankAccount(context.Context, *filesv1.DeleteBankAccountRequest, ...grpc.CallOption) (*filesv1.DeleteBankAccountResponse, error) {
	return &filesv1.DeleteBankAccountResponse{Success: true}, nil
}

type recordingOnboardingClient struct {
	listMembersReq *onboardingv1.ListProjectMembersRequest
}

func (c *recordingOnboardingClient) CreateProject(context.Context, *onboardingv1.CreateProjectRequest, ...grpc.CallOption) (*onboardingv1.CreateProjectResponse, error) {
	return &onboardingv1.CreateProjectResponse{}, nil
}
func (c *recordingOnboardingClient) InviteProjectMember(context.Context, *onboardingv1.InviteProjectMemberRequest, ...grpc.CallOption) (*onboardingv1.InviteProjectMemberResponse, error) {
	return &onboardingv1.InviteProjectMemberResponse{}, nil
}
func (c *recordingOnboardingClient) UpdateProjectMemberRole(context.Context, *onboardingv1.UpdateProjectMemberRoleRequest, ...grpc.CallOption) (*onboardingv1.UpdateProjectMemberRoleResponse, error) {
	return &onboardingv1.UpdateProjectMemberRoleResponse{}, nil
}
func (c *recordingOnboardingClient) ListProjectMembers(_ context.Context, in *onboardingv1.ListProjectMembersRequest, _ ...grpc.CallOption) (*onboardingv1.ListProjectMembersResponse, error) {
	c.listMembersReq = in
	return &onboardingv1.ListProjectMembersResponse{Members: []*onboardingv1.ProjectMember{}}, nil
}
func (c *recordingOnboardingClient) GetProject(context.Context, *onboardingv1.GetProjectRequest, ...grpc.CallOption) (*onboardingv1.GetProjectResponse, error) {
	return &onboardingv1.GetProjectResponse{}, nil
}

type recordingBillsClient struct {
	dashboardReq *billsv1.GetPaymentDashboardRequest
}

func (c *recordingBillsClient) GetPaymentDashboard(_ context.Context, in *billsv1.GetPaymentDashboardRequest, _ ...grpc.CallOption) (*billsv1.GetPaymentDashboardResponse, error) {
	c.dashboardReq = in
	return &billsv1.GetPaymentDashboardResponse{Entries: []*billsv1.PaymentDashboardEntry{}}, nil
}
func (c *recordingBillsClient) MarkBillPaid(context.Context, *billsv1.MarkBillPaidRequest, ...grpc.CallOption) (*billsv1.MarkBillPaidResponse, error) {
	return &billsv1.MarkBillPaidResponse{}, nil
}
func (c *recordingBillsClient) GetBill(context.Context, *billsv1.GetBillRequest, ...grpc.CallOption) (*billsv1.GetBillResponse, error) {
	return &billsv1.GetBillResponse{}, nil
}
func (c *recordingBillsClient) ListBills(context.Context, *billsv1.ListBillsRequest, ...grpc.CallOption) (*billsv1.ListBillsResponse, error) {
	return &billsv1.ListBillsResponse{}, nil
}

type recordingPaymentsClient struct {
	historyTimelineReq   *paymentsv1.GetHistoryTimelineRequest
	categoryBreakdownReq *paymentsv1.GetHistoryCategoryBreakdownRequest
	complianceReq        *paymentsv1.GetHistoryComplianceRequest
}

func (c *recordingPaymentsClient) GetCyclePreference(context.Context, *paymentsv1.GetCyclePreferenceRequest, ...grpc.CallOption) (*paymentsv1.GetCyclePreferenceResponse, error) {
	return &paymentsv1.GetCyclePreferenceResponse{}, nil
}
func (c *recordingPaymentsClient) SetCyclePreference(context.Context, *paymentsv1.SetCyclePreferenceRequest, ...grpc.CallOption) (*paymentsv1.SetCyclePreferenceResponse, error) {
	return &paymentsv1.SetCyclePreferenceResponse{}, nil
}
func (c *recordingPaymentsClient) GetHistoryTimeline(_ context.Context, in *paymentsv1.GetHistoryTimelineRequest, _ ...grpc.CallOption) (*paymentsv1.GetHistoryTimelineResponse, error) {
	c.historyTimelineReq = in
	return &paymentsv1.GetHistoryTimelineResponse{Entries: []*paymentsv1.MonthlyTimelineEntry{}}, nil
}
func (c *recordingPaymentsClient) GetHistoryCategoryBreakdown(_ context.Context, in *paymentsv1.GetHistoryCategoryBreakdownRequest, _ ...grpc.CallOption) (*paymentsv1.GetHistoryCategoryBreakdownResponse, error) {
	c.categoryBreakdownReq = in
	return &paymentsv1.GetHistoryCategoryBreakdownResponse{Entries: []*paymentsv1.CategoryBreakdownEntry{}}, nil
}
func (c *recordingPaymentsClient) GetHistoryCompliance(_ context.Context, in *paymentsv1.GetHistoryComplianceRequest, _ ...grpc.CallOption) (*paymentsv1.GetHistoryComplianceResponse, error) {
	c.complianceReq = in
	return &paymentsv1.GetHistoryComplianceResponse{Entries: []*paymentsv1.MonthlyComplianceEntry{}}, nil
}
func (c *recordingPaymentsClient) GetReconciliationSummary(context.Context, *paymentsv1.GetReconciliationSummaryRequest, ...grpc.CallOption) (*paymentsv1.GetReconciliationSummaryResponse, error) {
	return &paymentsv1.GetReconciliationSummaryResponse{}, nil
}
func (c *recordingPaymentsClient) CreateManualLink(context.Context, *paymentsv1.CreateManualLinkRequest, ...grpc.CallOption) (*paymentsv1.CreateManualLinkResponse, error) {
	return &paymentsv1.CreateManualLinkResponse{}, nil
}

var (
	_ bffinterfaces.FilesClient      = (*recordingFilesClient)(nil)
	_ bffinterfaces.OnboardingClient = (*recordingOnboardingClient)(nil)
)

func TestRestoreLoginSessionPropagation_TableDriven(t *testing.T) {
	t.Parallel()

	scenarios := []struct {
		name  string
		given string
		when  string
		then  string
		run   func(t *testing.T, ctx context.Context)
	}{
		{
			name:  "GivenAuthenticatedDocumentListWhenBFFCallsFilesThenSessionAndFallbackPaginationAreForwarded",
			given: "an authenticated seeded-owner context and the documents BFF service",
			when:  "the documents list flow is executed without explicit paging",
			then:  "the downstream files request carries the shared Session and the documented page_size=25 fallback",
			run: func(t *testing.T, ctx context.Context) {
				// Arrange
				filesClient := &recordingFilesClient{}
				svc := bffservices.NewDocumentsService(zaptest.NewLogger(t), filesClient)

				// Act
				_, err := svc.ListDocuments(ctx, seededOwnerProjectID, 0, "")

				// Assert
				require.NoError(t, err)
				require.NotNil(t, filesClient.listDocumentsReq)
				require.NotNil(t, filesClient.listDocumentsReq.GetSession())
				assert.Equal(t, seededOwnerUserID, filesClient.listDocumentsReq.GetSession().GetId())
				assert.Equal(t, "ralvescosta", filesClient.listDocumentsReq.GetSession().GetUsername())
				assert.EqualValues(t, 25, filesClient.listDocumentsReq.GetPagination().GetPageSize())
			},
		},
		{
			name:  "GivenAuthenticatedMemberListWhenBFFCallsOnboardingThenSessionAndFallbackPaginationAreForwarded",
			given: "an authenticated seeded-owner context and the projects BFF service",
			when:  "the project-members list flow is executed without explicit paging",
			then:  "the downstream onboarding request carries the shared Session and the documented page_size=25 fallback",
			run: func(t *testing.T, ctx context.Context) {
				// Arrange
				onboardingClient := &recordingOnboardingClient{}
				svc := bffservices.NewProjectsService(zaptest.NewLogger(t), onboardingClient)

				// Act
				_, err := svc.ListMembers(ctx, seededOwnerProjectID, seededOwnerUserID, "write", 0, "")

				// Assert
				require.NoError(t, err)
				require.NotNil(t, onboardingClient.listMembersReq)
				require.NotNil(t, onboardingClient.listMembersReq.GetSession())
				assert.Equal(t, seededOwnerUserID, onboardingClient.listMembersReq.GetSession().GetId())
				assert.Equal(t, "ralvescosta@local.dev", onboardingClient.listMembersReq.GetSession().GetEmail())
				assert.EqualValues(t, 25, onboardingClient.listMembersReq.GetPagination().GetPageSize())
			},
		},
		{
			name:  "GivenAuthenticatedPaymentDashboardWhenBFFCallsBillsThenSessionAndFallbackPaginationAreForwarded",
			given: "an authenticated seeded-owner context and the payments BFF service",
			when:  "the payment-dashboard flow is executed without explicit paging",
			then:  "the downstream bills request carries the shared Session and the documented page_size=20 fallback",
			run: func(t *testing.T, ctx context.Context) {
				// Arrange
				billsClient := &recordingBillsClient{}
				paymentsClient := &recordingPaymentsClient{}
				svc := bffservices.NewPaymentsService(zaptest.NewLogger(t), billsClient, paymentsClient)

				// Act
				_, err := svc.GetPaymentDashboard(ctx, seededOwnerProjectID, seededOwnerUserID, "", "", 0, "")

				// Assert
				require.NoError(t, err)
				require.NotNil(t, billsClient.dashboardReq)
				require.NotNil(t, billsClient.dashboardReq.GetSession())
				assert.Equal(t, seededOwnerUserID, billsClient.dashboardReq.GetSession().GetId())
				assert.EqualValues(t, 20, billsClient.dashboardReq.GetPagination().GetPageSize())
			},
		},
	}

	for _, scenario := range scenarios {
		scenario := scenario
		t.Run(scenario.name, func(t *testing.T) {
			// Given
			ctx := newAuthenticatedClaimsContext()

			// When / Then
			scenario.run(t, ctx)
		})
	}
}
