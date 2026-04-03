package controllers

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"

	bffinterfaces "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/interfaces"
	controllermappers "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/transport/http/controllers/mappers"
	bffmiddleware "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/transport/http/middleware"
	views "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/transport/http/views"
)

// ReconciliationController handles BFF reconciliation HTTP endpoints.
// It is a pure HTTP adapter: it extracts claims, delegates to ReconciliationService, and returns view types.
type ReconciliationController struct {
	BaseController
	svc bffinterfaces.ReconciliationService
}

// NewReconciliationController constructs a ReconciliationController.
func NewReconciliationController(logger *zap.Logger, validate *validator.Validate, svc bffinterfaces.ReconciliationService) *ReconciliationController {
	return &ReconciliationController{BaseController: BaseController{logger: logger, validate: validate}, svc: svc}
}

// HandleGetSummary returns the reconciliation summary for the project.
func (c *ReconciliationController) HandleGetSummary(ctx context.Context, input *views.ReconciliationSummaryInput) (*struct{ Body views.ReconciliationSummaryResponse }, error) {
	claims := bffmiddleware.ClaimsFromContext(ctx)
	if claims == nil {
		return nil, huma.Error401Unauthorized("missing authentication claims")
	}

	periodStart, periodEnd := controllermappers.ToReconciliationSummaryRequest(input)
	resp, err := c.svc.GetSummary(ctx, claims.GetProjectId(), periodStart, periodEnd)
	if err != nil {
		c.logger.Error("reconciliation_ctrl: get summary failed",
			zap.String("project_id", claims.GetProjectId()),
			zap.Error(err))
		return nil, huma.Error500InternalServerError("failed to retrieve reconciliation summary")
	}

	return &struct{ Body views.ReconciliationSummaryResponse }{Body: controllermappers.ToReconciliationSummaryResponse(resp)}, nil
}

// HandleCreateLink manually links a statement transaction to a bill record.
func (c *ReconciliationController) HandleCreateLink(ctx context.Context, input *views.CreateReconciliationLinkInput) (*struct{ Body views.ReconciliationLinkResponse }, error) {
	claims := bffmiddleware.ClaimsFromContext(ctx)
	if claims == nil {
		return nil, huma.Error401Unauthorized("missing authentication claims")
	}

	transactionLineID, billRecordID := controllermappers.ToCreateReconciliationLinkRequest(input)
	if transactionLineID == "" || billRecordID == "" {
		return nil, huma.Error400BadRequest("transactionLineId and billRecordId are required")
	}

	resp, err := c.svc.CreateManualLink(ctx, claims.GetProjectId(), transactionLineID, billRecordID, claims.GetSubject())
	if err != nil {
		c.logger.Error("reconciliation_ctrl: create manual link failed",
			zap.String("project_id", claims.GetProjectId()),
			zap.String("transaction_line_id", transactionLineID),
			zap.String("bill_record_id", billRecordID),
			zap.Error(err))
		return nil, huma.Error500InternalServerError("failed to create reconciliation link")
	}

	return &struct{ Body views.ReconciliationLinkResponse }{Body: controllermappers.ToReconciliationLinkResponse(resp)}, nil
}

