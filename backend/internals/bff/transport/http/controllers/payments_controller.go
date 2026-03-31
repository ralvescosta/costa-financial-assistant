package controllers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/danielgtaylor/huma/v2"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	bffmiddleware "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/transport/http/middleware"
	paymentsinterfaces "github.com/ralvescosta/costa-financial-assistant/backend/internals/payments/interfaces"
	billsv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/bills/v1"
	commonv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/common/v1"
)

// ─── Input / Output types ─────────────────────────────────────────────────────

// paymentBillRecordResponse is the JSON shape for a single bill record in payment routes.
type paymentBillRecordResponse struct {
	ID            string `json:"id"`
	ProjectID     string `json:"projectId"`
	DocumentID    string `json:"documentId"`
	BillTypeID    string `json:"billTypeId,omitempty"`
	DueDate       string `json:"dueDate"`
	AmountDue     string `json:"amountDue"`
	PixPayload    string `json:"pixPayload,omitempty"`
	PixQRImageRef string `json:"pixQrImageRef,omitempty"`
	Barcode       string `json:"barcode,omitempty"`
	PaymentStatus string `json:"paymentStatus"`
	PaidAt        string `json:"paidAt,omitempty"`
	MarkedPaidBy  string `json:"markedPaidBy,omitempty"`
	CreatedAt     string `json:"createdAt"`
	UpdatedAt     string `json:"updatedAt"`
}

// paymentBillTypeResponse is the JSON shape for a bill type label in payment routes.
type paymentBillTypeResponse struct {
	ID        string `json:"id"`
	ProjectID string `json:"projectId"`
	Name      string `json:"name"`
}

// paymentDashboardEntryResponse represents a single dashboard row.
type paymentDashboardEntryResponse struct {
	Bill         paymentBillRecordResponse `json:"bill"`
	BillType     *paymentBillTypeResponse  `json:"billType,omitempty"`
	IsOverdue    bool                      `json:"isOverdue"`
	DaysUntilDue int32                     `json:"daysUntilDue"`
}

// paymentDashboardResponse is the GET payment-dashboard response body.
type paymentDashboardResponse struct {
	Entries       []paymentDashboardEntryResponse `json:"entries"`
	NextPageToken string                          `json:"nextPageToken,omitempty"`
}

// markBillPaidInput carries mark-paid request parameters.
type markBillPaidInput struct {
	BillID string `path:"billId" doc:"Bill record UUID"`
}

// markBillPaidResponse is returned on success.
type markBillPaidResponse struct {
	Bill paymentBillRecordResponse `json:"bill"`
}

// cyclePreferenceResponse is the JSON shape for payment cycle preferences.
type cyclePreferenceResponse struct {
	ProjectID           string `json:"projectId"`
	PreferredDayOfMonth int    `json:"preferredDayOfMonth"`
	UpdatedAt           string `json:"updatedAt"`
}

// setPreferredDayInput carries the preferred day of month.
type setPreferredDayInput struct {
	Body struct {
		PreferredDayOfMonth int `json:"preferredDayOfMonth" minimum:"1" maximum:"28" doc:"Preferred payment day (1–28)"`
	}
}

// ─── Controller ───────────────────────────────────────────────────────────────

// PaymentsController registers and handles all payment-related HTTP routes.
type PaymentsController struct {
	logger       *zap.Logger
	billsClient  billsv1.BillsServiceClient
	cycleService paymentsinterfaces.PaymentCycleService
}

// NewPaymentsController constructs a PaymentsController.
func NewPaymentsController(
	logger *zap.Logger,
	billsClient billsv1.BillsServiceClient,
	cycleService paymentsinterfaces.PaymentCycleService,
) *PaymentsController {
	return &PaymentsController{
		logger:       logger,
		billsClient:  billsClient,
		cycleService: cycleService,
	}
}

