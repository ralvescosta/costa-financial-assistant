# Research: BFF HTTP Boundary Separation

## Decision 1: Use `transport/http/views` as the sole owner of HTTP request and response structs

- **Decision**: Move all transport-facing request and response structs out of controller files and into `backend/internals/bff/transport/http/views/`, organized by resource area.
- **Rationale**: Controllers currently mix handler logic with transport contract definitions. A dedicated views package makes the HTTP boundary explicit, gives routes and controllers a shared contract location, and supports the feature requirement that all JSON contracts live under one transport-layer ownership boundary.
- **Alternatives considered**:
  - Keep request/response structs in controller files: rejected because it preserves the current coupling between handler behavior and HTTP contracts.
  - Put view structs in the route package: rejected because route modules should own registration metadata, not transport payload definitions.

## Decision 2: Validate bound HTTP contracts with `go-playground/validator`, while preserving Huma binding and schema tags

- **Decision**: Treat `validator` tags as the runtime validation source for controller-side contract checks, while retaining the binding and OpenAPI-related Huma tags required for path, query, body, and schema metadata.
- **Rationale**: The new feature requires validator tags on fields that need validation, but the BFF still depends on Huma metadata for request binding and OpenAPI generation. Using both tag families on view structs satisfies runtime validation and keeps the existing OpenAPI contract quality.
- **Alternatives considered**:
  - Use only Huma tags: rejected because the spec explicitly requires validator-tag-based controller checks.
  - Use only validator tags: rejected because it would weaken Huma schema generation and path/query metadata already enforced by the existing integration tests.

## Decision 3: Move downstream gRPC and repository orchestration into BFF services

- **Decision**: Introduce BFF service interfaces and concrete services for documents, projects, settings, payments, reconciliation, and history. Controllers will depend on those service interfaces instead of generated gRPC clients or repository-backed services directly.
- **Rationale**: Current controllers call `filesv1`, `onboardingv1`, `billsv1`, and payments services/repositories directly. That violates the desired boundary because controllers become responsible for downstream orchestration instead of HTTP-only work.
- **Alternatives considered**:
  - Leave direct gRPC client usage in controllers and only move structs to views: rejected because it would not satisfy the controller-only responsibility requirement.
  - Hide downstream calls behind controller helper methods: rejected because it renames the problem without changing the ownership boundary.

## Decision 4: Route capability interfaces must consume `views` types, not controller-defined types

- **Decision**: Update `backend/internals/bff/transport/http/routes/contracts.go` so route capability interfaces refer to `views` inputs and outputs rather than controller package types.
- **Rationale**: The current route contracts import controller-defined request/response structs, which keeps the route layer coupled to controller-owned transport contracts. Moving contracts to `views` should also remove that import dependency.
- **Alternatives considered**:
  - Keep route capability interfaces importing controller types after moving structs: rejected because it would preserve the wrong package ownership relationship.
  - Replace capability interfaces with raw `huma.Register` closures in container wiring: rejected because it would weaken testability and consumer-defined contracts.

## Decision 5: Preserve the existing six route modules and 20 active operations unchanged at the HTTP surface

- **Decision**: Keep the existing route-module inventory intact and preserve all current operation IDs, methods, paths, tags, and middleware ordering while changing only the internal controller/view/service boundaries.
- **Rationale**: The current codebase already has dedicated route modules and integration tests validating registration and metadata for 20 active operations. The value of this feature is finishing the boundary separation, not redesigning the public HTTP API.
- **Alternatives considered**:
  - Merge or split route modules while refactoring: rejected because it would expand scope and complicate regression analysis.
  - Rename operation IDs during the refactor: rejected because it would break existing smoke and OpenAPI contract tests without delivering value to this feature.

## Decision 6: Keep resource-scoped route suites and formalize an explicit route coverage matrix

- **Decision**: Continue using the existing resource-scoped integration suites for documents, projects, settings, payments, reconciliation, and history, and pair them with a maintained route coverage matrix that maps every active operation to its validating suite.
- **Rationale**: The repository already has route-specific integration suites plus global smoke and OpenAPI tests. The missing piece is a planning artifact that makes route coverage auditable when transport contracts and controller responsibilities change.
- **Alternatives considered**:
  - Collapse coverage into one monolithic integration suite: rejected because it would obscure resource ownership and make maintenance harder.
  - Rely only on the smoke and OpenAPI tests: rejected because those tests validate structural registration but not the resource-level route intent.

## Decision 7: Refactor all six active route groups in one coordinated migration

- **Decision**: Apply the new controller, views, and service boundaries across all active BFF route groups in one coordinated implementation.
- **Rationale**: The spec clarification explicitly sets the scope to all active BFF HTTP routes. Partial migration would leave mixed transport boundaries, duplicate contract patterns, and incomplete governance coverage.
- **Alternatives considered**:
  - Migrate one route group at a time across multiple specs: rejected because it conflicts with the clarified scope and prolongs the mixed-state architecture.
