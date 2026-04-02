# Migration Baseline: Integration Suite Inventory

## Pre-migration Inventory (legacy flat layout)

- backend/tests/integration/auth_token_rejection_test.go
- backend/tests/integration/bff_metrics_test.go
- backend/tests/integration/bff_route_contract_wiring_test.go
- backend/tests/integration/bff_route_registration_smoke_test.go
- backend/tests/integration/bff_route_test_helpers.go
- backend/tests/integration/documents_routes_integration_test.go
- backend/tests/integration/history_routes_integration_test.go
- backend/tests/integration/identity_jwks_contract_test.go
- backend/tests/integration/openapi_contract_test.go
- backend/tests/integration/payments_routes_integration_test.go
- backend/tests/integration/projects_routes_integration_test.go
- backend/tests/integration/reconciliation_routes_integration_test.go
- backend/tests/integration/settings_routes_integration_test.go
- backend/tests/integration/us1_upload_classify_test.go
- backend/tests/integration/us2_analysis_pipeline_test.go
- backend/tests/integration/us3_bank_accounts_test.go
- backend/tests/integration/us4_mark_paid_idempotency_test.go
- backend/tests/integration/us4_payment_dashboard_test.go
- backend/tests/integration/us5_auto_reconciliation_test.go
- backend/tests/integration/us5_manual_reconciliation_test.go
- backend/tests/integration/us6_history_metrics_test.go
- backend/tests/integration/us6_history_timeline_test.go
- backend/tests/integration/us7_project_isolation_test.go
- backend/tests/integration/us7_role_enforcement_test.go

## Post-migration Inventory (canonical segmented layout)

- backend/tests/integration/bff/bff_route_contract_wiring_test.go
- backend/tests/integration/bff/bff_route_registration_smoke_test.go
- backend/tests/integration/bff/bff_route_test_helpers.go
- backend/tests/integration/bff/documents_routes_registration_test.go
- backend/tests/integration/bff/expose_metrics_endpoint_test.go
- backend/tests/integration/bff/history_routes_registration_test.go
- backend/tests/integration/bff/payments_routes_registration_test.go
- backend/tests/integration/bff/projects_routes_registration_test.go
- backend/tests/integration/bff/reconciliation_routes_registration_test.go
- backend/tests/integration/bff/reject_invalid_token_test.go
- backend/tests/integration/bff/settings_routes_registration_test.go
- backend/tests/integration/bff/validate_openapi_metadata_test.go
- backend/tests/integration/cross_service/auto_reconcile_transactions_test.go
- backend/tests/integration/cross_service/create_manual_reconciliation_link_test.go
- backend/tests/integration/cross_service/enforce_project_isolation_test.go
- backend/tests/integration/cross_service/enforce_role_permissions_test.go
- backend/tests/integration/cross_service/get_history_metrics_test.go
- backend/tests/integration/cross_service/get_history_timeline_test.go
- backend/tests/integration/files/manage_bank_accounts_crud_test.go
- backend/tests/integration/files/process_analysis_pipeline_test.go
- backend/tests/integration/files/upload_and_classify_document_test.go
- backend/tests/integration/identity/expose_jwks_metadata_contract_test.go
- backend/tests/integration/payments/mark_bill_paid_idempotency_test.go
- backend/tests/integration/payments/view_payment_dashboard_test.go

## Package Listing Parity Check

- Legacy integration test files mapped: 24/24
- Canonical migrated test files: 24/24
- Additional helper and suite lifecycle files added for compliance:
  - backend/tests/integration/helpers/lifecycle.go
  - backend/tests/integration/helpers/scenario_helpers_test.go
  - backend/tests/integration/files/suite_test.go
  - backend/tests/integration/payments/suite_test.go
  - backend/tests/integration/cross_service/suite_test.go

## Integration Run Notes

- Focused package checks executed:
  - `go test ./tests/integration/helpers ./tests/integration/bff ./tests/integration/identity -tags=integration` ✅
  - `go test ./tests/integration/files -tags=integration -run TestUS1_UploadAndClassifyDocument -count=1` ❌ (`documents` relation missing)
- Current blocker:
  - service migration directories are present but do not currently contain SQL migration files in this branch, so DB-backed integration suites cannot reach schema parity yet.
- Convention validator status:
  - `./backend/scripts/validate_integration_test_conventions.sh` currently fails by design until US2 table-driven BDD + AAA refactor tasks (T018-T020) are completed across all migrated tests.
