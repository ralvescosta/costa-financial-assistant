# Quickstart: BFF HTTP Boundary Separation

## Goal

Validate that all active BFF HTTP routes use the same boundary model: routes own registration, views own HTTP contracts, controllers stay HTTP-only, and services own downstream orchestration.

## Prerequisites

- Go 1.25.6 available locally
- Backend module dependencies installed from `backend/go.mod`
- PostgreSQL integration database reachable through `TEST_DB_DSN` or the existing local integration setup for full integration runs
- Downstream gRPC test doubles or existing integration dependencies available through the current backend integration bootstrap

## Implementation Outline

1. Create `backend/internals/bff/transport/http/views/` and move all HTTP request/response structs there by resource.
2. Add BFF service interfaces and concrete services in `backend/internals/bff/interfaces/` and `backend/internals/bff/services/` so controllers no longer call gRPC clients or repositories directly.
3. Update controllers to read request context, validate view contracts with `validator`, call BFF services, and translate service results back into view responses.
4. Update `backend/internals/bff/transport/http/routes/contracts.go` so route capability interfaces depend on `views` types rather than controller-owned types.
5. Update `backend/cmd/bff/container.go` to wire validators, BFF services, controllers, and existing route modules through Dig.
6. Update governance and integration tests so the views layer, controller/service boundary, and route coverage matrix become the enforced standard.

## Validation Commands

```bash
cd backend

# Build verification after service and view extraction
go build ./...

# Package tests
go test ./...

# Route registration and OpenAPI validation
go test -tags=integration -run 'Test(BFFRouteRegistrationSmoke|OpenAPIOperationMetadataCompleteness|.*RouteIntegration)$' ./tests/integration/...

# Full integration suite
go test -tags=integration ./tests/integration/...

# Sanity checks for the refactor shape
rg 'type .*Input|type .*Response' internals/bff/transport/http/controllers
rg 'ServiceClient|New.*Repository|repositories\.' internals/bff/transport/http/controllers
```

## Focused Validation

- Confirm `internals/bff/transport/http/controllers/` no longer owns HTTP request/response struct definitions.
- Confirm controllers do not import generated gRPC client packages or repository implementations directly.
- Confirm all route capability interfaces in `routes/contracts.go` reference `views` types.
- Confirm the 20 active routes listed in `contracts/route-coverage-matrix.md` remain registered with the same methods, paths, and operation IDs.
- Confirm OpenAPI metadata tests still pass after moving contracts to `views/`.
- Confirm all 6 BFF service test files in `backend/internals/bff/services/` pass with `go test ./internals/bff/services/...`.

## Expected Outcome

- The views package becomes the canonical home for all BFF HTTP transport contracts.
- Controllers become thin HTTP adapters that validate requests, invoke services, and format responses.
- BFF services become the only layer that orchestrates downstream gRPC clients and repository-backed operations.
- The route inventory remains stable and fully covered by integration tests.

## SC-006 Timed Maintainer Placement Check

**Goal**: Verify that the controller/service/views boundary is self-evident to a new contributor within 2 minutes using only 3 prompts.

**Steps**:
1. Open `backend/internals/bff/transport/http/controllers/payments_controller.go` — confirm it contains no gRPC client imports and no struct definitions (only call to a service method and view type mapping). Time: ~30s.
2. Open `backend/internals/bff/services/payments_service.go` — confirm it contains all downstream gRPC orchestration and returns transport-agnostic types. Time: ~30s.
3. Open `backend/internals/bff/transport/http/views/payments_views.go` — confirm it contains all HTTP request and response struct definitions for the payments resource. Time: ~30s.

**Pass criterion**: All three files are found, confirm expected content, and total elapsed time is under 2 minutes.
