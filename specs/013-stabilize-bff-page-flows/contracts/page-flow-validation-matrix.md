# Contract: Page Flow Validation Matrix

## Purpose

This matrix defines the concrete frontend-to-BFF-to-service paths that must be validated to consider the `documents`, `payments`, `analyses`, and `settings` screens stable for the default user.

## Flow Matrix

| Page | Frontend files | BFF endpoint(s) | Downstream owner | Canonical seeded expectation | Preferred integration-test placement |
|---|---|---|---|---|---|
| `documents` | `frontend/src/pages/UploadPage.tsx`, `frontend/src/hooks/useDocuments.ts`, `frontend/src/hooks/useUploadDocument.ts`, `frontend/src/services/documentsApi.ts` | `GET /api/v1/documents`, `POST /api/v1/documents/upload` | `files.v1` | At least one classified or analyzable document record is visible | `backend/tests/integration/bff/load_documents_page_success_test.go` |
| `payments` | `frontend/src/pages/PaymentDashboardPage.tsx`, `frontend/src/hooks/usePaymentDashboard.ts`, `frontend/src/hooks/usePreferredDay.ts`, `frontend/src/services/paymentsApi.ts` | `GET /api/v1/bills/payment-dashboard`, `POST /api/v1/bills/{billId}/mark-paid`, `GET/POST /api/v1/payment-cycle/preferred-day` | `bills.v1` + `payments.v1` | The user sees unpaid/overdue bill examples plus a valid cycle preference | `backend/tests/integration/cross_service/load_payments_page_success_test.go` |
| `analyses` (history) | `frontend/src/pages/HistoryDashboardPage.tsx`, `frontend/src/hooks/useHistoryDashboard.ts`, `frontend/src/services/historyApi.ts` | `GET /api/v1/history/timeline`, `GET /api/v1/history/categories` | `payments.v1` | Timeline and category breakdown return populated analytical data | `backend/tests/integration/cross_service/load_analyses_history_success_test.go` |
| `analyses` (reconciliation) | `frontend/src/pages/ReconciliationPage.tsx`, `frontend/src/hooks/useReconciliationSummary.ts`, `frontend/src/services/reconciliationApi.ts` | `GET /api/v1/reconciliation/summary`, `POST /api/v1/reconciliation/links` | `payments.v1` | Summary contains representative matched/unmatched/ambiguous transaction states | `backend/tests/integration/cross_service/load_analyses_reconciliation_success_test.go` |
| `settings` | `frontend/src/pages/SettingsPage.tsx`, `frontend/src/hooks/useBankAccounts.ts`, `frontend/src/services/bankAccountsApi.ts` | `GET/POST/DELETE /api/v1/bank-accounts` | `files.v1` | At least one bank-account label exists and CRUD does not crash the page | `backend/tests/integration/bff/load_settings_page_success_test.go` |

## Required Negative Cases

Each in-scope page flow must also verify these outcomes where applicable:

| Scenario | Expected behavior |
|---|---|
| Unauthenticated request | Return the intended auth failure rather than a generic 500 |
| Wrong-project or missing membership | Return the intended forbidden/access-denied result |
| Legitimate no-data state | Return a supported empty-state payload instead of a generic internal error |
| Downstream failure | Return a sanitized failure with AppError-consistent behavior; do not bypass service ownership |

## Notes

- The file names above are the recommended canonical placements for new coverage and must follow the existing backend integration-test standard.
- The matrix is acceptance-driven: if any listed page flow remains unvalidated, the feature is not complete.
