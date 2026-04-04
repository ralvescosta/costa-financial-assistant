// Package views defines all HTTP request and response contracts for BFF routes.
package views

// InviteMemberInput carries the invite request body.
type InviteMemberInput struct {
	Body struct {
		Email string `json:"email" format:"email" doc:"Email address of the user to invite" validate:"required,email"`
		Role  string `json:"role" enum:"read_only,update,write" doc:"Role to assign to the invited member" validate:"required,oneof=read_only update write"`
	}
}

// UpdateMemberRoleInput carries the member ID and new role.
type UpdateMemberRoleInput struct {
	MemberID string `path:"memberId" doc:"Project member UUID" validate:"required,uuid4"`
	Body     struct {
		Role string `json:"role" enum:"read_only,update,write" doc:"New role for the member" validate:"required,oneof=read_only update write"`
	}
}

// ListMembersInput carries optional pagination for member listing.
type ListMembersInput struct {
	PageSize  int32  `query:"pageSize"  minimum:"1" maximum:"100" doc:"Page size (default 25 for project-member lists)"`
	PageToken string `query:"pageToken" doc:"Opaque cursor from a previous list response"`
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

// ListMembersResponse is the JSON body for the list members endpoint.
type ListMembersResponse struct {
	Items         []*ProjectMemberResponse `json:"items"`
	NextPageToken string                   `json:"nextPageToken,omitempty"`
}
