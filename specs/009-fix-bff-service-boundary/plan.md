# Implementation Plan: Enforce Service Boundary Contracts

**Branch**: `009-fix-bff-service-boundary` | **Date**: 2026-04-03 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/009-fix-bff-service-boundary/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

Refactor the BFF so service-layer orchestration no longer depends on HTTP view contracts, enforcing strict transport-to-service mapping boundaries across all active BFF routes/services. In parallel, establish and apply backend-wide pointer-signature conventions for structs crossing function boundaries, then synchronize both architecture memory and instruction artifacts so the pattern remains mandatory in future implementation cycles.

## Technical Context

**Language/Version**: Go 1.25.6  
**Primary Dependencies**: `github.com/labstack/echo/v4`, `github.com/danielgtaylor/huma/v2`, `go.uber.org/dig`, generated gRPC clients from `backend/protos/generated`, `go.uber.org/zap`  
**Storage**: PostgreSQL and existing storage integrations are unchanged by this feature  
**Testing**: `go test ./...`, targeted service tests, and canonical backend integration tests under `backend/tests/integration/`  
**Target Platform**: Linux backend services executed in containerized local/dev and CI environments
**Project Type**: Backend multi-service monorepo refactor/governance feature  
**Performance Goals**: Reduce avoidable struct copy overhead on hot call paths while preserving endpoint behavior and latency characteristics  
**Constraints**: No HTTP view types in BFF services, full active-route rollout, pointer policy threshold (`reference-like fields` OR `size > 3 machine words`), no API behavior regressions  
**Scale/Scope**: All active BFF routes/services plus backend convention propagation across `bff`, `files`, `bills`, `identity`, `onboarding`, `payments`

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Checked against `.specify/memory/constitution.md` and feature spec `009-fix-bff-service-boundary`.

Mandatory gates:
- [x] `spec.md` includes `Architecture & Memory Diagram Flow Impact` with impacted
  `.specify/memory/*.md` file list or explicit no-impact rationale.
- [x] End-of-execution tasks include updates to all impacted service-flow memory files
  and `.specify/memory/architecture-diagram.md` when cross-service flow changes exist.
- [x] If the feature is refactor/reorganization, plan includes explicit instruction
  update tasks for impacted `.github/instructions/*.instructions.md` files.
- [x] If workflow behavior changes, plan includes updates to impacted
  `.specify/templates/*.md` files.
- [x] If backend integration behavior is in scope, plan includes canonical test-placement
  and naming compliance tasks (`backend/tests/integration/<service>/` or
  `backend/tests/integration/cross_service/`, behavior-based snake_case, BDD + AAA).

Gate status: PASS.

## Project Structure

### Documentation (this feature)

```text
specs/009-fix-bff-service-boundary/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── bff-service-boundary-contract.md
│   ├── bff-route-behavior-regression-matrix.md
│   └── nil-safety-boundary-checklist.md
├── checklists/
│   └── requirements.md
└── tasks.md
```

### Source Code (repository root)

```text
backend/
├── cmd/
│   ├── bff/
│   ├── bills/
│   ├── files/
│   ├── identity/
│   ├── onboarding/
│   └── payments/
├── internals/
│   ├── bff/
│   │   ├── interfaces/
│   │   ├── services/
│   │   └── transport/http/
│   │     ├── controllers/
│   │     ├── middleware/
│   │     ├── routes/
│   │     └── views/
│   ├── bills/
│   ├── files/
│   ├── identity/
│   ├── onboarding/
│   └── payments/
├── pkgs/
│   └── errors/
└── tests/
  └── integration/

.github/
└── instructions/

.specify/
└── memory/

specs/
└── 009-fix-bff-service-boundary/
```

**Structure Decision**: Backend-centric refactor with primary code impact in `backend/internals/bff` and governance impact in `.github/instructions/`, `.specify/memory/`, and `/memories/repo/`; no frontend code-path changes are planned in this feature.

## Mandatory End-of-Execution Sync

**Memory Diagram Sync (required)**:
- Impacted memory files: `.specify/memory/architecture-diagram.md`, `.specify/memory/bff-flows.md`, `.specify/memory/files-service-flows.md`, `.specify/memory/bills-service-flows.md`, `.specify/memory/identity-service-flows.md`, `.specify/memory/onboarding-service-flows.md`, `.specify/memory/payments-service-flows.md` (create if absent)
- Impacted repository memory files: `/memories/repo/backend-go-module-root.md`, `/memories/repo/bff-service-boundary-conventions.md` (create or update)
- Update tasks required in `tasks.md`: Yes
- If `No`, rationale: N/A

**Instruction Sync (required for refactor/reorganization)**:
- Refactor/reorganization in scope: Yes
- If `Yes`, impacted instruction files to update:
  - `.github/instructions/architecture.instructions.md`
  - `.github/instructions/project-structure.instructions.md`
  - `.github/instructions/golang.instructions.md`
  - `.github/instructions/coding-conventions.instructions.md`
  - `.github/instructions/testing.instructions.md`
  - `.github/instructions/ai-behavior.instructions.md`
- If `Yes`, impacted workflow templates to update:
  - `.specify/templates/spec-template.md` (if new reusable clarification rules are added)
  - `.specify/templates/plan-template.md` (if recurring governance checks are added)

**Completion gate**:
Implementation is not complete until all required sync tasks above are executed in the
same feature cycle.

## Phase 0: Research and Decisions

1. Define BFF boundary ownership rules for views, controllers, service contracts, and mappers.
2. Define contract-shape selection order (`proto message` first, `service-owned struct` when proto is insufficient).
3. Define pointer policy semantics and exception documentation standards for backend signatures.
4. Define endpoint behavior-preservation verification scope (status semantics and response shapes) for all active BFF routes.
5. Define mapper/service nil-safety boundary expectations and test obligations.
6. Define governance-sync obligations for `.specify/memory/*`, `/memories/repo/*`, and instruction files.

## Phase 1: Design Outputs

1. Produce `research.md` with final decisions and rejected alternatives.
2. Produce `data-model.md` with entities: `TransportViewContract`, `ServiceContract`, `BoundaryMapper`, `PointerConventionRule`, and `PointerExceptionRecord`.
3. Produce `contracts/bff-service-boundary-contract.md` with enforceable cross-layer invariants and acceptance mapping.
4. Produce `contracts/bff-route-behavior-regression-matrix.md` with per-route status/response assertions.
5. Produce `contracts/nil-safety-boundary-checklist.md` with mapper/service nil-case coverage expectations.
6. Produce `quickstart.md` with implementation order and validation commands.
7. Run `.specify/scripts/bash/update-agent-context.sh copilot`.

## Post-Design Constitution Re-check

- Memory impact enforcement: PASS.
- Instruction-sync explicitness for refactor scope: PASS.
- Canonical integration-test placement/naming requirement acknowledged in execution tasks: PASS.
- Workflow-template sync conditional check included: PASS.

## Final Readiness Decision

- Decision: READY for `/speckit.tasks`.
- Evidence:
  - Clarifications resolved in `spec.md`.
  - Phase 0 and Phase 1 artifacts generated.
  - Constitution gates passed pre-design and post-design.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

No constitution violations requiring justification.
