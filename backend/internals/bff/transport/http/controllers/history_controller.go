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

// historyQueryInput carries the optional look-back window for all history endpoints.
type historyQueryInput struct {
	Months int `query:"months" doc:"Number of calendar months to look back; 0 = all history. Default: 12" minimum:"0"`
}

// monthlyTimelineEntryResponse is a single row of the expenditure timeline.
type monthlyTimelineEntryResponse struct {
	Month       string `json:"month"`
	TotalAmount string `json:"totalAmount"`
	BillCount   int    `json:"billCount"`
}

// timelineResponse is the body for GET /history/timeline.
type timelineResponse struct {
	ProjectID string                         `json:"projectId"`
	Months    int                            `json:"months"`
	Timeline  []monthlyTimelineEntryResponse `json:"timeline"`
}

// categoryBreakdownEntryResponse is a single row of the category breakdown.
type categoryBreakdownEntryResponse struct {
	Month        string `json:"month"`
	BillTypeName string `json:"billTypeName"`
	TotalAmount  string `json:"totalAmount"`
	BillCount    int    `json:"billCount"`
}

// categoryBreakdownResponse is the body for GET /history/categories.
type categoryBreakdownResponse struct {
	ProjectID  string                            `json:"projectId"`
	Months     int                               `json:"months"`
	Categories []categoryBreakdownEntryResponse  `json:"categories"`
}

// monthlyComplianceEntryResponse is a single row of the compliance metrics.
type monthlyComplianceEntryResponse struct {
	Month          string `json:"month"`
	TotalBills     int    `json:"totalBills"`
	PaidOnTime     int    `json:"paidOnTime"`
	Overdue        int    `json:"overdue"`
	ComplianceRate string `json:"complianceRate"`
}

// complianceResponse is the body for GET /history/compliance.
type complianceResponse struct {
	ProjectID  string                           `json:"projectId"`
	Months     int                              `json:"months"`
	Compliance []monthlyComplianceEntryResponse `json:"compliance"`
}

// ─── Controller ───────────────────────────────────────────────────────────────

// HistoryController registers and handles the financial history analytics endpoints.
type HistoryController struct {
	logger     *zap.Logger
	historyRepo paymentsinterfaces.HistoryRepository
}

// NewHistoryController constructs a HistoryController.
func NewHistoryController(
	logger *zap.Logger,
	historyRepo paymentsinterfaces.HistoryRepository,
) *HistoryController {
	return &HistoryController{logger: logger, historyRepo: historyRepo}
}

// Register mounts all history analytics routes on the Huma API.
func (c *HistoryController) Register(api huma.API, auth func(huma.Context, func(huma.Context))) {
	c.registerTimeline(api, auth)
	c.registerCategories(api, auth)
	c.registerCompliance(api, auth)
}

// defaultMonths returns 12 when months=0 is not an explicit "all history" signal.
// Tasks require that 0 = all history (passed through to repository).
func defaultMonthsParam(m int) int {
	if m < 0 {
		return 12
	}
	return m
}

// registerTimeline handles GET /api/v1/history/timeline.
func (c *HistoryController) registerTimeline(api huma.API, auth func(huma.Context, func(huma.Context))) {
	type input struct {
		historyQueryInput
	}

	huma.Register(api, huma.Operation{
		OperationID: "get-history-timeline",
		Summary:     "Monthly expenditure timeline",
		Description: "Returns aggregated bill amounts per calendar month for the authenticated project.",
		Tags:        []string{"History"},
		Method:      http.MethodGet,
		Path:        "/api/v1/history/timeline",
		Middlewares: huma.Middlewares{auth},
	}, func(ctx context.Context, in *input) (*struct{ Body timelineResponse }, error) {
		claims := bffmiddleware.ClaimsFromContext(ctx)
		if claims == nil {
			return nil, huma.Error401Unauthorized("missing authentication")
		}

		months := defaultMonthsParam(in.Months)
		entries, err := c.historyRepo.GetTimeline(ctx, claims.ProjectId, months)
		if err != nil {
			c.logger.Error("history: get timeline failed",
				zap.String("project_id", claims.ProjectId),
				zap.Error(err),
			)
			return nil, huma.Error500InternalServerError("failed to load timeline")
		}

		rows := make([]monthlyTimelineEntryResponse, 0, len(entries))
		for _, e := range entries {
			rows = append(rows, monthlyTimelineEntryResponse{
				Month:       e.Month,
				TotalAmount: e.TotalAmount,
				BillCount:   e.BillCount,
			})
		}

		return &struct{ Body timelineResponse }{Body: timelineResponse{
			ProjectID: claims.ProjectId,
			Months:    months,
			Timeline:  rows,
		}}, nil
	})
}

