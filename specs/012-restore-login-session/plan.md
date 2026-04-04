# Implementation Plan: Restore Seeded Login & Session Propagation

**Branch**: `012-restore-login-session` | **Date**: 2026-04-04 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/012-restore-login-session/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

Restore the login-only bootstrap path by wiring the missing BFF auth gateway to the existing identity gRPC service, fixing the persistent seed data for the default owner user (`ralvescosta` / `mudar@1234`, `ralvescosta@local.dev`), and verifying that the authenticated result works across all currently exposed authenticated BFF routes/screens. The implementation will preserve the BFF gateway boundary, complete `common.v1.Session` propagation on authenticated downstream requests, standardize BFF pagination forwarding to `page_size = 20` and `page_token = ""` when query params are omitted, and add end-to-end regression coverage plus the required memory/instruction synchronization.

## Technical Context

**Language/Version**: Go 1.25.6 for backend services; TypeScript/React (Vite) for frontend contract verification  
**Primary Dependencies**: `github.com/labstack/echo/v4`, `github.com/danielgtaylor/huma/v2`, `go.uber.org/dig`, `google.golang.org/grpc`, generated protobuf clients in `backend/protos/generated/`, `go.uber.org/zap`, `github.com/go-playground/validator/v10`, frontend `zod` validation in `frontend/src/types/auth-response.schema.ts`  
**Storage**: PostgreSQL-backed identity and onboarding seed/migration flows; JWT/JWKS signing handled by the identity service  
**Testing**: `cd backend && go test ./...`, canonical BFF integration tests in `backend/tests/integration/bff/`, identity-focused integration tests in `backend/tests/integration/identity/`, cross-service auth/session tests in `backend/tests/integration/cross_service/`, and `make frontend/test` for login-contract verification  
**Target Platform**: Linux containerized backend services plus the browser-based frontend used in local/CI validation  
**Project Type**: Full-stack auth-restoration and contract-standardization feature inside a Go + React monorepo  
**Performance Goals**: Fresh-environment login succeeds without manual DB repair; protected routes keep their current behavior after sign-in; list/select flows return deterministic first-page results when the UI omits pagination  
**Constraints**: No registration screen or register endpoint; BFF must remain a thin auth/orchestration gateway; AppError-first boundary handling is mandatory; `Session` must supplement rather than replace `ProjectContext`; verification scope covers all currently exposed authenticated BFF route groups  
**Scale/Scope**: `backend/cmd/bff`, `backend/internals/bff/**`, `backend/internals/identity/**`, `backend/internals/onboarding/**`, `backend/protos/common/v1/messages.proto`, in-scope authenticated request protos under `backend/protos/{files,bills,onboarding,payments}/v1/`, relevant frontend auth files, and the required `.specify/memory/*`, `/memories/repo/*`, and `.github/instructions/*` sync artifacts  
**Verified Route Scope**: `documents`, `history`, `payments`, `projects`, `reconciliation`, and `settings` authenticated BFF route groups, plus the restored auth login/refresh gateway.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Checked against `.specify/memory/constitution.md` and `specs/012-restore-login-session/spec.md`.

Mandatory gates:
- [x] `spec.md` includes `Architecture & Memory Diagram Flow Impact` with impacted
  `.specify/memory/*.md` file list or explicit no-impact rationale.
- [x] End-of-execution tasks include updates to all impacted service-flow memory files
  and `.specify/memory/architecture-diagram.md` when cross-service flow changes exist.
- [x] If the feature is refactor/reorganization, plan includes explicit instruction
  update tasks for impacted `.github/instructions/*.instructions.md` files.
- [x] If workflow behavior changes, plan includes updates to impacted
  `.specify/templates/*.md` files. (N/A here; no workflow template changes are planned.)
- [x] If backend integration behavior is in scope, plan includes canonical test-placement
  and naming compliance tasks (`backend/tests/integration/<service>/` or
  `backend/tests/integration/cross_service/`, behavior-based snake_case, BDD + AAA).
- [x] If BFF boundaries are modified, plan includes explicit tasks for service-contract ownership (`services/contracts`), mapper-boundary enforcement (`controllers/mappers`), pointer-policy checks, and exception documentation.

Gate status: PASS.

## Project Structure

### Documentation (this feature)

```text
specs/012-restore-login-session/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── auth-bootstrap-contract.md
│   ├── governance-sync-checklist.md
│   ├── grpc-session-pagination-adoption-matrix.md
│   └── pointer-exceptions.md
├── checklists/
│   └── requirements.md
└── tasks.md
```

### Source Code (repository root)

