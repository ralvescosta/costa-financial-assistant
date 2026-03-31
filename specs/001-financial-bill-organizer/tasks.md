# Tasks: Financial Bill Organizer

**Input**: Design documents from `/specs/001-financial-bill-organizer/`
**Prerequisites**: `plan.md`, `spec.md`, `research.md`, `data-model.md`, `contracts/`

**Tests**: Included by request. Frontend tests are hook-only (BDD + Triple-A). Backend includes unit tests and transport-level integration tests with ephemeral DB lifecycle.

**Organization**: Tasks are grouped by user story so each story can be implemented and validated independently.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependency on incomplete tasks)
- **[Story]**: User story label (`[US1]..[US7]`) for story-phase tasks only
- Every task includes a concrete file path

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Initialize app skeletons and baseline tooling.

- [X] T001 Scaffold React + Vite + TypeScript frontend project in `frontend/package.json`
- [X] T002 Configure Tailwind + Vite plugin and base CSS entry in `frontend/vite.config.ts`
- [X] T003 [P] Create frontend app shell/routing skeleton in `frontend/src/app/router.tsx`
- [X] T004 [P] Add centralized design token scaffold in `frontend/src/styles/tokens.ts`
- [X] T005 Create root automation targets (`dev-up`, service run/test/migrate/proto) in `Makefile`
- [X] T006 [P] Add deterministic proto generation target in `Makefile`
- [X] T007 Add local integration-test DB compose profile in `docker-compose.yml`

### Phase 1 Verification

- [ ] T108 Run backend Go module build to confirm zero compile errors in `backend/` (`go build ./...`)
- [ ] T109 [P] Run frontend install and Vite production build to confirm compilation passes in `frontend/` (`npm ci && npm run build`)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core architecture and platform building blocks required by all stories.

**⚠️ CRITICAL**: No user-story implementation starts before this phase completes.

- [X] T008 Create multi-tenant bootstrap schema migration (`users/projects/project_members`) in `backend/onboarding/migrations/000001_create_identity_and_projects.up.sql`
- [X] T009 Create reversible down migration for tenant bootstrap schema in `backend/onboarding/migrations/000001_create_identity_and_projects.down.sql`
- [X] T010 Create seed migration for bootstrap user/project/member in `backend/onboarding/migrations/000002_seed_bootstrap_tenant.up.sql`
- [X] T011 Create down migration for bootstrap seed data in `backend/onboarding/migrations/000002_seed_bootstrap_tenant.down.sql`
- [X] T012 [P] Add onboarding v1 domain messages contract in `backend/protos/onboarding/v1/messages.proto`
- [X] T013 [P] Add onboarding v1 grpc service contract in `backend/protos/onboarding/v1/grpc.proto`
- [X] T014 [P] Add identity v1 domain messages contract in `backend/protos/identity/v1/messages.proto`
- [X] T015 [P] Add identity v1 grpc service contract in `backend/protos/identity/v1/grpc.proto`
- [X] T016 Add common shared proto messages in `backend/protos/common/v1/messages.proto`
- [X] T017 Regenerate protobuf Go artifacts in `backend/protos/generated/`
- [X] T018 Implement identity bootstrap JWT + JWKS service wiring in `backend/cmd/identity/container.go`
- [X] T019 Implement BFF Echo+Huma bootstrap with otelecho middleware in `backend/cmd/bff/container.go`
- [X] T020 [P] Implement JWT/JWKS validation middleware for BFF in `backend/internals/bff/transport/http/middleware/auth_middleware.go`
- [X] T021 [P] Implement project-membership and role guard middleware in `backend/internals/bff/transport/http/middleware/project_guard.go`
- [X] T022 Implement UnitOfWork for files service in `backend/internals/files/repositories/unit_of_work.go`
- [X] T023 Implement ephemeral integration-test DB harness in `backend/tests/integration/testmain_test.go`
- [X] T024 [P] Configure frontend query client/provider bootstrap in `frontend/src/app/providers.tsx`
- [X] T098 Implement UnitOfWork for bills service in `backend/internals/bills/repositories/unit_of_work.go`
- [X] T099 Implement UnitOfWork for payments service in `backend/internals/payments/repositories/unit_of_work.go`
- [X] T100 Implement UnitOfWork for onboarding service in `backend/internals/onboarding/repositories/unit_of_work.go`
- [X] T101 [P] Implement frontend theme provider with semantic light/dark token mapping in `frontend/src/app/theme/ThemeProvider.tsx`
- [X] T102 [P] Implement persisted theme preference + OS fallback bootstrap in `frontend/src/app/theme/themeStorage.ts`
- [X] T103 Implement no-reload theme toggle wiring in `frontend/src/components/ThemeToggle.tsx`
- [X] T104 [P] Add hook tests for theme preference/bootstrap flows in `frontend/src/hooks/useThemePreference.test.ts`
- [X] T105 [P] Add integration test for identity JWKS metadata endpoint contract in `backend/tests/integration/identity_jwks_contract_test.go`
- [X] T106 [P] Add integration test matrix for invalid and expired token rejection in `backend/tests/integration/auth_token_rejection_test.go`
- [X] T107 Implement JWKS cache/refresh component for auth middleware in `backend/internals/bff/transport/http/middleware/jwks_cache.go`
- [ ] T152 [P] Add shared project-scoped idempotency key migration for mutating bill and reconciliation workflows in `backend/bills/migrations/000003_create_idempotency_keys.up.sql`
- [ ] T153 [P] Add down migration for shared project-scoped idempotency keys in `backend/bills/migrations/000003_create_idempotency_keys.down.sql`
- [ ] T132 [P] Define identity service and repository interfaces in `backend/internals/identity/interfaces/identity_service.go`
- [ ] T133 [P] Define BFF gRPC client interfaces for downstream service consumers in `backend/internals/bff/interfaces/grpc_clients.go`
- [ ] T134 Wire otelgrpc unary and streaming server interceptors to identity gRPC server bootstrap in `backend/cmd/identity/container.go`
- [ ] T135 Add `<service>_up` health gauge and `build_info` OTel metrics to all six service bootstraps in `backend/cmd/*/container.go`
- [ ] T136 [P] Generate mockgen mocks for identity interfaces in `backend/internals/identity/mocks/`

