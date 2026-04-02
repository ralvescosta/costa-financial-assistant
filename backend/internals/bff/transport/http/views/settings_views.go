// Package views defines all HTTP request and response contracts for BFF routes.
package views

// CreateBankAccountInput carries the label for a new bank account.
type CreateBankAccountInput struct {
	Body struct {
		Label string `json:"label" minLength:"1" maxLength:"100" doc:"Display label for the bank account" validate:"required,min=1,max=100"`
	}
}

// DeleteBankAccountInput carries the bank account ID path parameter.
type DeleteBankAccountInput struct {
	BankAccountID string `path:"bankAccountId" doc:"Bank account UUID" validate:"required,uuid4"`
}

// BankAccountResponse is the JSON shape returned for a single bank account.
type BankAccountResponse struct {
	ID        string `json:"id"`
	ProjectID string `json:"projectId"`
	Label     string `json:"label"`
	CreatedBy string `json:"createdBy,omitempty"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

// ListBankAccountsResponse is the JSON body for the list endpoints.
type ListBankAccountsResponse struct {
	Items []*BankAccountResponse `json:"items"`
}
