# Tasks: Stabilize Broken BFF Page Flows

**Input**: Design documents from `/specs/013-stabilize-bff-page-flows/`  
**Prerequisites**: `plan.md` (required), `spec.md` (required), `research.md`, `data-model.md`, `contracts/`, `quickstart.md`

**Tests**: Included. The spec explicitly requires end-to-end integration coverage for the BFF and downstream-service flows plus seeded default-user verification for the four broken pages.

**Organization**: Tasks are grouped by user story so each increment remains independently testable and can be delivered in priority order. Every feature task list includes a final mandatory governance sync phase.

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Prepare the shared validation matrix and integration helpers used by all stories.

- [ ] T001 Refresh the page-flow audit checklist and reproduction notes in `specs/013-stabilize-bff-page-flows/contracts/page-flow-validation-matrix.md` and `specs/013-stabilize-bff-page-flows/quickstart.md`
- [ ] T002 [P] Extend shared authenticated scenario helpers in `backend/tests/integration/helpers/scenario_helpers_test.go` and `backend/tests/integration/helpers/lifecycle.go`
- [ ] T003 [P] Wire common BFF page-bootstrap test helpers in `backend/tests/integration/bff/bff_route_test_helpers.go` and `backend/tests/integration/bff/bff_route_contract_wiring_test.go`

**Checkpoint**: The feature has a shared request inventory and reusable test infrastructure for the affected page flows.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Establish the default-user bootstrap, project access, and shared failure-handling rules before story work begins.

**⚠️ CRITICAL**: No user story work should begin until this phase is complete.

- [ ] T004 Align the default user’s auth bootstrap in `backend/internals/identity/migrations/dml/local/000001_seed_default_user.up.sql`, `backend/internals/identity/migrations/dml/local/000002_seed_bootstrap_owner.up.sql`, and `backend/tests/integration/identity/bootstrap_login_seed_test.go`
- [ ] T005 [P] Align the default project-membership bootstrap in `backend/internals/onboarding/migrations/dml/local/000001_seed_default_project.up.sql`, `backend/internals/onboarding/migrations/dml/local/000002_seed_bootstrap_owner_membership.up.sql`, and `backend/tests/integration/cross_service/restore_login_session_propagation_test.go`
- [ ] T006 [P] Normalize authenticated request context and project-guard behavior in `backend/internals/bff/transport/http/middleware/auth_middleware.go`, `backend/internals/bff/transport/http/controllers/base_controller.go`, and `backend/tests/integration/cross_service/enforce_project_isolation_test.go`
- [ ] T007 [P] Build shared default-user demo-data builders in `backend/tests/integration/helpers/scenario_helpers_test.go`, `backend/tests/integration/payments/suite_test.go`, and `backend/tests/integration/cross_service/suite_test.go`
- [ ] T008 Audit and normalize AppError-first fallback handling for page bootstrap requests in `backend/internals/bff/services/documents_service.go`, `backend/internals/bff/services/payments_service.go`, `backend/internals/bff/services/history_service.go`, `backend/internals/bff/services/reconciliation_service.go`, and `backend/internals/bff/services/settings_service.go`

**Checkpoint**: Auth, membership, seeded-data builders, and shared error-handling behavior are ready for page-specific work.

---

## Phase 3: User Story 1 - Open key screens without backend errors (Priority: P1) 🎯 MVP

**Goal**: Make the `documents`, `payments`, `analyses`, and `settings` screens load successfully for the default user without blocker-level backend request failures.

**Independent Test**: Sign in as the default user, open each in-scope screen, and confirm the required page-load requests return supported populated or empty-state responses instead of generic backend failures.

### Tests for User Story 1 ⚠️

> Write these tests first and confirm the targeted flow is failing before finalizing the implementation.

- [ ] T009 [P] [US1] Add documents/settings page bootstrap integration coverage in `backend/tests/integration/bff/load_documents_page_success_test.go` and `backend/tests/integration/bff/load_settings_page_success_test.go`
- [ ] T010 [P] [US1] Add payments/analyses page bootstrap integration coverage in `backend/tests/integration/cross_service/load_payments_page_success_test.go`, `backend/tests/integration/cross_service/load_analyses_history_success_test.go`, and `backend/tests/integration/cross_service/load_analyses_reconciliation_success_test.go`
- [ ] T011 [P] [US1] Extend route-level unauthorized, forbidden, and supported-empty-state assertions in `backend/tests/integration/bff/documents_routes_registration_test.go`, `backend/tests/integration/bff/payments_routes_registration_test.go`, `backend/tests/integration/bff/history_routes_registration_test.go`, `backend/tests/integration/bff/reconciliation_routes_registration_test.go`, and `backend/tests/integration/bff/settings_routes_registration_test.go`

### Implementation for User Story 1

