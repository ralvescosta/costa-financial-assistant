package mappers

import (
	bffcontracts "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/services/contracts"
	views "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/transport/http/views"
)

func ToCreateBankAccountRequest(input *views.CreateBankAccountInput) string {
	if input == nil {
		return ""
	}
	return input.Body.Label
}

func ToDeleteBankAccountRequest(input *views.DeleteBankAccountInput) string {
	if input == nil {
		return ""
	}
	return input.BankAccountID
}

func ToListBankAccountsResponse(resp *bffcontracts.ListBankAccountsResponse) views.ListBankAccountsResponse {
	if resp == nil {
		return views.ListBankAccountsResponse{Items: []*views.BankAccountResponse{}}
	}
	return *resp
}

func ToBankAccountResponse(resp *bffcontracts.BankAccountResponse) views.BankAccountResponse {
	if resp == nil {
		return views.BankAccountResponse{}
	}
	return *resp
}
