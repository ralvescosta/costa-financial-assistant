# Research: BFF Route-Controller Segregation

## Decision 1: Create one route module per existing controller domain

- **Decision**: Introduce `backend/internals/bff/transport/http/routes/` with one route struct per existing domain resource: documents, projects, settings, payments, reconciliation, and history.
- **Rationale**: This mirrors the current controller decomposition, keeps the migration incremental in code shape even if implemented in one feature, and makes route ownership easy to locate during review.
- **Alternatives considered**:
  - One monolithic route registry file: rejected because it would centralize too many unrelated operations and recreate an oversized transport entrypoint.
  - Keep route declarations in controller files but move them to separate sections: rejected because it would not actually separate responsibilities.

## Decision 2: Use narrow consumer-defined controller capability interfaces plus an embeddable base struct

- **Decision**: Model the “default controller interface” as a minimal shared contract and a set of narrow route-specific capability interfaces defined in the consuming route package. Shared reusable behavior should live in an embeddable base struct rather than a broad interface with unused methods.
- **Rationale**: Go interfaces do not support default method implementations. Small consumer-driven interfaces align with repository rules and avoid forcing every controller to implement unrelated behavior.
- **Alternatives considered**:
  - One broad controller interface with every possible handler method: rejected because it violates the small-interface rule and creates brittle, transport-wide coupling.
  - No shared controller contract at all: rejected because the feature explicitly requires standardized controller behavior expectations.

## Decision 3: Use a single shared route contract for registration

- **Decision**: Each route struct implements a shared route contract responsible for registering operations against a Huma API using injected controller dependencies and middleware factories.
- **Rationale**: Route registration is the common responsibility across all route modules, so a single contract keeps container wiring simple and discoverable.
- **Alternatives considered**:
  - Controller-specific ad hoc registration methods: rejected because it would preserve divergence and weaken compile-time consistency.
  - Global package functions for route registration: rejected because it weakens DI boundaries and makes testing more awkward.

## Decision 4: Preserve middleware order and operation metadata exactly

- **Decision**: Route modules must preserve the current middleware chain order of auth before project guard and must keep existing `OperationID`, `Method`, `Path`, `Summary`, `Description`, and `Tags` unchanged unless a separate change explicitly updates the public API.
- **Rationale**: The refactor is structural. Changing operation metadata, middleware order, or paths would turn it into a behavior change and expand regression risk.
- **Alternatives considered**:
  - Revisit authorization roles during refactor: rejected because that is a separate behavioral concern.
  - Rename operations while moving them: rejected because it would break compatibility and obscure regression analysis.

## Decision 5: Standardize route coverage with resource-scoped integration suites plus a route matrix

- **Decision**: Cover all 20 BFF routes through resource-scoped integration test suites, using BDD-named subtests per route and maintaining an explicit route-to-test matrix artifact.
- **Rationale**: The repository already uses `backend/tests/integration/`; grouping by route resource gives a predictable test layout without requiring one file per endpoint. The matrix makes coverage auditable.
- **Alternatives considered**:
  - Continue relying on broad user-story integration files only: rejected because route coverage is currently implicit and incomplete.
  - Create one test file per route endpoint: rejected because it would over-fragment the suite and add noise without improving maintainability.

## Decision 6: Migrate all six BFF route groups in one coordinated refactor

- **Decision**: Implement the route/controller separation for all six BFF route groups in one coordinated change.
- **Rationale**: Container wiring currently invokes all controller registrations centrally. Mixing old and new registration patterns would create avoidable transitional complexity in the same transport layer.
- **Alternatives considered**:
  - Migrate one controller at a time over multiple features: rejected because it would keep the transport layer inconsistent and prolong governance mismatch with the constitution.