### Phase 2 Verification

- [ ] T110 Run full backend build and vet to confirm zero compile errors across all services in `backend/` (`go build ./... && go vet ./...`)
- [ ] T111 [P] Run foundational integration tests to confirm JWT/JWKS auth flows pass in `backend/tests/integration/` (`go test ./tests/integration/ -run 'JWKS|Token'`)
- [ ] T112 [P] Run frontend theme hook unit tests to confirm all pass in `frontend/src/hooks/useThemePreference.test.ts` (`npm test -- --run useThemePreference`)

**Checkpoint**: Foundation complete. User stories can now proceed.

---

## Phase 3: User Story 1 - Upload and Classify a PDF Document (Priority: P1) 🎯 MVP

**Goal**: Upload PDF, classify as bill/statement, persist metadata, and show pending-analysis state.

**Independent Test**: Upload a valid PDF, classify it, and verify project-scoped listing shows `Pending Analysis` with label.

### Tests for User Story 1

- [X] T025 [P] [US1] Add BFF integration test for upload + classify flow in `backend/tests/integration/us1_upload_classify_test.go`
- [X] T026 [P] [US1] Add files service unit tests (duplicate detection + PDF validation) in `backend/internals/files/services/upload_service_test.go`
- [X] T027 [P] [US1] Add hook test for upload flow (BDD + Triple-A) in `frontend/src/hooks/useUploadDocument.test.ts`
- [X] T028 [P] [US1] Add hook test for classification flow (BDD + Triple-A) in `frontend/src/hooks/useClassifyDocument.test.ts`

### Implementation for User Story 1

