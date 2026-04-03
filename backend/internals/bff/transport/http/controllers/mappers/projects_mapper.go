package mappers

import (
	bffcontracts "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/services/contracts"
	views "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/transport/http/views"
)

func ToListMembersRequest(input *views.ListMembersInput) (int32, string) {
	if input == nil {
		page := ToServicePage(0, "")
		return page.Size, page.Token
	}

	page := ToServicePage(input.PageSize, input.PageToken)
	return page.Size, page.Token
}

func ToInviteMemberRequest(input *views.InviteMemberInput) (string, string) {
	if input == nil {
		return "", ""
	}
	return input.Body.Email, input.Body.Role
}

func ToUpdateMemberRoleRequest(input *views.UpdateMemberRoleInput) (string, string) {
	if input == nil {
		return "", ""
	}
	return input.MemberID, input.Body.Role
}

func ToProjectResponse(resp *bffcontracts.ProjectResponse) views.ProjectResponse {
	if resp == nil {
		return views.ProjectResponse{}
	}
	return *resp
}

func ToListMembersResponse(resp *bffcontracts.ListMembersResponse) views.ListMembersResponse {
	if resp == nil {
		return views.ListMembersResponse{Items: []*views.ProjectMemberResponse{}}
	}
	return *resp
}

func ToProjectMemberResponse(resp *bffcontracts.ProjectMemberResponse) views.ProjectMemberResponse {
	if resp == nil {
		return views.ProjectMemberResponse{}
	}
	return *resp
}
