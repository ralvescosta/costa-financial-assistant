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

// HistoryController handles the financial history analytics HTTP endpoints.
// It is a pure HTTP adapter: it extracts claims, delegates to HistoryService, and returns view types.
type HistoryController struct {
	BaseController
	svc bffinterfaces.HistoryService
}

// NewHistoryController constructs a HistoryController.
func NewHistoryController(logger *zap.Logger, validate *validator.Validate, svc bffinterfaces.HistoryService) *HistoryController {
	return &HistoryController{BaseController: BaseController{logger: logger, validate: validate}, svc: svc}
}

// HandleGetTimeline returns aggregated bill amounts per calendar month.
func (c *HistoryController) HandleGetTimeline(ctx context.Context, in *views.HistoryQueryInput) (*struct{ Body views.TimelineResponse }, error) {
	claims := bffmiddleware.ClaimsFromContext(ctx)
	if claims == nil {
		return nil, huma.Error401Unauthorized("missing authentication")
	}

	months := controllermappers.ToHistoryMonths(in)
	resp, err := c.svc.GetTimeline(ctx, claims.GetProjectId(), months)
	if err != nil {
		c.logger.Error("history: get timeline failed",
			zap.String("project_id", claims.GetProjectId()),
			zap.Error(err))
		return nil, huma.Error500InternalServerError("failed to load timeline")
	}

	return &struct{ Body views.TimelineResponse }{Body: controllermappers.ToTimelineResponse(resp)}, nil
}

// HandleGetCategories returns bill amounts grouped by bill type and calendar month.
func (c *HistoryController) HandleGetCategories(ctx context.Context, in *views.HistoryQueryInput) (*struct {
	Body views.CategoryBreakdownResponse
}, error) {
	claims := bffmiddleware.ClaimsFromContext(ctx)
	if claims == nil {
		return nil, huma.Error401Unauthorized("missing authentication")
	}

	months := controllermappers.ToHistoryMonths(in)
	resp, err := c.svc.GetCategoryBreakdown(ctx, claims.GetProjectId(), months)
	if err != nil {
		c.logger.Error("history: get category breakdown failed",
			zap.String("project_id", claims.GetProjectId()),
			zap.Error(err))
		return nil, huma.Error500InternalServerError("failed to load category breakdown")
	}

	return &struct {
		Body views.CategoryBreakdownResponse
	}{Body: controllermappers.ToCategoryBreakdownResponse(resp)}, nil
}

// HandleGetCompliance returns on-time vs overdue bill counts and compliance rate.
func (c *HistoryController) HandleGetCompliance(ctx context.Context, in *views.HistoryQueryInput) (*struct{ Body views.ComplianceResponse }, error) {
	claims := bffmiddleware.ClaimsFromContext(ctx)
	if claims == nil {
		return nil, huma.Error401Unauthorized("missing authentication")
	}

	months := controllermappers.ToHistoryMonths(in)
	resp, err := c.svc.GetComplianceMetrics(ctx, claims.GetProjectId(), months)
	if err != nil {
		c.logger.Error("history: get compliance failed",
			zap.String("project_id", claims.GetProjectId()),
			zap.Error(err))
		return nil, huma.Error500InternalServerError("failed to load compliance metrics")
	}

	return &struct{ Body views.ComplianceResponse }{Body: controllermappers.ToComplianceResponse(resp)}, nil
}
