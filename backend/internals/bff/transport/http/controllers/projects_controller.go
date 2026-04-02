package controllers

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"

	bffinterfaces "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/interfaces"
	bffmiddleware "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/transport/http/middleware"
	views "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/transport/http/views"
)

// ProjectsController handles BFF project collaboration HTTP endpoints.
// It is a pure HTTP adapter: it extracts claims, delegates to ProjectsService, and returns view types.
type ProjectsController struct {
	BaseController
	svc bffinterfaces.ProjectsService
}

// NewProjectsController constructs a ProjectsController.
func NewProjectsController(logger *zap.Logger, validate *validator.Validate, svc bffinterfaces.ProjectsService) *ProjectsController {
	return &ProjectsController{BaseController: BaseController{logger: logger, validate: validate}, svc: svc}
}

// HandleGetCurrent returns the project for the caller's JWT project_id.
func (c *ProjectsController) HandleGetCurrent(ctx context.Context, _ *struct{}) (*struct{ Body views.ProjectResponse }, error) {
	claims := bffmiddleware.ClaimsFromContext(ctx)
	if claims == nil {
		return nil, huma.Error403Forbidden("missing project context")
	}

	resp, err := c.svc.GetCurrentProject(ctx, claims.GetProjectId(), claims.GetSubject(), claims.GetRole())
	if err != nil {
		return nil, c.grpcToHumaError(err, "get current project")
	}

	return &struct{ Body views.ProjectResponse }{Body: *resp}, nil
}

// HandleListMembers returns all members for the caller's project.
func (c *ProjectsController) HandleListMembers(ctx context.Context, input *views.ListMembersInput) (*struct{ Body views.ListMembersResponse }, error) {
	claims := bffmiddleware.ClaimsFromContext(ctx)
	if claims == nil {
		return nil, huma.Error403Forbidden("missing project context")
	}

	pageSize := input.PageSize
	if pageSize == 0 {
		pageSize = 25
	}

	resp, err := c.svc.ListMembers(ctx, claims.GetProjectId(), claims.GetSubject(), claims.GetRole(), pageSize, input.PageToken)
	if err != nil {
		return nil, c.grpcToHumaError(err, "list members")
	}

	return &struct{ Body views.ListMembersResponse }{Body: *resp}, nil
}

// HandleInvite adds a user by email to the project.
func (c *ProjectsController) HandleInvite(ctx context.Context, input *views.InviteMemberInput) (*struct{ Body views.ProjectMemberResponse }, error) {
	claims := bffmiddleware.ClaimsFromContext(ctx)
	if claims == nil {
		return nil, huma.Error403Forbidden("missing project context")
	}

	resp, err := c.svc.InviteMember(ctx, claims.GetProjectId(), claims.GetSubject(), claims.GetRole(), input.Body.Email, input.Body.Role)
	if err != nil {
		return nil, c.grpcToHumaError(err, "invite member")
	}

	return &struct{ Body views.ProjectMemberResponse }{Body: *resp}, nil
}

// HandleUpdateRole changes the role of an existing member.
func (c *ProjectsController) HandleUpdateRole(ctx context.Context, input *views.UpdateMemberRoleInput) (*struct{ Body views.ProjectMemberResponse }, error) {
	claims := bffmiddleware.ClaimsFromContext(ctx)
	if claims == nil {
		return nil, huma.Error403Forbidden("missing project context")
	}

	resp, err := c.svc.UpdateMemberRole(ctx, claims.GetProjectId(), claims.GetSubject(), claims.GetRole(), input.MemberID, input.Body.Role)
	if err != nil {
		return nil, c.grpcToHumaError(err, "update member role")
	}

	return &struct{ Body views.ProjectMemberResponse }{Body: *resp}, nil
}
