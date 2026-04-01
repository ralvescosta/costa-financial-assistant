# Implementation Plan: BFF Route-Controller Segregation

**Branch**: `[004-segregate-bff-routing]` | **Date**: 2026-04-01 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/004-segregate-bff-routing/spec.md`

## Summary

Refactor the BFF HTTP transport so route registration lives in dedicated route modules while controllers remain behavior-only structs. The implementation will add a route layer adjacent to the existing controllers, define narrow controller capability contracts consumed by routes, preserve all 20 existing Huma operations and middleware semantics, and standardize integration coverage with an explicit route-to-test matrix.

## Technical Context

**Language/Version**: Go 1.25.6  
**Primary Dependencies**: Echo v4.15.1, Huma v2.37.3, Dig v1.19.0, Zap v1.27.1, gRPC generated clients, Testify v1.11.1  
**Storage**: PostgreSQL-backed integration environment for existing BFF dependencies; no new persistent store introduced  
**Testing**: `go test`, integration build tag `integration`, `httptest`, Testify, existing ephemeral DB `TestMain` bootstrap  
**Target Platform**: Linux backend services in the monorepo local/dev pipeline  
**Project Type**: Backend web-service feature inside the BFF service  
**Performance Goals**: Preserve existing request/response behavior for 20 BFF routes and avoid observable startup or routing regressions  
**Constraints**: No public path or method changes, no middleware order regressions, no business logic inside routes, all declared routes must have passing integration coverage  
**Scale/Scope**: 6 existing BFF controller modules, 20 Huma routes, BFF container wiring, and integration coverage updates in `backend/tests/integration/`

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Modular monorepo boundaries**: PASS. Changes stay inside the BFF service and the shared integration-test area.
- **BFF framework constraint (Echo + Huma)**: PASS. The plan preserves Echo + Huma and only relocates route registration ownership.
- **MVC / controller responsibility rule**: CONDITIONAL PASS. The current constitution still says controllers own `Register(api huma.API)`. This feature intentionally replaces that rule. Implementation must include a constitution and instruction update that codifies dedicated route modules as registration owners before merge.
- **SOLID / consumer-driven interfaces**: PASS. Route modules will define narrow capability interfaces that concrete controllers satisfy, avoiding broad transport interfaces.
- **Security / middleware continuity**: PASS. The route contract keeps auth middleware and project-guard middleware on the same operations in the same order.
- **Testing discipline**: PASS. Integration tests remain under `backend/tests/integration/` and will be expanded to cover all 20 routes with BDD-style scenarios and a maintainable coverage matrix.

**Post-Design Re-check**: PASS, provided implementation includes the governance update for the new route layer and keeps route inventory coverage synchronized with integration tests.

## Project Structure

### Documentation (this feature)

```text
specs/004-segregate-bff-routing/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── route-controller-contract.md
│   └── route-coverage-matrix.md
└── tasks.md
```

### Source Code (repository root)

```text
backend/
├── cmd/
│   └── bff/
│       └── container.go
├── internals/
│   └── bff/
│       └── transport/
│           └── http/
│               ├── controllers/
│               │   ├── documents_controller.go
│               │   ├── history_controller.go
│               │   ├── payments_controller.go
│               │   ├── projects_controller.go
│               │   ├── reconciliation_controller.go
│               │   └── settings_controller.go
│               ├── middleware/
│               └── routes/
│                   ├── contracts.go
│                   ├── documents_routes.go
│                   ├── history_routes.go
│                   ├── payments_routes.go
│                   ├── projects_routes.go
│                   ├── reconciliation_routes.go
│                   └── settings_routes.go
└── tests/
    └── integration/
        ├── openapi_contract_test.go
        ├── documents_routes_integration_test.go
        ├── history_routes_integration_test.go
        ├── payments_routes_integration_test.go
        ├── projects_routes_integration_test.go
        ├── reconciliation_routes_integration_test.go
        ├── settings_routes_integration_test.go
        └── testmain_test.go
```

**Structure Decision**: Keep the existing BFF transport layer intact and add a dedicated `routes/` package beside `controllers/`. Controllers keep request-handling behavior, while route modules own Huma operation declarations and dependency-driven registration.

## Phase 0: Research Focus

- Confirm the least disruptive route-module structure for the six existing BFF resources.
- Resolve how to model “default controller interface” in a Go-idiomatic way without creating a broad, unused contract.
- Define the route coverage strategy that proves all 20 existing Huma operations remain reachable and guarded correctly.

## Phase 1: Design Focus

- Define the route/controller contract and DI boundaries.
- Define the route inventory and route-to-test coverage mapping.
- Define the migration path for controller files, container wiring, and metadata validation tests.

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| Constitution currently assigns route registration to controllers | The feature exists specifically to move route declaration into dedicated route modules | Keeping `Register()` on controllers would fail the primary requirement and preserve the current responsibility mixing |