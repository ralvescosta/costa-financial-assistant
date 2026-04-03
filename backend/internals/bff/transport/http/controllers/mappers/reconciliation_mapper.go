package mappers

import (
	bffcontracts "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/services/contracts"
	views "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/transport/http/views"
)

func ToReconciliationSummaryRequest(input *views.ReconciliationSummaryInput) (string, string) {
	if input == nil {
		return "", ""
	}
	return input.PeriodStart, input.PeriodEnd
}

func ToCreateReconciliationLinkRequest(input *views.CreateReconciliationLinkInput) (string, string) {
	if input == nil {
		return "", ""
	}
	return input.Body.TransactionLineID, input.Body.BillRecordID
}

func ToReconciliationSummaryResponse(resp *bffcontracts.ReconciliationSummaryResponse) views.ReconciliationSummaryResponse {
	if resp == nil {
		return views.ReconciliationSummaryResponse{Entries: []*views.ReconciliationEntryResponse{}}
	}
	return *resp
}

func ToReconciliationLinkResponse(resp *bffcontracts.ReconciliationLinkResponse) views.ReconciliationLinkResponse {
	if resp == nil {
		return views.ReconciliationLinkResponse{}
	}
	return *resp
}