// Register wires all payment routes to the Huma API with auth + role middleware.
func (c *PaymentsController) Register(api huma.API, auth func(huma.Context, func(huma.Context))) {
	huma.Register(api, huma.Operation{
		OperationID: "get-payment-dashboard",
		Method:      http.MethodGet,
		Path:        "/api/v1/bills/payment-dashboard",
		Summary:     "List outstanding bills for current cycle",
		Description: "Returns outstanding and overdue bills for the project's active payment cycle.",
		Tags:        []string{"payments"},
		Middlewares: huma.Middlewares{auth, bffmiddleware.NewProjectGuard("read_only", c.logger)},
	}, c.handleGetDashboard)

	huma.Register(api, huma.Operation{
		OperationID: "mark-bill-paid",
		Method:      http.MethodPost,
		Path:        "/api/v1/bills/{billId}/mark-paid",
		Summary:     "Mark bill as paid (idempotent)",
		Description: "Idempotently marks the bill identified by billId as paid in the caller's project.",
		Tags:        []string{"payments"},
		Middlewares: huma.Middlewares{auth, bffmiddleware.NewProjectGuard("update", c.logger)},
	}, c.handleMarkPaid)

	huma.Register(api, huma.Operation{
		OperationID: "get-preferred-payment-day",
		Method:      http.MethodGet,
		Path:        "/api/v1/payment-cycle/preferred-day",
		Summary:     "Get preferred payment day for active project",
		Description: "Returns the project's configured preferred payment day of month.",
		Tags:        []string{"settings"},
		Middlewares: huma.Middlewares{auth, bffmiddleware.NewProjectGuard("read_only", c.logger)},
	}, c.handleGetPreferredDay)

	huma.Register(api, huma.Operation{
		OperationID: "set-preferred-payment-day",
		Method:      http.MethodPut,
		Path:        "/api/v1/payment-cycle/preferred-day",
		Summary:     "Set preferred payment day for active project",
		Description: "Creates or updates the project's preferred payment day of month (1–28).",
		Tags:        []string{"settings"},
		Middlewares: huma.Middlewares{auth, bffmiddleware.NewProjectGuard("update", c.logger)},
	}, c.handleSetPreferredDay)
}

// ─── Handlers ─────────────────────────────────────────────────────────────────

func (c *PaymentsController) handleGetDashboard(ctx context.Context, input *struct {
	CycleStart string `query:"cycleStart" doc:"ISO-8601 cycle start date (YYYY-MM-DD)"`
	CycleEnd   string `query:"cycleEnd" doc:"ISO-8601 cycle end date (YYYY-MM-DD)"`
	PageSize   string `query:"pageSize" doc:"Number of results per page"`
	PageToken  string `query:"pageToken" doc:"Opaque pagination cursor"`
}) (*struct{ Body paymentDashboardResponse }, error) {
	claims := bffmiddleware.ClaimsFromContext(ctx)
	if claims == nil {
		return nil, huma.Error403Forbidden("missing project context")
	}

	pageSize := int32(20)
	if input.PageSize != "" {
		if n, err := strconv.Atoi(input.PageSize); err == nil && n > 0 {
			pageSize = int32(n)
		}
	}

	resp, err := c.billsClient.GetPaymentDashboard(ctx, &billsv1.GetPaymentDashboardRequest{
		Ctx: &commonv1.ProjectContext{
			ProjectId: claims.GetProjectId(),
			UserId:    claims.GetSubject(),
		},
		CycleStart: input.CycleStart,
		CycleEnd:   input.CycleEnd,
		Pagination: &commonv1.Pagination{PageSize: pageSize, PageToken: input.PageToken},
	})
	if err != nil {
		return nil, c.grpcToHumaError(err, "get payment dashboard failed")
	}

	entries := make([]paymentDashboardEntryResponse, 0, len(resp.GetEntries()))
	for _, e := range resp.GetEntries() {
		entry := paymentDashboardEntryResponse{
			Bill:         protoBillRecordToResponse(e.GetBill()),
			IsOverdue:    e.GetIsOverdue(),
			DaysUntilDue: e.GetDaysUntilDue(),
		}
		if bt := e.GetBillType(); bt != nil {
			entry.BillType = &paymentBillTypeResponse{
				ID:        bt.GetId(),
				ProjectID: bt.GetProjectId(),
				Name:      bt.GetName(),
			}
		}
		entries = append(entries, entry)
	}

	var nextToken string
	if resp.GetPagination() != nil {
		nextToken = resp.GetPagination().GetNextPageToken()
	}

	return &struct{ Body paymentDashboardResponse }{
		Body: paymentDashboardResponse{
			Entries:       entries,
			NextPageToken: nextToken,
		},
	}, nil
}

func (c *PaymentsController) handleMarkPaid(ctx context.Context, input *markBillPaidInput) (*struct{ Body markBillPaidResponse }, error) {
	claims := bffmiddleware.ClaimsFromContext(ctx)
	if claims == nil {
		return nil, huma.Error403Forbidden("missing project context")
	}

	resp, err := c.billsClient.MarkBillPaid(ctx, &billsv1.MarkBillPaidRequest{
		Ctx: &commonv1.ProjectContext{
			ProjectId: claims.GetProjectId(),
			UserId:    claims.GetSubject(),
		},
		BillId: input.BillID,
		Audit:  &commonv1.AuditMetadata{PerformedBy: claims.GetSubject()},
	})
	if err != nil {
		return nil, c.grpcToHumaError(err, "mark bill paid failed")
	}

	c.logger.Info("payments: bill marked paid",
		zap.String("bill_id", input.BillID),
		zap.String("project_id", claims.GetProjectId()))

	return &struct{ Body markBillPaidResponse }{
		Body: markBillPaidResponse{Bill: protoBillRecordToResponse(resp.GetBill())},
	}, nil
}

