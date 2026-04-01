package controllers

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"go.uber.org/zap"

	bffmiddleware "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/transport/http/middleware"
	commonv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/common/v1"
	onboardingv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/onboarding/v1"
)

// ─── Input / Output types ─────────────────────────────────────────────────────

// InviteMemberInput carries the invite request body.
type InviteMemberInput struct {
	Body struct {
		Email string `json:"email" format:"email" doc:"Email address of the user to invite"`
		Role  string `json:"role" enum:"read_only,update,write" doc:"Role to assign to the invited member"`
	}
}

// UpdateMemberRoleInput carries the member ID and new role.
type UpdateMemberRoleInput struct {
	MemberID string `path:"memberId" doc:"Project member UUID"`
	Body     struct {
		Role string `json:"role" enum:"read_only,update,write" doc:"New role for the member"`
	}
}

// ProjectResponse is the JSON shape for a single project.
type ProjectResponse struct {
	ID        string `json:"id"`
	OwnerID   string `json:"ownerId"`
	Name      string `json:"name"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

// ProjectMemberResponse is the JSON shape for a single project member.
type ProjectMemberResponse struct {
	ID        string `json:"id"`
	ProjectID string `json:"projectId"`
	UserID    string `json:"userId"`
	Role      string `json:"role"`
	InvitedBy string `json:"invitedBy,omitempty"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

// ListMembersInput carries optional pagination for member listing.
type ListMembersInput struct {
	PageSize  int32  `query:"pageSize"  minimum:"1" maximum:"100" doc:"Page size (default 25)"`
	PageToken string `query:"pageToken" doc:"Opaque cursor from a previous list response"`
}

// ListMembersResponse is the JSON body for the list members endpoint.
type ListMembersResponse struct {
	Items         []ProjectMemberResponse `json:"items"`
	NextPageToken string                  `json:"nextPageToken,omitempty"`
}

// ─── Controller ───────────────────────────────────────────────────────────────

// ProjectsController handles BFF project collaboration HTTP endpoints.
type ProjectsController struct {
	BaseController
	onboardingClient onboardingv1.OnboardingServiceClient
}

// NewProjectsController constructs a ProjectsController.
func NewProjectsController(logger *zap.Logger, onboardingClient onboardingv1.OnboardingServiceClient) *ProjectsController {
	return &ProjectsController{BaseController: BaseController{logger: logger}, onboardingClient: onboardingClient}
}

// ─── Handlers ─────────────────────────────────────────────────────────────────

// HandleGetCurrent returns the project for the caller's JWT project_id.
func (c *ProjectsController) HandleGetCurrent(ctx context.Context, _ *struct{}) (*struct{ Body ProjectResponse }, error) {
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
		return nil, c.grpcToHumaError(err, "get current project")
	}

	p := resp.GetProject()
	body := ProjectResponse{
		ID:        p.GetId(),
		OwnerID:   p.GetOwnerId(),
		Name:      p.GetName(),
		CreatedAt: p.GetCreatedAt(),
		UpdatedAt: p.GetUpdatedAt(),
	}
	return &struct{ Body ProjectResponse }{Body: body}, nil
}

// HandleListMembers returns all members for the caller's project.
func (c *ProjectsController) HandleListMembers(ctx context.Context, input *ListMembersInput) (*struct{ Body ListMembersResponse }, error) {
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
		return nil, c.grpcToHumaError(err, "list members")
	}

	items := make([]ProjectMemberResponse, 0, len(resp.GetMembers()))
	for _, m := range resp.GetMembers() {
		items = append(items, ProjectMemberResponse{
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

	body := ListMembersResponse{Items: items, NextPageToken: nextToken}
	return &struct{ Body ListMembersResponse }{Body: body}, nil
}

// HandleInvite adds a user by email to the project.
func (c *ProjectsController) HandleInvite(ctx context.Context, input *InviteMemberInput) (*struct{ Body ProjectMemberResponse }, error) {
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
		return nil, c.grpcToHumaError(err, "invite member")
	}

	m := resp.GetMember()
	body := ProjectMemberResponse{
		ID:        m.GetId(),
		ProjectID: m.GetProjectId(),
		UserID:    m.GetUserId(),
		Role:      protoRoleToString(m.GetRole()),
		InvitedBy: m.GetInvitedBy(),
		CreatedAt: m.GetCreatedAt(),
		UpdatedAt: m.GetUpdatedAt(),
	}
	return &struct{ Body ProjectMemberResponse }{Body: body}, nil
}

// HandleUpdateRole changes the role of an existing member.
func (c *ProjectsController) HandleUpdateRole(ctx context.Context, input *UpdateMemberRoleInput) (*struct{ Body ProjectMemberResponse }, error) {
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
		return nil, c.grpcToHumaError(err, "update member role")
	}

	m := resp.GetMember()
	body := ProjectMemberResponse{
		ID:        m.GetId(),
		ProjectID: m.GetProjectId(),
		UserID:    m.GetUserId(),
		Role:      protoRoleToString(m.GetRole()),
		InvitedBy: m.GetInvitedBy(),
		CreatedAt: m.GetCreatedAt(),
		UpdatedAt: m.GetUpdatedAt(),
	}
	return &struct{ Body ProjectMemberResponse }{Body: body}, nil
}

// ── helpers ──────────────────────────────────────────────────────────────────────────────

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
