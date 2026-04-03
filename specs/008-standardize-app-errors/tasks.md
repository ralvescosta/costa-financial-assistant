# Tasks: Standardize Backend App Errors

**Input**: Design documents from `/specs/008-standardize-app-errors/`
**Prerequisites**: `plan.md`, `spec.md`, `research.md`, `data-model.md`, `contracts/backend-error-propagation-contract.md`

**Tests**: Included. `spec.md` requires behavior coverage for propagation, retryability, and non-leakage (FR-010).

**Organization**: Tasks are grouped by user story so each story can be implemented and validated independently.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependency on incomplete tasks)
- **[Story]**: User story label (`[US1]`, `[US2]`, `[US3]`) for story-phase tasks only

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Establish shared backend error primitives and baseline implementation scaffolding.

- [ ] T001 Audit current backend error-leak points and record baseline in specs/008-standardize-app-errors/contracts/current-error-leaks.md
- [ ] T002 Expand shared error contract behavior in backend/pkgs/errors/error.go
- [ ] T003 [P] Create translation helper primitives in backend/pkgs/errors/translate.go
- [ ] T004 [P] Expand centralized error catalog entries in backend/pkgs/errors/consts.go
- [ ] T005 [P] Add package-level unit tests for shared error primitives in backend/pkgs/errors/error_test.go

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Define cross-service rules and shared mapping utilities that all stories depend on.

**⚠️ CRITICAL**: No user story work starts before this phase is complete.

- [ ] T006 Define deterministic translation policy and category mapping in backend/pkgs/errors/mapping.go
- [ ] T007 [P] Add shared SQL and gRPC native-error classification helpers in backend/pkgs/errors/native_classifiers.go
- [ ] T008 [P] Add reusable AppError test assertions for integration suites in backend/tests/integration/helpers/assert_app_error.go
- [ ] T009 [P] Define service adoption checklist in specs/008-standardize-app-errors/contracts/service-adoption-checklist.md
- [ ] T010 Create implementation progress matrix for all backend services in specs/008-standardize-app-errors/contracts/service-coverage-matrix.md

**Checkpoint**: Shared policy and mapping foundation complete.

---

## Phase 3: User Story 1 - Consistent Error Contract Across Layers (Priority: P1) 🎯 MVP

**Goal**: Enforce `AppError` as the only error type crossing backend layer boundaries.

**Independent Test**: Trigger repository/service/transport failures and verify no raw dependency error crosses boundaries.

### Tests for User Story 1

- [ ] T011 [P] [US1] Add files-repository translation tests in backend/internals/files/repositories/error_translation_test.go
- [ ] T012 [P] [US1] Add cross-service propagation integration test in backend/tests/integration/cross_service/app_error_propagation_test.go
- [ ] T061 [P] [US1] Add async publisher propagation integration test in backend/tests/integration/cross_service/app_error_async_publisher_propagation_test.go

### Implementation for User Story 1

- [ ] T013 [US1] Implement repository-to-service translation in backend/internals/files/repositories/document_repository.go
- [ ] T014 [US1] Implement repository-to-service translation in backend/internals/bills/repositories/payment_repository.go
- [ ] T015 [US1] Implement repository-to-service translation in backend/internals/onboarding/repositories/project_members_repository.go
- [ ] T016 [US1] Implement repository-to-service translation in backend/internals/payments/repositories/payment_cycle_repository.go
- [ ] T017 [US1] Implement repository-to-service translation in backend/internals/payments/repositories/reconciliation_repository.go
- [ ] T018 [US1] Implement service boundary propagation contract in backend/internals/files/services/document_service.go
- [ ] T019 [US1] Implement service boundary propagation contract in backend/internals/bills/services/payment_service.go
- [ ] T020 [US1] Implement service boundary propagation contract in backend/internals/onboarding/services/project_members_service.go
- [ ] T021 [US1] Implement transport boundary sanitization in backend/internals/files/transport/grpc/server.go
- [ ] T022 [US1] Implement transport boundary sanitization in backend/internals/bills/transport/grpc/server.go
- [ ] T023 [US1] Implement transport boundary sanitization in backend/internals/onboarding/transport/grpc/server.go
- [ ] T024 [US1] Implement transport boundary sanitization in backend/internals/identity/transport/grpc/server.go
- [ ] T025 [US1] Implement async consumer boundary sanitization in backend/internals/files/transport/rmq/analysis_consumer.go
- [ ] T062 [US1] Implement async producer boundary sanitization for affected RMQ publisher paths in backend/internals/files/transport/rmq/

