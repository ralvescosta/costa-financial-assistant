# Tasks: Restore BFF gRPC Gateway Boundary

**Input**: Design documents from `/specs/011-fix-bff-grpc-boundary/`  
**Prerequisites**: `plan.md` (required), `spec.md` (required), `research.md`, `data-model.md`, `contracts/`, `quickstart.md`

**Tests**: Regression and runtime verification are explicitly required by the spec, so test tasks are included.

**Organization**: Tasks are grouped by user story to keep each increment independently testable and aligned with BFF/domain ownership boundaries.
Every feature task list includes a final mandatory governance sync phase for memory-flow diagrams and instruction updates.

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Prepare the repository for the new payments gRPC module and generation workflow.

- [ ] T001 Update `Makefile` to add `payments/v1` to `PROTO_MODULES` and proto generation workflow
- [ ] T002 [P] Define payments-owned domain messages in `backend/protos/payments/v1/messages.proto`
- [ ] T003 [P] Define the new payments gRPC service RPCs in `backend/protos/payments/v1/grpc.proto`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Add the missing transport layer and shared contracts before any user story implementation starts.

**⚠️ CRITICAL**: No user story work should begin until this phase is complete.

- [ ] T004 Generate and commit `backend/protos/generated/payments/v1/*.pb.go` via `make proto/generate`
- [ ] T005 Implement the payments gRPC server and handler registration in `backend/internals/payments/transport/grpc/server.go`
- [ ] T006 Wire the payments gRPC server lifecycle and graceful shutdown in `backend/cmd/payments/container.go`
- [ ] T007 [P] Add/adjust payments transport tests in `backend/internals/payments/transport/grpc/server_test.go` and `backend/internals/payments/services/error_logging_test.go`
- [ ] T008 [P] Extend BFF-facing contracts for the new downstream payments client in `backend/internals/bff/interfaces/services.go` and `backend/internals/bff/services/contracts/*.go`

**Checkpoint**: The payments service exposes the missing gRPC surface and the BFF can start consuming it.

---

## Phase 3: User Story 1 - Keep the BFF as a secure gateway (Priority: P1) 🎯 MVP

**Goal**: Restore supported payment-cycle, history, and reconciliation screens through downstream gRPC only, while keeping BFF authentication, authorization, and response composition intact.

**Independent Test**: With `bff`, `payments`, and `bills` running, the payment-cycle, history, and reconciliation endpoints return their normal responses through downstream gRPC calls and do not require direct BFF domain access.

### Tests for User Story 1 ⚠️

> Write or update these tests first and confirm the relevant behavior is covered before implementation changes are finalized.

- [ ] T009 [P] [US1] Add route behavior regression coverage for payment-cycle endpoints in `backend/tests/integration/bff/payments_routes_registration_test.go`
- [ ] T010 [P] [US1] Add route behavior regression coverage for history endpoints in `backend/tests/integration/bff/history_routes_registration_test.go` and `backend/tests/integration/cross_service/get_history_timeline_test.go`
- [ ] T011 [P] [US1] Add route behavior regression coverage for reconciliation endpoints in `backend/tests/integration/bff/reconciliation_routes_registration_test.go` and `backend/tests/integration/cross_service/create_manual_reconciliation_link_test.go`

### Implementation for User Story 1

- [ ] T012 [US1] Add the payments gRPC client connection/provider in `backend/cmd/bff/container.go`
- [ ] T013 [US1] Migrate payment-cycle BFF logic in `backend/internals/bff/services/payments_service.go` to call the new `payments.v1` RPCs
- [ ] T014 [US1] Migrate history analytics BFF logic in `backend/internals/bff/services/history_service.go` to call the new `payments.v1` RPCs
- [ ] T015 [US1] Migrate reconciliation BFF logic in `backend/internals/bff/services/reconciliation_service.go` to call the new `payments.v1` RPCs
- [ ] T016 [P] [US1] Update response mapping helpers in `backend/internals/bff/transport/http/controllers/mappers/payments_mapper.go`, `backend/internals/bff/transport/http/controllers/mappers/history_mapper.go`, and `backend/internals/bff/transport/http/controllers/mappers/reconciliation_mapper.go`
- [ ] T017 [US1] Update the BFF unit tests for successful gRPC-backed behavior in `backend/internals/bff/services/payments_service_test.go`, `backend/internals/bff/services/history_service_test.go`, and `backend/internals/bff/services/reconciliation_service_test.go`