- [X] T029 [P] [US1] Add files v1 message definitions for document upload/classification in `backend/protos/files/v1/messages.proto`
- [X] T030 [P] [US1] Add files v1 grpc service methods for upload/classify/list/get in `backend/protos/files/v1/grpc.proto`
- [X] T031 [US1] Regenerate files protobuf artifacts in `backend/protos/generated/`
- [ ] T154 [P] [US1] Add migration for project-scoped bill-type labels in `backend/bills/migrations/000001_create_bill_types.up.sql`
- [ ] T155 [P] [US1] Add down migration for bill-type labels in `backend/bills/migrations/000001_create_bill_types.down.sql`
- [ ] T137 [US1] Define files service and document repository interfaces in `backend/internals/files/interfaces/document_service.go`
- [ ] T138 [P] [US1] Generate mockgen mocks for files service interfaces in `backend/internals/files/mocks/`
- [X] T032 [US1] Implement document repository with project-scoped uniqueness by hash in `backend/internals/files/repositories/document_repository.go`
- [X] T033 [US1] Implement upload/classification application service in `backend/internals/files/services/document_service.go`
- [X] T034 [US1] Implement BFF documents controller (`upload`, `classify`, `list`, `get`) with Huma metadata in `backend/internals/bff/financial/controllers/documents_controller.go`
- [X] T035 [US1] Implement frontend upload hook with classification mutation chain in `frontend/src/hooks/useUploadDocument.ts`
- [X] T036 [US1] Implement frontend document list/query hook in `frontend/src/hooks/useDocuments.ts`
- [X] T037 [US1] Implement upload/classification screen composition in `frontend/src/pages/UploadPage.tsx`

### Phase 3 Verification

- [ ] T113 Run files service unit tests to confirm upload/classification logic passes in `backend/internals/files/services/` (`go test ./internals/files/...`)
- [ ] T114 [P] Run US1 integration test suite to confirm upload/classify/list flow passes in `backend/tests/integration/us1_upload_classify_test.go` (`go test ./tests/integration/ -run 'Upload|Classify'`)
- [ ] T115 [P] Run frontend US1 hook tests to confirm all assertions pass in `frontend/src/hooks/` (`npm test -- --run 'useUploadDocument|useClassifyDocument|useDocuments'`)

**Checkpoint**: US1 is independently functional and testable.

---

## Phase 4: User Story 2 - Asynchronous PDF Analysis and Data Extraction (Priority: P2)

**Goal**: Process uploaded PDFs asynchronously and persist extracted bill/statement data with status transitions.

**Independent Test**: Submit a known bill PDF, process async job, and verify extracted fields + status changes.

### Tests for User Story 2

- [ ] T139 [P] [US2] Define extraction and analysis service interfaces in `backend/internals/files/interfaces/extraction_service.go`
- [ ] T140 [P] [US2] Generate mockgen mocks for files service after adding extraction interfaces (depends on T137 and T139) in `backend/internals/files/mocks/`
- [ ] T038 [P] [US2] Add integration test for analysis status transitions and failure handling in `backend/tests/integration/us2_analysis_pipeline_test.go`
- [ ] T039 [P] [US2] Add unit tests for extraction orchestration service in `backend/internals/files/services/analysis_service_test.go`
- [ ] T040 [P] [US2] Add hook test for document analysis polling/status in `frontend/src/hooks/useDocumentStatus.test.ts`

### Implementation for User Story 2

- [ ] T041 [P] [US2] Add migration for only the `analysis_jobs`, `statement_records`, and `transaction_lines` tables in `backend/files/migrations/000001_create_analysis_tables.up.sql`
- [ ] T042 [P] [US2] Add down migration for only the `analysis_jobs`, `statement_records`, and `transaction_lines` tables in `backend/files/migrations/000001_create_analysis_tables.down.sql`
- [ ] T156 [P] [US2] Add migration for project-scoped `bill_records` table in `backend/bills/migrations/000002_create_bill_records.up.sql`
- [ ] T157 [P] [US2] Add down migration for project-scoped `bill_records` table in `backend/bills/migrations/000002_create_bill_records.down.sql`
- [ ] T043 [US2] Implement async analysis job publisher/consumer pipeline with W3C trace-context injection and extraction in `backend/internals/files/transport/rmq/analysis_consumer.go`
- [ ] T044 [US2] Implement extraction service for due date/amount/pix/barcode and statement lines in `backend/internals/files/services/extraction_service.go`
- [ ] T045 [US2] Implement BFF document-detail projection for extracted fields after T034 in `backend/internals/bff/financial/controllers/documents_controller.go`
- [ ] T046 [US2] Implement frontend hook for document detail/status polling in `frontend/src/hooks/useDocumentStatus.ts`
- [ ] T047 [US2] Implement document detail page with extraction results and not-found markers in `frontend/src/pages/DocumentDetailPage.tsx`

### Phase 4 Verification

