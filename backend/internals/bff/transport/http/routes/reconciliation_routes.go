package routes

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"go.uber.org/zap"

	controllers "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/transport/http/controllers"
	bffmiddleware "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/transport/http/middleware"
)

// Compile-time assertion: *ReconciliationController satisfies ReconciliationCapability.
var _ ReconciliationCapability = (*controllers.ReconciliationController)(nil)

// ReconciliationRoute owns all Huma operation registrations for the reconciliation resource.
type ReconciliationRoute struct {
	ctrl   ReconciliationCapability
	logger *zap.Logger
}

// NewReconciliationRoute constructs a ReconciliationRoute.
func NewReconciliationRoute(ctrl ReconciliationCapability, logger *zap.Logger) *ReconciliationRoute {
	return &ReconciliationRoute{ctrl: ctrl, logger: logger}
}

// Register wires all reconciliation routes to the Huma API.
func (r *ReconciliationRoute) Register(api huma.API, auth func(huma.Context, func(huma.Context))) {
	huma.Register(api, huma.Operation{
		OperationID: "get-reconciliation-summary",
		Method:      http.MethodGet,
		Path:        "/api/v1/reconciliation/summary",
		Summary:     "Return matched/unmatched entries for a billing cycle",
		Description: "Computes a reconciliation summary of bill vs. statement entries for the specified cycle.",
		Tags:        []string{"Reconciliation"},
		Middlewares: huma.Middlewares{auth, bffmiddleware.NewProjectGuard("read_only", r.logger)},
	}, r.ctrl.HandleGetSummary)

	huma.Register(api, huma.Operation{
		OperationID: "create-reconciliation-link",
		Method:      http.MethodPost,
		Path:        "/api/v1/reconciliation/links",
		Summary:     "Link a bill record to a statement transaction line",
		Description: "Creates a manual reconciliation link between a bill record and a statement transaction.",
		Tags:        []string{"Reconciliation"},
		Middlewares: huma.Middlewares{auth, bffmiddleware.NewProjectGuard("update", r.logger)},
	}, r.ctrl.HandleCreateLink)
}