// registerCategories handles GET /api/v1/history/categories.
func (c *HistoryController) registerCategories(api huma.API, auth func(huma.Context, func(huma.Context))) {
	type input struct {
		historyQueryInput
	}

	huma.Register(api, huma.Operation{
		OperationID: "get-history-categories",
		Summary:     "Per-category monthly breakdown",
		Description: "Returns bill amounts grouped by bill type and calendar month for the authenticated project.",
		Tags:        []string{"History"},
		Method:      http.MethodGet,
		Path:        "/api/v1/history/categories",
		Middlewares: huma.Middlewares{auth},
	}, func(ctx context.Context, in *input) (*struct{ Body categoryBreakdownResponse }, error) {
		claims := bffmiddleware.ClaimsFromContext(ctx)
		if claims == nil {
			return nil, huma.Error401Unauthorized("missing authentication")
		}

		months := defaultMonthsParam(in.Months)
		entries, err := c.historyRepo.GetCategoryBreakdown(ctx, claims.ProjectId, months)
		if err != nil {
			c.logger.Error("history: get category breakdown failed",
				zap.String("project_id", claims.ProjectId),
				zap.Error(err),
			)
			return nil, huma.Error500InternalServerError("failed to load category breakdown")
		}

		rows := make([]categoryBreakdownEntryResponse, 0, len(entries))
		for _, e := range entries {
			rows = append(rows, categoryBreakdownEntryResponse{
				Month:        e.Month,
				BillTypeName: e.BillTypeName,
				TotalAmount:  e.TotalAmount,
				BillCount:    e.BillCount,
			})
		}

		return &struct{ Body categoryBreakdownResponse }{Body: categoryBreakdownResponse{
			ProjectID:  claims.ProjectId,
			Months:     months,
			Categories: rows,
		}}, nil
	})
}

// registerCompliance handles GET /api/v1/history/compliance.
func (c *HistoryController) registerCompliance(api huma.API, auth func(huma.Context, func(huma.Context))) {
	type input struct {
		historyQueryInput
	}

	huma.Register(api, huma.Operation{
		OperationID: "get-history-compliance",
		Summary:     "Monthly payment compliance metrics",
		Description: "Returns on-time vs overdue bill counts and compliance rate per calendar month.",
		Tags:        []string{"History"},
		Method:      http.MethodGet,
		Path:        "/api/v1/history/compliance",
		Middlewares: huma.Middlewares{auth},
	}, func(ctx context.Context, in *input) (*struct{ Body complianceResponse }, error) {
		claims := bffmiddleware.ClaimsFromContext(ctx)
		if claims == nil {
			return nil, huma.Error401Unauthorized("missing authentication")
		}

		months := defaultMonthsParam(in.Months)
		entries, err := c.historyRepo.GetComplianceMetrics(ctx, claims.ProjectId, months)
		if err != nil {
			c.logger.Error("history: get compliance failed",
				zap.String("project_id", claims.ProjectId),
				zap.Error(err),
			)
			return nil, huma.Error500InternalServerError("failed to load compliance metrics")
		}

		rows := make([]monthlyComplianceEntryResponse, 0, len(entries))
		for _, e := range entries {
			rows = append(rows, monthlyComplianceEntryResponse{
				Month:          e.Month,
				TotalBills:     e.TotalBills,
				PaidOnTime:     e.PaidOnTime,
				Overdue:        e.Overdue,
				ComplianceRate: e.ComplianceRate,
			})
		}

		return &struct{ Body complianceResponse }{Body: complianceResponse{
			ProjectID:  claims.ProjectId,
			Months:     months,
			Compliance: rows,
		}}, nil
	})
}
