# Tasks: Standardize Integration Test System

**Input**: Design documents from `/specs/005-standardize-integration-tests/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/integration-test-conventions.md, quickstart.md

**Tests**: This feature is test-centric. Validation and migration parity checks are mandatory.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Initialize migration workspace and baseline inventory for integration test standardization.

- [ ] T001 Create service-segment directories in backend/tests/integration/bff, backend/tests/integration/bills, backend/tests/integration/files, backend/tests/integration/identity, backend/tests/integration/onboarding, backend/tests/integration/payments, and backend/tests/integration/cross_service
- [ ] T002 Create migration traceability table in specs/005-standardize-integration-tests/migration-mapping.md with legacy path, target path, status, and coverage note columns
- [ ] T003 [P] Capture baseline integration suite inventory in specs/005-standardize-integration-tests/migration-baseline.md from backend/tests/integration/*.go

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Define standards and suite-wide helpers that all user stories depend on.

**CRITICAL**: No user-story migration should start until these tasks are complete.

- [ ] T004 Finalize canonical integration standard checklist in specs/005-standardize-integration-tests/contracts/integration-test-conventions.md
- [ ] T005 [P] Add compliance review checklist section to specs/005-standardize-integration-tests/quickstart.md
- [ ] T006 [P] Introduce shared BDD scenario struct and helper utilities in backend/tests/integration/helpers/scenario_helpers_test.go
- [ ] T007 [P] Implement testcontainers-go based ephemeral PostgreSQL lifecycle in backend/tests/integration/testmain_test.go with deterministic setup, migration application, and teardown
- [ ] T008 Record approved-library policy note in specs/005-standardize-integration-tests/research.md referencing testing, testify, and testcontainers-go
- [ ] T031 [P] Add integration-test convention validation script in backend/scripts/validate_integration_test_conventions.sh to enforce scenario structure and required file layout checks

**Checkpoint**: Foundation ready. US1, US2, and US3 can proceed.

---

## Phase 3: User Story 1 - Enforce One Integration Test Standard (Priority: P1) MVP

**Goal**: Standardize directory placement and behavior-based filenames for all existing integration tests.

**Independent Test**: A maintainer can locate test ownership and naming intent for every migrated file without using user story IDs.

### Tests for User Story 1

- [ ] T009 [US1] Validate migration map completeness by reconciling backend/tests/integration/*.go with specs/005-standardize-integration-tests/migration-mapping.md
- [ ] T010 [US1] Capture pre-migration integration package listing baseline in specs/005-standardize-integration-tests/migration-baseline.md from backend/tests/integration

### Implementation for User Story 1

- [ ] T011 [P] [US1] Migrate BFF integration tests to canonical names under backend/tests/integration/bff/ from backend/tests/integration/auth_token_rejection_test.go, backend/tests/integration/bff_metrics_test.go, and backend/tests/integration/openapi_contract_test.go
- [ ] T012 [P] [US1] Migrate files-domain integration tests to canonical names under backend/tests/integration/files/ from backend/tests/integration/us1_upload_classify_test.go, backend/tests/integration/us2_analysis_pipeline_test.go, and backend/tests/integration/us3_bank_accounts_test.go
- [ ] T013 [P] [US1] Migrate payments-domain integration tests to canonical names under backend/tests/integration/payments/ from backend/tests/integration/us4_mark_paid_idempotency_test.go and backend/tests/integration/us4_payment_dashboard_test.go
- [ ] T014 [P] [US1] Migrate reconciliation/history/project-guard scenarios to canonical names under backend/tests/integration/cross_service/ from backend/tests/integration/us5_auto_reconciliation_test.go, backend/tests/integration/us5_manual_reconciliation_test.go, backend/tests/integration/us6_history_metrics_test.go, backend/tests/integration/us6_history_timeline_test.go, backend/tests/integration/us7_project_isolation_test.go, and backend/tests/integration/us7_role_enforcement_test.go
- [ ] T015 [US1] Migrate identity contract test to canonical name under backend/tests/integration/identity/ from backend/tests/integration/identity_jwks_contract_test.go and update specs/005-standardize-integration-tests/migration-mapping.md statuses to moved
- [ ] T032 [US1] Run post-migration package listing check and compare against baseline in specs/005-standardize-integration-tests/migration-baseline.md

**Checkpoint**: All integration tests use canonical placement and behavior-based snake_case filenames.

---

## Phase 4: User Story 2 - Define BDD Integration Test Authoring Pattern (Priority: P2)

**Goal**: Normalize all migrated integration tests to table-driven BDD scenarios with explicit given/when/then and AAA readability.

**Independent Test**: Reviewers can read any migrated test and identify Given, When, Then, Arrange, Act, and Assert consistently.

### Tests for User Story 2

- [ ] T016 [US2] Add compliance verification cases for BDD scenario shape in backend/tests/integration/helpers/scenario_helpers_test.go
- [ ] T017 [US2] Execute backend integration suite with integration build tag from backend/tests/integration and record result in specs/005-standardize-integration-tests/migration-baseline.md

### Implementation for User Story 2

- [ ] T018 [P] [US2] Refactor BFF and identity integration tests to table-driven BDD scenarios in backend/tests/integration/bff/*.go and backend/tests/integration/identity/*.go
- [ ] T019 [P] [US2] Refactor files and payments integration tests to table-driven BDD scenarios in backend/tests/integration/files/*.go and backend/tests/integration/payments/*.go
- [ ] T020 [P] [US2] Refactor cross-service integration tests to table-driven BDD scenarios in backend/tests/integration/cross_service/*.go
- [ ] T021 [US2] Update specs/005-standardize-integration-tests/migration-mapping.md statuses to verified with coverage parity notes

**Checkpoint**: All migrated tests follow required BDD table-driven structure and pass suite validation.

---

## Phase 5: User Story 3 - Govern Future Compliance Through Project Rules (Priority: P3)

**Goal**: Encode integration-test standard in governance and Copilot instruction sources for future features.

**Independent Test**: New feature work receives explicit guidance to follow canonical integration test conventions without relying on tribal knowledge.

### Tests for User Story 3

- [ ] T022 [US3] Validate governance consistency across .specify/memory/constitution.md, .github/instructions/testing.instructions.md, and .github/instructions/ai-behavior.instructions.md against specs/005-standardize-integration-tests/contracts/integration-test-conventions.md

### Implementation for User Story 3

- [ ] T023 [P] [US3] Update backend integration testing policy in .specify/memory/constitution.md to require canonical directories, behavior-based filenames, and table-driven BDD scenarios
- [ ] T024 [P] [US3] Update enforcement rules in .github/instructions/testing.instructions.md for canonical placement, naming, and required given/when/then plus AAA structure
- [ ] T025 [P] [US3] Update AI deterministic-generation policy in .github/instructions/ai-behavior.instructions.md to require adherence to the integration-test standard when generating or editing integration tests
- [ ] T026 [US3] Add maintainer-facing compliance section to specs/005-standardize-integration-tests/quickstart.md linking governance rules and review expectations
- [ ] T033 [P] [US3] Update .specify/templates/spec-template.md to require integration-testing standards reference for backend behavior features
- [ ] T034 [P] [US3] Update .specify/templates/plan-template.md and .specify/templates/tasks-template.md to include integration-testing compliance gates for future feature planning

**Checkpoint**: Governance and instruction systems enforce the new integration-testing standard for future feature delivery.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Final consistency checks, documentation cleanup, and handoff readiness.

- [ ] T027 [P] Normalize package comments and file headers in backend/tests/integration/**/*.go after migration/refactor
- [ ] T028 [P] Validate task-to-artifact traceability by reconciling specs/005-standardize-integration-tests/spec.md, specs/005-standardize-integration-tests/plan.md, and specs/005-standardize-integration-tests/tasks.md
- [ ] T029 Run full backend test command in backend via go test ./... and capture integration-related outcomes in specs/005-standardize-integration-tests/migration-baseline.md
- [ ] T030 Run quickstart workflow verification from specs/005-standardize-integration-tests/quickstart.md and record final adoption notes in specs/005-standardize-integration-tests/research.md

---

## Dependencies & Execution Order

### Phase Dependencies

- Setup (Phase 1): No dependencies.
- Foundational (Phase 2): Depends on Phase 1 and blocks all user-story implementation.
- User Story phases (Phase 3-5): Depend on Phase 2 completion.
- Polish (Phase 6): Depends on selected user stories being complete.

### User Story Dependencies

- US1 (P1): Starts after Foundational; no dependency on other user stories.
- US2 (P2): Starts after US1 file migration tasks complete (T011-T015).
- US3 (P3): Starts after Foundational and can run in parallel with US2, but final governance validation (T022) should occur after T023-T026 and T033-T034 edits.

### Within Each User Story

- Validation task(s) first for baseline and traceability.
- Structural/file migration before BDD refactor.
- BDD refactor before verification status updates.
- Baseline package listing before migration and post-migration listing comparison before story sign-off.
- Governance file updates before governance consistency validation.

### Parallel Opportunities

- Phase 1: T003 can run in parallel with T001/T002.
- Phase 2: T005, T006, T007, and T008 can run in parallel after T004 starts.
- US1: T011-T015 are parallelizable by service/cross-service segment.
- US2: T018-T020 are parallelizable by segment.
- US3: T023-T025 and T033-T034 are parallelizable by file.
- Polish: T027 and T028 can run in parallel before T029/T030.

---

## Parallel Example: User Story 1

```bash
# Parallel migration streams after mapping is prepared
Task T011: Migrate BFF integration files to backend/tests/integration/bff/
Task T012: Migrate files integration files to backend/tests/integration/files/
Task T013: Migrate payments integration files to backend/tests/integration/payments/
Task T014: Migrate cross-service integration files to backend/tests/integration/cross_service/
Task T015: Migrate identity integration files to backend/tests/integration/identity/
```

---

## Parallel Example: User Story 2

```bash
# Parallel BDD refactor streams by ownership segment
Task T018: Refactor backend/tests/integration/bff/*.go and backend/tests/integration/identity/*.go
Task T019: Refactor backend/tests/integration/files/*.go and backend/tests/integration/payments/*.go
Task T020: Refactor backend/tests/integration/cross_service/*.go
```

---

## Implementation Strategy

### MVP First (User Story 1)

1. Complete Setup (Phase 1).
2. Complete Foundational tasks (Phase 2).
3. Complete US1 migration and naming standardization (Phase 3).
4. Validate independent test criteria for US1 before expanding scope.

### Incremental Delivery

1. Deliver US1 canonical placement + naming.
2. Deliver US2 BDD table-driven normalization and verification.
3. Deliver US3 governance enforcement updates.
4. Finish with cross-cutting validation and quickstart verification.

### Parallel Team Strategy

1. Team A: US1 structural migration by service segment.
2. Team B: US2 BDD refactor after each segment lands.
3. Team C: US3 governance and instruction updates in parallel.
4. Integrate with Phase 6 validation gates before merge.

---

## Notes

- [P] tasks indicate no direct file-level conflict and can run concurrently.
- [US1], [US2], [US3] labels maintain traceability to prioritized user stories.
- Every task includes explicit file paths for execution clarity.
- Keep commits scoped by task group to simplify review and rollback if needed.
