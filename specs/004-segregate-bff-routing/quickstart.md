# Quickstart: BFF Route-Controller Segregation

## Goal

Validate that the BFF transport layer can be refactored to dedicated route modules without changing public endpoint behavior and with complete route-level integration coverage.

## Prerequisites

- Go 1.25.6 available locally
- Backend dependencies installed via the existing module setup
- PostgreSQL integration database reachable through `TEST_DB_DSN` or the default local integration DSN
- Existing gRPC dependencies or test doubles available through the current integration setup

## Implementation Outline

1. Create `backend/internals/bff/transport/http/routes/` and define the shared route contract.
2. Move `huma.Register(...)` declarations out of each controller file and into a matching route module.
3. Keep concrete handler logic inside controller structs and expose only the capability methods required by the consuming route module.
4. Update `backend/cmd/bff/container.go` so Dig wires route modules and invokes route registration instead of controller registration.
5. Update BFF integration tests so every route in the route coverage matrix is exercised by at least one passing BDD-style scenario.
6. Update governance and instruction documents so future BFF changes follow the same route-module pattern.

## Validation Commands

```bash
cd backend

# Unit/package-level build verification
go build ./...

# Unit tests (no integration DB needed)
go test ./...

# Route-level integration tests (requires no live DB — uses stub capabilities)
go test -tags=integration -run 'TestBFF' ./tests/integration/...

# Full integration suite (requires TEST_DB_DSN or default local DSN)
go test -tags=integration ./tests/integration/...
```

## Focused Validation

- Confirm all 20 existing BFF routes still appear in OpenAPI metadata validation.
- Confirm auth and project guard middleware are still applied to the same operations.
- Confirm route modules, not controllers, are the sole owners of `huma.Register(...)` declarations.
- Confirm every route listed in `contracts/route-coverage-matrix.md` maps to a passing integration suite.

## Expected Outcome

- The BFF container registers route modules instead of controller `Register()` methods.
- Controller files contain only request-handling behavior and transport translations.
- Route registration becomes discoverable by resource.
- Integration coverage explicitly accounts for every BFF route.