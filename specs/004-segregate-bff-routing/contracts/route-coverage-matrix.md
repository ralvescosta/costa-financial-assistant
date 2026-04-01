# Route Coverage Matrix

## Purpose

Track the route-to-test mapping required by the specification so every declared BFF route has explicit integration coverage.

| Resource | Operation ID | Method | Path | Planned Coverage Suite |
|----------|--------------|--------|------|------------------------|
| documents | upload-document | POST | /api/v1/documents/upload | documents_routes_integration_test.go |
| documents | classify-document | POST | /api/v1/documents/{documentId}/classify | documents_routes_integration_test.go |
| documents | list-documents | GET | /api/v1/documents | documents_routes_integration_test.go |
| documents | get-document | GET | /api/v1/documents/{documentId} | documents_routes_integration_test.go |
| projects | get-current-project | GET | /api/v1/projects/current | projects_routes_integration_test.go |
| projects | list-project-members | GET | /api/v1/projects/members | projects_routes_integration_test.go |
| projects | invite-project-member | POST | /api/v1/projects/members/invite | projects_routes_integration_test.go |
| projects | update-project-member-role | PATCH | /api/v1/projects/members/{memberId}/role | projects_routes_integration_test.go |
| settings | list-bank-accounts | GET | /api/v1/bank-accounts | settings_routes_integration_test.go |
| settings | create-bank-account | POST | /api/v1/bank-accounts | settings_routes_integration_test.go |
| settings | delete-bank-account | DELETE | /api/v1/bank-accounts/{bankAccountId} | settings_routes_integration_test.go |
| payments | get-payment-dashboard | GET | /api/v1/bills/payment-dashboard | payments_routes_integration_test.go |
| payments | mark-bill-paid | POST | /api/v1/bills/{billId}/mark-paid | payments_routes_integration_test.go |
| payments | get-preferred-payment-day | GET | /api/v1/payment-cycle/preferred-day | payments_routes_integration_test.go |
| payments | set-preferred-payment-day | PUT | /api/v1/payment-cycle/preferred-day | payments_routes_integration_test.go |
| reconciliation | get-reconciliation-summary | GET | /api/v1/reconciliation/summary | reconciliation_routes_integration_test.go |
| reconciliation | create-reconciliation-link | POST | /api/v1/reconciliation/links | reconciliation_routes_integration_test.go |
| history | get-history-timeline | GET | /api/v1/history/timeline | history_routes_integration_test.go |
| history | get-history-categories | GET | /api/v1/history/categories | history_routes_integration_test.go |
| history | get-history-compliance | GET | /api/v1/history/compliance | history_routes_integration_test.go |

## Maintenance Rule

Any new BFF route is incomplete until this matrix is updated and the referenced integration suite contains at least one passing scenario for that route.