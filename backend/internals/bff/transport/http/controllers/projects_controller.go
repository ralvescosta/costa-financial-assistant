package controllers

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	bffmiddleware "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/transport/http/middleware"
	commonv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/common/v1"
	onboardingv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/onboarding/v1"
)

// ─── Input / Output types ─────────────────────────────────────────────────────

// inviteMemberInput carries the invite request body.
type inviteMemberInput struct {
	Body struct {
		Email string `json:"email" format:"email" doc:"Email address of the user to invite"`
		Role  string `json:"role" enum:"read_only,update,write" doc:"Role to assign to the invited member"`
	}
}

// updateMemberRoleInput carries the member ID and new role.
type updateMemberRoleInput struct {
	MemberID string `path:"memberId" doc:"Project member UUID"`
	Body     struct {
		Role string `json:"role" enum:"read_only,update,write" doc:"New role for the member"`
	}
}

// projectResponse is the JSON shape for a single project.
type projectResponse struct {
	ID        string `json:"id"`
	OwnerID   string `json:"ownerId"`
	Name      string `json:"name"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

// projectMemberResponse is the JSON shape for a single project member.
type projectMemberResponse struct {
	ID         string `json:"id"`
	ProjectID  string `json:"projectId"`
	UserID     string `json:"userId"`
	Role       string `json:"role"`
	InvitedBy  string `json:"invitedBy,omitempty"`
	CreatedAt  string `json:"createdAt"`
	UpdatedAt  string `json:"updatedAt"`
}

// listMembersInput carries optional pagination for member listing.
type listMembersInput struct {
	PageSize  int32  `query:"pageSize"  minimum:"1" maximum:"100" doc:"Page size (default 25)"`
	PageToken string `query:"pageToken" doc:"Opaque cursor from a previous list response"`
}

// listMembersResponse is the JSON body for the list members endpoint.
type listMembersResponse struct {
	Items         []projectMemberResponse `json:"items"`
	NextPageToken string                  `json:"nextPageToken,omitempty"`
}

// ─── Controller ───────────────────────────────────────────────────────────────

// ProjectsController registers and handles all project collaboration HTTP routes.
type ProjectsController struct {
	logger           *zap.Logger
	onboardingClient onboardingv1.OnboardingServiceClient
}

// NewProjectsController constructs a ProjectsController.
func NewProjectsController(logger *zap.Logger, onboardingClient onboardingv1.OnboardingServiceClient) *ProjectsController {
	return &ProjectsController{logger: logger, onboardingClient: onboardingClient}
}

// Register wires all project routes to the Huma API with auth + role middleware.
func (c *ProjectsController) Register(api huma.API, auth func(huma.Context, func(huma.Context))) {
	huma.Register(api, huma.Operation{
		OperationID: "get-current-project",
		Method:      http.MethodGet,
		Path:        "/api/v1/projects/current",
		Summary:     "Get the current project details",
		Description: "Returns project details for the project in the caller's JWT claims.",
		Tags:        []string{"projects"},
		Middlewares: huma.Middlewares{auth, bffmiddleware.NewProjectGuard("read_only", c.logger)},
	}, c.handleGetCurrent)

	huma.Register(api, huma.Operation{
		OperationID: "list-project-members",
		Method:      http.MethodGet,
		Path:        "/api/v1/projects/members",
		Summary:     "List members of the current project",
		Description: "Returns all project members with their roles.",
		Tags:        []string{"projects"},
		Middlewares: huma.Middlewares{auth, bffmiddleware.NewProjectGuard("read_only", c.logger)},
	}, c.handleListMembers)

	huma.Register(api, huma.Operation{
		OperationID: "invite-project-member",
		Method:      http.MethodPost,
		Path:        "/api/v1/projects/members/invite",
		Summary:     "Invite a user to the current project",
		Description: "Resolves the user by email and adds them as a project member with the given role.",
		Tags:        []string{"projects"},
		Middlewares: huma.Middlewares{auth, bffmiddleware.NewProjectGuard("write", c.logger)},
	}, c.handleInvite)

	huma.Register(api, huma.Operation{
		OperationID: "update-project-member-role",
		Method:      http.MethodPatch,
		Path:        "/api/v1/projects/members/{memberId}/role",
		Summary:     "Update the role of a project member",
		Description: "Changes the role of an existing project member.",
		Tags:        []string{"projects"},
		Middlewares: huma.Middlewares{auth, bffmiddleware.NewProjectGuard("write", c.logger)},
	}, c.handleUpdateRole)
}

// handleGetCurrent returns the project for the caller's JWT project_id.
func (c *ProjectsController) handleGetCurrent(ctx context.Context, _ *struct{}) (*struct{ Body projectResponse }, error) {
	claims := bffmiddleware.ClaimsFromContext(ctx)
	if claims == nil {
		return nil, huma.Error403Forbidden("missing project context")
	}

	resp, err := c.onboardingClient.GetProject(ctx, &onboardingv1.GetProjectRequest{
		Ctx: &commonv1.ProjectContext{
			ProjectId: claims.GetProjectId(),
			UserId:    claims.GetSubject(),
			Role:      claims.GetRole(),
		},
	})
	if err != nil {
		return nil, grpcToProjectsHumaError(err, "get current project", c.logger)
	}

	p := resp.GetProject()
	body := projectResponse{
		ID:        p.GetId(),
		OwnerID:   p.GetOwnerId(),
		Name:      p.GetName(),
		CreatedAt: p.GetCreatedAt(),
		UpdatedAt: p.GetUpdatedAt(),
	}
	return &struct{ Body projectResponse }{Body: body}, nil
}

// handleListMembers returns all members for the caller's project.
func (c *ProjectsController) handleListMembers(ctx context.Context, input *listMembersInput) (*struct{ Body listMembersResponse }, error) {
	claims := bffmiddleware.ClaimsFromContext(ctx)
	if claims == nil {
		return nil, huma.Error403Forbidden("missing project context")
	}

	pageSize := input.PageSize
	if pageSize == 0 {
		pageSize = 25
	}

	resp, err := c.onboardingClient.ListProjectMembers(ctx, &onboardingv1.ListProjectMembersRequest{
		Ctx: &commonv1.ProjectContext{
			ProjectId: claims.GetProjectId(),
			UserId:    claims.GetSubject(),
			Role:      claims.GetRole(),
		},
		Pagination: &commonv1.Pagination{
			PageSize:  pageSize,
			PageToken: input.PageToken,
		},
	})
	if err != nil {
		return nil, grpcToProjectsHumaError(err, "list members", c.logger)
	}

	items := make([]projectMemberResponse, 0, len(resp.GetMembers()))
	for _, m := range resp.GetMembers() {
		items = append(items, projectMemberResponse{
			ID:        m.GetId(),
			ProjectID: m.GetProjectId(),
			UserID:    m.GetUserId(),
			Role:      protoRoleToString(m.GetRole()),
			InvitedBy: m.GetInvitedBy(),
			CreatedAt: m.GetCreatedAt(),
			UpdatedAt: m.GetUpdatedAt(),
		})
	}

	nextToken := ""
	if resp.GetPagination() != nil {
		nextToken = resp.GetPagination().GetNextPageToken()
	}

	body := listMembersResponse{Items: items, NextPageToken: nextToken}
	return &struct{ Body listMembersResponse }{Body: body}, nil
}

// handleInvite adds a user by email to the project.
func (c *ProjectsController) handleInvite(ctx context.Context, input *inviteMemberInput) (*struct{ Body projectMemberResponse }, error) {
	claims := bffmiddleware.ClaimsFromContext(ctx)
	if claims == nil {
		return nil, huma.Error403Forbidden("missing project context")
	}

	role := roleStringToProto(input.Body.Role)

	resp, err := c.onboardingClient.InviteProjectMember(ctx, &onboardingv1.InviteProjectMemberRequest{
		Ctx: &commonv1.ProjectContext{
			ProjectId: claims.GetProjectId(),
			UserId:    claims.GetSubject(),
			Role:      claims.GetRole(),
		},
		InviteeEmail: input.Body.Email,
		Role:         role,
		Audit:        &commonv1.AuditMetadata{PerformedBy: claims.GetSubject()},
	})
	if err != nil {
		return nil, grpcToProjectsHumaError(err, "invite member", c.logger)
	}

	m := resp.GetMember()
	body := projectMemberResponse{
		ID:        m.GetId(),
		ProjectID: m.GetProjectId(),
		UserID:    m.GetUserId(),
		Role:      protoRoleToString(m.GetRole()),
		InvitedBy: m.GetInvitedBy(),
		CreatedAt: m.GetCreatedAt(),
		UpdatedAt: m.GetUpdatedAt(),
	}
	return &struct{ Body projectMemberResponse }{Body: body}, nil
}

// handleUpdateRole changes the role of an existing member.
func (c *ProjectsController) handleUpdateRole(ctx context.Context, input *updateMemberRoleInput) (*struct{ Body projectMemberResponse }, error) {
	claims := bffmiddleware.ClaimsFromContext(ctx)
	if claims == nil {
		return nil, huma.Error403Forbidden("missing project context")
	}

	role := roleStringToProto(input.Body.Role)

	resp, err := c.onboardingClient.UpdateProjectMemberRole(ctx, &onboardingv1.UpdateProjectMemberRoleRequest{
		Ctx: &commonv1.ProjectContext{
			ProjectId: claims.GetProjectId(),
			UserId:    claims.GetSubject(),
			Role:      claims.GetRole(),
		},
		MemberId: input.MemberID,
		NewRole:  role,
		Audit:    &commonv1.AuditMetadata{PerformedBy: claims.GetSubject()},
	})
	if err != nil {
		return nil, grpcToProjectsHumaError(err, "update member role", c.logger)
	}

	m := resp.GetMember()
	body := projectMemberResponse{
		ID:        m.GetId(),
		ProjectID: m.GetProjectId(),
		UserID:    m.GetUserId(),
		Role:      protoRoleToString(m.GetRole()),
		InvitedBy: m.GetInvitedBy(),
		CreatedAt: m.GetCreatedAt(),
		UpdatedAt: m.GetUpdatedAt(),
	}
	return &struct{ Body projectMemberResponse }{Body: body}, nil
}

// grpcToProjectsHumaError converts a gRPC status error to an appropriate Huma HTTP error.
func grpcToProjectsHumaError(err error, op string, logger *zap.Logger) error {
	st, ok := status.FromError(err)
	if !ok {
		logger.Error("projects controller: unexpected error", zap.String("op", op), zap.Error(err))
		return huma.NewError(http.StatusInternalServerError, op+" failed")
	}
	switch st.Code() {
	case codes.NotFound:
		return huma.NewError(http.StatusNotFound, st.Message())
	case codes.AlreadyExists:
		return huma.NewError(http.StatusConflict, st.Message())
	case codes.InvalidArgument:
		return huma.NewError(http.StatusBadRequest, st.Message())
	case codes.PermissionDenied:
		return huma.NewError(http.StatusForbidden, st.Message())
	default:
		logger.Error("projects controller: grpc error", zap.String("op", op), zap.Error(err))
		return huma.NewError(http.StatusInternalServerError, op+" failed")
	}
}

// roleStringToProto converts a role string to the proto enum.
func roleStringToProto(role string) onboardingv1.ProjectMemberRole {
	switch role {
	case "update":
		return onboardingv1.ProjectMemberRole_PROJECT_MEMBER_ROLE_UPDATE
	case "write":
		return onboardingv1.ProjectMemberRole_PROJECT_MEMBER_ROLE_WRITE
	default:
		return onboardingv1.ProjectMemberRole_PROJECT_MEMBER_ROLE_READ_ONLY
	}
}

// protoRoleToString converts a proto ProjectMemberRole enum to its user-facing string.
func protoRoleToString(r onboardingv1.ProjectMemberRole) string {
	switch r {
	case onboardingv1.ProjectMemberRole_PROJECT_MEMBER_ROLE_UPDATE:
		return "update"
	case onboardingv1.ProjectMemberRole_PROJECT_MEMBER_ROLE_WRITE:
		return "write"
	default:
		return "read_only"
	}
}