func (c *PaymentsController) handleGetPreferredDay(ctx context.Context, _ *struct{}) (*struct{ Body cyclePreferenceResponse }, error) {
	claims := bffmiddleware.ClaimsFromContext(ctx)
	if claims == nil {
		return nil, huma.Error403Forbidden("missing project context")
	}

	pref, err := c.cycleService.GetCyclePreference(ctx, claims.GetProjectId())
	if err != nil {
		c.logger.Error("payments: get preferred day failed", zap.Error(err))
		return nil, huma.Error500InternalServerError("get preferred day failed")
	}
	if pref == nil {
		return nil, huma.Error404NotFound("no payment cycle preference configured")
	}

	return &struct{ Body cyclePreferenceResponse }{
		Body: cyclePreferenceResponse{
			ProjectID:           pref.ProjectID,
			PreferredDayOfMonth: pref.PreferredDayOfMonth,
			UpdatedAt:           pref.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		},
	}, nil
}

func (c *PaymentsController) handleSetPreferredDay(ctx context.Context, input *setPreferredDayInput) (*struct{ Body cyclePreferenceResponse }, error) {
	claims := bffmiddleware.ClaimsFromContext(ctx)
	if claims == nil {
		return nil, huma.Error403Forbidden("missing project context")
	}

	day := input.Body.PreferredDayOfMonth
	if day < 1 || day > 28 {
		return nil, huma.Error400BadRequest(fmt.Sprintf("preferredDayOfMonth must be between 1 and 28, got %d", day))
	}

	pref, err := c.cycleService.UpsertCyclePreference(ctx, claims.GetProjectId(), day, claims.GetSubject())
	if err != nil {
		c.logger.Error("payments: set preferred day failed", zap.Error(err))
		return nil, huma.Error500InternalServerError("set preferred day failed")
	}

	c.logger.Info("payments: preferred day set",
		zap.String("project_id", claims.GetProjectId()),
		zap.Int("day", day))

	return &struct{ Body cyclePreferenceResponse }{
		Body: cyclePreferenceResponse{
			ProjectID:           pref.ProjectID,
			PreferredDayOfMonth: pref.PreferredDayOfMonth,
			UpdatedAt:           pref.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		},
	}, nil
}

// ── helpers ──────────────────────────────────────────────────────────────────

func protoBillRecordToResponse(b *billsv1.BillRecord) paymentBillRecordResponse {
	if b == nil {
		return paymentBillRecordResponse{}
	}
	return paymentBillRecordResponse{
		ID:            b.GetId(),
		ProjectID:     b.GetProjectId(),
		DocumentID:    b.GetDocumentId(),
		BillTypeID:    b.GetBillTypeId(),
		DueDate:       b.GetDueDate(),
		AmountDue:     b.GetAmountDue(),
		PixPayload:    b.GetPixPayload(),
		PixQRImageRef: b.GetPixQrImageRef(),
		Barcode:       b.GetBarcode(),
		PaymentStatus: b.GetPaymentStatus().String(),
		PaidAt:        b.GetPaidAt(),
		MarkedPaidBy:  b.GetMarkedPaidBy(),
		CreatedAt:     b.GetCreatedAt(),
		UpdatedAt:     b.GetUpdatedAt(),
	}
}

// grpcToHumaError maps gRPC status codes to Huma HTTP errors.
func (c *PaymentsController) grpcToHumaError(err error, fallback string) error {
	st, ok := status.FromError(err)
	if !ok {
		c.logger.Error(fallback, zap.Error(err))
		return huma.Error500InternalServerError(fallback)
	}
	switch st.Code() {
	case codes.NotFound:
		return huma.Error404NotFound(st.Message())
	case codes.AlreadyExists:
		return huma.Error409Conflict(st.Message())
	case codes.InvalidArgument:
		return huma.Error400BadRequest(st.Message())
	case codes.FailedPrecondition:
		return huma.Error409Conflict(st.Message())
	case codes.PermissionDenied:
		return huma.Error403Forbidden(st.Message())
	case codes.Unauthenticated:
		return huma.Error401Unauthorized(st.Message())
	default:
		c.logger.Error(fallback, zap.Error(err))
		return huma.Error500InternalServerError(fallback)
	}
}
