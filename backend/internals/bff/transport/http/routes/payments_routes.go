package routes

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"go.uber.org/zap"

	controllers "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/transport/http/controllers"
	bffmiddleware "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/transport/http/middleware"
)

// Compile-time assertion: *PaymentsController satisfies PaymentsCapability.
var _ PaymentsCapability = (*controllers.PaymentsController)(nil)

// PaymentsRoute owns all Huma operation registrations for the payments and payment-cycle resources.
type PaymentsRoute struct {
	ctrl   PaymentsCapability
	logger *zap.Logger
}

// NewPaymentsRoute constructs a PaymentsRoute.
func NewPaymentsRoute(ctrl PaymentsCapability, logger *zap.Logger) *PaymentsRoute {
	return &PaymentsRoute{ctrl: ctrl, logger: logger}
}

// Register wires all payment routes to the Huma API.
func (r *PaymentsRoute) Register(api huma.API, auth func(huma.Context, func(huma.Context))) {
	huma.Register(api, huma.Operation{
		OperationID: "get-payment-dashboard",
		Method:      http.MethodGet,
		Path:        "/api/v1/bills/payment-dashboard",
		Summary:     "Get paginated payment dashboard entries for a billing cycle",
		Description: "Returns the payment status and amounts for all bills in the specified billing cycle.",
		Tags:        []string{"payments"},
		Middlewares: huma.Middlewares{auth, bffmiddleware.NewProjectGuard("read_only", r.logger)},
	}, r.ctrl.HandleGetDashboard)

	huma.Register(api, huma.Operation{
		OperationID: "mark-bill-paid",
		Method:      http.MethodPost,
		Path:        "/api/v1/bills/{billId}/mark-paid",
		Summary:     "Mark a recurring bill as paid for the current cycle",
		Description: "Records a payment event for the specified bill in the active billing cycle.",
		Tags:        []string{"payments"},
		Middlewares: huma.Middlewares{auth, bffmiddleware.NewProjectGuard("update", r.logger)},
	}, r.ctrl.HandleMarkPaid)

	huma.Register(api, huma.Operation{
		OperationID: "get-preferred-payment-day",
		Method:      http.MethodGet,
		Path:        "/api/v1/payment-cycle/preferred-day",
		Summary:     "Retrieve the preferred monthly payment day for this project",
		Description: "Returns the day-of-month setting used to calculate billing cycle boundaries.",
		Tags:        []string{"settings"},
		Middlewares: huma.Middlewares{auth, bffmiddleware.NewProjectGuard("read_only", r.logger)},
	}, r.ctrl.HandleGetPreferredDay)

	huma.Register(api, huma.Operation{
		OperationID: "set-preferred-payment-day",
		Method:      http.MethodPut,
		Path:        "/api/v1/payment-cycle/preferred-day",
		Summary:     "Update the preferred monthly payment day for this project",
		Description: "Persists the day-of-month preference used to compute billing cycle boundaries.",
		Tags:        []string{"settings"},
		Middlewares: huma.Middlewares{auth, bffmiddleware.NewProjectGuard("update", r.logger)},
	}, r.ctrl.HandleSetPreferredDay)
}