- [ ] T116 Run analysis service unit tests to confirm extraction orchestration passes in `backend/internals/files/services/` (`go test ./internals/files/services/... -run Analysis`)
- [ ] T117 [P] Run US2 integration tests to confirm async pipeline and status transitions pass in `backend/tests/integration/us2_analysis_pipeline_test.go` (`go test ./tests/integration/ -run 'Analysis|Pipeline'`)
- [ ] T118 [P] Run frontend document status hook tests to confirm polling behavior passes in `frontend/src/hooks/useDocumentStatus.test.ts` (`npm test -- --run useDocumentStatus`)

**Checkpoint**: US2 independently testable on top of foundational + US1 data ingestion.

---

## Phase 5: User Story 3 - Bank Account Registration (Priority: P2)

**Goal**: Manage project-scoped bank account labels used by statement classification.

**Independent Test**: Create/update/delete labels and validate classification dialog uses those labels.

### Tests for User Story 3

- [ ] T141 [P] [US3] Define files-service bank-account service and repository interfaces in `backend/internals/files/interfaces/bank_account_service.go`
- [ ] T142 [P] [US3] Generate mockgen mocks for files service after adding bank-account interfaces (depends on T137, T139, and T141) in `backend/internals/files/mocks/`
- [ ] T048 [P] [US3] Add integration test for bank-account CRUD and attribution guard in `backend/tests/integration/us3_bank_accounts_test.go`
- [ ] T049 [P] [US3] Add unit tests for bank-account repository validation rules in `backend/internals/files/repositories/bank_account_repository_test.go`
- [ ] T050 [P] [US3] Add hook test for bank-account CRUD flows in `frontend/src/hooks/useBankAccounts.test.ts`

### Implementation for User Story 3

- [ ] T051 [P] [US3] Add migration for project-scoped bank-account labels in `backend/files/migrations/000002_create_bank_accounts.up.sql`
- [ ] T052 [P] [US3] Add down migration for bank-account labels in `backend/files/migrations/000002_create_bank_accounts.down.sql`
- [ ] T053 [US3] Implement bank-account repository with project scope + duplicate checks in `backend/internals/files/repositories/bank_account_repository.go`
- [ ] T054 [US3] Implement BFF settings controller for bank-account endpoints in `backend/internals/bff/financial/controllers/settings_controller.go`
- [ ] T055 [US3] Implement frontend hook for bank-account queries/mutations in `frontend/src/hooks/useBankAccounts.ts`
- [ ] T056 [US3] Implement settings page for bank-account label management in `frontend/src/pages/SettingsPage.tsx`

### Phase 5 Verification

- [ ] T119 Run US3 integration tests to confirm bank-account CRUD and attribution guard pass in `backend/tests/integration/us3_bank_accounts_test.go` (`go test ./tests/integration/ -run BankAccount`)
- [ ] T120 [P] Run frontend bank-account hook tests to confirm CRUD flow assertions pass in `frontend/src/hooks/useBankAccounts.test.ts` (`npm test -- --run useBankAccounts`)

**Checkpoint**: US3 independently functional and usable by statement flows.

---

## Phase 6: User Story 7 - Project-Based Collaboration and Access Control (Priority: P2)

**Goal**: Enforce project isolation and role-based permissions with collaboration endpoints.

**Independent Test**: Two projects + mixed roles demonstrate strict data isolation and permission boundaries.

### Tests for User Story 7

- [ ] T143 [P] [US7] Define onboarding project-members service and gRPC handler interfaces in `backend/internals/onboarding/interfaces/project_service.go`
- [ ] T144 [P] [US7] Generate mockgen mocks for onboarding service interfaces in `backend/internals/onboarding/mocks/`
- [ ] T057 [P] [US7] Add integration test for cross-project isolation on list/get endpoints in `backend/tests/integration/us7_project_isolation_test.go`
- [ ] T058 [P] [US7] Add integration test for role permission matrix (`read_only/update/write`) in `backend/tests/integration/us7_role_enforcement_test.go`
- [ ] T059 [P] [US7] Add hook test for project switching behavior in `frontend/src/hooks/useCurrentProject.test.ts`

### Implementation for User Story 7

