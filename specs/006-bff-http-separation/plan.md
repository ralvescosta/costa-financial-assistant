# Implementation Plan: BFF HTTP Boundary Separation

**Branch**: `[006-bff-http-separation]` | **Date**: 2026-04-02 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/006-bff-http-separation/spec.md`

## Summary

Complete the BFF transport refactor so all 20 active BFF HTTP routes follow the same boundary model: route modules own Huma registration, controllers perform only HTTP parsing/validation/response translation, all HTTP request and response structs move into `transport/http/views`, and downstream gRPC or repository orchestration shifts into BFF service interfaces. The implementation keeps the current Echo + Huma transport and existing route surface stable while adding a validator-driven contract check path and explicit route-to-test coverage documentation.

## Technical Context

**Language/Version**: Go 1.25.6  
**Primary Dependencies**: Echo v4.15.1, Huma v2.37.3, Dig v1.19.0, Zap v1.27.1, gRPC v1.80.0 generated clients, Testify v1.11.1, go-playground/validator for controller-side HTTP contract validation  
**Storage**: Existing downstream PostgreSQL-backed services and payments repositories; no new persistent store introduced  
**Testing**: `go test`, integration build tag `integration`, `httptest`, Testify, OpenAPI metadata checks, route registration smoke tests, existing integration `TestMain` bootstrap  
**Target Platform**: Linux backend services inside the monorepo local/dev pipeline  
**Project Type**: Backend web-service feature inside the BFF service  
**Performance Goals**: Preserve externally visible behavior for all 20 active BFF routes, add no extra network hops in request handling, and avoid observable startup or routing regressions  
**Constraints**: No public path or method changes, no middleware order regressions, controllers cannot call gRPC clients or repositories directly, all HTTP contracts must live in `transport/http/views`, validator tags are required on fields needing validation, and integration coverage must remain complete for all active routes  
**Scale/Scope**: 6 active BFF route modules, 6 controller domains, 20 Huma operations, BFF DI wiring in `backend/cmd/bff/container.go`, new BFF service contracts in `backend/internals/bff/services` and `backend/internals/bff/interfaces`, and integration coverage updates in `backend/tests/integration/`

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Modular monorepo boundaries**: PASS. All implementation work stays inside the BFF service, its interfaces/services, and the shared backend integration-test area.
- **BFF framework constraint (Echo + Huma)**: PASS. The plan preserves Echo + Huma and keeps Huma operation registration in dedicated route modules.
- **MVC / controller responsibility rule**: CONDITIONAL PASS. The constitution already requires routes to own `huma.Register(...)` and controllers to avoid business logic, but it does not yet codify the `transport/http/views` layer or explicitly forbid direct downstream client orchestration from controllers. Implementation must update governance and instruction files to make the new HTTP boundary model the required standard.
- **SOLID / clean architecture**: PASS. Controllers will depend on BFF service interfaces, services will own downstream orchestration, and route capability interfaces will stay consumer-defined and narrow.
- **OpenAPI documentation continuity**: CONDITIONAL PASS. The feature adds controller-side validation with `validator` tags; design must preserve Huma-compatible transport metadata on view structs so generated OpenAPI remains complete.
- **Testing discipline**: PASS. Integration tests remain under `backend/tests/integration/`, and the route coverage matrix keeps all 20 active routes explicitly mapped to at least one passing suite.

**Post-Design Re-check**: PASS, provided implementation includes governance updates for the views layer and controller/service boundary, preserves OpenAPI metadata on moved view structs, and keeps the route coverage matrix synchronized with integration suites.

## Project Structure

### Documentation (this feature)

```text
specs/006-bff-http-separation/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── route-controller-service-contract.md
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
│       ├── interfaces/
│       │   ├── grpc_clients.go
│       │   └── services.go
│       ├── services/
│       │   ├── documents_service.go
│       │   ├── history_service.go
│       │   ├── payments_service.go
│       │   ├── projects_service.go
│       │   ├── reconciliation_service.go
│       │   └── settings_service.go
│       └── transport/
│           └── http/
│               ├── controllers/
│               │   ├── base_controller.go
│               │   ├── documents_controller.go
│               │   ├── history_controller.go
│               │   ├── payments_controller.go
│               │   ├── projects_controller.go
│               │   ├── reconciliation_controller.go
│               │   └── settings_controller.go
│               ├── middleware/
│               ├── routes/
│               │   ├── contracts.go
│               │   ├── documents_routes.go
│               │   ├── history_routes.go
│               │   ├── payments_routes.go
│               │   ├── projects_routes.go
│               │   ├── reconciliation_routes.go
│               │   └── settings_routes.go
│               └── views/
│                   ├── documents_views.go
│                   ├── history_views.go
│                   ├── payments_views.go
│                   ├── projects_views.go
│                   ├── reconciliation_views.go
│                   └── settings_views.go
└── tests/
    └── integration/
        ├── bff_route_registration_smoke_test.go
        ├── documents_routes_integration_test.go
        ├── history_routes_integration_test.go
        ├── openapi_contract_test.go
        ├── payments_routes_integration_test.go
        ├── projects_routes_integration_test.go
        ├── reconciliation_routes_integration_test.go
        ├── settings_routes_integration_test.go
        └── testmain_test.go
```

**Structure Decision**: Keep the existing BFF transport decomposition, preserve the `routes/` package, add a dedicated `views/` package as the sole HTTP contract owner, and introduce BFF service/interface files so controllers stop orchestrating gRPC clients and repositories directly.

## Phase 0: Research Focus

- Define the least disruptive way to move all HTTP request/response structs from controllers into `transport/http/views` without weakening Huma binding or OpenAPI generation.
- Define the controller-to-service boundary so controllers no longer call generated gRPC clients or repositories directly.
- Define the route capability contract shape after views move out of controllers and ensure route coverage remains explicit for all 20 active operations.

## Phase 1: Design Focus

- Design the `views/` package layout, validation rules, and transport-to-service mapping responsibilities.
- Design BFF service interfaces and concrete services per route domain, including how they wrap downstream gRPC clients and existing payments services.
- Design route-controller-service contracts, route coverage documentation, container wiring changes, and validation updates needed in tests and governance docs.

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| Constitution and instruction set do not yet codify the `transport/http/views` layer or explicit controller ban on direct downstream orchestration | The feature exists specifically to make that boundary mandatory across all active BFF routes | Keeping the current governance text would leave the new structure optional and allow regression to mixed controller responsibilities |
