# Tasks: Restore Seeded Login & Session Propagation

**Input**: Design documents from `/specs/012-restore-login-session/`  
**Prerequisites**: `plan.md` (required), `spec.md` (required), `research.md`, `data-model.md`, `contracts/`, `quickstart.md`

**Tests**: Included. The spec requires end-to-end evidence for bootstrap login success, invalid-login failure, protected-route access, and default pagination behavior.

**Organization**: Tasks are grouped by user story so each increment remains independently testable and can be delivered in dependency order. Every feature task list includes a final mandatory governance sync phase.

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Prepare the shared auth/session contract surface and implementation tracking artifacts.

- [X] T001 Verify and preserve the canonical `common.v1.Session` and `common.v1.Pagination` contracts in `backend/protos/common/v1/messages.proto`
- [X] T002 [P] Define the login/refresh gRPC contract updates in `backend/protos/identity/v1/grpc.proto` and `backend/protos/identity/v1/messages.proto`
- [X] T003 [P] Refresh the implementation checklist and rollout notes in `specs/012-restore-login-session/contracts/auth-bootstrap-contract.md`, `specs/012-restore-login-session/contracts/grpc-session-pagination-adoption-matrix.md`, and `specs/012-restore-login-session/quickstart.md`

**Checkpoint**: The shared auth/session contract intent is explicit and the feature rollout remains traceable.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Establish the shared auth gateway, code generation, and seed plumbing before user story work begins.

**⚠️ CRITICAL**: No user story work should begin until this phase is complete.

- [X] T004 Generate and commit the updated protobuf artifacts in `backend/protos/generated/common/v1/*.pb.go` and `backend/protos/generated/identity/v1/*.pb.go` via `make proto/generate`
- [X] T005 [P] Wire the identity gRPC client and auth-service dependency injection in `backend/cmd/bff/container.go`, `backend/internals/bff/interfaces/services.go`, and `backend/internals/bff/transport/http/routes/contracts.go`
- [X] T006 [P] Create the BFF auth transport scaffolding in `backend/internals/bff/transport/http/routes/auth_routes.go`, `backend/internals/bff/transport/http/controllers/auth_controller.go`, `backend/internals/bff/transport/http/views/auth_views.go`, and `backend/internals/bff/transport/http/controllers/mappers/auth_mapper.go`
- [X] T007 Add persistent bootstrap seed plumbing for the owner user and membership in `backend/internals/identity/migrations/ddl/000001_create_users_table.up.sql`, `backend/internals/identity/migrations/dml/local/000001_seed_default_user.up.sql`, `backend/internals/identity/migrations/dml/local/000001_seed_default_user.down.sql`, and `backend/internals/onboarding/migrations/dml/local/000001_seed_default_project.up.sql`
- [X] T008 [P] Add baseline auth/session regression scaffolding in `backend/internals/identity/services/token_service_test.go` and `backend/internals/bff/services/auth_service_test.go`

**Checkpoint**: Proto generation, auth wiring, and seed prerequisites are ready for story implementation.

---

## Phase 3: User Story 1 - Sign in with the seeded owner account (Priority: P1) 🎯 MVP

**Goal**: Restore a real login-only bootstrap flow so a fresh environment can be accessed immediately with the seeded owner user.

**Independent Test**: Apply migrations/seed data in a fresh environment, sign in with `ralvescosta` / `mudar@1234`, and confirm the BFF returns a usable authenticated result without manual database edits.

### Tests for User Story 1 ⚠️

> Write or update these tests first and confirm the relevant failure is covered before implementation is finalized.

- [X] T009 [P] [US1] Add seed-backed login integration coverage in `backend/tests/integration/identity/bootstrap_login_seed_test.go`
- [X] T010 [P] [US1] Add BFF auth route registration and response-contract coverage in `backend/tests/integration/bff/auth_routes_registration_test.go`
- [X] T011 [P] [US1] Add hook-only frontend login regression coverage for the seeded owner flow in `frontend/src/hooks/useAuthSession.test.tsx`, `frontend/src/hooks/usePersistentSession.test.ts`, and `frontend/src/hooks/useTokenRefreshInterceptor.test.tsx`

### Implementation for User Story 1

