# Implementation Plan: Restore BFF gRPC Gateway Boundary

**Branch**: `011-fix-bff-grpc-boundary` | **Date**: 2026-04-04 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/011-fix-bff-grpc-boundary/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

Complete the BFF boundary correction by removing the last direct domain-data shortcuts and replacing them with proper downstream gRPC contracts. The implementation will add a new `payments/v1` proto and gRPC transport for cycle-preference, history, and reconciliation capabilities, wire the payments gRPC server and client into `backend/cmd/payments/container.go` and `backend/cmd/bff/container.go`, perform an explicit audit of `backend/internals/bff/**` for any additional non-payments direct-access violations, restore supported BFF flows to normal data-returning behavior, add auth/project-membership regression coverage for the migrated routes, and finalize the required instruction/memory synchronization.

## Technical Context

**Language/Version**: Go 1.25.6  
**Primary Dependencies**: `github.com/labstack/echo/v4`, `github.com/danielgtaylor/huma/v2`, `go.uber.org/dig`, `google.golang.org/grpc`, generated protobuf clients in `backend/protos/generated/`, `go.uber.org/zap`, `github.com/go-playground/validator/v10`  
**Storage**: PostgreSQL for domain data; existing Redis/S3/RabbitMQ integrations remain unchanged for this feature  
**Testing**: `cd backend && go test ./...`, targeted BFF and payments unit tests, canonical BFF integration tests in `backend/tests/integration/bff/`, cross-service integration tests in `backend/tests/integration/cross_service/`, and explicit `401/403` + project-membership regression coverage for the migrated BFF routes  
**Target Platform**: Linux containerized backend services run locally via `make svc/run/<service>` and in CI  
**Project Type**: Backend multi-service monorepo refactor + contract-expansion feature  
**Performance Goals**: Preserve current supported-screen behavior and startup health while replacing in-process data shortcuts with the canonical single gRPC hop to the owning service  
**Constraints**: BFF must not access domain DB/repositories; new inter-service behavior must be gRPC-first and proto-versioned; AppError-first error propagation must remain intact; supported routes cannot remain on permanent dependency-error fallbacks  
**Scale/Scope**: `backend/internals/bff/**`, `backend/internals/payments/**`, `backend/protos/payments/v1/**`, `backend/protos/generated/payments/v1/**`, related tests, plus required `.specify/memory/*`, `/memories/repo/*`, and `.github/instructions/*` synchronization  
**Supported Scope Boundary**: Mandatory implementation and verification cover `bff`, `payments`, and `bills` for payment-cycle preference, history timeline, reconciliation, and the associated bills-backed dashboard/mark-paid interactions. `files`, `onboarding`, and `identity` are audit-in-scope and become implementation-in-scope only if the BFF audit uncovers direct-access violations in touched routes.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Checked against `.specify/memory/constitution.md` and `specs/011-fix-bff-grpc-boundary/spec.md`.

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
specs/011-fix-bff-grpc-boundary/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── payments-grpc-service-contract.md
│   ├── bff-route-migration-matrix.md
│   ├── governance-sync-checklist.md
│   └── pointer-exceptions.md
├── checklists/
│   └── requirements.md
└── tasks.md
```

### Source Code (repository root)

```text
backend/
├── cmd/
│   ├── bff/
│   │   └── container.go
│   └── payments/
│       └── container.go
├── internals/
│   ├── bff/
│   │   ├── interfaces/services.go
│   │   ├── services/
│   │   │   ├── payments_service.go
│   │   │   ├── history_service.go
│   │   │   └── reconciliation_service.go
│   │   └── transport/http/
│   │       ├── controllers/
│   │       ├── controllers/mappers/
│   │       ├── routes/
│   │       └── views/
│   └── payments/
│       ├── interfaces/
│       ├── services/
│       ├── repositories/
│       └── transport/grpc/        # new server implementation replaces current `.gitkeep`
├── protos/
│   ├── payments/v1/               # new proto module to add
│   │   ├── messages.proto
│   │   └── grpc.proto
│   └── generated/payments/v1/     # regenerated artifacts
└── tests/
    ├── integration/bff/
    └── integration/cross_service/
