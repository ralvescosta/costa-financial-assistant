// Package views defines all HTTP request and response contracts for BFF routes.
package views

// GetPaymentDashboardInput carries query parameters for the payment dashboard.
type GetPaymentDashboardInput struct {
	CycleStart string `query:"cycleStart" doc:"ISO-8601 cycle start date (YYYY-MM-DD)"`
	CycleEnd   string `query:"cycleEnd" doc:"ISO-8601 cycle end date (YYYY-MM-DD)"`
	PageSize   string `query:"pageSize" doc:"Number of results per page"`
	PageToken  string `query:"pageToken" doc:"Opaque pagination cursor"`
}

// PaymentBillRecordResponse is the JSON shape for a single bill record in payment routes.
type PaymentBillRecordResponse struct {
	ID            string `json:"id"`
	ProjectID     string `json:"projectId"`
	DocumentID    string `json:"documentId"`
	BillTypeID    string `json:"billTypeId,omitempty"`
	DueDate       string `json:"dueDate"`
	AmountDue     string `json:"amountDue"`
	PixPayload    string `json:"pixPayload,omitempty"`
	PixQRImageRef string `json:"pixQrImageRef,omitempty"`
	Barcode       string `json:"barcode,omitempty"`
	PaymentStatus string `json:"paymentStatus"`
	PaidAt        string `json:"paidAt,omitempty"`
	MarkedPaidBy  string `json:"markedPaidBy,omitempty"`
	CreatedAt     string `json:"createdAt"`
	UpdatedAt     string `json:"updatedAt"`
}

// PaymentBillTypeResponse is the JSON shape for a bill type label in payment routes.
type PaymentBillTypeResponse struct {
	ID        string `json:"id"`
	ProjectID string `json:"projectId"`
	Name      string `json:"name"`
}

// PaymentDashboardEntryResponse represents a single dashboard row.
type PaymentDashboardEntryResponse struct {
	Bill         PaymentBillRecordResponse `json:"bill"`
	BillType     *PaymentBillTypeResponse   `json:"billType,omitempty"`
	IsOverdue    bool                       `json:"isOverdue"`
	DaysUntilDue int32                      `json:"daysUntilDue"`
}

// PaymentDashboardResponse is the GET payment-dashboard response body.
type PaymentDashboardResponse struct {
	Entries       []*PaymentDashboardEntryResponse `json:"entries"`
	NextPageToken string                           `json:"nextPageToken,omitempty"`
}

// MarkBillPaidInput carries mark-paid request parameters.
type MarkBillPaidInput struct {
	BillID string `path:"billId" doc:"Bill record UUID" validate:"required,uuid4"`
}

// MarkBillPaidResponse is returned on success.
type MarkBillPaidResponse struct {
	Bill PaymentBillRecordResponse `json:"bill"`
}

// CyclePreferenceResponse is the JSON shape for payment cycle preferences.
type CyclePreferenceResponse struct {
	ProjectID           string `json:"projectId"`
	PreferredDayOfMonth int    `json:"preferredDayOfMonth"`
	UpdatedAt           string `json:"updatedAt"`
}

// SetPreferredDayInput carries the preferred day of month.
type SetPreferredDayInput struct {
	Body struct {
		PreferredDayOfMonth int `json:"preferredDayOfMonth" minimum:"1" maximum:"28" doc:"Preferred payment day (1–28)" validate:"required,min=1,max=28"`
	}
}
