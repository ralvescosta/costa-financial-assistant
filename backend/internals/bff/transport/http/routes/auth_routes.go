package routes

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"go.uber.org/zap"

	controllers "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/transport/http/controllers"
)

// Compile-time assertion: *AuthController satisfies AuthCapability.
var _ AuthCapability = (*controllers.AuthController)(nil)

// AuthRoute owns all Huma operation registrations for login and refresh.
type AuthRoute struct {
	ctrl   AuthCapability
	logger *zap.Logger
}

// NewAuthRoute constructs an AuthRoute.
func NewAuthRoute(ctrl AuthCapability, logger *zap.Logger) *AuthRoute {
	return &AuthRoute{ctrl: ctrl, logger: logger}
}

// Register wires all authentication routes to the Huma API.
func (r *AuthRoute) Register(api huma.API, _ func(huma.Context, func(huma.Context))) {
	huma.Register(api, huma.Operation{
		OperationID: "login-user",
		Method:      http.MethodPost,
		Path:        "/api/auth/login",
		Summary:     "Authenticate the seeded owner account",
		Description: "Validates the submitted bootstrap credentials and starts the BFF session.",
		Tags:        []string{"auth"},
	}, r.ctrl.HandleLogin)

	huma.Register(api, huma.Operation{
		OperationID: "refresh-session",
		Method:      http.MethodPost,
		Path:        "/api/auth/refresh",
		Summary:     "Rotate the current session token",
		Description: "Validates the current cookie or bearer token and returns refreshed session metadata.",
		Tags:        []string{"auth"},
	}, r.ctrl.HandleRefresh)
}