- [ ] T060 [P] [US7] Implement onboarding service for member invite/role update in `backend/internals/onboarding/services/project_members_service.go`
- [ ] T061 [P] [US7] Implement onboarding grpc handlers for collaboration lifecycle in `backend/internals/onboarding/transport/grpc/server.go`
- [ ] T145 [US7] Wire otelgrpc unary and streaming server interceptors to onboarding gRPC server after T061 in `backend/internals/onboarding/transport/grpc/server.go`
- [ ] T062 [US7] Implement BFF projects controller (`current`, `invite`, `update-role`) in `backend/internals/bff/financial/controllers/projects_controller.go`
- [ ] T063 [US7] Integrate role checks into mutating controllers/services in `backend/internals/bff/transport/http/middleware/project_guard.go`
- [ ] T064 [US7] Implement frontend current-project and role-context hook in `frontend/src/hooks/useCurrentProject.ts`
- [ ] T065 [US7] Implement project switcher/invite UI entrypoints in `frontend/src/components/ProjectSwitcher.tsx`

### Phase 6 Verification

- [ ] T121 Run US7 integration tests to confirm project isolation and role enforcement pass in `backend/tests/integration/` (`go test ./tests/integration/ -run 'Isolation|Role'`)
- [ ] T122 [P] Run frontend current-project hook tests to confirm project-switching context passes in `frontend/src/hooks/useCurrentProject.test.ts` (`npm test -- --run useCurrentProject`)

**Checkpoint**: US7 independently validates isolation and collaboration rules.

---

## Phase 7: User Story 4 - Preferred Payment Day and Bill Payment Dashboard (Priority: P3)

**Goal**: Render outstanding bills and allow idempotent mark-as-paid workflow.

**Independent Test**: Configure preferred day, view dashboard, mark bill paid, verify outstanding list updates.

### Tests for User Story 4

- [ ] T146 [P] [US4] Define payment cycle and bill payment service interfaces in `backend/internals/payments/interfaces/payment_cycle_service.go` and `backend/internals/bills/interfaces/payment_service.go`
- [ ] T147 [P] [US4] Generate mockgen mocks for payments and bills service interfaces in `backend/internals/payments/mocks/` and `backend/internals/bills/mocks/`
- [ ] T066 [P] [US4] Add integration test for payment dashboard + overdue styling flags in `backend/tests/integration/us4_payment_dashboard_test.go`
- [ ] T067 [P] [US4] Add integration test for idempotent mark-paid endpoint in `backend/tests/integration/us4_mark_paid_idempotency_test.go`
- [ ] T068 [P] [US4] Add hook test for payment dashboard filtering/sorting in `frontend/src/hooks/usePaymentDashboard.test.ts`

### Implementation for User Story 4

- [X] T069 [P] [US4] Add bills v1 message fields for payment status and cycle views in `backend/protos/bills/v1/messages.proto`
- [X] T070 [P] [US4] Add bills v1 grpc methods for dashboard and mark-paid in `backend/protos/bills/v1/grpc.proto`
- [X] T071 [US4] Regenerate bills protobuf artifacts in `backend/protos/generated/`
- [ ] T158 [P] [US4] Add migration for project-scoped payment cycle preferences in `backend/payments/migrations/000002_create_payment_cycle_preferences.up.sql`
- [ ] T159 [P] [US4] Add down migration for payment cycle preferences in `backend/payments/migrations/000002_create_payment_cycle_preferences.down.sql`
- [ ] T072 [US4] Implement payment cycle preference repository/service in `backend/internals/payments/services/payment_cycle_service.go`
- [ ] T073 [US4] Implement bill payment service with idempotency-key enforcement in `backend/internals/bills/services/payment_service.go`
- [ ] T074 [US4] Implement BFF payments controller (`dashboard`, `mark-paid`, preferred day) in `backend/internals/bff/financial/controllers/payments_controller.go`
- [ ] T075 [US4] Implement frontend payment dashboard hook in `frontend/src/hooks/usePaymentDashboard.ts`
- [ ] T076 [US4] Implement frontend payment dashboard page in `frontend/src/pages/PaymentDashboardPage.tsx`

### Phase 7 Verification

- [ ] T123 Run US4 integration tests to confirm payment dashboard and idempotent mark-paid pass in `backend/tests/integration/` (`go test ./tests/integration/ -run 'Payment|Dashboard|MarkPaid'`)
- [ ] T124 [P] Run frontend payment dashboard hook tests to confirm filter/sort assertions pass in `frontend/src/hooks/usePaymentDashboard.test.ts` (`npm test -- --run usePaymentDashboard`)