- [X] T012 [US1] Implement the idempotent bootstrap user, password-hash, and owner-membership seed path in `backend/internals/identity/migrations/ddl/000001_create_users_table.up.sql`, `backend/internals/identity/migrations/dml/local/000001_seed_default_user.up.sql`, `backend/internals/identity/migrations/dml/local/000001_seed_default_user.down.sql`, and `backend/internals/onboarding/migrations/dml/local/000001_seed_default_project.up.sql`
- [X] T013 [US1] Extend the identity bootstrap/login and refresh behavior in `backend/internals/identity/services/token_service.go` and `backend/internals/identity/transport/grpc/server.go`
- [X] T014 [US1] Implement BFF auth orchestration and service-owned contracts in `backend/internals/bff/services/auth_service.go` and `backend/internals/bff/services/contracts/auth_contracts.go`
- [X] T015 [US1] Implement the login/refresh HTTP adapter flow in `backend/internals/bff/transport/http/routes/auth_routes.go`, `backend/internals/bff/transport/http/controllers/auth_controller.go`, `backend/internals/bff/transport/http/views/auth_views.go`, and `backend/internals/bff/transport/http/controllers/mappers/auth_mapper.go`
- [X] T016 [US1] Restore the frontend login-only session mapping in `frontend/src/hooks/useAuthContext.tsx`, `frontend/src/hooks/useAuthSession.ts`, `frontend/src/pages/LoginPage.tsx`, `frontend/src/services/api.client.ts`, `frontend/src/types/auth-response.schema.ts`, and `frontend/src/types/auth.ts`
- [X] T017 [US1] Ensure refresh/logout failures clear ambiguous client state in `frontend/src/hooks/useTokenRefreshInterceptor.ts` and `frontend/src/hooks/usePersistentSession.ts`

**Checkpoint**: The seeded owner can sign in successfully on a fresh environment through the BFF with no registration screen.

---

## Phase 4: User Story 2 - Use the authenticated session across protected routes (Priority: P1)

**Goal**: Propagate the same authenticated caller identity through the BFF and downstream services so the seeded owner can use all protected routes in the verified scope.

**Independent Test**: Sign in as the seeded owner, then exercise representative protected routes across projects, documents, payments, history, and reconciliation; each route must succeed with the propagated session and fail safely when the session or membership is invalid.

### Tests for User Story 2 ⚠️

- [X] T018 [P] [US2] Add cross-service session-propagation regression coverage in `backend/tests/integration/cross_service/restore_login_session_propagation_test.go`
- [X] T019 [P] [US2] Extend owner-access assertions in `backend/tests/integration/bff/projects_routes_registration_test.go`, `backend/tests/integration/bff/documents_routes_registration_test.go`, `backend/tests/integration/bff/payments_routes_registration_test.go`, and `backend/tests/integration/bff/reconciliation_routes_registration_test.go`
- [X] T020 [P] [US2] Add BFF service unit tests for `Session` forwarding in `backend/internals/bff/services/documents_service_test.go`, `backend/internals/bff/services/projects_service_test.go`, `backend/internals/bff/services/payments_service_test.go`, and `backend/internals/bff/services/history_service_test.go`

### Implementation for User Story 2

- [X] T021 [US2] Verify and finish propagating `common.v1.Session` across the authenticated request contracts in `backend/protos/onboarding/v1/grpc.proto`, `backend/protos/files/v1/grpc.proto`, `backend/protos/bills/v1/grpc.proto`, and `backend/protos/payments/v1/grpc.proto`
- [X] T022 [US2] Regenerate and commit the updated downstream gRPC artifacts in `backend/protos/generated/onboarding/v1/*.pb.go`, `backend/protos/generated/files/v1/*.pb.go`, `backend/protos/generated/bills/v1/*.pb.go`, and `backend/protos/generated/payments/v1/*.pb.go`
- [X] T023 [US2] Build authenticated caller envelopes from validated JWT claims in `backend/internals/bff/transport/http/middleware/auth_middleware.go` and `backend/internals/bff/transport/http/controllers/base_controller.go`
- [X] T024 [US2] Forward `Session` through all protected BFF service calls in `backend/internals/bff/services/documents_service.go`, `backend/internals/bff/services/projects_service.go`, `backend/internals/bff/services/payments_service.go`, `backend/internals/bff/services/history_service.go`, `backend/internals/bff/services/reconciliation_service.go`, and `backend/internals/bff/services/settings_service.go`
- [X] T025 [US2] Accept and validate the propagated session with AppError-first handling in `backend/internals/onboarding/transport/grpc/server.go`, `backend/internals/files/transport/grpc/server.go`, `backend/internals/bills/transport/grpc/server.go`, and `backend/internals/payments/transport/grpc/server.go`

