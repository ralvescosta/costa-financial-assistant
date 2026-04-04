package mappers

import (
	bffcontracts "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/services/contracts"
	views "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/transport/http/views"
)

func ToLoginCredentials(input *views.LoginInput) (string, string) {
	if input == nil {
		return "", ""
	}
	return input.Body.Username, input.Body.Password
}

func ToLoginOutput(resp *bffcontracts.AuthSessionResponse, sessionCookie string) *views.LoginOutput {
	body := views.LoginResponse{StatusCode: 200}
	if resp != nil {
		body.Data = views.LoginSuccessData{
			ExpiresIn: resp.ExpiresIn,
			RefreshAt: resp.RefreshAt,
			CSRFToken: resp.CSRFToken,
			User: views.UserSummaryResponse{
				ID:       resp.User.ID,
				Username: resp.User.Username,
				Email:    resp.User.Email,
			},
		}
		if resp.ActiveProject != nil {
			body.Data.ActiveProject = &views.ProjectSummaryResponse{
				ID:   resp.ActiveProject.ID,
				Name: resp.ActiveProject.Name,
				Role: resp.ActiveProject.Role,
			}
		}
	}
	return &views.LoginOutput{SetCookie: sessionCookie, Body: body}
}

func ToRefreshOutput(resp *bffcontracts.RefreshSessionResponse, sessionCookie string) *views.RefreshOutput {
	body := views.RefreshResponse{StatusCode: 200}
	if resp != nil {
		body.Data = views.RefreshSuccessData{
			ExpiresIn: resp.ExpiresIn,
			RefreshAt: resp.RefreshAt,
			CSRFToken: resp.CSRFToken,
		}
	}
	return &views.RefreshOutput{SetCookie: sessionCookie, Body: body}
}