**Checkpoint**: User Story 1 is complete when the BFF behaves as a true gateway for the affected supported routes and returns normal downstream-backed responses.

---

## Phase 4: User Story 2 - Remove direct BFF domain-data ownership (Priority: P1)

**Goal**: Eliminate remaining direct domain injections/imports and enforce the architecture rule that business data lives behind the owning service boundary.

**Independent Test**: BFF service files and DI wiring no longer depend on `backend/internals/payments/interfaces`, repositories, or direct database-backed domain access for the supported flows; boundary tests and integration tests continue to pass.

### Tests for User Story 2 ⚠️

- [ ] T018 [P] [US2] Add ownership-regression assertions in `backend/internals/bff/services/payments_service_test.go`, `backend/internals/bff/services/reconciliation_service_test.go`, and `backend/tests/integration/cross_service/app_error_boundary_logging_test.go`

### Implementation for User Story 2

- [ ] T019 [US2] Remove any remaining payments-domain direct injections from `backend/cmd/bff/container.go` and `backend/internals/bff/services/*.go`
- [ ] T020 [P] [US2] Align payments-owned service/repository signatures with the pointer policy in `backend/internals/payments/interfaces/payment_cycle_service.go`, `backend/internals/payments/interfaces/history_repository.go`, and `backend/internals/payments/interfaces/reconciliation_service.go`
- [ ] T021 [P] [US2] Ensure AppError-first translation and logging for the new payments transport in `backend/internals/payments/transport/grpc/server.go` and `backend/pkgs/errors/*`
- [ ] T022 [US2] Record any approved value-semantics exception or confirm none in `specs/011-fix-bff-grpc-boundary/contracts/pointer-exceptions.md`

**Checkpoint**: User Story 2 is complete when the BFF no longer owns payments data access and the boundary rules are enforced in code and tests.

---

## Phase 5: User Story 3 - Preserve reliability after the boundary correction (Priority: P2)

**Goal**: Prove the corrected architecture still compiles, runs, and is operable for future contributors.

**Independent Test**: The backend test suite passes, the `bff`, `payments`, and `bills` services boot successfully with their standard commands, and the runbook reflects the final verification steps.

### Implementation for User Story 3

- [ ] T023 [US3] Run and stabilize the backend test suite by updating impacted tests in `backend/internals/bff/services/*_test.go` and `backend/tests/integration/cross_service/*.go`
- [ ] T024 [US3] Capture the final verification commands and expected outcomes in `specs/011-fix-bff-grpc-boundary/quickstart.md`
- [ ] T025 [US3] Verify short boot checks for `bff`, `payments`, and `bills` and document any required notes in `specs/011-fix-bff-grpc-boundary/quickstart.md`

**Checkpoint**: User Story 3 is complete when the feature has fresh run/verification evidence and an accurate quickstart for future work.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Clean up generated artifacts, formatting, and any remaining cross-story issues.

- [ ] T026 [P] Regenerate and format touched protobuf and Go files in `backend/protos/generated/payments/v1/*` and affected `backend/**/*.go` files
- [ ] T027 [P] Review and remove temporary fallback comments or migration-only stubs in `backend/internals/bff/services/*.go` and `backend/internals/payments/transport/grpc/server.go`

---

## Phase 7: Mandatory Governance Sync (Blocking)

**Purpose**: Ensure architecture guidance, memory-flow diagrams, and testing conventions remain aligned with the implementation.