**Checkpoint**: Protected BFF routes consistently honor the seeded owner session end-to-end and deny invalid sessions safely.

---

## Phase 5: User Story 3 - Keep list and search flows usable with default pagination (Priority: P2)

**Goal**: Ensure the BFF always forwards populated pagination for authenticated list/select flows so dashboards and list screens remain predictable when query params are omitted.

**Independent Test**: Call representative list/search endpoints with and without pagination query params and confirm the BFF forwards deterministic defaults (`20` fallback, with documented route-specific overrides such as `25` for documents/project members and `20` for the payment dashboard).

### Tests for User Story 3 ⚠️

- [X] T026 [P] [US3] Add pagination-default regression coverage in `backend/tests/integration/bff/documents_routes_registration_test.go`, `backend/tests/integration/bff/projects_routes_registration_test.go`, and `backend/tests/integration/bff/payments_routes_registration_test.go`
- [X] T027 [P] [US3] Add cross-service pagination propagation coverage in `backend/tests/integration/cross_service/get_history_timeline_test.go` and `backend/tests/integration/cross_service/get_history_metrics_test.go`
- [X] T028 [P] [US3] Add unit tests for route-specific page-size defaults in `backend/internals/bff/services/documents_service_test.go`, `backend/internals/bff/services/projects_service_test.go`, and `backend/internals/bff/services/payments_service_test.go`

### Implementation for User Story 3

- [X] T029 [US3] Verify and preserve `common.v1.Pagination` fields for multi-record requests in `backend/protos/files/v1/grpc.proto`, `backend/protos/bills/v1/grpc.proto`, `backend/protos/onboarding/v1/grpc.proto`, `backend/protos/payments/v1/grpc.proto`, and any additionally touched `backend/protos/*/v1/grpc.proto`
- [X] T030 [US3] Standardize default pagination builders in `backend/internals/bff/services/documents_service.go`, `backend/internals/bff/services/projects_service.go`, `backend/internals/bff/services/payments_service.go`, and `backend/internals/bff/services/history_service.go`
- [X] T031 [US3] Normalize pagination query parsing and mapper/view contracts in `backend/internals/bff/transport/http/controllers/mappers/documents_mapper.go`, `backend/internals/bff/transport/http/controllers/mappers/projects_mapper.go`, `backend/internals/bff/transport/http/controllers/mappers/payments_mapper.go`, and the corresponding `backend/internals/bff/transport/http/views/*.go` files

**Checkpoint**: Authenticated list and search screens return a predictable first page even when the frontend omits pagination values.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Finalize verification evidence, cleanup, and exception tracking across all stories.

- [X] T032 [P] Refresh the validation runbook and expected outputs in `specs/012-restore-login-session/quickstart.md` and `specs/012-restore-login-session/contracts/grpc-session-pagination-adoption-matrix.md`
- [X] T033 Run the required verification commands from `backend/` and `frontend/`, then fix any regressions in the touched suites under `backend/tests/integration/**` and `frontend/src/**/*.test.tsx`
- [X] T034 [P] Record any approved pointer/value-semantics exception or confirm none in `specs/012-restore-login-session/contracts/pointer-exceptions.md`

---

## Phase 7: Mandatory Governance Sync (Blocking)

**Purpose**: Keep memory artifacts, repository guidance, and testing conventions aligned with the restored login/session architecture.

- [X] T035 Update `.specify/memory/architecture-diagram.md`, `.specify/memory/bff-flows.md`, and `.specify/memory/identity-service-flows.md` for the seeded login and session propagation path
- [X] T036 [P] Update `.specify/memory/onboarding-service-flows.md`, `.specify/memory/files-service-flows.md`, `.specify/memory/bills-service-flows.md`, and `.specify/memory/payments-service-flows.md` for downstream `Session` and `Pagination` handling
- [X] T037 [P] Update `/memories/repo/bff-service-boundary-conventions.md` and re-check `.github/instructions/architecture.instructions.md`, `.github/instructions/ai-behavior.instructions.md`, and `.github/instructions/testing.instructions.md`
- [X] T038 Verify canonical integration-test placement, snake_case filenames, and BDD + AAA structure across `backend/tests/integration/bff/`, `backend/tests/integration/identity/`, and `backend/tests/integration/cross_service/`

