package controllers

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"go.uber.org/zap"

	bffmiddleware "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/transport/http/middleware"
	paymentsinterfaces "github.com/ralvescosta/costa-financial-assistant/backend/internals/payments/interfaces"
)

// ─── Input / Output types ─────────────────────────────────────────────────────

// HistoryQueryInput carries the optional look-back window for all history endpoints.
type HistoryQueryInput struct {
	Months int `query:"months" doc:"Number of calendar months to look back; 0 = all history. Default: 12" minimum:"0"`
}

// MonthlyTimelineEntryResponse is a single row of the expenditure timeline.
type MonthlyTimelineEntryResponse struct {
	Month       string `json:"month"`
	TotalAmount string `json:"totalAmount"`
	BillCount   int    `json:"billCount"`
}

// TimelineResponse is the body for GET /history/timeline.
type TimelineResponse struct {
	ProjectID string                         `json:"projectId"`
	Months    int                            `json:"months"`
	Timeline  []MonthlyTimelineEntryResponse `json:"timeline"`
}

// CategoryBreakdownEntryResponse is a single row of the category breakdown.
type CategoryBreakdownEntryResponse struct {
	Month        string `json:"month"`
	BillTypeName string `json:"billTypeName"`
	TotalAmount  string `json:"totalAmount"`
	BillCount    int    `json:"billCount"`
}

// CategoryBreakdownResponse is the body for GET /history/categories.
type CategoryBreakdownResponse struct {
	ProjectID  string                            `json:"projectId"`
	Months     int                               `json:"months"`
	Categories []CategoryBreakdownEntryResponse  `json:"categories"`
}

// MonthlyComplianceEntryResponse is a single row of the compliance metrics.
type MonthlyComplianceEntryResponse struct {
	Month          string `json:"month"`
	TotalBills     int    `json:"totalBills"`
	PaidOnTime     int    `json:"paidOnTime"`
	Overdue        int    `json:"overdue"`
	ComplianceRate string `json:"complianceRate"`
}

// ComplianceResponse is the body for GET /history/compliance.
type ComplianceResponse struct {
	ProjectID  string                           `json:"projectId"`
	Months     int                              `json:"months"`
	Compliance []MonthlyComplianceEntryResponse `json:"compliance"`
}

// ─── Controller ───────────────────────────────────────────────────────────────

// HistoryController handles the financial history analytics HTTP endpoints.
type HistoryController struct {
	BaseController
	historyRepo paymentsinterfaces.HistoryRepository
}

// NewHistoryController constructs a HistoryController.
func NewHistoryController(
	logger *zap.Logger,
	historyRepo paymentsinterfaces.HistoryRepository,
) *HistoryController {
	return &HistoryController{BaseController: BaseController{logger: logger}, historyRepo: historyRepo}
}

// defaultMonthsParam returns 12 when months=0 is not an explicit "all history" signal.
// Tasks require that 0 = all history (passed through to repository).
func defaultMonthsParam(m int) int {
	if m < 0 {
		return 12
	}
	return m
}

// ─── Handlers ─────────────────────────────────────────────────────────────────

// HandleGetTimeline returns aggregated bill amounts per calendar month.
func (c *HistoryController) HandleGetTimeline(ctx context.Context, in *HistoryQueryInput) (*struct{ Body TimelineResponse }, error) {
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

	rows := make([]MonthlyTimelineEntryResponse, 0, len(entries))
	for _, e := range entries {
		rows = append(rows, MonthlyTimelineEntryResponse{
			Month:       e.Month,
			TotalAmount: e.TotalAmount,
			BillCount:   e.BillCount,
		})
	}

	return &struct{ Body TimelineResponse }{Body: TimelineResponse{
		ProjectID: claims.ProjectId,
		Months:    months,
		Timeline:  rows,
	}}, nil
}

// HandleGetCategories returns bill amounts grouped by bill type and calendar month.
func (c *HistoryController) HandleGetCategories(ctx context.Context, in *HistoryQueryInput) (*struct{ Body CategoryBreakdownResponse }, error) {
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

	rows := make([]CategoryBreakdownEntryResponse, 0, len(entries))
	for _, e := range entries {
		rows = append(rows, CategoryBreakdownEntryResponse{
			Month:        e.Month,
			BillTypeName: e.BillTypeName,
			TotalAmount:  e.TotalAmount,
			BillCount:    e.BillCount,
		})
	}

	return &struct{ Body CategoryBreakdownResponse }{Body: CategoryBreakdownResponse{
		ProjectID:  claims.ProjectId,
		Months:     months,
		Categories: rows,
	}}, nil
}

// HandleGetCompliance returns on-time vs overdue bill counts and compliance rate.
func (c *HistoryController) HandleGetCompliance(ctx context.Context, in *HistoryQueryInput) (*struct{ Body ComplianceResponse }, error) {
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

	rows := make([]MonthlyComplianceEntryResponse, 0, len(entries))
	for _, e := range entries {
		rows = append(rows, MonthlyComplianceEntryResponse{
			Month:          e.Month,
			TotalBills:     e.TotalBills,
			PaidOnTime:     e.PaidOnTime,
			Overdue:        e.Overdue,
			ComplianceRate: e.ComplianceRate,
		})
	}

	return &struct{ Body ComplianceResponse }{Body: ComplianceResponse{
		ProjectID:  claims.ProjectId,
		Months:     months,
		Compliance: rows,
	}}, nil
}
