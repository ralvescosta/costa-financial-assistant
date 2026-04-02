// Package views defines all HTTP request and response contracts for BFF routes.
package views

// HistoryQueryInput carries the optional look-back window for all history endpoints.
type HistoryQueryInput struct {
	Months int `query:"months" doc:"Number of calendar months to look back; 0 = all history. Default: 12" minimum:"0"`
}

// MonthlyTimelineEntryResponse is a single row of the expenditure timeline.
type MonthlyTimelineEntryResponse struct {
	Month       string `json:"month"`
	TotalAmount string `json:"totalAmount"`
	BillCount   int    `json:"billCount"`
}

// TimelineResponse is the body for GET /history/timeline.
type TimelineResponse struct {
	ProjectID string                          `json:"projectId"`
	Months    int                             `json:"months"`
	Timeline  []*MonthlyTimelineEntryResponse `json:"timeline"`
}

// CategoryBreakdownEntryResponse is a single row of the category breakdown.
type CategoryBreakdownEntryResponse struct {
	Month        string `json:"month"`
	BillTypeName string `json:"billTypeName"`
	TotalAmount  string `json:"totalAmount"`
	BillCount    int    `json:"billCount"`
}

// CategoryBreakdownResponse is the body for GET /history/categories.
type CategoryBreakdownResponse struct {
	ProjectID  string                            `json:"projectId"`
	Months     int                               `json:"months"`
	Categories []*CategoryBreakdownEntryResponse `json:"categories"`
}

// MonthlyComplianceEntryResponse is a single row of the compliance metrics.
type MonthlyComplianceEntryResponse struct {
	Month          string `json:"month"`
	TotalBills     int    `json:"totalBills"`
	PaidOnTime     int    `json:"paidOnTime"`
	Overdue        int    `json:"overdue"`
	ComplianceRate string `json:"complianceRate"`
}

// ComplianceResponse is the body for GET /history/compliance.
type ComplianceResponse struct {
	ProjectID  string                            `json:"projectId"`
	Months     int                               `json:"months"`
	Compliance []*MonthlyComplianceEntryResponse `json:"compliance"`
}
