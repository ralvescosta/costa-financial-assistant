# Tasks: Enforce Service Boundary Contracts

**Input**: Design documents from /specs/009-fix-bff-service-boundary/
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/bff-service-boundary-contract.md

**Tests**: Included. The specification requires route/controller and service-level behavior verification (FR-010).

**Organization**: Tasks are grouped by user story so each story can be implemented and validated independently.

## Format: [ID] [P?] [Story] Description

- [P]: Can run in parallel (different files, no dependency on incomplete tasks)
- [Story]: User story label ([US1], [US2], [US3]) for story-phase tasks only

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Prepare tracking artifacts and baseline mapping inventory for the refactor.

- [X] T001 Create BFF boundary baseline inventory in specs/009-fix-bff-service-boundary/contracts/bff-boundary-baseline.md
- [X] T002 [P] Create active-route service mapping matrix in specs/009-fix-bff-service-boundary/contracts/bff-route-service-matrix.md
- [X] T003 [P] Create pointer-policy adoption matrix in specs/009-fix-bff-service-boundary/contracts/pointer-policy-adoption-matrix.md
- [X] T004 Create implementation decision log scaffold in specs/009-fix-bff-service-boundary/contracts/implementation-decisions.md

**Checkpoint**: Setup artifacts exist and implementation scope is traceable.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Establish shared contracts and mapper scaffolding required by all user stories.

**CRITICAL**: No user story work begins before this phase is complete.

- [X] T005 Add transport-agnostic BFF service contract package bootstrap in backend/internals/bff/services/contracts/contracts.go
- [X] T006 [P] Add service-contract mapper helper bootstrap in backend/internals/bff/transport/http/controllers/mappers/contracts_mapper.go
- [X] T007 [P] Define BFF contract ownership guard tests in backend/internals/bff/services/contract_ownership_guard_test.go
- [X] T008 Update BFF service interfaces to consume transport-agnostic contracts in backend/internals/bff/interfaces/services.go
- [X] T009 [P] Add pointer-policy helper and exception annotation primitives in backend/pkgs/errors/pointer_policy.go
- [X] T010 Document exception-record convention for value semantics in specs/009-fix-bff-service-boundary/contracts/pointer-exceptions.md

**Checkpoint**: Shared contracts, mappers, and policy primitives are in place.

---

## Phase 3: User Story 1 - Enforce BFF Layer Contract (Priority: P1) MVP

**Goal**: Remove HTTP view leakage from all active BFF services and enforce transport-to-service mapping.

**Independent Test**: Verify all active BFF service signatures are transport-agnostic and all routes remain reachable with expected behavior.

### Tests for User Story 1

- [X] T011 [P] [US1] Add/extend documents route registration and reachability assertions in backend/tests/integration/bff/documents_routes_registration_test.go
- [X] T012 [P] [US1] Add/extend history route registration and reachability assertions in backend/tests/integration/bff/history_routes_registration_test.go
- [X] T013 [P] [US1] Add/extend payments route registration and reachability assertions in backend/tests/integration/bff/payments_routes_registration_test.go
- [X] T014 [P] [US1] Add/extend projects route registration and reachability assertions in backend/tests/integration/bff/projects_routes_registration_test.go
- [X] T015 [P] [US1] Add/extend reconciliation route registration and reachability assertions in backend/tests/integration/bff/reconciliation_routes_registration_test.go
- [X] T016 [P] [US1] Add/extend settings route registration and reachability assertions in backend/tests/integration/bff/settings_routes_registration_test.go
- [X] T088 [P] [US1] Add documents endpoint status/response-semantics assertions in backend/tests/integration/bff/documents_routes_registration_test.go
- [X] T089 [P] [US1] Add history endpoint status/response-semantics assertions in backend/tests/integration/bff/history_routes_registration_test.go
- [X] T090 [P] [US1] Add payments endpoint status/response-semantics assertions in backend/tests/integration/bff/payments_routes_registration_test.go
- [X] T091 [P] [US1] Add projects endpoint status/response-semantics assertions in backend/tests/integration/bff/projects_routes_registration_test.go
- [X] T092 [P] [US1] Add reconciliation endpoint status/response-semantics assertions in backend/tests/integration/bff/reconciliation_routes_registration_test.go
- [X] T093 [P] [US1] Add settings endpoint status/response-semantics assertions in backend/tests/integration/bff/settings_routes_registration_test.go
- [X] T017 [P] [US1] Add service-boundary contract tests for documents and history services in backend/internals/bff/services/documents_service_test.go
- [X] T018 [P] [US1] Add service-boundary contract tests for payments and reconciliation services in backend/internals/bff/services/payments_service_test.go
- [X] T019 [P] [US1] Add service-boundary contract tests for projects and settings services in backend/internals/bff/services/projects_service_test.go
- [X] T094 [P] [US1] Add nil-safety boundary tests for documents and history mappers in backend/internals/bff/transport/http/controllers/mappers/documents_mapper_test.go
- [X] T095 [P] [US1] Add nil-safety boundary tests for payments/projects/reconciliation/settings mappers in backend/internals/bff/transport/http/controllers/mappers/payments_mapper_test.go

