package controllers

import (
	"context"
	"fmt"
	"strconv"

	"github.com/danielgtaylor/huma/v2"
	"go.uber.org/zap"

	bffmiddleware "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/transport/http/middleware"
	paymentsinterfaces "github.com/ralvescosta/costa-financial-assistant/backend/internals/payments/interfaces"
	billsv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/bills/v1"
	commonv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/common/v1"
)

// ─── Input / Output types ─────────────────────────────────────────────────────

// PaymentBillRecordResponse is the JSON shape for a single bill record in payment routes.
type PaymentBillRecordResponse struct {
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

// PaymentBillTypeResponse is the JSON shape for a bill type label in payment routes.
type PaymentBillTypeResponse struct {
	ID        string `json:"id"`
	ProjectID string `json:"projectId"`
	Name      string `json:"name"`
}

// PaymentDashboardEntryResponse represents a single dashboard row.
type PaymentDashboardEntryResponse struct {
	Bill         PaymentBillRecordResponse `json:"bill"`
	BillType     *PaymentBillTypeResponse  `json:"billType,omitempty"`
	IsOverdue    bool                      `json:"isOverdue"`
	DaysUntilDue int32                     `json:"daysUntilDue"`
}

// PaymentDashboardResponse is the GET payment-dashboard response body.
type PaymentDashboardResponse struct {
	Entries       []PaymentDashboardEntryResponse `json:"entries"`
	NextPageToken string                          `json:"nextPageToken,omitempty"`
}

// MarkBillPaidInput carries mark-paid request parameters.
type MarkBillPaidInput struct {
	BillID string `path:"billId" doc:"Bill record UUID"`
}

// MarkBillPaidResponse is returned on success.
type MarkBillPaidResponse struct {
	Bill PaymentBillRecordResponse `json:"bill"`
}

// CyclePreferenceResponse is the JSON shape for payment cycle preferences.
type CyclePreferenceResponse struct {
	ProjectID           string `json:"projectId"`
	PreferredDayOfMonth int    `json:"preferredDayOfMonth"`
	UpdatedAt           string `json:"updatedAt"`
}

// SetPreferredDayInput carries the preferred day of month.
type SetPreferredDayInput struct {
	Body struct {
		PreferredDayOfMonth int `json:"preferredDayOfMonth" minimum:"1" maximum:"28" doc:"Preferred payment day (1–28)"`
	}
}

// ─── Controller ───────────────────────────────────────────────────────────────

// GetPaymentDashboardInput carries query parameters for the payment dashboard.
type GetPaymentDashboardInput struct {
	CycleStart string `query:"cycleStart" doc:"ISO-8601 cycle start date (YYYY-MM-DD)"`
	CycleEnd   string `query:"cycleEnd" doc:"ISO-8601 cycle end date (YYYY-MM-DD)"`
	PageSize   string `query:"pageSize" doc:"Number of results per page"`
	PageToken  string `query:"pageToken" doc:"Opaque pagination cursor"`
}

// PaymentsController handles BFF payment HTTP endpoints.
type PaymentsController struct {
	BaseController
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
		BaseController: BaseController{logger: logger},
		billsClient:    billsClient,
		cycleService:   cycleService,
	}
}

// ─── Handlers ─────────────────────────────────────────────────────────────────

// HandleGetDashboard returns outstanding bills for the project's active payment cycle.
func (c *PaymentsController) HandleGetDashboard(ctx context.Context, input *GetPaymentDashboardInput) (*struct{ Body PaymentDashboardResponse }, error) {
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

	entries := make([]PaymentDashboardEntryResponse, 0, len(resp.GetEntries()))
	for _, e := range resp.GetEntries() {
		entry := PaymentDashboardEntryResponse{
			Bill:         protoBillRecordToResponse(e.GetBill()),
			IsOverdue:    e.GetIsOverdue(),
			DaysUntilDue: e.GetDaysUntilDue(),
		}
		if bt := e.GetBillType(); bt != nil {
			entry.BillType = &PaymentBillTypeResponse{
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

	return &struct{ Body PaymentDashboardResponse }{
		Body: PaymentDashboardResponse{
			Entries:       entries,
			NextPageToken: nextToken,
		},
	}, nil
}

// HandleMarkPaid idempotently marks a bill as paid.
func (c *PaymentsController) HandleMarkPaid(ctx context.Context, input *MarkBillPaidInput) (*struct{ Body MarkBillPaidResponse }, error) {
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

	return &struct{ Body MarkBillPaidResponse }{
		Body: MarkBillPaidResponse{Bill: protoBillRecordToResponse(resp.GetBill())},
	}, nil
}

// HandleGetPreferredDay returns the project's preferred payment day.
func (c *PaymentsController) HandleGetPreferredDay(ctx context.Context, _ *struct{}) (*struct{ Body CyclePreferenceResponse }, error) {
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

	return &struct{ Body CyclePreferenceResponse }{
		Body: CyclePreferenceResponse{
			ProjectID:           pref.ProjectID,
			PreferredDayOfMonth: pref.PreferredDayOfMonth,
			UpdatedAt:           pref.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		},
	}, nil
}

// HandleSetPreferredDay creates or updates the project's preferred payment day.
func (c *PaymentsController) HandleSetPreferredDay(ctx context.Context, input *SetPreferredDayInput) (*struct{ Body CyclePreferenceResponse }, error) {
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

	return &struct{ Body CyclePreferenceResponse }{
		Body: CyclePreferenceResponse{
			ProjectID:           pref.ProjectID,
			PreferredDayOfMonth: pref.PreferredDayOfMonth,
			UpdatedAt:           pref.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		},
	}, nil
}

// ── helpers ──────────────────────────────────────────────────────────────────

func protoBillRecordToResponse(b *billsv1.BillRecord) PaymentBillRecordResponse {
	if b == nil {
		return PaymentBillRecordResponse{}
	}
	return PaymentBillRecordResponse{
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