- [ ] T028 Update `.specify/memory/architecture-diagram.md`, `.specify/memory/bff-flows.md`, and `.specify/memory/payments-service-flows.md` with the final BFF → payments gRPC flow
- [ ] T029 Update `/memories/repo/bff-service-boundary-conventions.md` and finalize refactor guidance in `.github/instructions/architecture.instructions.md`, `.github/instructions/project-structure.instructions.md`, and `.github/instructions/ai-behavior.instructions.md`
- [ ] T030 Verify canonical integration-test placement, snake_case filenames, and BDD + AAA compliance in `backend/tests/integration/bff/` and `backend/tests/integration/cross_service/`

**Checkpoint**: The feature is not complete until this phase is complete.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1: Setup** → can start immediately
- **Phase 2: Foundational** → depends on Setup and blocks all user stories
- **Phase 3: US1** → depends on Foundational completion
- **Phase 4: US2** → depends on Foundational completion and should follow the new gRPC path introduced in US1 for the supported payments-owned flows
- **Phase 5: US3** → depends on US1 and US2 implementation being in place so verification reflects the final behavior
- **Phase 6: Polish** → depends on the desired user stories being complete
- **Phase 7: Mandatory Governance Sync** → depends on all implementation phases and must complete before merge

### User Story Dependencies

- **US1 (P1)**: MVP after Foundational phase; delivers the actual downstream gRPC migration for supported routes
- **US2 (P1)**: Depends on the new gRPC surface from Foundational/US1 to remove the last direct access patterns safely
- **US3 (P2)**: Verifies and documents the final working state after US1 and US2

### Parallel Opportunities

- `T002` and `T003` can run in parallel once the payments proto module name is agreed
- `T007` and `T008` can run in parallel after the core proto shape is defined
- `T009`, `T010`, and `T011` can run in parallel as route-specific regression tasks
- `T016`, `T020`, `T021`, `T026`, and `T027` are parallel-safe because they target different files or independent cleanup work
- Governance sync tasks can be split across teammates once implementation behavior is final

---

## Parallel Example: User Story 1

```bash
# Route-specific regression work can proceed together:
Task: "T009 [US1] Add route behavior regression coverage for payment-cycle endpoints in backend/tests/integration/bff/payments_routes_registration_test.go"
Task: "T010 [US1] Add route behavior regression coverage for history endpoints in backend/tests/integration/bff/history_routes_registration_test.go and backend/tests/integration/cross_service/get_history_timeline_test.go"
Task: "T011 [US1] Add route behavior regression coverage for reconciliation endpoints in backend/tests/integration/bff/reconciliation_routes_registration_test.go and backend/tests/integration/cross_service/create_manual_reconciliation_link_test.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 only)

1. Finish **Phase 1: Setup**
2. Finish **Phase 2: Foundational**
3. Complete **Phase 3: User Story 1**
4. Run the US1 verification commands from `quickstart.md`
5. Demonstrate that the BFF now gets payment-cycle/history/reconciliation data through downstream gRPC instead of direct domain access

### Incremental Delivery

1. Setup + Foundational make the missing payments transport available
2. US1 restores supported route behavior through the correct boundary
3. US2 removes the remaining direct-ownership code paths and locks in architecture enforcement
4. US3 captures fresh reliability evidence and operability notes
5. Phase 7 finalizes the governance sync required by the spec and constitution

### Parallel Team Strategy

With multiple developers:

1. One developer handles proto + payments transport (`T001`–`T006`)
2. One developer handles BFF route/service migration (`T009`–`T017`)
3. One developer handles governance/test enforcement (`T018`–`T030`) after the core gRPC path stabilizes

---

## Notes

- `[P]` tasks are safe to execute in parallel because they target different files or independent workstreams.
- `[US1]`, `[US2]`, and `[US3]` labels map tasks directly to the user stories in `spec.md`.
- Run backend verification from `backend/` because this repository uses a nested Go module.
- Prefer completing and validating one phase at a time; do not skip the final governance sync phase.
