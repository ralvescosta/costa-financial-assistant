package controllers

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"go.uber.org/zap"

	bffmiddleware "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/transport/http/middleware"
	paymentsinterfaces "github.com/ralvescosta/costa-financial-assistant/backend/internals/payments/interfaces"
)

// ─── Input / Output types ─────────────────────────────────────────────────────

// reconciliationSummaryInput carries optional date filters for the summary endpoint.
type reconciliationSummaryInput struct {
	PeriodStart string `query:"periodStart" doc:"ISO-8601 date (YYYY-MM-DD) — start of reconciliation window"`
	PeriodEnd   string `query:"periodEnd" doc:"ISO-8601 date (YYYY-MM-DD) — end of reconciliation window"`
}

// reconciliationEntryResponse is a single row in the reconciliation summary.
type reconciliationEntryResponse struct {
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

// reconciliationSummaryResponse is the body for the GET /reconciliation/summary endpoint.
type reconciliationSummaryResponse struct {
	ProjectID   string                        `json:"projectId"`
	PeriodStart string                        `json:"periodStart,omitempty"`
	PeriodEnd   string                        `json:"periodEnd,omitempty"`
	Entries     []reconciliationEntryResponse `json:"entries"`
}

// createReconciliationLinkInput carries the body for POST /reconciliation/links.
type createReconciliationLinkInput struct {
	Body struct {
		TransactionLineID string `json:"transactionLineId" doc:"UUID of the transaction line to link"`
		BillRecordID      string `json:"billRecordId" doc:"UUID of the bill record to link"`
	}
}

// reconciliationLinkResponse is returned on successful link creation.
type reconciliationLinkResponse struct {
	ID                string  `json:"id"`
	ProjectID         string  `json:"projectId"`
	TransactionLineID string  `json:"transactionLineId"`
	BillRecordID      string  `json:"billRecordId"`
	LinkType          string  `json:"linkType"`
	LinkedBy          *string `json:"linkedBy,omitempty"`
	CreatedAt         string  `json:"createdAt"`
}

// ─── Controller ───────────────────────────────────────────────────────────────

// ReconciliationController registers and handles all reconciliation HTTP routes.
type ReconciliationController struct {
	logger  *zap.Logger
	reconSvc paymentsinterfaces.ReconciliationService
}

// NewReconciliationController constructs a ReconciliationController.
func NewReconciliationController(
	logger *zap.Logger,
	reconSvc paymentsinterfaces.ReconciliationService,
) *ReconciliationController {
	return &ReconciliationController{logger: logger, reconSvc: reconSvc}
}

// Register mounts all reconciliation routes on the Huma API.
func (c *ReconciliationController) Register(
	api huma.API,
	auth func(huma.Context, func(huma.Context)),
) {
	huma.Register(api, huma.Operation{
		OperationID: "get-reconciliation-summary",
		Method:      http.MethodGet,
		Path:        "/api/v1/reconciliation/summary",
		Summary:     "Get reconciliation summary",
		Description: "Returns the transaction-to-bill reconciliation status for the project, optionally filtered by date range.",
		Tags:        []string{"Reconciliation"},
		Middlewares: huma.Middlewares{auth, bffmiddleware.NewProjectGuard("read_only", c.logger)},
	}, func(ctx context.Context, input *reconciliationSummaryInput) (*struct{ Body reconciliationSummaryResponse }, error) {
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

		entries := make([]reconciliationEntryResponse, 0, len(summary.Entries))
		for _, e := range summary.Entries {
			entry := reconciliationEntryResponse{
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

		return &struct{ Body reconciliationSummaryResponse }{
			Body: reconciliationSummaryResponse{
				ProjectID:   summary.ProjectID,
				PeriodStart: summary.PeriodStart,
				PeriodEnd:   summary.PeriodEnd,
				Entries:     entries,
			},
		}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "create-reconciliation-link",
		Method:      http.MethodPost,
		Path:        "/api/v1/reconciliation/links",
		Summary:     "Create manual reconciliation link",
		Description: "Manually links a statement transaction to a bill record as a user-confirmed match.",
		Tags:        []string{"Reconciliation"},
		Middlewares: huma.Middlewares{auth, bffmiddleware.NewProjectGuard("update", c.logger)},
	}, func(ctx context.Context, input *createReconciliationLinkInput) (*struct{ Body reconciliationLinkResponse }, error) {
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

		resp := reconciliationLinkResponse{
			ID:                link.ID,
			ProjectID:         link.ProjectID,
			TransactionLineID: link.TransactionLineID,
			BillRecordID:      link.BillRecordID,
			LinkType:          string(link.LinkType),
			LinkedBy:          link.LinkedBy,
			CreatedAt:         link.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}

		return &struct{ Body reconciliationLinkResponse }{Body: resp}, nil
	})
}
