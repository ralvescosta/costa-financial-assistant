# BFF Route Migration Matrix

This matrix records the supported BFF routes that must end this feature on a clean downstream-service contract, with no direct BFF domain-data access remaining.

| Route | BFF controller handler | BFF service method | Owning service | Target RPC / contract | Verification focus |
|---|---|---|---|---|---|
| `GET /api/v1/bills/payment-dashboard` | `PaymentsController.HandleGetDashboard` | `PaymentsService.GetPaymentDashboard` | `bills` | `bills.v1.BillsService.GetPaymentDashboard` | Existing behavior preserved |
| `POST /api/v1/bills/{billId}/mark-paid` | `PaymentsController.HandleMarkPaid` | `PaymentsService.MarkBillPaid` | `bills` | `bills.v1.BillsService.MarkBillPaid` | Existing behavior preserved |
| `GET /api/v1/payment-cycle/preferred-day` | `PaymentsController.HandleGetPreferredDay` | `PaymentsService.GetCyclePreference` | `payments` | new `payments.v1.PaymentsService.GetCyclePreference` | No direct fallback; response preserved |
| `PUT /api/v1/payment-cycle/preferred-day` | `PaymentsController.HandleSetPreferredDay` | `PaymentsService.SetCyclePreference` | `payments` | new `payments.v1.PaymentsService.SetCyclePreference` | Validation + mutation preserved |
| `GET /api/v1/history/timeline` | `HistoryController.HandleGetTimeline` | `HistoryService.GetTimeline` | `payments` | new `payments.v1.PaymentsService.GetHistoryTimeline` | Auth + analytics response preserved |
| `GET /api/v1/history/categories` | `HistoryController.HandleGetCategories` | `HistoryService.GetCategoryBreakdown` | `payments` | new `payments.v1.PaymentsService.GetHistoryCategoryBreakdown` | Category response preserved |
| `GET /api/v1/history/compliance` | `HistoryController.HandleGetCompliance` | `HistoryService.GetComplianceMetrics` | `payments` | new `payments.v1.PaymentsService.GetHistoryCompliance` | Compliance response preserved |
| `GET /api/v1/reconciliation/summary` | `ReconciliationController.HandleGetSummary` | `ReconciliationService.GetSummary` | `payments` | new `payments.v1.PaymentsService.GetReconciliationSummary` | Guard + summary behavior preserved |
| `POST /api/v1/reconciliation/links` | `ReconciliationController.HandleCreateLink` | `ReconciliationService.CreateManualLink` | `payments` | new `payments.v1.PaymentsService.CreateManualLink` | Guard + mutation behavior preserved |

## Required regression evidence

- Route registration and OpenAPI metadata continue to pass in `backend/tests/integration/bff/*_routes_registration_test.go`.
- Cross-service analytics/reconciliation behavior remains covered in `backend/tests/integration/cross_service/`.
- No BFF service file for the routes above imports payments repositories or `backend/internals/payments/interfaces` directly in the final state.
