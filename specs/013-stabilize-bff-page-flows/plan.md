# Implementation Plan: Stabilize Broken BFF Page Flows

**Branch**: `013-stabilize-bff-page-flows` | **Date**: 2026-04-04 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/013-stabilize-bff-page-flows/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

Stabilize the broken `documents`, `payments`, `analyses`, and `settings` frontend flows by auditing the exact frontend request inventory, fixing the failing BFF-to-downstream behavior at the owning service boundary, adding canonical end-to-end integration coverage, and defining a populated default-user seed path so the four pages render meaningful data without blocker-level request errors.

## Technical Context

**Language/Version**: Go 1.25.6 for backend services; TypeScript 5.8 / React 18 / Vite 6 for the frontend  
**Primary Dependencies**: `github.com/labstack/echo/v4`, `github.com/danielgtaylor/huma/v2`, `go.uber.org/dig`, `google.golang.org/grpc`, generated protobuf clients under `backend/protos/generated/`, `go.uber.org/zap`, `github.com/go-playground/validator/v10`, frontend `@tanstack/react-query`, `zod`, and `vitest`  
**Storage**: PostgreSQL-backed service databases (`identity`, `files`, `bills`, `payments`, `onboarding`) plus testcontainers-backed ephemeral Postgres for integration validation  
**Testing**: `cd backend && go test ./...`, `make test/integration`, canonical BFF tests in `backend/tests/integration/bff/`, payments and cross-service flows in `backend/tests/integration/payments/` and `backend/tests/integration/cross_service/`, and `make frontend/test` for hook-level verification  
**Target Platform**: Linux-based local/CI containers for backend services and browser-based frontend validation via Vite  
**Project Type**: Full-stack Go + React monorepo with a thin BFF gateway and gRPC-based downstream service ownership  
**Performance Goals**: The default user can open all four in-scope pages without blocker-level request errors; request failures degrade to supported auth/access/empty-state responses rather than generic 500s; a new tester can bring up the populated demo environment in under 10 minutes  
**Constraints**: Preserve the BFF gateway boundary, fix issues at the owning downstream/service or mapping layer, keep AppError-first translation behavior, use canonical backend integration-test placement and naming, and limit seed data changes to local/test/demo environments  
**Scale/Scope**: Frontend pages/hooks/services for `UploadPage.tsx`, `PaymentDashboardPage.tsx`, `HistoryDashboardPage.tsx`, `ReconciliationPage.tsx`, and `SettingsPage.tsx`; backend BFF routes/controllers/services for `documents`, `payments`, `history`, `reconciliation`, and `settings`; downstream `files`, `bills`, `payments`, `identity`, and `onboarding` only if auth/project membership gaps are part of the failure path  
**Verified Route Scope**: `GET /api/v1/documents`, `POST /api/v1/documents/upload`, `GET /api/v1/bills/payment-dashboard`, `POST /api/v1/bills/{billId}/mark-paid`, `GET/POST /api/v1/payment-cycle/preferred-day`, `GET /api/v1/history/timeline`, `GET /api/v1/history/categories`, `GET /api/v1/reconciliation/summary`, `POST /api/v1/reconciliation/links`, and `GET/POST/DELETE /api/v1/bank-accounts`

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Checked against `.specify/memory/constitution.md` and `specs/013-stabilize-bff-page-flows/spec.md`.

Mandatory gates:
- [x] `spec.md` includes `Architecture & Memory Diagram Flow Impact` with impacted
  `.specify/memory/*.md` file list or explicit no-impact rationale.
- [x] End-of-execution tasks include updates to all impacted service-flow memory files
  and `.specify/memory/architecture-diagram.md` when cross-service flow changes exist.
- [x] If the feature is refactor/reorganization, plan includes explicit instruction
  update tasks for impacted `.github/instructions/*.instructions.md` files. (N/A here; refactor is not the planned outcome.)
- [x] If workflow behavior changes, plan includes updates to impacted
  `.specify/templates/*.md` files. (N/A here unless implementation discovers a reusable planning-rule change.)
- [x] If backend integration behavior is in scope, plan includes canonical test-placement
  and naming compliance tasks (`backend/tests/integration/<service>/` or
  `backend/tests/integration/cross_service/`, behavior-based snake_case, BDD + AAA).
- [x] If BFF boundaries are modified, plan includes explicit tasks for service-contract ownership (`services/contracts`), mapper-boundary enforcement (`controllers/mappers`), pointer-policy checks, and exception documentation.

Gate status: PASS.

## Project Structure

### Documentation (this feature)

```text
specs/013-stabilize-bff-page-flows/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── default-user-seed-contract.md
│   ├── page-flow-validation-matrix.md
│   └── pointer-exceptions.md
├── checklists/
│   └── requirements.md
└── tasks.md
```

### Source Code (repository root)