**Checkpoint**: US4 independently delivers operational payment workflow.

---

## Phase 8: User Story 5 - Cross-Reference Account Statement with Bills (Priority: P4)

**Goal**: Automatically and manually reconcile statement transactions with bills.

**Independent Test**: Process statement + bills and verify matched/unmatched/ambiguous reconciliation behavior.

### Tests for User Story 5

- [ ] T148 [P] [US5] Define reconciliation service interface in `backend/internals/payments/interfaces/reconciliation_service.go`
- [ ] T149 [P] [US5] Generate mockgen mocks for payments service after adding reconciliation interface in `backend/internals/payments/mocks/`
- [ ] T077 [P] [US5] Add integration test for auto-reconciliation outcomes in `backend/tests/integration/us5_auto_reconciliation_test.go`
- [ ] T078 [P] [US5] Add integration test for manual reconciliation link creation in `backend/tests/integration/us5_manual_reconciliation_test.go`
- [ ] T079 [P] [US5] Add hook test for reconciliation summary and link mutation in `frontend/src/hooks/useReconciliation.test.ts`

### Implementation for User Story 5

- [ ] T080 [P] [US5] Add migration for reconciliation links and indexes in `backend/payments/migrations/000001_create_reconciliation_tables.up.sql`
- [ ] T081 [P] [US5] Add down migration for reconciliation tables in `backend/payments/migrations/000001_create_reconciliation_tables.down.sql`
- [ ] T082 [US5] Implement reconciliation matching service (auto + ambiguous routing) in `backend/internals/payments/services/reconciliation_service.go`
- [ ] T083 [US5] Implement BFF reconciliation controller (`summary`, `create-link`) in `backend/internals/bff/financial/controllers/reconciliation_controller.go`
- [ ] T084 [US5] Implement frontend reconciliation hook in `frontend/src/hooks/useReconciliation.ts`
- [ ] T085 [US5] Implement frontend reconciliation page in `frontend/src/pages/ReconciliationPage.tsx`

### Phase 8 Verification

- [ ] T125 Run US5 integration tests to confirm auto and manual reconciliation outcomes pass in `backend/tests/integration/` (`go test ./tests/integration/ -run Reconcil`)
- [ ] T126 [P] Run frontend reconciliation hook tests to confirm summary and link mutation assertions pass in `frontend/src/hooks/useReconciliation.test.ts` (`npm test -- --run useReconciliation`)

**Checkpoint**: US5 independently provides reconciliation oversight.

---

## Phase 9: User Story 6 - Financial History Dashboard (Priority: P5)

**Goal**: Provide monthly timeline, category breakdowns, and payment compliance analytics.

**Independent Test**: Seed historical data and verify timeline, monthly categories, and compliance calculations.

### Tests for User Story 6

- [ ] T150 [P] [US6] Define history repository interface in `backend/internals/payments/interfaces/history_repository.go`
- [ ] T151 [P] [US6] Generate mockgen mocks for payments service after adding history interface in `backend/internals/payments/mocks/`
- [ ] T086 [P] [US6] Add integration test for monthly timeline aggregation in `backend/tests/integration/us6_history_timeline_test.go`
- [ ] T087 [P] [US6] Add integration test for category breakdown + compliance metrics in `backend/tests/integration/us6_history_metrics_test.go`
- [ ] T088 [P] [US6] Add hook test for history query state and filters in `frontend/src/hooks/useHistoryDashboard.test.ts`

### Implementation for User Story 6

- [ ] T089 [US6] Implement analytics aggregation queries with project scoping in `backend/internals/payments/repositories/history_repository.go`
- [ ] T090 [US6] Implement BFF history controller (`timeline`, `categories`, `compliance`) in `backend/internals/bff/financial/controllers/history_controller.go`
- [ ] T091 [US6] Implement frontend history dashboard hook in `frontend/src/hooks/useHistoryDashboard.ts`
- [ ] T092 [US6] Implement frontend history dashboard page in `frontend/src/pages/HistoryDashboardPage.tsx`

### Phase 9 Verification

