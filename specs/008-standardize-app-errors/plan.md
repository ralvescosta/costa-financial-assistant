# Implementation Plan: Standardize Backend App Errors

**Branch**: `008-standardize-app-errors` | **Date**: 2026-04-03 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/008-standardize-app-errors/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

Standardize backend error propagation so all cross-layer failures use `AppError`, never raw dependency errors. The implementation will establish and enforce a centralized error catalog in `backend/pkgs/errors/consts.go`, define deterministic translation rules from dependency/library failures to cataloged `AppError` entries, and require structured one-time boundary logging (`zap.Error(err)`) before sanitizing propagation. Retryability semantics are captured now through explicit `Retryable` classification to support future retry orchestration.

## Technical Context

**Language/Version**: Go 1.25.6  
**Primary Dependencies**: `go.uber.org/zap`, `google.golang.org/grpc`, `github.com/lib/pq`, `github.com/labstack/echo/v4`, `github.com/danielgtaylor/huma/v2`, `go.uber.org/dig`  
**Storage**: PostgreSQL plus object/file storage paths already in backend services; no schema change in this feature  
**Testing**: `go test ./...`, targeted service tests via `make svc/test/<service>`, integration checks via `make test/integration`  
**Target Platform**: Linux backend services in Docker/local dev workflow
**Project Type**: Backend multi-service monorepo governance and implementation standardization  
**Performance Goals**: No regression to service latency/error throughput; preserve existing operational behavior while standardizing error contracts  
**Constraints**: No raw dependency error leakage across layers; one boundary log before translation; retryability classification required for each cataloged error; no breaking of clean-architecture boundaries  
**Scale/Scope**: Cross-cutting update across backend services (`bff`, `bills`, `files`, `identity`, `onboarding`, `payments`) and shared package `backend/pkgs/errors`

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Checked against `.specify/memory/constitution.md` and feature spec `008-standardize-app-errors`.

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
specs/008-standardize-app-errors/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── backend-error-propagation-contract.md
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
│   ├── bills/
│   ├── files/
│   ├── identity/
│   ├── onboarding/
│   └── payments/
├── pkgs/
│   └── errors/
│       ├── error.go
│       └── consts.go
└── tests/
  └── integration/

specs/
└── 008-standardize-app-errors/
  ├── plan.md
  ├── research.md
  ├── data-model.md
  ├── quickstart.md
  └── contracts/
    └── backend-error-propagation-contract.md
```

**Structure Decision**: Backend-focused cross-cutting feature spanning `backend/pkgs/errors` and service-layer propagation paths in each backend domain module, with governance artifacts under `specs/008-standardize-app-errors`.

## Mandatory End-of-Execution Sync

**Memory Diagram Sync (required)**:
- Impacted memory files: `.specify/memory/bff-flows.md`, `.specify/memory/files-service-flows.md`, `.specify/memory/bills-service-flows.md`, `.specify/memory/identity-service-flows.md`, `.specify/memory/onboarding-service-flows.md`.
- Update tasks required in `tasks.md`: Yes
- If `No`, rationale: N/A

**Instruction Sync (required for refactor/reorganization)**:
- Refactor/reorganization in scope: Yes
- If `Yes`, impacted instruction files to update:
  - `.github/instructions/observability.instructions.md`
  - `.github/instructions/security.instructions.md`
  - `.github/instructions/golang.instructions.md`
  - `.github/instructions/ai-behavior.instructions.md`
  - `.github/instructions/architecture.instructions.md`
- If `Yes`, impacted workflow templates to update:
  - none currently expected; validate and update only if implementation introduces reusable workflow changes.

**Constitution amendment obligations (required when constitution content changes)**:
- If `.specify/memory/constitution.md` is changed, the same feature cycle MUST include:
  - semantic version update in constitution header (`MAJOR`/`MINOR`/`PATCH` as applicable),
  - `SYNC IMPACT REPORT` update describing rationale and impacted templates/prompts,
  - explicit evidence of dependent template/prompt re-validation in feature contracts.

**Completion gate**:
Implementation is not complete until all required sync tasks above are executed in the
same feature cycle.

## Phase 0: Research and Decisions

1. Catalogue known backend failure categories per service and shared package boundaries.
2. Define deterministic retryability criteria for transient versus deterministic failures.
3. Define translation-boundary logging rule (single structured log at translation boundary).
4. Define unknown-failure fallback behavior and propagation guarantees.

## Phase 1: Design Outputs

1. Produce `research.md` with finalized decisions and rejected alternatives.
2. Produce `data-model.md` capturing entities: `AppError`, `ErrorCatalogEntry`, `TranslationRule`, and `ErrorEventLogRecord`.
3. Produce `contracts/backend-error-propagation-contract.md` with cross-layer invariants and acceptance mapping.
4. Produce `quickstart.md` with implementation and validation workflow.
5. Run `.specify/scripts/bash/update-agent-context.sh copilot`.

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
