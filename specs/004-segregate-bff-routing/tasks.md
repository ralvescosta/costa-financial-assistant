# Tasks: BFF Route-Controller Segregation

**Input**: Design documents from `/specs/004-segregate-bff-routing/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, quickstart.md, contracts/

**Tests**: Integration and regression tests are required because the specification explicitly requires route-level coverage for all BFF routes.

**Organization**: Tasks are grouped by user story so each story can be implemented and validated independently.

## Phase 1: Setup

**Purpose**: Create the shared files and test scaffolding that the route refactor will build on.

- [ ] T001 Create the route package scaffold in `backend/internals/bff/transport/http/routes/contracts.go`
- [ ] T002 [P] Create shared BFF route integration helpers in `backend/tests/integration/bff_route_test_helpers.go`

**Checkpoint**: Route package and shared test helpers exist.

---

## Phase 2: Foundational

**Purpose**: Establish the contracts, governance updates, and test baseline that block all user story work.

**⚠️ CRITICAL**: No user story work can begin until this phase is complete.

- [ ] T003 Define the shared route registration contract and middleware dependency types in `backend/internals/bff/transport/http/routes/contracts.go`
- [ ] T004 [P] Add shared controller base behavior in `backend/internals/bff/transport/http/controllers/base_controller.go`
- [ ] T005 [P] Update OpenAPI metadata validation to scan route files in `backend/tests/integration/openapi_contract_test.go`
- [ ] T006 Update the BFF routing governance rule in `.specify/memory/constitution.md`
- [ ] T007 [P] Update the route ownership architecture rule in `.github/instructions/architecture.instructions.md`
- [ ] T008 [P] Update backend route coverage testing guidance in `.github/instructions/testing.instructions.md`

**Checkpoint**: Shared route/controller contracts and governance rules are in place, and route metadata validation points at the new route layer.

---

## Phase 3: User Story 1 - Separate Route Registration from Controller Behavior (Priority: P1) 🎯 MVP

**Goal**: Move all BFF `huma.Register(...)` declarations into dedicated route modules while preserving current endpoint behavior.

**Independent Test**: Review the route modules and run the route-registration regression tests to confirm controllers no longer own registration and all operations remain mounted with unchanged metadata and middleware intent.

### Tests for User Story 1

- [ ] T009 [P] [US1] Create route registration regression coverage in `backend/tests/integration/bff_route_registration_smoke_test.go`

### Implementation for User Story 1

- [ ] T010 [P] [US1] Create the documents and projects route modules in `backend/internals/bff/transport/http/routes/documents_routes.go` and `backend/internals/bff/transport/http/routes/projects_routes.go`
- [ ] T011 [P] [US1] Create the settings and payments route modules in `backend/internals/bff/transport/http/routes/settings_routes.go` and `backend/internals/bff/transport/http/routes/payments_routes.go`
- [ ] T012 [P] [US1] Create the reconciliation and history route modules in `backend/internals/bff/transport/http/routes/reconciliation_routes.go` and `backend/internals/bff/transport/http/routes/history_routes.go`
- [ ] T013 [P] [US1] Remove `Register(...)` ownership from the documents, projects, and settings controllers in `backend/internals/bff/transport/http/controllers/documents_controller.go`, `backend/internals/bff/transport/http/controllers/projects_controller.go`, and `backend/internals/bff/transport/http/controllers/settings_controller.go`
- [ ] T014 [P] [US1] Remove `Register(...)` ownership from the payments, reconciliation, and history controllers in `backend/internals/bff/transport/http/controllers/payments_controller.go`, `backend/internals/bff/transport/http/controllers/reconciliation_controller.go`, and `backend/internals/bff/transport/http/controllers/history_controller.go`
- [ ] T015 [US1] Update Dig wiring to provide and register route modules in `backend/cmd/bff/container.go`

**Checkpoint**: All BFF route declarations live in dedicated route modules, controllers are behavior-only, and the BFF container registers route modules instead of controller `Register(...)` methods.

---

## Phase 4: User Story 2 - Standardize Controller and Route Contracts (Priority: P2)

**Goal**: Make route registration and controller behavior conform to a shared, Go-idiomatic contract model that future modules can reuse consistently.

**Independent Test**: Verify that route modules consume capability-specific controller interfaces, controllers satisfy those contracts, and the BFF can wire all route groups without ad hoc registration patterns.

### Tests for User Story 2

- [ ] T016 [P] [US2] Create route contract wiring coverage in `backend/tests/integration/bff_route_contract_wiring_test.go`

### Implementation for User Story 2

- [ ] T017 [US2] Define the shared controller capability interfaces in `backend/internals/bff/transport/http/routes/contracts.go`
- [ ] T018 [P] [US2] Update the documents, projects, and settings controllers to satisfy capability interfaces in `backend/internals/bff/transport/http/controllers/documents_controller.go`, `backend/internals/bff/transport/http/controllers/projects_controller.go`, and `backend/internals/bff/transport/http/controllers/settings_controller.go`
- [ ] T019 [P] [US2] Update the payments, reconciliation, and history controllers to satisfy capability interfaces in `backend/internals/bff/transport/http/controllers/payments_controller.go`, `backend/internals/bff/transport/http/controllers/reconciliation_controller.go`, and `backend/internals/bff/transport/http/controllers/history_controller.go`
- [ ] T020 [US2] Add compile-time assertions and consistent constructor wiring in `backend/internals/bff/transport/http/routes/documents_routes.go`, `backend/internals/bff/transport/http/routes/projects_routes.go`, `backend/internals/bff/transport/http/routes/settings_routes.go`, `backend/internals/bff/transport/http/routes/payments_routes.go`, `backend/internals/bff/transport/http/routes/reconciliation_routes.go`, `backend/internals/bff/transport/http/routes/history_routes.go`, and `backend/cmd/bff/container.go`

**Checkpoint**: Route modules and controllers follow the shared contract model, and future endpoint modules can adopt the same pattern without custom transport wiring.

---

## Phase 5: User Story 3 - Ensure Route-Level Integration Coverage (Priority: P3)

**Goal**: Make route coverage explicit and complete so every declared BFF route has at least one passing integration scenario.

**Independent Test**: Run the integration suite and verify every route in the route coverage matrix maps to at least one passing resource-scoped integration test.

### Tests for User Story 3

- [ ] T021 [P] [US3] Create documents and projects route integration suites in `backend/tests/integration/documents_routes_integration_test.go` and `backend/tests/integration/projects_routes_integration_test.go`
- [ ] T022 [P] [US3] Create settings and payments route integration suites in `backend/tests/integration/settings_routes_integration_test.go` and `backend/tests/integration/payments_routes_integration_test.go`
- [ ] T023 [P] [US3] Create reconciliation and history route integration suites in `backend/tests/integration/reconciliation_routes_integration_test.go` and `backend/tests/integration/history_routes_integration_test.go`

### Implementation for User Story 3

- [ ] T024 [P] [US3] Refactor legacy document and settings route assertions in `backend/tests/integration/us1_upload_classify_test.go` and `backend/tests/integration/us3_bank_accounts_test.go`
- [ ] T025 [P] [US3] Refactor legacy payments, reconciliation, and history route assertions in `backend/tests/integration/us4_payment_dashboard_test.go`, `backend/tests/integration/us4_mark_paid_idempotency_test.go`, `backend/tests/integration/us5_manual_reconciliation_test.go`, `backend/tests/integration/us5_auto_reconciliation_test.go`, `backend/tests/integration/us6_history_timeline_test.go`, and `backend/tests/integration/us6_history_metrics_test.go`
- [ ] T026 [US3] Update the route-to-test mapping in `specs/004-segregate-bff-routing/contracts/route-coverage-matrix.md`

**Checkpoint**: All 20 BFF routes are covered by explicit resource-scoped integration suites and the route coverage matrix reflects the actual tests.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Finalize documentation and validate the full feature end to end.

- [ ] T027 [P] Update the validation workflow in `specs/004-segregate-bff-routing/quickstart.md`
- [ ] T028 Run `go test ./...` from `backend/`
- [ ] T029 Run `go test -tags=integration ./tests/integration/...` from `backend/`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies; starts immediately.
- **Foundational (Phase 2)**: Depends on Setup and blocks all user stories.
- **User Story 1 (Phase 3)**: Depends on Foundational and is the MVP.
- **User Story 2 (Phase 4)**: Depends on User Story 1 because contracts must be applied to the new route modules.
- **User Story 3 (Phase 5)**: Depends on User Story 1 for the route layer and benefits from User Story 2 contract stabilization.
- **Polish (Phase 6)**: Depends on all desired user stories being complete.

### User Story Dependencies

- **US1**: No dependency on other user stories once Foundational is complete.
- **US2**: Depends on US1 route modules existing.
- **US3**: Depends on US1 route modules existing; should run after US2 if compile-time contracts alter route wiring.

### Within Each User Story

- Write the listed regression or integration tests before implementation and confirm they fail for the intended gap.
- Create route modules before removing controller `Register(...)` ownership.
- Update controllers before final container wiring.
- Update coverage mapping only after the route suites exist and pass.

### Parallel Opportunities

- `T002`, `T004`, `T005`, `T007`, and `T008` can run in parallel after `T001`.
- `T010`, `T011`, and `T012` can run in parallel once the foundational contracts exist.
- `T013` and `T014` can run in parallel after their corresponding route modules are in place.
- `T018` and `T019` can run in parallel after `T017` defines the shared capability interfaces.
- `T021`, `T022`, and `T023` can run in parallel once the route modules are stable.
- `T024` and `T025` can run in parallel while the new resource-scoped coverage suites are being finalized.

---

## Parallel Example: User Story 1

```bash
# Build route modules for different resource groups in parallel
Task: T010 backend/internals/bff/transport/http/routes/documents_routes.go + backend/internals/bff/transport/http/routes/projects_routes.go
Task: T011 backend/internals/bff/transport/http/routes/settings_routes.go + backend/internals/bff/transport/http/routes/payments_routes.go
Task: T012 backend/internals/bff/transport/http/routes/reconciliation_routes.go + backend/internals/bff/transport/http/routes/history_routes.go
```

## Parallel Example: User Story 3

```bash
# Build resource-scoped integration coverage in parallel
Task: T021 backend/tests/integration/documents_routes_integration_test.go + backend/tests/integration/projects_routes_integration_test.go
Task: T022 backend/tests/integration/settings_routes_integration_test.go + backend/tests/integration/payments_routes_integration_test.go
Task: T023 backend/tests/integration/reconciliation_routes_integration_test.go + backend/tests/integration/history_routes_integration_test.go
```

---

## Implementation Strategy

### MVP First

1. Complete Phase 1: Setup.
2. Complete Phase 2: Foundational.
3. Complete Phase 3: User Story 1.
4. Validate that route modules fully replace controller registration without breaking the BFF API.

### Incremental Delivery

1. Deliver US1 to establish dedicated route ownership.
2. Deliver US2 to harden the shared controller/route contract model.
3. Deliver US3 to make route coverage explicit and complete.
4. Finish with cross-cutting validation and documentation updates.

### Suggested MVP Scope

- **MVP**: Setup + Foundational + User Story 1
- **Next increment**: User Story 2
- **Final increment**: User Story 3 + Polish

---

## Notes

- All tasks use the `004-segregate-bff-routing` design artifacts, not the empty active 005 branch context.
- Tasks marked `[P]` touch different files and can be split across multiple contributors.
- Route coverage is not complete until `specs/004-segregate-bff-routing/contracts/route-coverage-matrix.md` matches passing integration suites.