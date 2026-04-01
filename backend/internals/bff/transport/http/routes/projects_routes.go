package routes

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"go.uber.org/zap"

	controllers "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/transport/http/controllers"
	bffmiddleware "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/transport/http/middleware"
)

// Compile-time assertion: *ProjectsController satisfies ProjectsCapability.
var _ ProjectsCapability = (*controllers.ProjectsController)(nil)

// ProjectsRoute owns all Huma operation registrations for the projects resource.
type ProjectsRoute struct {
	ctrl   ProjectsCapability
	logger *zap.Logger
}

// NewProjectsRoute constructs a ProjectsRoute.
func NewProjectsRoute(ctrl ProjectsCapability, logger *zap.Logger) *ProjectsRoute {
	return &ProjectsRoute{ctrl: ctrl, logger: logger}
}

// Register wires all project routes to the Huma API.
func (r *ProjectsRoute) Register(api huma.API, auth func(huma.Context, func(huma.Context))) {
	huma.Register(api, huma.Operation{
		OperationID: "get-current-project",
		Method:      http.MethodGet,
		Path:        "/api/v1/projects/current",
		Summary:     "Get caller's active project context",
		Description: "Returns the project that the authenticated user is currently operating under.",
		Tags:        []string{"projects"},
		Middlewares: huma.Middlewares{auth, bffmiddleware.NewProjectGuard("read_only", r.logger)},
	}, r.ctrl.HandleGetCurrent)

	huma.Register(api, huma.Operation{
		OperationID: "list-project-members",
		Method:      http.MethodGet,
		Path:        "/api/v1/projects/members",
		Summary:     "List members of the active project",
		Description: "Returns all members and their roles for the caller's active project.",
		Tags:        []string{"projects"},
		Middlewares: huma.Middlewares{auth, bffmiddleware.NewProjectGuard("read_only", r.logger)},
	}, r.ctrl.HandleListMembers)

	huma.Register(api, huma.Operation{
		OperationID: "invite-project-member",
		Method:      http.MethodPost,
		Path:        "/api/v1/projects/members/invite",
		Summary:     "Invite a new member to the active project",
		Description: "Sends an invitation to the provided email address with the specified role.",
		Tags:        []string{"projects"},
		Middlewares: huma.Middlewares{auth, bffmiddleware.NewProjectGuard("write", r.logger)},
	}, r.ctrl.HandleInvite)

	huma.Register(api, huma.Operation{
		OperationID: "update-project-member-role",
		Method:      http.MethodPatch,
		Path:        "/api/v1/projects/members/{memberId}/role",
		Summary:     "Update the role of a project member",
		Description: "Changes the role of an existing project member within the active project.",
		Tags:        []string{"projects"},
		Middlewares: huma.Middlewares{auth, bffmiddleware.NewProjectGuard("write", r.logger)},
	}, r.ctrl.HandleUpdateRole)
}
