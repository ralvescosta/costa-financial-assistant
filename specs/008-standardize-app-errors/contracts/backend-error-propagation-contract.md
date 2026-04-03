# Contract: Backend Error Propagation Standard

## Purpose

Define mandatory invariants for translating dependency-native failures into `AppError` across backend layers.

## Scope

- Services: `bff`, `bills`, `files`, `identity`, `onboarding`, `payments`
- Paths: synchronous request flow (HTTP/gRPC) and asynchronous messaging flow (RMQ consumers/producers)
- Shared package: `backend/pkgs/errors`

## Contract Invariants

1. Boundary propagation invariant:
   - Any error crossing repository->service, service->controller/handler, or transport->caller boundaries MUST be represented as `AppError`.
2. Non-leakage invariant:
   - Raw dependency-native errors (database/grpc/library) MUST NOT be propagated beyond translation boundary.
3. Logging invariant:
   - Translation boundary MUST emit one structured log event including `zap.Error(err)` and contextual fields.
4. Catalog invariant:
   - Propagated `AppError` MUST come from centralized catalog entries in `backend/pkgs/errors/consts.go`.
5. Retryability invariant:
   - Every catalog entry MUST define `Retryable` intent.
6. Unknown fallback invariant:
   - Unmapped failures MUST map to a generic safe fallback `AppError`.

## Acceptance Mapping

- FR-001, FR-005 -> Invariants 1 and 2
- FR-002, FR-003 -> Invariants 4 and 5
- FR-004 -> Invariant 3
- FR-007 -> Invariant 6
- FR-009 -> Invariants 1-6 applied to sync and async paths

## Verification Checklist

- [ ] No raw dependency error reaches upper layers in reviewed paths.
- [ ] Translation boundary logs include native cause and context.
- [ ] Catalog coverage includes known categories and unknown fallback.
- [ ] Retryability field is set for all catalog entries.
- [ ] Tests assert translation behavior and non-leakage.
