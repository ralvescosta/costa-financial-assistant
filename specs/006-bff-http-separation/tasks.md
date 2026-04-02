# Tasks: BFF HTTP Boundary Separation

**Input**: Design documents from `/specs/006-bff-http-separation/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, quickstart.md, contracts/

**Tests**: Unit and integration tests are required because the specification explicitly requires controller/service boundary validation and route-level coverage for all active BFF routes.

**Organization**: Tasks are grouped by user story so each story can be implemented and validated independently.

## Phase 1: Setup

**Purpose**: Add the shared package scaffolding and dependencies the refactor will build on.

- [ ] T001 Add the validator dependency for BFF HTTP contract checks in `backend/go.mod`
- [ ] T002 [P] Create the BFF service contract scaffold in `backend/internals/bff/interfaces/services.go`
- [ ] T003 [P] Create the HTTP views package scaffold in `backend/internals/bff/transport/http/views/documents_views.go`, `backend/internals/bff/transport/http/views/projects_views.go`, `backend/internals/bff/transport/http/views/settings_views.go`, `backend/internals/bff/transport/http/views/payments_views.go`, `backend/internals/bff/transport/http/views/reconciliation_views.go`, and `backend/internals/bff/transport/http/views/history_views.go`
- [ ] T004 [P] Create the BFF service implementation scaffold in `backend/internals/bff/services/documents_service.go`, `backend/internals/bff/services/projects_service.go`, `backend/internals/bff/services/settings_service.go`, `backend/internals/bff/services/payments_service.go`, `backend/internals/bff/services/reconciliation_service.go`, and `backend/internals/bff/services/history_service.go`

**Checkpoint**: The dependency, interfaces, service files, and views files exist and are ready for shared refactor work.

---

## Phase 2: Foundational

**Purpose**: Establish shared contracts, controller helpers, dependency injection, and governance rules that block all user story work.

**⚠️ CRITICAL**: No user story work can begin until this phase is complete.

- [ ] T005 Update shared controller helpers for validator-driven request checks and service error translation in `backend/internals/bff/transport/http/controllers/base_controller.go`
- [ ] T006 Define the shared BFF service interfaces for all active route groups in `backend/internals/bff/interfaces/services.go`
- [ ] T007 Update route capability interfaces to depend on view contracts in `backend/internals/bff/transport/http/routes/contracts.go`
- [ ] T008 [P] Wire the validator instance and BFF service providers through Dig in `backend/cmd/bff/container.go`
- [ ] T009 [P] Update the BFF route test bootstrap for service-backed route modules in `backend/tests/integration/bff_route_test_helpers.go`
- [ ] T010 [P] Update the architecture governance for the `transport/http/views` layer and controller/service boundary in `.specify/memory/constitution.md` and `.github/instructions/architecture.instructions.md`
- [ ] T011 [P] Update the route coverage testing rule for active BFF routes in `.github/instructions/testing.instructions.md`

**Checkpoint**: The shared controller, route, service, DI, and governance foundations are in place, and every user story can build on the same boundary model.

---

## Phase 3: User Story 1 - Keep Controllers Focused on HTTP Work (Priority: P1) 🎯 MVP

**Goal**: Remove downstream orchestration from controllers so every active BFF controller becomes an HTTP-only adapter.

**Independent Test**: Review one refactored controller and run the BFF service tests to confirm controllers only read context, validate input, call services, and format responses.

### Tests for User Story 1

- [ ] T012 [P] [US1] Add documents and projects BFF service unit tests in `backend/internals/bff/services/documents_service_test.go` and `backend/internals/bff/services/projects_service_test.go`
- [ ] T013 [P] [US1] Add settings, payments, reconciliation, and history BFF service unit tests in `backend/internals/bff/services/settings_service_test.go`, `backend/internals/bff/services/payments_service_test.go`, `backend/internals/bff/services/reconciliation_service_test.go`, and `backend/internals/bff/services/history_service_test.go`

### Implementation for User Story 1

- [ ] T014 [P] [US1] Implement documents and projects BFF services in `backend/internals/bff/services/documents_service.go` and `backend/internals/bff/services/projects_service.go`
- [ ] T015 [P] [US1] Implement settings and payments BFF services in `backend/internals/bff/services/settings_service.go` and `backend/internals/bff/services/payments_service.go`
- [ ] T016 [P] [US1] Implement reconciliation and history BFF services in `backend/internals/bff/services/reconciliation_service.go` and `backend/internals/bff/services/history_service.go`
- [ ] T017 [P] [US1] Refactor the documents, projects, and settings controllers to depend on BFF services instead of direct downstream orchestration in `backend/internals/bff/transport/http/controllers/documents_controller.go`, `backend/internals/bff/transport/http/controllers/projects_controller.go`, and `backend/internals/bff/transport/http/controllers/settings_controller.go`
- [ ] T018 [P] [US1] Refactor the payments, reconciliation, and history controllers to depend on BFF services instead of direct downstream orchestration in `backend/internals/bff/transport/http/controllers/payments_controller.go`, `backend/internals/bff/transport/http/controllers/reconciliation_controller.go`, and `backend/internals/bff/transport/http/controllers/history_controller.go`
- [ ] T019 [US1] Update BFF container wiring to provide service-backed controllers and remove direct downstream constructor dependencies in `backend/cmd/bff/container.go`

**Checkpoint**: All active BFF controllers are HTTP-only adapters and no controller owns direct gRPC client or repository orchestration.

---

## Phase 4: User Story 2 - Centralize HTTP Contracts in a Dedicated View Layer (Priority: P2)

**Goal**: Move every active BFF HTTP request and response contract into the dedicated views layer and enforce validator tags for controller-side checks.

**Independent Test**: Inspect one route group and confirm all request and response structs live under `transport/http/views`, route contracts use those types, and invalid input fails controller-side validation.

### Tests for User Story 2

- [ ] T020 [P] [US2] Add HTTP contract validation coverage for documents, projects, and settings routes in `backend/tests/integration/documents_routes_integration_test.go`, `backend/tests/integration/projects_routes_integration_test.go`, and `backend/tests/integration/settings_routes_integration_test.go`
- [ ] T021 [P] [US2] Add HTTP contract validation coverage for payments, reconciliation, and history routes in `backend/tests/integration/payments_routes_integration_test.go`, `backend/tests/integration/reconciliation_routes_integration_test.go`, and `backend/tests/integration/history_routes_integration_test.go`

### Implementation for User Story 2

- [ ] T022 [P] [US2] Move documents and projects HTTP request/response contracts into `backend/internals/bff/transport/http/views/documents_views.go` and `backend/internals/bff/transport/http/views/projects_views.go`
- [ ] T023 [P] [US2] Move settings and payments HTTP request/response contracts into `backend/internals/bff/transport/http/views/settings_views.go` and `backend/internals/bff/transport/http/views/payments_views.go`
- [ ] T024 [P] [US2] Move reconciliation and history HTTP request/response contracts into `backend/internals/bff/transport/http/views/reconciliation_views.go` and `backend/internals/bff/transport/http/views/history_views.go`
- [ ] T025 [US2] Refactor route capability interfaces and all route modules to use view contracts in `backend/internals/bff/transport/http/routes/contracts.go`, `backend/internals/bff/transport/http/routes/documents_routes.go`, `backend/internals/bff/transport/http/routes/projects_routes.go`, `backend/internals/bff/transport/http/routes/settings_routes.go`, `backend/internals/bff/transport/http/routes/payments_routes.go`, `backend/internals/bff/transport/http/routes/reconciliation_routes.go`, and `backend/internals/bff/transport/http/routes/history_routes.go`
- [ ] T026 [US2] Refactor all BFF controllers to validate and return view contracts in `backend/internals/bff/transport/http/controllers/documents_controller.go`, `backend/internals/bff/transport/http/controllers/projects_controller.go`, `backend/internals/bff/transport/http/controllers/settings_controller.go`, `backend/internals/bff/transport/http/controllers/payments_controller.go`, `backend/internals/bff/transport/http/controllers/reconciliation_controller.go`, and `backend/internals/bff/transport/http/controllers/history_controller.go`
- [ ] T027 [US2] Add validator tags while preserving Huma binding and OpenAPI metadata in `backend/internals/bff/transport/http/views/documents_views.go`, `backend/internals/bff/transport/http/views/projects_views.go`, `backend/internals/bff/transport/http/views/settings_views.go`, `backend/internals/bff/transport/http/views/payments_views.go`, `backend/internals/bff/transport/http/views/reconciliation_views.go`, and `backend/internals/bff/transport/http/views/history_views.go`

**Checkpoint**: All active BFF HTTP contracts live in the dedicated views package, controllers validate those contracts, and route capability interfaces no longer depend on controller-owned transport types.

---

## Phase 5: User Story 3 - Preserve Route Clarity and Coverage During Refactor (Priority: P3)

**Goal**: Keep the active BFF route inventory stable and fully covered while the internal boundaries change.

**Independent Test**: Run the BFF route smoke, OpenAPI, and resource-scoped integration suites and verify every active route in the coverage matrix has at least one passing scenario.

### Tests for User Story 3

- [ ] T028 [P] [US3] Expand registration and metadata regression coverage in `backend/tests/integration/bff_route_registration_smoke_test.go` and `backend/tests/integration/openapi_contract_test.go`
- [ ] T029 [P] [US3] Complete resource-scoped coverage for documents, projects, settings, and payments routes in `backend/tests/integration/documents_routes_integration_test.go`, `backend/tests/integration/projects_routes_integration_test.go`, `backend/tests/integration/settings_routes_integration_test.go`, and `backend/tests/integration/payments_routes_integration_test.go`
- [ ] T030 [P] [US3] Complete resource-scoped coverage for reconciliation and history routes in `backend/tests/integration/reconciliation_routes_integration_test.go` and `backend/tests/integration/history_routes_integration_test.go`

### Implementation for User Story 3

- [ ] T031 [US3] Align the active 20-route inventory with the integration suites in `specs/006-bff-http-separation/contracts/route-coverage-matrix.md`
- [ ] T032 [US3] Update the route, controller, and service boundary contract to match the implemented BFF shape in `specs/006-bff-http-separation/contracts/route-controller-service-contract.md`
- [ ] T033 [US3] Update the feature validation workflow for the final route and contract checks in `specs/006-bff-http-separation/quickstart.md`
- [ ] T034 [US3] Preserve authentication, role enforcement, and project-isolation regression coverage in `backend/tests/integration/auth_token_rejection_test.go`, `backend/tests/integration/us7_role_enforcement_test.go`, and `backend/tests/integration/us7_project_isolation_test.go`

**Checkpoint**: Every active BFF route is still registered with the same external behavior and is explicitly covered by integration tests and feature documentation.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Final cleanup, repo guidance alignment, and end-to-end verification.

- [ ] T035 [P] Update agent guidance for the final 006 boundary model in `.github/agents/copilot-instructions.md`
- [ ] T036 Run `go test ./...` from `backend/` to validate the packages rooted at `backend/go.mod`
- [ ] T037 Run `go test -tags=integration ./tests/integration/...` from `backend/` to validate the suites bootstrapped by `backend/tests/integration/testmain_test.go`
- [ ] T038 Update the implementation guidance and living architecture references to match the delivered boundary model in `.github/instructions/architecture.instructions.md`, `.github/instructions/testing.instructions.md`, `.specify/memory/constitution.md`, `.specify/memory/bff-flows.md`, and `.specify/memory/architecture-diagram.md`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies and can start immediately.
- **Foundational (Phase 2)**: Depends on Setup and blocks all user stories.
- **User Story 1 (Phase 3)**: Depends on Foundational and is the MVP.
- **User Story 2 (Phase 4)**: Depends on User Story 1 because controllers must already depend on BFF services before views become the only HTTP contract types.
- **User Story 3 (Phase 5)**: Depends on User Story 2 because route and contract coverage must reflect the final views-based transport boundary.
- **Polish (Phase 6)**: Depends on all user stories being complete.
- **Polish (Phase 6)**: Depends on all user stories being complete.

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational and does not depend on the later stories.
- **User Story 2 (P2)**: Depends on the service-backed controller boundary from User Story 1.
- **User Story 3 (P3)**: Depends on the final controller, route, and view structure from User Stories 1 and 2.

### Within Each User Story

- Write the listed tests first and confirm they fail for the intended gap before implementing the story.
- Implement service logic before refactoring controllers to consume it.
- Move contracts into `views/` before switching route capability interfaces and controller signatures to those types.
- Update route coverage documents only after the matching tests pass.

### Parallel Opportunities

- `T002`, `T003`, and `T004` can run in parallel after `T001`.
- `T008`, `T009`, `T010`, and `T011` can run in parallel after `T005`, `T006`, and `T007`.
- `T012` and `T013` can run in parallel, followed by `T014`, `T015`, and `T016` in parallel.
- `T017` and `T018` can run in parallel after the matching services exist.
- `T020` and `T021` can run in parallel, followed by `T022`, `T023`, and `T024` in parallel.
- `T028`, `T029`, and `T030` can run in parallel after User Story 2 is implemented.

---

## Parallel Example: User Story 1

```bash
# Launch service implementation work for separate resource groups in parallel
Task: T014 backend/internals/bff/services/documents_service.go + backend/internals/bff/services/projects_service.go
Task: T015 backend/internals/bff/services/settings_service.go + backend/internals/bff/services/payments_service.go
Task: T016 backend/internals/bff/services/reconciliation_service.go + backend/internals/bff/services/history_service.go
```

## Parallel Example: User Story 2

```bash
# Move HTTP contracts into views by resource group in parallel
Task: T022 backend/internals/bff/transport/http/views/documents_views.go + backend/internals/bff/transport/http/views/projects_views.go
Task: T023 backend/internals/bff/transport/http/views/settings_views.go + backend/internals/bff/transport/http/views/payments_views.go
Task: T024 backend/internals/bff/transport/http/views/reconciliation_views.go + backend/internals/bff/transport/http/views/history_views.go
```

## Parallel Example: User Story 3

```bash
# Expand route coverage suites in parallel by resource group
Task: T029 backend/tests/integration/documents_routes_integration_test.go + backend/tests/integration/projects_routes_integration_test.go + backend/tests/integration/settings_routes_integration_test.go + backend/tests/integration/payments_routes_integration_test.go
Task: T030 backend/tests/integration/reconciliation_routes_integration_test.go + backend/tests/integration/history_routes_integration_test.go
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup.
2. Complete Phase 2: Foundational.
3. Complete Phase 3: User Story 1.
4. **STOP and VALIDATE**: Confirm controllers no longer orchestrate downstream dependencies directly.

### Incremental Delivery

1. Deliver User Story 1 to establish the controller-to-service boundary.
2. Deliver User Story 2 to centralize HTTP contracts in `transport/http/views`.
3. Deliver User Story 3 to lock route coverage and documentation to the final architecture.
4. Finish with full backend and integration validation.

### Suggested MVP Scope

- **MVP**: Setup + Foundational + User Story 1
- **Second increment**: User Story 2
- **Final increment**: User Story 3 + Polish

---

## Notes

- All tasks follow the strict `- [ ] T### [P] [US#] Description with file path` checklist format.
- `[P]` is used only where the tasks touch separate file groups with no incomplete dependency between them.
- Route coverage is not complete until the 20 rows in `specs/006-bff-http-separation/contracts/route-coverage-matrix.md` match passing integration suites.
- The final governance sync is not complete until `.github/instructions/architecture.instructions.md`, `.github/instructions/testing.instructions.md`, `.specify/memory/constitution.md`, `.specify/memory/bff-flows.md`, and `.specify/memory/architecture-diagram.md` reflect the implemented boundary model.