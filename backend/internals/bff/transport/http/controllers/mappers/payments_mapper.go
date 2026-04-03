package mappers

import (
	"strconv"

	bffcontracts "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/services/contracts"
	views "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/transport/http/views"
)

func ToPaymentDashboardRequest(input *views.GetPaymentDashboardInput) (string, string, int32, string) {
	if input == nil {
		return "", "", 20, ""
	}

	pageSize := int32(20)
	if input.PageSize != "" {
		if n, err := strconv.Atoi(input.PageSize); err == nil && n > 0 {
			pageSize = int32(n)
		}
	}
	return input.CycleStart, input.CycleEnd, pageSize, input.PageToken
}

func ToMarkBillPaidRequest(input *views.MarkBillPaidInput) string {
	if input == nil {
		return ""
	}
	return input.BillID
}

func ToSetPreferredDayRequest(input *views.SetPreferredDayInput) int {
	if input == nil {
		return 0
	}
	return input.Body.PreferredDayOfMonth
}

func ToPaymentDashboardResponse(resp *bffcontracts.PaymentDashboardResponse) views.PaymentDashboardResponse {
	if resp == nil {
		return views.PaymentDashboardResponse{Entries: []*views.PaymentDashboardEntryResponse{}}
	}
	return *resp
}

func ToMarkBillPaidResponse(resp *bffcontracts.MarkBillPaidResponse) views.MarkBillPaidResponse {
	if resp == nil {
		return views.MarkBillPaidResponse{}
	}
	return *resp
}

func ToCyclePreferenceResponse(resp *bffcontracts.CyclePreferenceResponse) views.CyclePreferenceResponse {
	if resp == nil {
		return views.CyclePreferenceResponse{}
	}
	return *resp
}