### Implementation for User Story 1

- [X] T020 [P] [US1] Define documents service contracts in backend/internals/bff/services/contracts/documents_contracts.go
- [X] T021 [P] [US1] Define history service contracts in backend/internals/bff/services/contracts/history_contracts.go
- [X] T022 [P] [US1] Define payments service contracts in backend/internals/bff/services/contracts/payments_contracts.go
- [X] T023 [P] [US1] Define projects service contracts in backend/internals/bff/services/contracts/projects_contracts.go
- [X] T024 [P] [US1] Define reconciliation service contracts in backend/internals/bff/services/contracts/reconciliation_contracts.go
- [X] T025 [P] [US1] Define settings service contracts in backend/internals/bff/services/contracts/settings_contracts.go
- [X] T026 [US1] Refactor documents service to remove HTTP view dependencies in backend/internals/bff/services/documents_service.go
- [X] T027 [US1] Refactor history service to remove HTTP view dependencies in backend/internals/bff/services/history_service.go
- [X] T028 [US1] Refactor payments service to remove HTTP view dependencies in backend/internals/bff/services/payments_service.go
- [X] T029 [US1] Refactor projects service to remove HTTP view dependencies in backend/internals/bff/services/projects_service.go
- [X] T030 [US1] Refactor reconciliation service to remove HTTP view dependencies in backend/internals/bff/services/reconciliation_service.go
- [X] T031 [US1] Refactor settings service to remove HTTP view dependencies in backend/internals/bff/services/settings_service.go
- [X] T032 [US1] Implement documents transport-to-service mappers in backend/internals/bff/transport/http/controllers/mappers/documents_mapper.go
- [X] T033 [US1] Implement history transport-to-service mappers in backend/internals/bff/transport/http/controllers/mappers/history_mapper.go
- [X] T034 [US1] Implement payments transport-to-service mappers in backend/internals/bff/transport/http/controllers/mappers/payments_mapper.go
- [X] T035 [US1] Implement projects transport-to-service mappers in backend/internals/bff/transport/http/controllers/mappers/projects_mapper.go
- [X] T036 [US1] Implement reconciliation transport-to-service mappers in backend/internals/bff/transport/http/controllers/mappers/reconciliation_mapper.go
- [X] T037 [US1] Implement settings transport-to-service mappers in backend/internals/bff/transport/http/controllers/mappers/settings_mapper.go
- [X] T096 [US1] Implement nil and empty-boundary guards in documents mapper flows in backend/internals/bff/transport/http/controllers/mappers/documents_mapper.go
- [X] T097 [US1] Implement nil and empty-boundary guards in history mapper flows in backend/internals/bff/transport/http/controllers/mappers/history_mapper.go
- [X] T098 [US1] Implement nil and empty-boundary guards in payments mapper flows in backend/internals/bff/transport/http/controllers/mappers/payments_mapper.go
- [X] T099 [US1] Implement nil and empty-boundary guards in projects mapper flows in backend/internals/bff/transport/http/controllers/mappers/projects_mapper.go
- [X] T100 [US1] Implement nil and empty-boundary guards in reconciliation mapper flows in backend/internals/bff/transport/http/controllers/mappers/reconciliation_mapper.go
- [X] T101 [US1] Implement nil and empty-boundary guards in settings mapper flows in backend/internals/bff/transport/http/controllers/mappers/settings_mapper.go
- [X] T038 [US1] Update documents controller to invoke mappers and service contracts in backend/internals/bff/transport/http/controllers/documents_controller.go
- [X] T039 [US1] Update history controller to invoke mappers and service contracts in backend/internals/bff/transport/http/controllers/history_controller.go
- [X] T040 [US1] Update payments controller to invoke mappers and service contracts in backend/internals/bff/transport/http/controllers/payments_controller.go
- [X] T080 [US1] Update projects controller to invoke mappers and service contracts in backend/internals/bff/transport/http/controllers/projects_controller.go
- [X] T081 [US1] Update reconciliation controller to invoke mappers and service contracts in backend/internals/bff/transport/http/controllers/reconciliation_controller.go
- [X] T082 [US1] Update settings controller to invoke mappers and service contracts in backend/internals/bff/transport/http/controllers/settings_controller.go
- [X] T041 [US1] Align route capability signatures with transport view contracts only in backend/internals/bff/transport/http/routes/contracts.go
- [X] T042 [US1] Validate and adjust documents route registrations for updated controller signatures in backend/internals/bff/transport/http/routes/documents_routes.go
- [X] T083 [US1] Validate and adjust history route registrations for updated controller signatures in backend/internals/bff/transport/http/routes/history_routes.go
- [X] T084 [US1] Validate and adjust payments route registrations for updated controller signatures in backend/internals/bff/transport/http/routes/payments_routes.go
- [X] T085 [US1] Validate and adjust projects route registrations for updated controller signatures in backend/internals/bff/transport/http/routes/projects_routes.go
- [X] T086 [US1] Validate and adjust reconciliation route registrations for updated controller signatures in backend/internals/bff/transport/http/routes/reconciliation_routes.go
- [X] T087 [US1] Validate and adjust settings route registrations for updated controller signatures in backend/internals/bff/transport/http/routes/settings_routes.go
- [X] T043 [US1] Refresh route registration smoke inventory for all active operations in backend/tests/integration/bff/bff_route_registration_smoke_test.go