```text
frontend/src/
├── pages/
│   ├── UploadPage.tsx
│   ├── PaymentDashboardPage.tsx
│   ├── HistoryDashboardPage.tsx
│   ├── ReconciliationPage.tsx
│   └── SettingsPage.tsx
├── hooks/
│   ├── useDocuments.ts
│   ├── useUploadDocument.ts
│   ├── usePaymentDashboard.ts
│   ├── usePreferredDay.ts
│   ├── useHistoryDashboard.ts
│   ├── useReconciliationSummary.ts
│   └── useBankAccounts.ts
└── services/
    ├── documentsApi.ts
    ├── paymentsApi.ts
    ├── historyApi.ts
    ├── reconciliationApi.ts
    └── bankAccountsApi.ts

backend/internals/bff/
├── services/
│   ├── documents_service.go
│   ├── payments_service.go
│   ├── history_service.go
│   ├── reconciliation_service.go
│   └── settings_service.go
└── transport/http/
    ├── controllers/
    ├── controllers/mappers/
    ├── routes/
    │   ├── documents_routes.go
    │   ├── payments_routes.go
    │   ├── history_routes.go
    │   ├── reconciliation_routes.go
    │   └── settings_routes.go
    └── views/

backend/tests/integration/
├── bff/
│   ├── documents_routes_registration_test.go
│   ├── payments_routes_registration_test.go
│   ├── history_routes_registration_test.go
│   ├── reconciliation_routes_registration_test.go
│   └── settings_routes_registration_test.go
├── payments/
│   ├── view_payment_dashboard_test.go
│   └── mark_bill_paid_idempotency_test.go
└── cross_service/
    ├── auto_reconcile_transactions_test.go
    ├── create_manual_reconciliation_link_test.go
    └── get_history_timeline_test.go
```

**Structure Decision**: This feature is a backend-first stabilization effort anchored in the existing monorepo layout. The frontend pages and hooks define the request inventory, the BFF remains the single HTTP gateway, and the owning downstream services (`files`, `bills`, `payments`, `identity`, and possibly `onboarding`) keep control of business data while the new artifacts capture the flow map, seed expectations, and verification strategy.

## Mandatory End-of-Execution Sync

**Memory Diagram Sync (required)**:
- Impacted memory files: `.specify/memory/architecture-diagram.md`, `.specify/memory/bff-flows.md`, `.specify/memory/files-service-flows.md`, `.specify/memory/bills-service-flows.md`, `.specify/memory/payments-service-flows.md`, `.specify/memory/identity-service-flows.md`, and `.specify/memory/onboarding-service-flows.md` if the auth/project bootstrap path is part of the root cause.
- Impacted repository memory files: `/memories/repo/bff-service-boundary-conventions.md` and any new repo note needed to preserve the verified default-user seed/runbook.
- Update tasks required in `tasks.md`: Yes
- If `No`, rationale: N/A

**Instruction Sync (required for refactor/reorganization)**:
- Refactor/reorganization in scope: No
- If `Yes`, impacted instruction files to update:
  - None planned
- If `Yes`, impacted workflow templates to update:
  - None planned

**Completion gate**:
Implementation is not complete until the four in-scope pages are validated with the default user, the new integration coverage is added in canonical locations, and all required memory-sync updates are completed in the same feature cycle.

## Phase 0: Research and Decisions

1. Build the exact request inventory for `documents`, `payments`, `analyses` (`history` + `reconciliation`), and `settings` by tracing frontend page → hook → service → BFF route → downstream gRPC ownership.
2. Reproduce the current failures with the default user and classify each one as an auth/session gap, project-guard problem, downstream contract mismatch, missing seed data, or AppError translation/empty-state issue.
3. Audit the default-user data dependencies across `identity`, `onboarding`, `files`, `bills`, and `payments` so the canonical acceptance path produces populated content on all four pages.
4. Preserve the BFF gateway boundary: fix failures in the owning downstream service or in the service/mapper translation layer rather than bypassing the boundary in controllers or the frontend.
5. Define the regression matrix for successful access, unauthenticated access, wrong-project access, expected empty-state behavior, and the most likely failure hotspots already suggested by the codebase (`history` project guarding, settings/files mapping, payment-cycle edge handling, and duplicate/double error translation behavior).

## Phase 1: Design Outputs

1. Produce `research.md` with the confirmed flow inventory, failure hypotheses, and chosen stabilization strategy.
2. Produce `data-model.md` for the authenticated default user, project membership, documents, payments/dashboard state, history analytics, settings bank-account labels, and reconciliation entities touched by the four pages.
3. Produce `contracts/page-flow-validation-matrix.md` mapping every in-scope page request to its owning BFF route, downstream service, expected success state, and intended integration-test placement.
4. Produce `contracts/default-user-seed-contract.md` defining the representative populated data required for `documents`, `payments`, `analyses`, and `settings` in local/test/demo environments.
5. Produce `contracts/pointer-exceptions.md` to record any approved value-semantics exceptions if implementation introduces them.
6. Produce `quickstart.md` with the recommended implementation order, validation commands, and developer runbook for reproducing the stabilized experience.
7. Run `.specify/scripts/bash/update-agent-context.sh copilot` so the Copilot context reflects the planning decisions.

## Post-Design Constitution Re-check

- Modular monorepo and clean-architecture constraints: PASS.
- BFF gateway ownership rule: PASS — the plan fixes page failures while keeping domain data owned by downstream services.
- Proto-first and shared-session/pagination expectations: PASS — the plan assumes continued gRPC contract ownership and validates list flows through the BFF rather than bypassing them.
- Canonical integration-test placement/naming obligations: PASS — the new regression coverage will live under `backend/tests/integration/bff/`, `backend/tests/integration/payments/`, and `backend/tests/integration/cross_service/` using behavior-based snake_case files and BDD + AAA scenarios.
- Memory and repo-instruction sync completion gate: PASS — explicit sync targets are listed above and must remain in-scope until the feature closes.

## Final Readiness Decision

- Decision: READY for `/speckit.tasks`.
- Evidence:
  - The in-scope page flows, owning routes, and downstream service boundaries are now identified.
  - The seeded-data expectation is clarified: the canonical default user must see populated content on all four pages.
  - The research/design outputs define how to stabilize the backend behavior without violating the BFF boundary or test-organization rules.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

No constitution violations requiring justification.
