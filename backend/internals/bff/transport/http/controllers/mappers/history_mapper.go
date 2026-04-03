package mappers

import (
	bffcontracts "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/services/contracts"
	views "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/transport/http/views"
)

func ToHistoryMonths(input *views.HistoryQueryInput) int {
	if input == nil {
		return 0
	}
	return input.Months
}

func ToTimelineResponse(resp *bffcontracts.TimelineResponse) views.TimelineResponse {
	if resp == nil {
		return views.TimelineResponse{Timeline: []*views.MonthlyTimelineEntryResponse{}}
	}
	return *resp
}

func ToCategoryBreakdownResponse(resp *bffcontracts.CategoryBreakdownResponse) views.CategoryBreakdownResponse {
	if resp == nil {
		return views.CategoryBreakdownResponse{Categories: []*views.CategoryBreakdownEntryResponse{}}
	}
	return *resp
}

func ToComplianceResponse(resp *bffcontracts.ComplianceResponse) views.ComplianceResponse {
	if resp == nil {
		return views.ComplianceResponse{Compliance: []*views.MonthlyComplianceEntryResponse{}}
	}
	return *resp
}
