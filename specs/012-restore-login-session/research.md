# Research: Restore Seeded Login & Session Propagation

## Decision 1: Keep a login-only UX and do not reintroduce self-registration

- **Decision**: Preserve `frontend/src/app/router.tsx` + `frontend/src/pages/LoginPage.tsx` as the only auth entry path and do not add any `/register` screen or BFF registration endpoint.
- **Rationale**: The spec forbids a registration screen, and the existing frontend already expects `POST /api/auth/login` through `frontend/src/hooks/useAuthContext.tsx`.
- **Alternatives considered**:
  - Re-enable registration flow: rejected because it violates the feature intent.
  - Hard-code automatic authentication: rejected because protected-route authorization still needs a real session.

## Decision 2: The owner user must be seeded persistently and idempotently

- **Decision**: Seed the bootstrap owner account and its credential/membership records through the existing migration/bootstrap path (`backend/cmd/migrations/`, `backend/internals/identity/**`, `backend/internals/onboarding/**`) so a fresh environment can sign in without manual SQL.
- **Rationale**: The current identity service in `backend/internals/identity/services/token_service.go` can issue and validate JWTs, but it does not persist or verify username/password credentials on its own.
- **Alternatives considered**:
  - Keep the user only in memory: rejected because login would still fail after restart.
  - Require developers to insert rows manually: rejected because it breaks the “fresh environment works immediately” requirement.

## Decision 3: The BFF should expose auth endpoints but remain a thin gateway

- **Decision**: Add or restore BFF auth route, controller, view, and service modules under `backend/internals/bff/transport/http/{routes,controllers,views}/` and `backend/internals/bff/services/`, with downstream calls made via gRPC clients only.
- **Rationale**: `backend/cmd/bff/container.go` currently wires business routes but no auth route module; any login restoration must preserve the existing Echo + Huma + controller/service boundary rules.
- **Alternatives considered**:
  - Call identity logic directly from controllers: rejected because controllers must remain HTTP adapters only.
  - Perform auth directly in the frontend: rejected because token issuance and cookie handling belong to the BFF/identity boundary.

## Decision 4: Preserve the existing `common.v1.Session` contract and finish its adoption

- **Decision**: Keep `Session` in `backend/protos/common/v1/messages.proto` as the canonical authenticated caller envelope and verify/complete its presence on every authenticated gRPC request in the verified scope while preserving `common.v1.ProjectContext` for tenant isolation.
- **Rationale**: The shared proto already contains `Session` with the required `id`, `email`, and `username` fields; the feature work is to make the downstream adoption and BFF propagation complete and reliable, not to invent a second identity envelope.
- **Alternatives considered**:
  - Overload `ProjectContext` with email/username: rejected because tenant scope and caller identity are separate concerns.
  - Forward only raw JWT tokens: rejected because downstream services need a stable, transport-level identity contract.

## Decision 5: Standardize pagination defaults in the BFF request-building layer

- **Decision**: Keep pagination normalization in the BFF request-building path and always forward a populated `common.v1.Pagination` on list/select requests, using the clarified verified-scope defaults of `page_size = 20` and `page_token = ""` when the frontend omits query params.
- **Rationale**: This matches the repo rule that the BFF owns query-param parsing/defaulting and removes the current `25` vs `20` drift between route groups.
- **Alternatives considered**:
  - Make each downstream service invent its own defaults: rejected because callers would get inconsistent behavior.
  - Require the frontend to always send pagination: rejected because the feature explicitly requires predictable behavior when the UI omits it.

## Decision 6: Failure handling must remain AppError-first and explicit

- **Decision**: Treat wrong credentials as auth failures, missing bootstrap seed data as a setup/dependency failure, and invalid project membership as an authorization failure; each boundary must translate to `AppError` before crossing service or transport layers.
- **Rationale**: The constitution and architecture instructions require sanitized, category-aware error propagation.
- **Alternatives considered**:
  - Bubble raw SQL or gRPC errors to the UI: rejected because it leaks internals.
  - Return a generic “something went wrong” for every auth failure: rejected because operators need clear bootstrap/setup signals.

## Decision 7: Verification must cover login, protected routes, and pagination behavior

- **Decision**: Plan validation across `cd backend && go test ./...`, canonical integration suites in `backend/tests/integration/bff/`, `backend/tests/integration/identity/`, and `backend/tests/integration/cross_service/`, plus `make frontend/test` for the login-only frontend flow.
- **Rationale**: This feature touches full-stack auth state, gRPC contracts, and route access; compile-only checks would be insufficient.
- **Alternatives considered**:
  - Backend-only verification: rejected because the frontend contract must stay compatible.
  - Manual smoke testing only: rejected because route/pagination regressions would be easy to miss.

## Decision 8: Governance sync is part of the feature completion gate

- **Decision**: Treat `.specify/memory/architecture-diagram.md`, the impacted service-flow memory files, and the repo memory note `/memories/repo/bff-service-boundary-conventions.md` as required follow-up updates in the same feature cycle, with instruction-file updates if new auth/session rules need codification.
- **Rationale**: Both the spec and the constitution mark memory/instruction synchronization as mandatory for this scope.
- **Alternatives considered**:
  - Defer docs until after implementation: rejected because this feature changes long-lived cross-service conventions.