- [ ] T012 [US1] Fix the document page bootstrap and upload/list response mapping in `backend/internals/bff/services/documents_service.go`, `backend/internals/bff/transport/http/controllers/documents_controller.go`, and `backend/internals/bff/transport/http/controllers/mappers/documents_mapper.go`
- [ ] T013 [US1] Fix the settings bank-account page flow in `backend/internals/bff/services/settings_service.go`, `backend/internals/bff/transport/http/controllers/settings_controller.go`, and `backend/internals/bff/transport/http/controllers/mappers/settings_mapper.go`
- [ ] T014 [US1] Fix the payments dashboard and preferred-day flow in `backend/internals/bff/services/payments_service.go`, `backend/internals/bff/transport/http/controllers/payments_controller.go`, and `backend/internals/bff/transport/http/controllers/mappers/payments_mapper.go`
- [ ] T015 [US1] Fix the analyses history and reconciliation page flows in `backend/internals/bff/services/history_service.go`, `backend/internals/bff/services/reconciliation_service.go`, `backend/internals/bff/transport/http/controllers/history_controller.go`, and `backend/internals/bff/transport/http/controllers/reconciliation_controller.go`
- [ ] T016 [US1] Correct any downstream contract or empty-state behavior needed by the four pages in `backend/internals/files/transport/grpc/server.go`, `backend/internals/bills/transport/grpc/server.go`, and `backend/internals/payments/transport/grpc/server.go`

**Checkpoint**: The four in-scope screens are functional for the default user and no longer fail at page bootstrap.

---

## Phase 4: User Story 2 - Validate the full frontend-to-service flow (Priority: P1)

**Goal**: Prove that each in-scope frontend request is mapped, validated, and protected across the full BFF-to-downstream-service path.

**Independent Test**: Run the BFF and cross-service integration suites and confirm each request path has explicit success, auth failure, project-access failure, and sanitized dependency-failure evidence.

### Tests for User Story 2 ⚠️

- [ ] T017 [P] [US2] Add BFF service regression tests for downstream error/access translation in `backend/internals/bff/services/documents_service_test.go`, `backend/internals/bff/services/payments_service_test.go`, `backend/internals/bff/services/history_service_test.go`, `backend/internals/bff/services/reconciliation_service_test.go`, and `backend/internals/bff/services/settings_service_test.go`
- [ ] T018 [P] [US2] Extend cross-service validation for project isolation and role enforcement in `backend/tests/integration/cross_service/enforce_project_isolation_test.go`, `backend/tests/integration/cross_service/enforce_role_permissions_test.go`, and `backend/tests/integration/cross_service/app_error_boundary_logging_test.go`

### Implementation for User Story 2

- [ ] T019 [US2] Trace and finalize the request inventory and owning-route documentation in `specs/013-stabilize-bff-page-flows/research.md` and `specs/013-stabilize-bff-page-flows/contracts/page-flow-validation-matrix.md`
- [ ] T020 [US2] Normalize request-context propagation and access semantics for history and reconciliation in `backend/internals/bff/transport/http/routes/history_routes.go`, `backend/internals/bff/services/history_service.go`, and `backend/internals/bff/services/reconciliation_service.go`
- [ ] T021 [US2] Align BFF response contracts and mapper/view behavior for the four page flows in `backend/internals/bff/services/contracts/documents_contracts.go`, `backend/internals/bff/services/contracts/reconciliation_contracts.go`, `backend/internals/bff/transport/http/views/documents_views.go`, `backend/internals/bff/transport/http/views/payments_views.go`, `backend/internals/bff/transport/http/views/history_views.go`, `backend/internals/bff/transport/http/views/reconciliation_views.go`, `backend/internals/bff/transport/http/views/settings_views.go`, and `backend/internals/bff/transport/http/controllers/mappers/documents_mapper.go`, `backend/internals/bff/transport/http/controllers/mappers/payments_mapper.go`, `backend/internals/bff/transport/http/controllers/mappers/history_mapper.go`, `backend/internals/bff/transport/http/controllers/mappers/reconciliation_mapper.go`, `backend/internals/bff/transport/http/controllers/mappers/settings_mapper.go`
- [ ] T022 [US2] Reconcile any frontend service or hook contract mismatches with the stabilized BFF payloads in `frontend/src/services/documentsApi.ts`, `frontend/src/services/paymentsApi.ts`, `frontend/src/services/historyApi.ts`, `frontend/src/services/reconciliationApi.ts`, `frontend/src/services/bankAccountsApi.ts`, `frontend/src/hooks/useDocuments.ts`, `frontend/src/hooks/useUploadDocument.ts`, `frontend/src/hooks/usePaymentDashboard.ts`, `frontend/src/hooks/usePreferredDay.ts`, `frontend/src/hooks/useHistoryDashboard.ts`, `frontend/src/hooks/useReconciliationSummary.ts`, and `frontend/src/hooks/useBankAccounts.ts`