**Checkpoint**: All active BFF routes/services use strict transport-to-service mapping, service layer is transport-agnostic, and endpoint behavior semantics are preserved.

---

## Phase 4: User Story 2 - Reduce Struct Copy Overhead (Priority: P2)

**Goal**: Apply measurable pointer policy across modified backend boundaries with explicit exception handling.

**Independent Test**: Validate pointer-threshold compliance on touched boundaries and document every approved value-semantic exception.

### Tests for User Story 2

- [X] T044 [P] [US2] Add pointer-policy compliance tests for BFF service contracts in backend/internals/bff/services/pointer_policy_test.go
- [X] T045 [P] [US2] Add pointer-policy compliance tests for files and bills service boundaries in backend/internals/files/services/pointer_policy_test.go
- [X] T046 [P] [US2] Add pointer-policy compliance tests for identity, onboarding, and payments boundaries in backend/internals/payments/services/pointer_policy_test.go

### Implementation for User Story 2

- [X] T047 [US2] Apply pointer-signature policy to BFF service contract boundaries in backend/internals/bff/interfaces/services.go
- [X] T048 [US2] Apply pointer-signature policy to files service boundaries in backend/internals/files/services/document_service.go
- [X] T049 [US2] Apply pointer-signature policy to bills service boundaries in backend/internals/bills/services/payment_service.go
- [X] T050 [US2] Apply pointer-signature policy to identity service boundaries in backend/internals/identity/services/token_service.go
- [X] T051 [US2] Apply pointer-signature policy to onboarding service boundaries in backend/internals/onboarding/services/project_members_service.go
- [X] T052 [US2] Apply pointer-signature policy to payments service boundaries in backend/internals/payments/services/payment_cycle_service.go
- [X] T053 [US2] Record approved value-semantic exceptions with rationale in specs/009-fix-bff-service-boundary/contracts/pointer-exceptions.md
- [X] T054 [US2] Update pointer-policy adoption matrix with per-module status in specs/009-fix-bff-service-boundary/contracts/pointer-policy-adoption-matrix.md

**Checkpoint**: Pointer-threshold policy is consistently applied on scoped backend boundaries with explicit exception traceability.

---

## Phase 5: User Story 3 - Preserve Refactor Rules in Project Guidance (Priority: P3)

**Goal**: Encode boundary and pointer conventions in instructions and memory artifacts for future deterministic reuse.

**Independent Test**: Confirm instructions and memory artifacts explicitly encode the implemented rules and update obligations.

### Tests for User Story 3

- [X] T055 [P] [US3] Add governance consistency checklist for memory and instruction sync in specs/009-fix-bff-service-boundary/contracts/governance-sync-checklist.md
- [X] T056 [P] [US3] Add integration convention validation evidence record in specs/009-fix-bff-service-boundary/contracts/integration-convention-validation.md

### Implementation for User Story 3

- [X] T057 [US3] Update architecture boundary rules for BFF transport/service separation in .github/instructions/architecture.instructions.md
- [X] T058 [US3] Update project-structure conventions for service-owned contracts and mapper placement in .github/instructions/project-structure.instructions.md
- [X] T059 [US3] Update Go conventions for pointer-threshold signature policy in .github/instructions/golang.instructions.md
- [X] T060 [US3] Update coding conventions for documented value-semantics exceptions in .github/instructions/coding-conventions.instructions.md
- [X] T061 [US3] Update testing conventions for boundary and route verification requirements in .github/instructions/testing.instructions.md
- [X] T062 [US3] Update deterministic AI behavior rules for boundary and pointer policy preservation in .github/instructions/ai-behavior.instructions.md
- [X] T063 [US3] Update BFF flow ownership and mapper responsibilities in .specify/memory/bff-flows.md
- [X] T064 [US3] Update cross-service flow guidance for pointer policy in .specify/memory/files-service-flows.md
- [X] T065 [US3] Update cross-service flow guidance for pointer policy in .specify/memory/bills-service-flows.md
- [X] T066 [US3] Update cross-service flow guidance for pointer policy in .specify/memory/identity-service-flows.md
- [X] T067 [US3] Update cross-service flow guidance for pointer policy in .specify/memory/onboarding-service-flows.md
- [X] T068 [US3] Add or update payments flow guidance for pointer policy in .specify/memory/payments-service-flows.md
- [X] T069 [US3] Update architecture diagram ownership notes for refactored boundaries in .specify/memory/architecture-diagram.md
- [X] T070 [US3] Add or update repository memory note for BFF boundary conventions in /memories/repo/bff-service-boundary-conventions.md

