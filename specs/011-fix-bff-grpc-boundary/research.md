# Research: Restore BFF gRPC Gateway Boundary

## Decision 1: Introduce a dedicated `payments/v1` gRPC contract surface

- **Decision**: Create `backend/protos/payments/v1/messages.proto` and `backend/protos/payments/v1/grpc.proto`, then generate `backend/protos/generated/payments/v1/*`.
- **Rationale**: The codebase currently has no `backend/protos/payments/` module while the Payments domain already owns cycle preference, history analytics, and reconciliation business logic under `backend/internals/payments/`. The BFF cannot respect service ownership for those flows until a real transport boundary exists.
- **Alternatives considered**:
  - Keep the BFF on in-process `internals/payments/interfaces` dependencies: rejected because it reproduces the exact anti-pattern this feature is meant to remove.
  - Permanently return `dependency_grpc` errors from the BFF for these flows: rejected because the clarified spec requires full migration for supported paths in the final state.
  - Move the flows into `bills` just because some payment routes already use `bills.v1`: rejected because the domain logic and repositories already live in `payments`.

## Decision 2: Preserve the existing domain ownership split between `bills` and `payments`

- **Decision**: Keep `GetPaymentDashboard` and `MarkBillPaid` in `bills.v1.BillsService`, while adding cycle preference, history analytics, and reconciliation RPCs to the new `payments.v1.PaymentsService`.
- **Rationale**: Bills already owns bill records and payment-status mutations via existing gRPC contracts, but the analytics/reconciliation services are implemented in `backend/internals/payments/services/` and should remain there.
- **Alternatives considered**:
  - Consolidate every payment-adjacent route into `payments`: rejected because it would churn already-correct bills-owned behavior.
  - Split ownership ad hoc by route name: rejected because ownership should follow the domain service that implements the business rules.

## Decision 3: Add a real payments gRPC transport and wire it from `cmd/payments/container.go`

- **Decision**: Implement `backend/internals/payments/transport/grpc/server.go` and register/start a gRPC server from `backend/cmd/payments/container.go`.
- **Rationale**: `backend/internals/payments/transport/grpc/` currently contains only `.gitkeep`, and the payments command currently logs startup without exposing a gRPC listener.
- **Alternatives considered**:
  - Keep payments as an in-process only service: rejected because the BFF would still lack a contract-safe access path.
  - Expose these capabilities over HTTP: rejected because the repository’s inter-service standard is gRPC + proto-first contracts.

## Decision 4: Keep the BFF’s HTTP mapper boundary intact while swapping service internals to gRPC clients

- **Decision**: Leave route/view/controller contracts in place and update `backend/cmd/bff/container.go` plus BFF services (`payments_service.go`, `history_service.go`, `reconciliation_service.go`) to consume the generated payments gRPC client only.
- **Rationale**: This preserves the repository’s enforced boundary of `views` → `controllers/mappers` → `services/contracts` → gRPC client.
- **Alternatives considered**:
  - Put proto-to-view mapping inside BFF services: rejected because mapper logic belongs in `backend/internals/bff/transport/http/controllers/mappers/`.
  - Let controllers call gRPC directly: rejected because controllers must stay as thin HTTP adapters.

## Decision 5: Use proto-first contract design with service-owned semantics

- **Decision**: Place payments-owned domain messages in `backend/protos/payments/v1/messages.proto` and the new RPC wrappers/service in `backend/protos/payments/v1/grpc.proto`, reusing `common/v1.ProjectContext` and pagination patterns where relevant.
- **Rationale**: This matches the repo’s proto layout rules and creates a stable inter-service contract for the BFF.
- **Alternatives considered**:
  - Encode business payloads only in `grpc.proto`: rejected because the repo requires domain messages and service declarations to be split.
  - Reuse HTTP view structs for inter-service transport: rejected because HTTP views are transport-owned and must not cross this boundary.

## Decision 6: AppError-first migration safety remains mandatory

- **Decision**: All new transport and BFF boundary code must translate failures through `backend/pkgs/errors` and keep sanitized error propagation across repository → service → transport transitions.
- **Rationale**: The constitution and current instructions make `AppError` the mandatory cross-layer contract.
- **Alternatives considered**:
  - Bubble raw gRPC/SQL errors upward: rejected because it violates non-leakage and retryability governance.
  - Collapse everything into `unknown`: rejected because category-specific behavior and observability would be lost.

## Decision 7: Verification must combine unit, integration, and startup checks

- **Decision**: Validate this feature with `cd backend && go test ./...`, targeted BFF/payments tests, canonical BFF integration tests in `backend/tests/integration/bff/`, cross-service integration tests in `backend/tests/integration/cross_service/`, and short boot checks for `bff`, `payments`, and `bills`.
- **Rationale**: This feature changes DI wiring, proto surfaces, transport behavior, and service startup; compile-only verification is insufficient.
- **Alternatives considered**:
  - Unit tests only: rejected because runtime container wiring and gRPC startup issues would be missed.
  - Startup checks only: rejected because mapper and contract regressions would be missed.

## Decision 8: Governance sync is part of the deliverable

- **Decision**: Final implementation tasks must update `.specify/memory/architecture-diagram.md`, `.specify/memory/bff-flows.md`, `.specify/memory/payments-service-flows.md`, `/memories/repo/bff-service-boundary-conventions.md`, and the impacted `.github/instructions/*.instructions.md` files in the same feature cycle.
- **Rationale**: The user explicitly requested rule and flow-diagram updates, and the constitution makes that a completion gate.
- **Alternatives considered**:
  - Defer guidance updates until after merge: rejected because stale instructions would continue to encourage regressions.
