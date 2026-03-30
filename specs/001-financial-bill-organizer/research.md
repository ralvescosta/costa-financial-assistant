# Phase 0 Research - Financial Bill Organizer

## Decision 1: Frontend stack with minimal libraries

- Decision: Use React + Vite + TypeScript + Tailwind CSS with a minimal runtime library set:
  - `react-router-dom` for routing
  - `@tanstack/react-query` for server-state fetching/cache
  - `zod` for form/payload schema validation
- Rationale:
  - Matches constitution and user request (React/Vite/Tailwind).
  - Keeps dependencies small while still handling async API state and validation cleanly.
  - Avoids introducing large state frameworks for MVP scope.
- Alternatives considered:
  - Redux Toolkit: rejected as unnecessary for current feature complexity.
  - Zustand: lightweight but duplicates concerns already handled by React Query + local hook state.
  - No data-fetch library: rejected due to repeated loading/error/retry boilerplate.

## Decision 2: Frontend layout and interaction model

- Decision: Use a mobile-first app-shell layout:
  - Top app bar + project switcher + theme toggle.
  - Bottom tab navigation on mobile (`Upload`, `Payments`, `Reconcile`, `History`, `Settings`).
  - Left rail navigation from `breakpointLg` upward.
  - Dashboard cards with overdue/paid emphasis via semantic tokens.
- Rationale:
  - Supports frequent flows (upload + payment dashboard) with one-tap access on mobile.
  - Aligns with responsive and typography token mandates.
  - Allows gradual expansion for collaboration/project-level controls.
- Alternatives considered:
  - Desktop-first sidebar-only layout: rejected (violates mobile-first intent).
  - Wizard-only navigation: rejected (poor discoverability for recurring usage).

## Decision 3: BFF API design and contract strategy

- Decision: BFF uses Echo + Huma with MVC layering and OpenAPI-first route definitions.
- Rationale:
  - Required by constitution (Principle I and gate 13).
  - Huma operation metadata ensures consistent `/openapi.json` quality.
  - MVC keeps controllers thin and service layer transport-agnostic.
- Alternatives considered:
  - Pure Echo handlers without Huma: rejected (would not satisfy OpenAPI-first and metadata rules).
  - Gin/Fiber: rejected by constitution.

## Decision 4: Multi-tenancy and identity bootstrap approach

- Decision:
  - Every domain table includes `project_id` FK.
  - Project collaboration roles: `read_only`, `update`, `write`.
  - Phase-1 uses seeded user/project and bootstrap JWT issued by `identity-grpc`.
  - All services validate JWT via `identity-grpc` JWKS endpoint.
- Rationale:
  - Meets Principle X now while preserving upgrade path to full auth later.
  - Enforces tenant isolation at schema + query + transport boundaries.
- Alternatives considered:
  - Single-tenant schema for MVP then migration later: rejected due to high migration and risk cost.
  - Hardcoded bypass without JWT validation: rejected by constitution gate 30.

## Decision 5: Async analysis and reconciliation processing model

- Decision:
  - Upload flow writes document metadata + object storage pointer, then publishes job event.
  - Background workers perform extraction and reconciliation.
  - Status transitions are explicit and queryable from BFF endpoints.
- Rationale:
  - Required for responsive UX and scalable processing.
  - Supports retries and observability instrumentation per stage.
- Alternatives considered:
  - Synchronous extraction during upload: rejected (violates UX/performance goals).

## Decision 6: Backend test strategy

- Decision:
  - Unit tests for services/repositories use BDD + Triple-A + `uber/mock` generated mocks.
  - Integration tests execute at HTTP/gRPC transport boundaries.
  - DB integration tests use ephemeral database lifecycle in `TestMain`:
    provision -> migrate up -> test -> teardown.
- Rationale:
  - Matches Principle V and gate 31.
  - Prevents flaky shared-state test pollution.
- Alternatives considered:
  - Shared long-lived test DB: rejected (state leakage and non-determinism).

## Decision 7: Proto and contract versioning approach

- Decision:
  - Keep one major-version folder per module under `backend/protos/<module>/v1`.
  - Split `messages.proto` (domain messages) and `grpc.proto` (service and I/O wrappers).
  - Keep shared messages in `backend/protos/common/v1`.
- Rationale:
  - Required by constitution Principle II.
  - Supports non-breaking evolution and clean generated code ownership.
- Alternatives considered:
  - Single monolithic proto file: rejected (poor modularity and versioning control).