**Checkpoint**: The feature is not complete until this phase is complete.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1: Setup** → can start immediately
- **Phase 2: Foundational** → depends on Setup and blocks all user stories
- **Phase 3: US1** → depends on Foundational completion
- **Phase 4: US2** → depends on Foundational and the auth/session primitives established by US1
- **Phase 5: US3** → depends on the shared session/pagination contract updates from US2 for the touched list/select flows
- **Phase 6: Polish** → depends on the desired user stories being complete
- **Phase 7: Mandatory Governance Sync** → depends on all implementation phases and must complete before merge

### User Story Dependencies

- **US1 (P1)**: Restores the bootstrap login path and is the first deliverable needed to make the application accessible.
- **US2 (P1)**: Builds on the authenticated result from US1 and propagates the caller identity across protected routes.
- **US3 (P2)**: Finalizes predictable list/select behavior once the shared session contract is in place.

### Within Each User Story

- Tests and regression harnesses first
- Proto or contract changes before code generation
- Service-layer orchestration before controller/route adaptation
- BFF propagation before downstream server enforcement
- Story verification before moving to the next dependent story

### Parallel Opportunities

- `T002` and `T003` can run in parallel once the session envelope fields are agreed.
- `T005`, `T006`, and `T008` can run in parallel after the foundational proto shape is settled.
- US1 test tasks `T009`–`T011` can run in parallel.
- US2 test tasks `T018`–`T020` can run in parallel, followed by `T023` and `T025` on separate files.
- US3 test tasks `T026`–`T028` can run in parallel, and `T032`, `T034`, `T036`, and `T037` are safe parallel follow-up work.

---

## Parallel Example: User Story 1

```bash
# Auth regression work that can proceed together:
Task: "T009 [US1] Add seed-backed login integration coverage in backend/tests/integration/identity/bootstrap_login_seed_test.go"
Task: "T010 [US1] Add BFF auth route registration and response-contract coverage in backend/tests/integration/bff/auth_routes_registration_test.go"
Task: "T011 [US1] Add frontend login regression coverage in frontend/src/hooks/useAuthSession.test.tsx and frontend/src/pages/LoginPage.integration.test.tsx"
```

---

## Parallel Example: User Story 2

```bash
# Session propagation checks that can proceed together after login works:
Task: "T018 [US2] Add cross-service session-propagation regression coverage in backend/tests/integration/cross_service/restore_login_session_propagation_test.go"
Task: "T019 [US2] Extend owner-access assertions in backend/tests/integration/bff/projects_routes_registration_test.go, documents_routes_registration_test.go, payments_routes_registration_test.go, and reconciliation_routes_registration_test.go"
Task: "T020 [US2] Add BFF service unit tests for Session forwarding in backend/internals/bff/services/*.go test files"
```

---

## Implementation Strategy

### MVP First

1. Complete **Phase 1: Setup**
2. Complete **Phase 2: Foundational**
3. Complete **Phase 3: User Story 1**
4. Complete **Phase 4: User Story 2**
5. **Stop and validate**: confirm the seeded owner can sign in and use protected routes end-to-end

> For this feature, the practical MVP is **US1 + US2**, because restored login is only useful once protected routes accept the same authenticated session.

### Incremental Delivery

1. Setup + Foundational establish the shared proto, seed, and BFF auth scaffolding
2. US1 restores a deterministic, login-only bootstrap path
3. US2 propagates the authenticated session through all protected routes in scope
4. US3 standardizes default pagination for list/select behavior
5. Polish and Governance Sync close the loop with verification evidence and documentation

### Parallel Team Strategy

With multiple developers:

1. One developer handles proto/auth gateway scaffolding (`T001`–`T008`)
2. One developer handles the seed + identity/BFF login work (`T009`–`T017`)
3. One developer handles downstream session/pagination propagation and regression enforcement (`T018`–`T031`)
4. One developer can finalize verification + governance sync once implementation stabilizes (`T032`–`T038`)

---

## Notes

- All tasks follow the required checklist format with an ID and explicit file path.
- `[US1]`, `[US2]`, and `[US3]` map directly to the user stories in `spec.md`.
- `[P]` tasks are safe to run in parallel when their prerequisite phase is complete.
- Run backend verification from `backend/` because this repository uses a nested Go module rooted at `backend/go.mod`.
- Preserve the login-only experience; do **not** reintroduce a registration route or registration screen in this feature.