**Checkpoint**: `AppError` is the only cross-layer propagated error contract for P1 paths.

---

## Phase 4: User Story 2 - Dependency Error Logging and Sanitization (Priority: P2)

**Goal**: Ensure one structured boundary log of native errors before propagation as sanitized `AppError`.

**Independent Test**: Trigger DB/gRPC failures and verify one boundary `zap.Error(err)` log plus sanitized propagated contract.

### Tests for User Story 2

- [ ] T026 [P] [US2] Add boundary-logging unit tests for file services in backend/internals/files/services/error_logging_test.go
- [ ] T027 [P] [US2] Add boundary-logging unit tests for payments services in backend/internals/payments/services/error_logging_test.go

### Implementation for User Story 2

- [ ] T028 [US2] Apply one-boundary logging in backend/internals/files/services/extraction_service.go
- [ ] T029 [US2] Apply one-boundary logging in backend/internals/files/services/bank_account_service.go
- [ ] T030 [US2] Apply one-boundary logging in backend/internals/bff/services/documents_service.go
- [ ] T031 [US2] Apply one-boundary logging in backend/internals/bff/services/payments_service.go
- [ ] T032 [US2] Apply one-boundary logging in backend/internals/bff/services/reconciliation_service.go
- [ ] T033 [US2] Apply one-boundary logging in backend/internals/bff/services/history_service.go
- [ ] T034 [US2] Apply one-boundary logging in backend/internals/identity/services/token_service.go
- [ ] T035 [US2] Apply one-boundary logging in backend/internals/payments/services/payment_cycle_service.go
- [ ] T036 [US2] Apply one-boundary logging in backend/internals/payments/services/reconciliation_service.go

**Checkpoint**: Native dependency errors are logged once at translation boundaries and never leaked.

---

## Phase 5: User Story 3 - Retryability Classification for Future Policies (Priority: P3)

**Goal**: Classify all cataloged error entries as retryable/non-retryable with mandatory unknown fallback.

**Independent Test**: Validate all catalog entries include retryability and unknown fallback behavior is deterministic.

### Tests for User Story 3

- [ ] T037 [P] [US3] Add retryability coverage tests for centralized catalog in backend/pkgs/errors/consts_retryability_test.go
- [ ] T038 [P] [US3] Add unknown-fallback integration test in backend/tests/integration/cross_service/app_error_unknown_fallback_test.go
- [ ] T063 [P] [US3] Add wrapped and nil-native-error translation tests in backend/pkgs/errors/translate_test.go verifying errors.Is/errors.As compatibility and deterministic fallback behavior

### Implementation for User Story 3

- [ ] T039 [US3] Implement explicit retryability categories for all catalog entries in backend/pkgs/errors/consts.go
- [ ] T040 [US3] Implement unknown-fallback translation behavior in backend/pkgs/errors/translate.go
- [ ] T041 [US3] Apply retryability-aware translation for gRPC dependency failures in backend/internals/bff/services/projects_service.go
- [ ] T042 [US3] Apply retryability-aware translation for database dependency failures in backend/internals/payments/repositories/history_repository.go
- [ ] T043 [US3] Apply retryability-aware translation for unit-of-work failures in backend/internals/files/repositories/unit_of_work.go

**Checkpoint**: Retryability classification and unknown fallback rules are fully enforced.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Final verification and documentation alignment.

- [ ] T044 [P] Update implementation notes and validation evidence in specs/008-standardize-app-errors/quickstart.md
- [ ] T045 Run full backend test suite and targeted service suites for error-standardization changes from backend/go.mod scope
- [ ] T046 [P] Update coverage matrix with final status in specs/008-standardize-app-errors/contracts/service-coverage-matrix.md
- [ ] T064 Add cross-service structured logging verification for bff, bills, identity, and onboarding failure traces in backend/tests/integration/cross_service/app_error_boundary_logging_test.go
- [ ] T067 Configure CI gate to validate zero dependency-error leakage across layer-boundary test coverage and document enforcement evidence in specs/008-standardize-app-errors/contracts/ci-enforcement-config.md

---

## Phase 7: Mandatory Governance Sync (Blocking)

**Purpose**: Keep memory-flow and instruction guidance synchronized with implemented behavior.

