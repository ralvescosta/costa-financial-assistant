# Migration Mapping: Integration Tests

| Legacy Path | Target Path | Status | Coverage Note |
|-------------|-------------|--------|---------------|
| backend/tests/integration/auth_token_rejection_test.go | backend/tests/integration/bff/reject_invalid_token_test.go | moved | Authentication rejection behavior preserved. |
| backend/tests/integration/bff_metrics_test.go | backend/tests/integration/bff/expose_metrics_endpoint_test.go | moved | Metrics endpoint reachability preserved. |
| backend/tests/integration/openapi_contract_test.go | backend/tests/integration/bff/validate_openapi_metadata_test.go | moved | OpenAPI metadata completeness assertions preserved. |
| backend/tests/integration/bff_route_contract_wiring_test.go | backend/tests/integration/bff/bff_route_contract_wiring_test.go | moved | Route capability wiring checks preserved. |
| backend/tests/integration/bff_route_registration_smoke_test.go | backend/tests/integration/bff/bff_route_registration_smoke_test.go | moved | Route registration smoke coverage preserved. |
| backend/tests/integration/bff_route_test_helpers.go | backend/tests/integration/bff/bff_route_test_helpers.go | moved | Shared BFF test server helper preserved. |
| backend/tests/integration/documents_routes_integration_test.go | backend/tests/integration/bff/documents_routes_registration_test.go | moved | Documents route registration coverage preserved. |
| backend/tests/integration/history_routes_integration_test.go | backend/tests/integration/bff/history_routes_registration_test.go | moved | History route registration coverage preserved. |
| backend/tests/integration/payments_routes_integration_test.go | backend/tests/integration/bff/payments_routes_registration_test.go | moved | Payments route registration coverage preserved. |
| backend/tests/integration/projects_routes_integration_test.go | backend/tests/integration/bff/projects_routes_registration_test.go | moved | Projects route registration coverage preserved. |
| backend/tests/integration/reconciliation_routes_integration_test.go | backend/tests/integration/bff/reconciliation_routes_registration_test.go | moved | Reconciliation route registration coverage preserved. |
| backend/tests/integration/settings_routes_integration_test.go | backend/tests/integration/bff/settings_routes_registration_test.go | moved | Settings route registration coverage preserved. |
| backend/tests/integration/identity_jwks_contract_test.go | backend/tests/integration/identity/expose_jwks_metadata_contract_test.go | moved | JWKS contract assertions preserved. |
| backend/tests/integration/us1_upload_classify_test.go | backend/tests/integration/files/upload_and_classify_document_test.go | moved | Upload/classify flow preserved. |
| backend/tests/integration/us2_analysis_pipeline_test.go | backend/tests/integration/files/process_analysis_pipeline_test.go | moved | Analysis pipeline scenarios preserved. |
| backend/tests/integration/us3_bank_accounts_test.go | backend/tests/integration/files/manage_bank_accounts_crud_test.go | moved | Bank account CRUD behavior preserved. |
| backend/tests/integration/us4_mark_paid_idempotency_test.go | backend/tests/integration/payments/mark_bill_paid_idempotency_test.go | moved | Mark-paid idempotency behavior preserved. |
| backend/tests/integration/us4_payment_dashboard_test.go | backend/tests/integration/payments/view_payment_dashboard_test.go | moved | Payment dashboard behavior preserved. |
| backend/tests/integration/us5_auto_reconciliation_test.go | backend/tests/integration/cross_service/auto_reconcile_transactions_test.go | moved | Automatic reconciliation behavior preserved. |
| backend/tests/integration/us5_manual_reconciliation_test.go | backend/tests/integration/cross_service/create_manual_reconciliation_link_test.go | moved | Manual reconciliation link behavior preserved. |
| backend/tests/integration/us6_history_metrics_test.go | backend/tests/integration/cross_service/get_history_metrics_test.go | moved | History metrics behavior preserved. |
| backend/tests/integration/us6_history_timeline_test.go | backend/tests/integration/cross_service/get_history_timeline_test.go | moved | History timeline behavior preserved. |
| backend/tests/integration/us7_project_isolation_test.go | backend/tests/integration/cross_service/enforce_project_isolation_test.go | moved | Project isolation behavior preserved. |
| backend/tests/integration/us7_role_enforcement_test.go | backend/tests/integration/cross_service/enforce_role_permissions_test.go | moved | Role enforcement matrix behavior preserved. |
