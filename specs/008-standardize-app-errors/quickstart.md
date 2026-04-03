# Quickstart: Standardize Backend App Errors

## 1. Preconditions

- Use branch `008-standardize-app-errors`.
- Ensure backend dependencies are available.
- Confirm baseline files exist:
  - `backend/pkgs/errors/error.go`
  - `backend/pkgs/errors/consts.go`

## 2. Implement Error Catalog and Semantics

1. Expand `backend/pkgs/errors/consts.go` with catalog entries for all currently known failure categories.
2. Classify each catalog entry with deterministic retryability semantics.
3. Keep one mandatory unknown-fallback catalog entry.

## 3. Apply Translation Rules Across Services

1. For each backend service path (sync and async), translate dependency-native errors before crossing layer boundaries.
2. Ensure propagation uses only `AppError`.
3. Ensure translation boundaries log native errors exactly once with structured context (`zap.Error(err)`).

## 4. Validate Behavior

1. Run backend tests:

```bash
cd backend
go test ./...
```

2. Run service-focused tests as needed:

```bash
make svc/test/bff
make svc/test/files
make svc/test/bills
make svc/test/identity
make svc/test/onboarding
make svc/test/payments
```

3. Run backend integration tests when integration behavior coverage is added:

```bash
make test/integration
```

## 4.1 Validation Evidence (2026-04-03)

- Repaired and re-ran boundary logging/retryability tests:
  - `go test ./internals/files/services ./internals/payments/services ./pkgs/errors`
- Re-ran package-level verification for touched layer boundaries:
  - `go test ./internals/files/services ./internals/payments/services ./internals/bff/services ./internals/identity/services`
  - `go test ./internals/files/repositories ./internals/payments/repositories ./internals/bills/repositories ./internals/onboarding/repositories`
  - `go test ./internals/files/transport/grpc ./internals/bills/transport/grpc ./internals/onboarding/transport/grpc ./internals/identity/transport/grpc ./internals/files/transport/rmq`
- Re-ran cross-service integration suites:
  - `go test ./tests/integration/cross_service`

Validation outcome: all above commands completed successfully in this implementation cycle.

## 5. Verify Non-Leakage and Classification

- Confirm no raw database/grpc/library errors are propagated across layer boundaries.
- Confirm unknown/unmapped failures resolve to generic fallback `AppError`.
- Confirm retryability flags are set for all catalog entries.

## 7. CI Enforcement Note

- CI gate coverage for non-leakage and boundary-contract tests is documented in:
  - `specs/008-standardize-app-errors/contracts/ci-enforcement-config.md`

## 6. Sync Governance Artifacts

- Update impacted memory files listed in plan:
  - `.specify/memory/bff-flows.md`
  - `.specify/memory/files-service-flows.md`
  - `.specify/memory/bills-service-flows.md`
  - `.specify/memory/identity-service-flows.md`
  - `.specify/memory/onboarding-service-flows.md`
- Update instruction files if implementation changes reusable patterns.