- [ ] T127 Run US6 integration tests to confirm timeline and metrics aggregation pass in `backend/tests/integration/` (`go test ./tests/integration/ -run 'History|Timeline|Compliance'`)
- [ ] T128 [P] Run frontend history hook tests to confirm query state and filter assertions pass in `frontend/src/hooks/useHistoryDashboard.test.ts` (`npm test -- --run useHistoryDashboard`)

**Checkpoint**: US6 independently delivers analytics/reporting value.

---

## Phase 10: Polish & Cross-Cutting Concerns

**Purpose**: Final hardening across all stories.

- [ ] T093 [P] Add OpenAPI operation metadata completeness checks in `backend/tests/integration/openapi_contract_test.go`
- [ ] T094 [P] Add BFF RED metrics middleware coverage test in `backend/tests/integration/bff_metrics_test.go`
- [ ] T095 Validate tokenized responsive theming and typography usage in `frontend/src/styles/tokens.ts`
- [ ] T096 Add quickstart validation script for local bootstrap workflow in `scripts/validate-financial-bill-organizer.sh`
- [ ] T097 Run end-to-end manual validation checklist updates in `specs/001-financial-bill-organizer/quickstart.md`

### Phase 10 Verification

- [ ] T129 Run full backend test suite to confirm all unit and integration tests pass in `backend/` (`go test ./...`)
- [ ] T130 [P] Run full frontend test suite to confirm all hook and component tests pass in `frontend/` (`npm test`)
- [ ] T131 [P] Run golangci-lint to confirm zero lint violations across all services in `backend/` (`make lint`)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: No dependencies.
- **Phase 2 (Foundational)**: Depends on Phase 1; blocks all user stories.
- **Phase 3+ (User Stories)**: Depend on Phase 2 completion.
- **Phase 10 (Polish)**: Depends on completed target user stories.

### User Story Dependencies

- **US1 (P1)**: Starts immediately after foundational phase.
- **US2 (P2)**: Depends on US1 ingestion pipeline and foundational async infrastructure.
- **US3 (P2)**: Depends on foundational phase only.
- **US7 (P2)**: Depends on foundational identity/project middleware and bootstrap schema.
- **US4 (P3)**: Depends on US1 + US2 extracted bill data.
- **US5 (P4)**: Depends on US2 statement extraction + US4 bill/payment data.
- **US6 (P5)**: Depends on US4 + US5 aggregated historical data.

### Within Each User Story

- Interface definitions come first (required by both mocks and implementations).
- Mock generation (`mockgen`) runs after interfaces and before tests.
- Tests are written next and must fail before implementation.
- Contracts/migrations before services.
- Services before controllers/endpoints.
- Backend endpoints before frontend hooks/pages consuming them.

### Parallel Opportunities

- Setup and foundational tasks marked `[P]` can run concurrently.
- In each story, test tasks and independent contract/migration tasks marked `[P]` can run concurrently.
- US3 and US7 can proceed in parallel after foundational completion.

---

## Parallel Example: User Story 1

```bash
# Step 1: Interface definitions and mock generation
T137 backend/internals/files/interfaces/document_service.go
T138 backend/internals/files/mocks/  (mockgen output)

# Step 2: Parallel test authoring (needs mocks from T138)
T025 backend/tests/integration/us1_upload_classify_test.go
T026 backend/internals/files/services/upload_service_test.go
T027 frontend/src/hooks/useUploadDocument.test.ts
T028 frontend/src/hooks/useClassifyDocument.test.ts

# Step 3: Parallel contract work (can overlap with tests)
T029 backend/protos/files/v1/messages.proto
T030 backend/protos/files/v1/grpc.proto
```

---

## Implementation Strategy

### MVP First (US1)

1. Complete Phase 1 and Phase 2.
2. Deliver US1 (upload/classification + pending state).
3. Validate US1 independently and demo.

### Incremental Delivery

1. US1 (MVP ingestion)
2. US2 + US3 + US7 (processing + settings + access control)
3. US4 (payment operations)
4. US5 (reconciliation)
5. US6 (history analytics)
6. Polish phase and release-readiness checks

### Parallel Team Strategy

1. Team A: platform/foundational backend + contracts
2. Team B: frontend shell/tokens/hooks
3. Team C: integration tests + observability/performance hardening
4. After Phase 2, split by story ownership with contract-first sync points