**Checkpoint**: Instruction and memory guidance preserves the implemented architecture and policy conventions.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Final cross-story validation and release-readiness checks.

- [X] T071 [P] Update validation evidence and command outcomes in specs/009-fix-bff-service-boundary/quickstart.md
- [X] T072 Run full backend regression suite from module root in backend/go.mod scope via go test ./...
- [X] T073 [P] Run canonical integration convention validator and record outcomes in backend/scripts/validate_integration_test_conventions.sh
- [X] T074 [P] Update implementation decision log with final trade-offs in specs/009-fix-bff-service-boundary/contracts/implementation-decisions.md

---

## Phase 7: Mandatory Governance Sync (Blocking)

**Purpose**: Ensure memory and instruction artifacts remain synchronized with delivered implementation.

- [X] T075 Update architecture-diagram maintenance metadata and change-trigger notes only in .specify/memory/architecture-diagram-maintenance.md
- [X] T076 [P] Verify architecture-diagram reference consistency in .specify/memory/architecture-diagram-reference.md
- [X] T077 [P] Confirm workflow-template impact decision and update if needed in .specify/templates/spec-template.md
- [X] T078 [P] Confirm workflow-template impact decision and update if needed in .specify/templates/plan-template.md
- [X] T079 Verify canonical backend integration test placement/naming and BDD AAA compliance in backend/tests/integration/bff/bff_route_registration_smoke_test.go

**Checkpoint**: Feature cannot be marked complete until this phase is complete.

---

## Dependencies & Execution Order

### Phase Dependencies

- Phase 1 (Setup): no dependencies.
- Phase 2 (Foundational): depends on Phase 1; blocks all user stories.
- Phase 3 (US1): depends on Phase 2; delivers MVP.
- Phase 4 (US2): depends on Phase 2; can run after US1 or in parallel once capacity allows.
- Phase 5 (US3): depends on Phase 2; can run after US1 or in parallel once implementation stabilizes.
- Phase 6 (Polish): depends on selected story completion.
- Phase 7 (Mandatory Governance Sync): depends on all implementation phases and blocks completion.

### User Story Dependencies

- US1 (P1): independent after foundational completion.
- US2 (P2): independent after foundational completion; may share files with US1 and should be sequenced by file ownership.
- US3 (P3): independent after foundational completion; should start when behavior-impacting refactors settle.

### Within Each User Story

- Test tasks first, implementation second.
- Service contract definitions before service refactors.
- Mapper implementation before controller adaptation.
- Nil-safety guards before controller adaptation.
- Controller and route alignment before smoke/integration finalization.

---

## Parallel Execution Examples

### User Story 1

- Parallel test batch: T011, T012, T013, T014, T015, T016, T017, T018, T019, T088, T089, T090, T091, T092, T093, T094, T095
- Parallel contract batch: T020, T021, T022, T023, T024, T025

### User Story 2

- Parallel test batch: T044, T045, T046
- Parallel service-boundary batch by service ownership: T048, T049, T050, T051, T052

### User Story 3

- Parallel instruction updates: T057, T058, T059, T060, T061, T062
- Parallel memory updates: T064, T065, T066, T067, T068

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1 and Phase 2.
2. Complete Phase 3 (US1).
3. Validate route reachability and service-boundary decoupling.
4. Demo or merge MVP increment if stable.

### Incremental Delivery

1. Deliver US1 to enforce architectural separation.
2. Deliver US2 to standardize pointer semantics.
3. Deliver US3 to preserve conventions in instructions and memory.
4. Finish with Phase 6 and mandatory Phase 7 before final completion.

### Parallel Team Strategy

1. Engineer A: BFF contract and mapper refactor (US1).
2. Engineer B: Pointer policy propagation across non-BFF services (US2).
3. Engineer C: Instruction and memory synchronization plus governance checks (US3 + Phase 7).

---

## Notes

- All tasks follow the required checklist format with ID and explicit file path.
- Story tasks include labels [US1], [US2], [US3].
- Tasks marked [P] are parallelizable with dependency rules respected.
- Suggested MVP scope: complete through Phase 3 (US1).
