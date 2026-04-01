package routes

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"go.uber.org/zap"

	controllers "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/transport/http/controllers"
	bffmiddleware "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/transport/http/middleware"
)

// Compile-time assertion: *SettingsController satisfies SettingsCapability.
var _ SettingsCapability = (*controllers.SettingsController)(nil)

// SettingsRoute owns all Huma operation registrations for the settings resource.
type SettingsRoute struct {
	ctrl   SettingsCapability
	logger *zap.Logger
}

// NewSettingsRoute constructs a SettingsRoute.
func NewSettingsRoute(ctrl SettingsCapability, logger *zap.Logger) *SettingsRoute {
	return &SettingsRoute{ctrl: ctrl, logger: logger}
}

// Register wires all settings routes to the Huma API.
func (r *SettingsRoute) Register(api huma.API, auth func(huma.Context, func(huma.Context))) {
	huma.Register(api, huma.Operation{
		OperationID: "list-bank-accounts",
		Method:      http.MethodGet,
		Path:        "/api/v1/bank-accounts",
		Summary:     "List bank accounts linked to the active project",
		Description: "Returns all bank accounts configured for the caller's active project.",
		Tags:        []string{"settings"},
		Middlewares: huma.Middlewares{auth, bffmiddleware.NewProjectGuard("read_only", r.logger)},
	}, r.ctrl.HandleList)

	huma.Register(api, huma.Operation{
		OperationID: "create-bank-account",
		Method:      http.MethodPost,
		Path:        "/api/v1/bank-accounts",
		Summary:     "Register a new bank account for the active project",
		Description: "Creates a bank account record scoped to the caller's active project.",
		Tags:        []string{"settings"},
		Middlewares: huma.Middlewares{auth, bffmiddleware.NewProjectGuard("update", r.logger)},
	}, r.ctrl.HandleCreate)

	huma.Register(api, huma.Operation{
		OperationID: "delete-bank-account",
		Method:      http.MethodDelete,
		Path:        "/api/v1/bank-accounts/{bankAccountId}",
		Summary:     "Remove a bank account from the active project",
		Description: "Deletes the specified bank account scoped to the caller's active project.",
		Tags:        []string{"settings"},
		Middlewares: huma.Middlewares{auth, bffmiddleware.NewProjectGuard("update", r.logger)},
	}, r.ctrl.HandleDelete)
}
