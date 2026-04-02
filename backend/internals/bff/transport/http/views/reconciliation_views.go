// Package views defines all HTTP request and response contracts for BFF routes.
package views

// ReconciliationSummaryInput carries optional date filters for the summary endpoint.
type ReconciliationSummaryInput struct {
	PeriodStart string `query:"periodStart" doc:"ISO-8601 date (YYYY-MM-DD) — start of reconciliation window"`
	PeriodEnd   string `query:"periodEnd" doc:"ISO-8601 date (YYYY-MM-DD) — end of reconciliation window"`
}

// ReconciliationEntryResponse is a single row in the reconciliation summary.
type ReconciliationEntryResponse struct {
	TransactionLineID    string  `json:"transactionLineId"`
	TransactionDate      string  `json:"transactionDate"`
	Description          string  `json:"description"`
	Amount               string  `json:"amount"`
	Direction            string  `json:"direction"`
	ReconciliationStatus string  `json:"reconciliationStatus"`
	LinkedBillID         *string `json:"linkedBillId,omitempty"`
	LinkedBillDueDate    *string `json:"linkedBillDueDate,omitempty"`
	LinkedBillAmount     *string `json:"linkedBillAmount,omitempty"`
	LinkType             *string `json:"linkType,omitempty"`
}

// ReconciliationSummaryResponse is the body for the GET /reconciliation/summary endpoint.
type ReconciliationSummaryResponse struct {
	ProjectID   string                         `json:"projectId"`
	PeriodStart string                         `json:"periodStart,omitempty"`
	PeriodEnd   string                         `json:"periodEnd,omitempty"`
	Entries     []*ReconciliationEntryResponse `json:"entries"`
}

// CreateReconciliationLinkInput carries the body for POST /reconciliation/links.
type CreateReconciliationLinkInput struct {
	Body struct {
		TransactionLineID string `json:"transactionLineId" doc:"UUID of the transaction line to link" validate:"required,uuid4"`
		BillRecordID      string `json:"billRecordId" doc:"UUID of the bill record to link" validate:"required,uuid4"`
	}
}

// ReconciliationLinkResponse is returned on successful link creation.
type ReconciliationLinkResponse struct {
	ID                string  `json:"id"`
	ProjectID         string  `json:"projectId"`
	TransactionLineID string  `json:"transactionLineId"`
	BillRecordID      string  `json:"billRecordId"`
	LinkType          string  `json:"linkType"`
	LinkedBy          *string `json:"linkedBy,omitempty"`
	CreatedAt         string  `json:"createdAt"`
}