**Checkpoint**: Every in-scope request path is explicitly validated and traceable to the correct backend owner.

---

## Phase 5: User Story 3 - Use realistic default-user data for investigation and demos (Priority: P2)

**Goal**: Provide representative, populated default-user data across all four pages so the system can be understood and validated without manual setup.

**Independent Test**: Bootstrap the local/test environment and confirm the default user sees meaningful content on `documents`, `payments`, `analyses`, and `settings`.

### Tests for User Story 3 ⚠️

- [ ] T023 [P] [US3] Add seeded-data verification coverage for the default user in `backend/tests/integration/identity/bootstrap_login_seed_test.go`, `backend/tests/integration/files/upload_and_classify_document_test.go`, `backend/tests/integration/payments/view_payment_dashboard_test.go`, and `backend/tests/integration/files/manage_bank_accounts_crud_test.go`

### Implementation for User Story 3

- [ ] T024 [US3] Populate the default user’s identity/auth seed state in `backend/internals/identity/migrations/dml/local/000001_seed_default_user.up.sql` and `backend/internals/identity/migrations/dml/local/000002_seed_bootstrap_owner.up.sql`
- [ ] T025 [US3] Populate the default project and membership seed state in `backend/internals/onboarding/migrations/dml/local/000001_seed_default_project.up.sql` and `backend/internals/onboarding/migrations/dml/local/000002_seed_bootstrap_owner_membership.up.sql`
- [ ] T026 [US3] Populate representative documents and settings demo data in `backend/internals/files/migrations/dml/local/000001_seed_default_documents.up.sql` and `backend/tests/integration/helpers/scenario_helpers_test.go`
- [ ] T027 [US3] Populate representative bills, transactions, payment-cycle, and reconciliation demo data in `backend/internals/bills/migrations/dml/local/000001_seed_bill_types.up.sql`, `backend/internals/payments/migrations/dml/local/000001_seed_transaction_types.up.sql`, `backend/tests/integration/payments/suite_test.go`, `backend/tests/integration/cross_service/suite_test.go`, and `backend/tests/integration/helpers/scenario_helpers_test.go`
- [ ] T028 [US3] Add or document a deterministic local/demo seed loader or migration-backed bootstrap path for the populated payments and analyses experience in `specs/013-stabilize-bff-page-flows/contracts/default-user-seed-contract.md`, `specs/013-stabilize-bff-page-flows/quickstart.md`, and the touched local seed files under `backend/internals/{bills,payments}/migrations/dml/local/`

**Checkpoint**: The canonical default user sees populated, connected demo data across all four in-scope pages in local/test/demo environments.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Finalize verification evidence, frontend smoke coverage, and cleanup across all stories.

- [ ] T029 [P] Add missing frontend hook smoke coverage for page bootstrap states in `frontend/src/hooks/useDocuments.test.ts`, `frontend/src/hooks/usePaymentDashboard.test.ts`, `frontend/src/hooks/useBankAccounts.test.ts`, `frontend/src/hooks/useHistoryDashboard.test.ts`, and `frontend/src/hooks/useReconciliationSummary.test.ts`
- [ ] T030 Run the required verification suite from `backend/` and `frontend/`, then fix any regressions in `backend/tests/integration/**`, `backend/internals/bff/services/*_test.go`, and `frontend/src/**/*.test.ts`
- [ ] T031 [P] Add and execute an explicit real-screen smoke-validation step for `documents`, `payments`, `analyses`, and `settings` in `specs/013-stabilize-bff-page-flows/quickstart.md` and the final implementation checklist
- [ ] T032 [P] Confirm or update exception tracking and the final validation runbook in `specs/013-stabilize-bff-page-flows/contracts/pointer-exceptions.md` and `specs/013-stabilize-bff-page-flows/quickstart.md`

---

## Phase 7: Mandatory Governance Sync (Blocking)

**Purpose**: Keep memory artifacts, repository guidance, and testing conventions aligned with the stabilized implementation.

- [ ] T033 Update `.specify/memory/architecture-diagram.md` and `.specify/memory/bff-flows.md` with the verified page-to-BFF-to-service request paths
- [ ] T034 [P] Update `.specify/memory/files-service-flows.md`, `.specify/memory/bills-service-flows.md`, `.specify/memory/payments-service-flows.md`, `.specify/memory/identity-service-flows.md`, and `.specify/memory/onboarding-service-flows.md` for the final root cause and seed behavior
- [ ] T035 [P] Update `/memories/repo/bff-service-boundary-conventions.md` and verify canonical integration-test placement, snake_case filenames, and BDD + AAA compliance across `backend/tests/integration/bff/`, `backend/tests/integration/files/`, `backend/tests/integration/payments/`, and `backend/tests/integration/cross_service/`

