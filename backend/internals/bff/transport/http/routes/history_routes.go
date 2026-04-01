package routes

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"go.uber.org/zap"

	controllers "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/transport/http/controllers"
)

// Compile-time assertion: *HistoryController satisfies HistoryCapability.
var _ HistoryCapability = (*controllers.HistoryController)(nil)

// HistoryRoute owns all Huma operation registrations for the history resource.
type HistoryRoute struct {
	ctrl   HistoryCapability
	logger *zap.Logger
}

// NewHistoryRoute constructs a HistoryRoute.
func NewHistoryRoute(ctrl HistoryCapability, logger *zap.Logger) *HistoryRoute {
	return &HistoryRoute{ctrl: ctrl, logger: logger}
}

// Register wires all history routes to the Huma API.
// History routes require only authentication — no project guard is applied.
func (r *HistoryRoute) Register(api huma.API, auth func(huma.Context, func(huma.Context))) {
	huma.Register(api, huma.Operation{
		OperationID: "get-history-timeline",
		Method:      http.MethodGet,
		Path:        "/api/v1/history/timeline",
		Summary:     "Monthly paid/unpaid bill timeline",
		Description: "Returns a month-by-month breakdown of bill payment history for trend analysis.",
		Tags:        []string{"History"},
		Middlewares: huma.Middlewares{auth},
	}, r.ctrl.HandleGetTimeline)

	huma.Register(api, huma.Operation{
		OperationID: "get-history-categories",
		Method:      http.MethodGet,
		Path:        "/api/v1/history/categories",
		Summary:     "Category-level spend breakdown over a date range",
		Description: "Returns spend totals grouped by bill category for the specified date range.",
		Tags:        []string{"History"},
		Middlewares: huma.Middlewares{auth},
	}, r.ctrl.HandleGetCategories)

	huma.Register(api, huma.Operation{
		OperationID: "get-history-compliance",
		Method:      http.MethodGet,
		Path:        "/api/v1/history/compliance",
		Summary:     "Month-over-month on-time payment compliance ratio",
		Description: "Returns the percentage of bills paid on time per month over the specified range.",
		Tags:        []string{"History"},
		Middlewares: huma.Middlewares{auth},
	}, r.ctrl.HandleGetCompliance)
}