```

**Structure Decision**: This is a backend-only contract and transport expansion centered on the BFF/payments boundary. The BFF keeps its existing HTTP views/controllers/mappers structure while the payments domain gains the missing gRPC transport and proto surface it needs to remain the owner of cycle-preference, history, and reconciliation flows.

## Mandatory End-of-Execution Sync

**Memory Diagram Sync (required)**:
- Impacted memory files: `.specify/memory/architecture-diagram.md`, `.specify/memory/bff-flows.md`, `.specify/memory/payments-service-flows.md`, plus any additional `.specify/memory/*-service-flows.md` file touched if new violations are discovered outside payments.
- Impacted repository memory files: `/memories/repo/bff-service-boundary-conventions.md` and any related repo-memory note updated to reflect the final gRPC path.
- Update tasks required in `tasks.md`: Yes
- If `No`, rationale: N/A

**Instruction Sync (required for refactor/reorganization)**:
- Refactor/reorganization in scope: Yes
- If `Yes`, impacted instruction files to update:
  - `.github/instructions/architecture.instructions.md`
  - `.github/instructions/project-structure.instructions.md`
  - `.github/instructions/ai-behavior.instructions.md`
- If `Yes`, impacted workflow templates to update:
  - None planned unless implementation uncovers a reusable planning-rule change

**Completion gate**:
Implementation is not complete until the new payments gRPC boundary is in place for supported flows and all required sync tasks above are executed in the same feature cycle.

## Phase 0: Research and Decisions

1. Inventory every remaining BFF direct-access violation across `backend/internals/bff/**` and classify it by owning service, even if the final migration work is concentrated on the supported payment-facing routes.
2. Define the new `payments/v1` gRPC surface needed for cycle-preference, history analytics, and reconciliation flows.
3. Preserve the current `bills` ownership for bill-dashboard and mark-paid operations already served by `bills.v1.BillsService`.
4. Determine the BFF migration sequence so controllers stay thin and `services/contracts`/`controllers/mappers` ownership remains intact during the swap to gRPC.
5. Define AppError-first transport behavior, rollout sequencing, and regression verification obligations.
6. Define explicit auth, authorization, and project-membership regression checks for the migrated BFF routes so the gateway responsibilities remain provable after the transport change.

## Phase 1: Design Outputs

1. Produce `research.md` with the chosen migration and contract decisions.
2. Produce `data-model.md` for the payments-owned analytics and reconciliation projections that will cross the new gRPC boundary.
3. Produce `contracts/payments-grpc-service-contract.md` describing the new inter-service RPC surface and domain ownership split.
4. Produce `contracts/bff-route-migration-matrix.md` mapping each affected BFF route to its owning downstream RPC and required regression tests.
5. Produce `contracts/governance-sync-checklist.md` for mandatory instruction/memory updates.
6. Produce `contracts/pointer-exceptions.md` to track any approved value-semantics exceptions on modified backend boundaries.
7. Produce `quickstart.md` with the recommended implementation order and validation commands.
8. Run `.specify/scripts/bash/update-agent-context.sh copilot`.

## Post-Design Constitution Re-check

- Modular monorepo and clean-architecture constraints: PASS.
- BFF gateway ownership rule: PASS — design requires BFF → gRPC → owning service, not BFF → repo/DB.
- Proto-first inter-service contract rule: PASS — new cross-service capability is modeled as a versioned `payments/v1` proto surface.
- Canonical integration-test placement/naming obligations: PASS — route and cross-service verification remains under `backend/tests/integration/bff/` and `backend/tests/integration/cross_service/`.
- Memory and instruction sync completion gate: PASS — explicit sync artifacts and tasks are included.

## Final Readiness Decision

- Decision: READY for `/speckit.tasks`.
- Evidence:
  - The missing-scope clarification is resolved in `spec.md`.
  - Research and design artifacts define the required payments gRPC contract surface and migration path.
  - Constitution gates pass before and after design.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

No constitution violations requiring justification.