- [ ] T047 Update BFF flow error propagation guidance in .specify/memory/bff-flows.md
- [ ] T048 [P] Update files service flow error propagation guidance in .specify/memory/files-service-flows.md
- [ ] T049 [P] Update bills service flow error propagation guidance in .specify/memory/bills-service-flows.md
- [ ] T050 [P] Update identity service flow error propagation guidance in .specify/memory/identity-service-flows.md
- [ ] T051 [P] Update onboarding service flow error propagation guidance in .specify/memory/onboarding-service-flows.md
- [ ] T052 Verify architecture topology impact and record no-change or update in .specify/memory/architecture-diagram.md
- [ ] T053 Update observability error-logging rules in .github/instructions/observability.instructions.md
- [ ] T054 [P] Update security non-leakage rules in .github/instructions/security.instructions.md
- [ ] T055 [P] Update Go error-handling conventions in .github/instructions/golang.instructions.md
- [ ] T056 Validate backend integration test naming/placement conventions via backend/scripts/validate_integration_test_conventions.sh
- [ ] T057 Document workflow-template impact decision in specs/008-standardize-app-errors/contracts/workflow-template-impact.md
- [ ] T058 Update deterministic AI implementation rules to enforce AppError-first propagation for all future backend features in .github/instructions/ai-behavior.instructions.md
- [ ] T059 Update architecture governance rules to mandate AppError translation boundaries between backend layers in .github/instructions/architecture.instructions.md
- [ ] T060 Update constitution memory governance to codify AppError standard as a required backend implementation pattern in .specify/memory/constitution.md
- [ ] T065 Apply constitution amendment procedure for AppError governance updates in .specify/memory/constitution.md (semantic version bump + SYNC IMPACT REPORT update)
- [ ] T066 [P] Re-validate dependent templates/prompts after constitution update and record evidence in specs/008-standardize-app-errors/contracts/workflow-template-impact.md

**Checkpoint**: Feature is not complete until this phase is complete.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: starts immediately.
- **Phase 2 (Foundational)**: depends on Phase 1 completion and blocks all user stories.
- **Phase 3 (US1)**: depends on Phase 2; delivers MVP.
- **Phase 4 (US2)**: depends on Phase 2 and can run after or in parallel with US1 when staffing allows.
- **Phase 5 (US3)**: depends on Phase 2 and can run after or in parallel with US1/US2 when staffing allows.
- **Phase 6 (Polish)**: depends on completion of selected story phases; includes T067 CI enforcement validation.
- **Phase 7 (Governance Sync)**: depends on all implementation phases (including T067 completion) and is merge-blocking.

### User Story Dependencies

- **US1 (P1)**: independent after foundational completion; no dependency on US2/US3.
- **US2 (P2)**: independent after foundational completion; can be validated without US3.
- **US3 (P3)**: independent after foundational completion; can be validated without US1/US2 business flows.

### Within Each User Story

- **All test tasks MUST complete before implementation tasks begin** (e.g., T011+T012+T061 must be done before T013–T062 in US1).
- Repository/service translation before transport adaptation where both are involved.
- Complete story checkpoint before declaring story done.

---

## Parallel Execution Examples

### User Story 1

- Run in parallel: T011, T012, and T061 (all test tasks together).
- **After ALL tests complete** (T011 AND T012 AND T061): run T013, T014, T015, T016, T017, and T062 in parallel.

### User Story 2

- Run in parallel: T026 and T027.
- Run in parallel after logging policy is stable: T028, T029, T030, T031, T032, T033, T034, T035, T036.

### User Story 3

- Run in parallel: T037, T038, and T063.
- Run in parallel after catalog updates: T041, T042, T043.

### Governance Sync

- Run in parallel: T048, T049, T050, T051.
- Run in parallel: T054, T055, T058, T059, and T066.

---

## Implementation Strategy

### MVP First (US1)

1. Complete Phases 1 and 2.
2. Complete US1 (Phase 3).
3. Validate propagation contract with T011 and T012.
4. Demonstrate no dependency-error leakage on MVP paths.

### Incremental Delivery

1. Deliver US1 as MVP.
2. Add US2 boundary logging behavior.
3. Add US3 retryability classification and unknown fallback hardening.
4. Finalize with Phase 6 and Phase 7 mandatory sync.

### Team Parallelization

1. One engineer owns shared package/error catalog tasks (T001-T010).
2. Service/domain engineers implement US1/US2 in parallel by domain.
3. Governance owner executes Phase 7 once implementation converges.

---

## Notes

- `[P]` tasks are safe to run concurrently when dependencies are respected.
- Every story task includes `[USx]` label for traceability.
- Keep commits scoped to logical task groups to simplify review.
