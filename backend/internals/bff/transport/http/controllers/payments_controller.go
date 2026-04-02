package controllers

import (
	"context"
	"strconv"

	"github.com/danielgtaylor/huma/v2"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"

	bffinterfaces "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/interfaces"
	bffmiddleware "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/transport/http/middleware"
	views "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/transport/http/views"
)

// PaymentsController handles BFF payment HTTP endpoints.
// It is a pure HTTP adapter: it extracts claims, delegates to PaymentsService, and returns view types.
type PaymentsController struct {
	BaseController
	svc bffinterfaces.PaymentsService
}

// NewPaymentsController constructs a PaymentsController.
func NewPaymentsController(logger *zap.Logger, validate *validator.Validate, svc bffinterfaces.PaymentsService) *PaymentsController {
	return &PaymentsController{BaseController: BaseController{logger: logger, validate: validate}, svc: svc}
}

// HandleGetDashboard returns outstanding bills for the project's active payment cycle.
func (c *PaymentsController) HandleGetDashboard(ctx context.Context, input *views.GetPaymentDashboardInput) (*struct{ Body views.PaymentDashboardResponse }, error) {
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

	resp, err := c.svc.GetPaymentDashboard(ctx, claims.GetProjectId(), claims.GetSubject(), input.CycleStart, input.CycleEnd, pageSize, input.PageToken)
	if err != nil {
		return nil, c.grpcToHumaError(err, "get payment dashboard failed")
	}

	return &struct{ Body views.PaymentDashboardResponse }{Body: *resp}, nil
}

// HandleMarkPaid idempotently marks a bill as paid.
func (c *PaymentsController) HandleMarkPaid(ctx context.Context, input *views.MarkBillPaidInput) (*struct{ Body views.MarkBillPaidResponse }, error) {
	claims := bffmiddleware.ClaimsFromContext(ctx)
	if claims == nil {
		return nil, huma.Error403Forbidden("missing project context")
	}

	resp, err := c.svc.MarkBillPaid(ctx, claims.GetProjectId(), input.BillID, claims.GetSubject())
	if err != nil {
		return nil, c.grpcToHumaError(err, "mark bill paid failed")
	}

	c.logger.Info("payments: bill marked paid",
		zap.String("bill_id", input.BillID),
		zap.String("project_id", claims.GetProjectId()))
	return &struct{ Body views.MarkBillPaidResponse }{Body: *resp}, nil
}

// HandleGetPreferredDay returns the project's preferred payment day.
func (c *PaymentsController) HandleGetPreferredDay(ctx context.Context, _ *struct{}) (*struct{ Body views.CyclePreferenceResponse }, error) {
	claims := bffmiddleware.ClaimsFromContext(ctx)
	if claims == nil {
		return nil, huma.Error403Forbidden("missing project context")
	}

	resp, err := c.svc.GetCyclePreference(ctx, claims.GetProjectId())
	if err != nil {
		c.logger.Error("payments: get preferred day failed", zap.Error(err))
		return nil, huma.Error500InternalServerError("get preferred day failed")
	}
	if resp == nil {
		return nil, huma.Error404NotFound("no payment cycle preference configured")
	}

	return &struct{ Body views.CyclePreferenceResponse }{Body: *resp}, nil
}

// HandleSetPreferredDay creates or updates the project's preferred payment day.
func (c *PaymentsController) HandleSetPreferredDay(ctx context.Context, input *views.SetPreferredDayInput) (*struct{ Body views.CyclePreferenceResponse }, error) {
	claims := bffmiddleware.ClaimsFromContext(ctx)
	if claims == nil {
		return nil, huma.Error403Forbidden("missing project context")
	}

	resp, err := c.svc.SetCyclePreference(ctx, claims.GetProjectId(), input.Body.PreferredDayOfMonth, claims.GetSubject())
	if err != nil {
		c.logger.Error("payments: set preferred day failed", zap.Error(err))
		return nil, huma.Error500InternalServerError("set preferred day failed")
	}

	c.logger.Info("payments: preferred day set",
		zap.String("project_id", claims.GetProjectId()),
		zap.Int("day", input.Body.PreferredDayOfMonth))
	return &struct{ Body views.CyclePreferenceResponse }{Body: *resp}, nil
}

