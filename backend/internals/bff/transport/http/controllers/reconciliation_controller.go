package controllers

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"go.uber.org/zap"

	bffmiddleware "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/transport/http/middleware"
	paymentsinterfaces "github.com/ralvescosta/costa-financial-assistant/backend/internals/payments/interfaces"
)

// ─── Input / Output types ─────────────────────────────────────────────────────

// ReconciliationSummaryInput carries optional date filters for the summary endpoint.
type ReconciliationSummaryInput struct {
	PeriodStart string `query:"periodStart" doc:"ISO-8601 date (YYYY-MM-DD) — start of reconciliation window"`
	PeriodEnd   string `query:"periodEnd" doc:"ISO-8601 date (YYYY-MM-DD) — end of reconciliation window"`
}

// ReconciliationEntryResponse is a single row in the reconciliation summary.
type ReconciliationEntryResponse struct {
	TransactionLineID    string  `json:"transactionLineId"`
	TransactionDate      string  `json:"transactionDate"`
	Description          string  `json:"description"`
	Amount               string  `json:"amount"`
	Direction            string  `json:"direction"`
	ReconciliationStatus string  `json:"reconciliationStatus"`
	LinkedBillID         *string `json:"linkedBillId,omitempty"`
	LinkedBillDueDate    *string `json:"linkedBillDueDate,omitempty"`
	LinkedBillAmount     *string `json:"linkedBillAmount,omitempty"`
	LinkType             *string `json:"linkType,omitempty"`
}

// ReconciliationSummaryResponse is the body for the GET /reconciliation/summary endpoint.
type ReconciliationSummaryResponse struct {
	ProjectID   string                         `json:"projectId"`
	PeriodStart string                         `json:"periodStart,omitempty"`
	PeriodEnd   string                         `json:"periodEnd,omitempty"`
	Entries     []ReconciliationEntryResponse  `json:"entries"`
}

// CreateReconciliationLinkInput carries the body for POST /reconciliation/links.
type CreateReconciliationLinkInput struct {
	Body struct {
		TransactionLineID string `json:"transactionLineId" doc:"UUID of the transaction line to link"`
		BillRecordID      string `json:"billRecordId" doc:"UUID of the bill record to link"`
	}
}

// ReconciliationLinkResponse is returned on successful link creation.
type ReconciliationLinkResponse struct {
	ID                string  `json:"id"`
	ProjectID         string  `json:"projectId"`
	TransactionLineID string  `json:"transactionLineId"`
	BillRecordID      string  `json:"billRecordId"`
	LinkType          string  `json:"linkType"`
	LinkedBy          *string `json:"linkedBy,omitempty"`
	CreatedAt         string  `json:"createdAt"`
}

// ─── Controller ───────────────────────────────────────────────────────────────

// ReconciliationController handles BFF reconciliation HTTP endpoints.
type ReconciliationController struct {
	BaseController
	reconSvc paymentsinterfaces.ReconciliationService
}

// NewReconciliationController constructs a ReconciliationController.
func NewReconciliationController(
	logger *zap.Logger,
	reconSvc paymentsinterfaces.ReconciliationService,
) *ReconciliationController {
	return &ReconciliationController{BaseController: BaseController{logger: logger}, reconSvc: reconSvc}
}

// ─── Handlers ─────────────────────────────────────────────────────────────────

// HandleGetSummary returns the reconciliation summary for the project.
func (c *ReconciliationController) HandleGetSummary(ctx context.Context, input *ReconciliationSummaryInput) (*struct{ Body ReconciliationSummaryResponse }, error) {
	claims := bffmiddleware.ClaimsFromContext(ctx)
	if claims == nil {
		return nil, huma.Error401Unauthorized("missing authentication claims")
	}

	summary, err := c.reconSvc.GetSummary(ctx, claims.GetProjectId(), input.PeriodStart, input.PeriodEnd)
	if err != nil {
		c.logger.Error("reconciliation_ctrl: get summary failed",
			zap.String("project_id", claims.GetProjectId()),
			zap.Error(err))
		return nil, huma.Error500InternalServerError("failed to retrieve reconciliation summary")
	}

	entries := make([]ReconciliationEntryResponse, 0, len(summary.Entries))
	for _, e := range summary.Entries {
		entry := ReconciliationEntryResponse{
			TransactionLineID:    e.TransactionLineID,
			TransactionDate:      e.TransactionDate,
			Description:          e.Description,
			Amount:               e.Amount,
			Direction:            e.Direction,
			ReconciliationStatus: string(e.ReconciliationStatus),
			LinkedBillID:         e.LinkedBillID,
			LinkedBillDueDate:    e.LinkedBillDueDate,
			LinkedBillAmount:     e.LinkedBillAmount,
		}
		if e.LinkType != nil {
			lt := string(*e.LinkType)
			entry.LinkType = &lt
		}
		entries = append(entries, entry)
	}

	return &struct{ Body ReconciliationSummaryResponse }{
		Body: ReconciliationSummaryResponse{
			ProjectID:   summary.ProjectID,
			PeriodStart: summary.PeriodStart,
			PeriodEnd:   summary.PeriodEnd,
			Entries:     entries,
		},
	}, nil
}

// HandleCreateLink manually links a statement transaction to a bill record.
func (c *ReconciliationController) HandleCreateLink(ctx context.Context, input *CreateReconciliationLinkInput) (*struct{ Body ReconciliationLinkResponse }, error) {
	claims := bffmiddleware.ClaimsFromContext(ctx)
	if claims == nil {
		return nil, huma.Error401Unauthorized("missing authentication claims")
	}

	if input.Body.TransactionLineID == "" || input.Body.BillRecordID == "" {
		return nil, huma.Error400BadRequest("transactionLineId and billRecordId are required")
	}

	link, err := c.reconSvc.CreateManualLink(
		ctx,
		claims.GetProjectId(),
		input.Body.TransactionLineID,
		input.Body.BillRecordID,
		claims.GetSubject(),
	)
	if err != nil {
		c.logger.Error("reconciliation_ctrl: create manual link failed",
			zap.String("project_id", claims.GetProjectId()),
			zap.String("transaction_line_id", input.Body.TransactionLineID),
			zap.String("bill_record_id", input.Body.BillRecordID),
			zap.Error(err))
		return nil, huma.Error500InternalServerError("failed to create reconciliation link")
	}

	resp := ReconciliationLinkResponse{
		ID:                link.ID,
		ProjectID:         link.ProjectID,
		TransactionLineID: link.TransactionLineID,
		BillRecordID:      link.BillRecordID,
		LinkType:          string(link.LinkType),
		LinkedBy:          link.LinkedBy,
		CreatedAt:         link.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}

	return &struct{ Body ReconciliationLinkResponse }{Body: resp}, nil
}