**Checkpoint**: The feature is not complete until this phase is complete.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1: Setup** → can start immediately
- **Phase 2: Foundational** → depends on Setup and blocks all user stories
- **Phase 3: US1** → depends on Foundational completion
- **Phase 4: US2** → depends on US1 stabilizing the visible page flows
- **Phase 5: US3** → depends on Foundational bootstrap data and can overlap with late US1/US2 validation once the seed shape is stable
- **Phase 6: Polish** → depends on the desired user stories being complete
- **Phase 7: Mandatory Governance Sync** → depends on all implementation phases and must complete before merge

### User Story Dependencies

- **US1 (P1)**: First deliverable and MVP; restores the broken screens for the default user.
- **US2 (P1)**: Builds on US1 by making each route traceable, protected, and regression-covered end-to-end.
- **US3 (P2)**: Uses the stabilized flows to provide the populated demo/seed experience required for reliable validation and demos.

### Within Each User Story

- Write the regression tests first and confirm they capture the failure mode.
- Fix the BFF service/mapper/controller behavior before changing the frontend contract shape.
- Update downstream behavior only where the owning service must correct the response or seeded data.
- Re-run the story-specific checks before moving to the next dependent story.

### Parallel Opportunities

- `T002` and `T003` can run in parallel once the validation matrix is agreed.
- `T005`, `T006`, and `T007` can run in parallel during the Foundational phase.
- US1 test tasks `T009`–`T011` can run in parallel.
- US2 test tasks `T017` and `T018` can run in parallel.
- US3 seed tasks `T024`–`T028` can be split across identity/onboarding/files/payments workstreams.
- Governance tasks `T034` and `T035` can run in parallel after implementation behavior is final.

---

## Parallel Example: User Story 1

```bash
# Page-bootstrap regression work that can proceed together:
Task: "T009 [US1] Add documents/settings page bootstrap integration coverage in backend/tests/integration/bff/load_documents_page_success_test.go and backend/tests/integration/bff/load_settings_page_success_test.go"
Task: "T010 [US1] Add payments/analyses page bootstrap integration coverage in backend/tests/integration/cross_service/load_payments_page_success_test.go, load_analyses_history_success_test.go, and load_analyses_reconciliation_success_test.go"
Task: "T011 [US1] Extend route-level unauthorized, forbidden, and supported-empty-state assertions in backend/tests/integration/bff/*_routes_registration_test.go"
```

---

## Parallel Example: User Story 3

```bash
# Default-user seed work that can proceed together once the shape is agreed:
Task: "T024 [US3] Populate the default user’s identity/auth seed state in backend/internals/identity/migrations/dml/local/000001_seed_default_user.up.sql and 000002_seed_bootstrap_owner.up.sql"
Task: "T025 [US3] Populate the default project and membership seed state in backend/internals/onboarding/migrations/dml/local/000001_seed_default_project.up.sql and 000002_seed_bootstrap_owner_membership.up.sql"
Task: "T026 [US3] Populate representative documents and settings demo data in backend/internals/files/migrations/dml/local/000001_seed_default_documents.up.sql and backend/tests/integration/helpers/scenario_helpers_test.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete **Phase 1: Setup**
2. Complete **Phase 2: Foundational**
3. Complete **Phase 3: User Story 1**
4. **Stop and validate**: sign in as the default user and confirm the four in-scope pages no longer fail at bootstrap

### Incremental Delivery

1. Setup + Foundational establish the shared auth/seed/integration baseline
2. US1 restores the visible page behavior for the default user
3. US2 locks in route-by-route validation and downstream ownership correctness
4. US3 finalizes the representative seeded demo state across all four pages
5. Polish and Governance Sync close the loop with verification evidence and architecture-memory updates

### Parallel Team Strategy

With multiple developers:

1. One developer handles shared test infrastructure and bootstrap/auth context (`T001`–`T008`)
2. One developer stabilizes the visible BFF page flows (`T009`–`T016`)
3. One developer drives route validation and frontend/BFF contract alignment (`T017`–`T022`)
4. One developer finalizes seeded demo data, validation evidence, and sync artifacts (`T023`–`T035`)

---

## Notes

- All tasks follow the required checklist format with an ID and explicit file path.
- `[US1]`, `[US2]`, and `[US3]` map directly to the user stories in `spec.md`.
- `[P]` tasks are safe to execute in parallel once their prerequisite phase is complete.
- Run backend verification from `backend/` because this repository uses a nested Go module rooted at `backend/go.mod`.
- Preserve the BFF gateway boundary while fixing the failures; do not move domain ownership into the frontend or controller layer.