```text
backend/
├── cmd/
│   └── bff/
│       └── container.go
├── internals/
│   ├── bff/
│   │   ├── interfaces/
│   │   │   ├── grpc_clients.go
│   │   │   └── services.go
│   │   ├── services/
│   │   │   ├── auth_service.go
│   │   │   ├── documents_service.go
│   │   │   ├── history_service.go
│   │   │   ├── payments_service.go
│   │   │   ├── projects_service.go
│   │   │   └── reconciliation_service.go
│   │   └── transport/http/
│   │       ├── controllers/
│   │       ├── controllers/mappers/
│   │       ├── routes/
│   │       └── views/
│   ├── identity/
│   │   ├── migrations/
│   │   ├── repositories/
│   │   ├── services/
│   │   └── transport/grpc/
│   └── onboarding/
│       ├── migrations/
│       ├── services/
│       └── transport/grpc/
├── protos/
│   ├── common/v1/messages.proto
│   ├── identity/v1/
│   ├── files/v1/
│   ├── bills/v1/
│   ├── onboarding/v1/
│   └── payments/v1/
└── tests/
    ├── integration/bff/
    ├── integration/identity/
    └── integration/cross_service/

frontend/
└── src/
    ├── app/router.tsx
    ├── hooks/useAuthContext.tsx
    ├── pages/LoginPage.tsx
    └── types/auth-response.schema.ts
```

**Structure Decision**: This is a backend-first auth restoration feature that preserves the existing login-only frontend contract. The BFF remains the single HTTP gateway, identity/onboarding remain the owners of credential and membership data, and the planning artifacts above capture the cross-service contract, testing, and governance work needed to deliver the feature safely.

## Mandatory End-of-Execution Sync

**Memory Diagram Sync (required)**:
- Impacted memory files: `.specify/memory/architecture-diagram.md`, `.specify/memory/bff-flows.md`, `.specify/memory/identity-service-flows.md`, `.specify/memory/onboarding-service-flows.md`, plus `.specify/memory/files-service-flows.md`, `.specify/memory/bills-service-flows.md`, and `.specify/memory/payments-service-flows.md` for the session/pagination propagation changes on protected requests.
- Impacted repository memory files: `/memories/repo/bff-service-boundary-conventions.md` and any additional repo note updated to preserve the final login/session pattern.
- Update tasks required in `tasks.md`: Yes
- If `No`, rationale: N/A

**Instruction Sync (required for refactor/reorganization)**:
- Refactor/reorganization in scope: Yes
- If `Yes`, impacted instruction files to update:
  - `.github/instructions/architecture.instructions.md`
  - `.github/instructions/ai-behavior.instructions.md`
  - `.github/instructions/testing.instructions.md`
- If `Yes`, impacted workflow templates to update:
  - None planned unless implementation uncovers a reusable planning-rule change

**Completion gate**:
Implementation is not complete until the persistent bootstrap login path works end-to-end and all required sync tasks above are executed in the same feature cycle.

## Phase 0: Research and Decisions

1. Audit the current auth gap and preserve the working pieces already in place: the frontend login-only contract, the BFF `AuthService`, and the identity token service/JWKS flow.
2. Fix the persistent bootstrap-data path by aligning identity seed data, credential hashing, seeded email, and onboarding owner membership so a fresh environment signs in without manual SQL.
3. Verify that `common.v1.Session` already exists in `backend/protos/common/v1/messages.proto` and complete its adoption on every authenticated gRPC request in the verified scope.
4. Standardize BFF request-building so list/select flows always forward a populated `common.v1.Pagination`, using the clarified defaults of `page_size = 20` and `page_token = ""` when the frontend omits them.
5. Define the BFF login/refresh transport surface using the existing controller/service/view/mapper ownership rules and maintain compatibility with `frontend/src/types/auth-response.schema.ts`.
6. Define the regression test matrix for login success, invalid credentials, protected-route access, session propagation, and default pagination behavior across the canonical backend integration-test locations.

## Phase 1: Design Outputs

1. Produce `research.md` documenting the login-restoration and contract-propagation decisions.
2. Produce `data-model.md` for the bootstrap owner, project membership, `Session`, and pagination-default concepts.
3. Produce `contracts/auth-bootstrap-contract.md` describing the BFF login/refresh contract expected by the frontend.
4. Produce `contracts/grpc-session-pagination-adoption-matrix.md` mapping every in-scope service request to its `Session`/`Pagination` obligations.
5. Produce `contracts/governance-sync-checklist.md` capturing the required memory, instruction, and verification updates.
6. Produce `contracts/pointer-exceptions.md` to record any approved value-semantics exceptions if implementation introduces them.
7. Produce `quickstart.md` with the recommended implementation order and validation commands.
8. Run `.specify/scripts/bash/update-agent-context.sh copilot` so the agent context reflects the planning decisions.

## Post-Design Constitution Re-check

- Modular monorepo and clean-architecture constraints: PASS.
- BFF gateway ownership rule: PASS — the plan keeps auth and protected-route orchestration in the BFF while identity/onboarding/downstream services remain the domain owners.
- Proto-first inter-service contract rule: PASS — the plan preserves `common.v1.Session` and `common.v1.Pagination` as the canonical shared contracts.
- Canonical integration-test placement/naming obligations: PASS — all verification remains under `backend/tests/integration/bff/`, `backend/tests/integration/identity/`, and `backend/tests/integration/cross_service/` with BDD + AAA expectations.
- Memory and instruction sync completion gate: PASS — explicit end-of-execution sync artifacts and tasks are included.

## Final Readiness Decision

- Decision: READY for `/speckit.tasks`.
- Evidence:
  - The spec’s route-scope, seeded-email, and pagination-default clarifications are resolved.
  - Research and design artifacts identify the concrete seed, BFF gateway, session-propagation, and test work required for implementation.
  - Constitution gates pass before and after design.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

No constitution violations requiring justification.
