# Route Coverage Matrix

## Purpose

Track the active BFF route inventory and the integration suites that must remain green while the route, views, controller, and service boundaries are refactored.

| Resource | Operation ID | Method | Path | Primary Coverage Suite | Structural Coverage |
|----------|--------------|--------|------|------------------------|---------------------|
| Documents | `upload-document` | `POST` | `/api/v1/documents/upload` | `backend/tests/integration/bff/documents_routes_registration_test.go` | `backend/tests/integration/bff/bff_route_registration_smoke_test.go`, `backend/tests/integration/bff/validate_openapi_metadata_test.go` |
| Documents | `classify-document` | `POST` | `/api/v1/documents/{documentId}/classify` | `backend/tests/integration/bff/documents_routes_registration_test.go` | `backend/tests/integration/bff/bff_route_registration_smoke_test.go`, `backend/tests/integration/bff/validate_openapi_metadata_test.go` |
| Documents | `list-documents` | `GET` | `/api/v1/documents` | `backend/tests/integration/bff/documents_routes_registration_test.go` | `backend/tests/integration/bff/bff_route_registration_smoke_test.go`, `backend/tests/integration/bff/validate_openapi_metadata_test.go` |
| Documents | `get-document` | `GET` | `/api/v1/documents/{documentId}` | `backend/tests/integration/bff/documents_routes_registration_test.go` | `backend/tests/integration/bff/bff_route_registration_smoke_test.go`, `backend/tests/integration/bff/validate_openapi_metadata_test.go` |
| Projects | `get-current-project` | `GET` | `/api/v1/projects/current` | `backend/tests/integration/bff/projects_routes_registration_test.go` | `backend/tests/integration/bff/bff_route_registration_smoke_test.go`, `backend/tests/integration/bff/validate_openapi_metadata_test.go` |
| Projects | `list-project-members` | `GET` | `/api/v1/projects/members` | `backend/tests/integration/bff/projects_routes_registration_test.go` | `backend/tests/integration/bff/bff_route_registration_smoke_test.go`, `backend/tests/integration/bff/validate_openapi_metadata_test.go` |
| Projects | `invite-project-member` | `POST` | `/api/v1/projects/members/invite` | `backend/tests/integration/bff/projects_routes_registration_test.go` | `backend/tests/integration/bff/bff_route_registration_smoke_test.go`, `backend/tests/integration/bff/validate_openapi_metadata_test.go` |
| Projects | `update-project-member-role` | `PATCH` | `/api/v1/projects/members/{memberId}/role` | `backend/tests/integration/bff/projects_routes_registration_test.go` | `backend/tests/integration/bff/bff_route_registration_smoke_test.go`, `backend/tests/integration/bff/validate_openapi_metadata_test.go` |
| Settings | `list-bank-accounts` | `GET` | `/api/v1/bank-accounts` | `backend/tests/integration/bff/settings_routes_registration_test.go` | `backend/tests/integration/bff/bff_route_registration_smoke_test.go`, `backend/tests/integration/bff/validate_openapi_metadata_test.go` |
| Settings | `create-bank-account` | `POST` | `/api/v1/bank-accounts` | `backend/tests/integration/bff/settings_routes_registration_test.go` | `backend/tests/integration/bff/bff_route_registration_smoke_test.go`, `backend/tests/integration/bff/validate_openapi_metadata_test.go` |
| Settings | `delete-bank-account` | `DELETE` | `/api/v1/bank-accounts/{bankAccountId}` | `backend/tests/integration/bff/settings_routes_registration_test.go` | `backend/tests/integration/bff/bff_route_registration_smoke_test.go`, `backend/tests/integration/bff/validate_openapi_metadata_test.go` |
| Payments | `get-payment-dashboard` | `GET` | `/api/v1/bills/payment-dashboard` | `backend/tests/integration/bff/payments_routes_registration_test.go` | `backend/tests/integration/bff/bff_route_registration_smoke_test.go`, `backend/tests/integration/bff/validate_openapi_metadata_test.go` |
| Payments | `mark-bill-paid` | `POST` | `/api/v1/bills/{billId}/mark-paid` | `backend/tests/integration/bff/payments_routes_registration_test.go` | `backend/tests/integration/bff/bff_route_registration_smoke_test.go`, `backend/tests/integration/bff/validate_openapi_metadata_test.go` |
| Payments | `get-preferred-payment-day` | `GET` | `/api/v1/payment-cycle/preferred-day` | `backend/tests/integration/bff/payments_routes_registration_test.go` | `backend/tests/integration/bff/bff_route_registration_smoke_test.go`, `backend/tests/integration/bff/validate_openapi_metadata_test.go` |
| Payments | `set-preferred-payment-day` | `PUT` | `/api/v1/payment-cycle/preferred-day` | `backend/tests/integration/bff/payments_routes_registration_test.go` | `backend/tests/integration/bff/bff_route_registration_smoke_test.go`, `backend/tests/integration/bff/validate_openapi_metadata_test.go` |
| Reconciliation | `get-reconciliation-summary` | `GET` | `/api/v1/reconciliation/summary` | `backend/tests/integration/bff/reconciliation_routes_registration_test.go` | `backend/tests/integration/bff/bff_route_registration_smoke_test.go`, `backend/tests/integration/bff/validate_openapi_metadata_test.go` |
| Reconciliation | `create-reconciliation-link` | `POST` | `/api/v1/reconciliation/links` | `backend/tests/integration/bff/reconciliation_routes_registration_test.go` | `backend/tests/integration/bff/bff_route_registration_smoke_test.go`, `backend/tests/integration/bff/validate_openapi_metadata_test.go` |
| History | `get-history-timeline` | `GET` | `/api/v1/history/timeline` | `backend/tests/integration/bff/history_routes_registration_test.go` | `backend/tests/integration/bff/bff_route_registration_smoke_test.go`, `backend/tests/integration/bff/validate_openapi_metadata_test.go` |
| History | `get-history-categories` | `GET` | `/api/v1/history/categories` | `backend/tests/integration/bff/history_routes_registration_test.go` | `backend/tests/integration/bff/bff_route_registration_smoke_test.go`, `backend/tests/integration/bff/validate_openapi_metadata_test.go` |
| History | `get-history-compliance` | `GET` | `/api/v1/history/compliance` | `backend/tests/integration/bff/history_routes_registration_test.go` | `backend/tests/integration/bff/bff_route_registration_smoke_test.go`, `backend/tests/integration/bff/validate_openapi_metadata_test.go` |

## Maintenance Rules

- A route change is incomplete until this matrix is updated if the active route inventory changes.
- A new active BFF route must add both a primary resource-scoped integration suite assertion and structural coverage through the smoke and OpenAPI tests.
- If a route is removed or renamed, the matching row and its validating suites must change in the same pull request.