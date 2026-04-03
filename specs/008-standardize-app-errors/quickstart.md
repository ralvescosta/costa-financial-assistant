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

## 5. Verify Non-Leakage and Classification

- Confirm no raw database/grpc/library errors are propagated across layer boundaries.
- Confirm unknown/unmapped failures resolve to generic fallback `AppError`.
- Confirm retryability flags are set for all catalog entries.

## 6. Sync Governance Artifacts

- Update impacted memory files listed in plan:
  - `.specify/memory/bff-flows.md`
  - `.specify/memory/files-service-flows.md`
  - `.specify/memory/bills-service-flows.md`
  - `.specify/memory/identity-service-flows.md`
  - `.specify/memory/onboarding-service-flows.md`
- Update instruction files if implementation changes reusable patterns.
