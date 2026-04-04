# Quickstart: Restore BFF gRPC Gateway Boundary

## 1. Preconditions

- Work on branch `011-fix-bff-grpc-boundary`.
- Confirm the active design artifacts exist:
  - `specs/011-fix-bff-grpc-boundary/spec.md`
  - `specs/011-fix-bff-grpc-boundary/plan.md`
  - `specs/011-fix-bff-grpc-boundary/research.md`
- Run Go commands from the backend module root (`backend/`) when validating.

## 2. Add the missing payments gRPC surface

1. Create the new proto module:
   - `backend/protos/payments/v1/messages.proto`
   - `backend/protos/payments/v1/grpc.proto`
2. Define RPCs for the payments-owned BFF flows:
   - `GetCyclePreference`
   - `SetCyclePreference`
   - `GetHistoryTimeline`
   - `GetHistoryCategoryBreakdown`
   - `GetHistoryCompliance`
   - `GetReconciliationSummary`
   - `CreateManualLink`
3. Re-generate Go code:

```bash
make proto/generate
```

## 3. Wire the payments service transport

1. Implement `backend/internals/payments/transport/grpc/server.go`.
2. Register the payments gRPC server and graceful shutdown behavior in `backend/cmd/payments/container.go`.
3. Keep business logic in `backend/internals/payments/services/`; do not move repository logic into transport code.

## 4. Migrate the BFF to the new boundary

1. Dial the payments gRPC client in `backend/cmd/bff/container.go`.
2. Update the BFF services to use only downstream gRPC clients for business data:
   - `backend/internals/bff/services/payments_service.go`
   - `backend/internals/bff/services/history_service.go`
   - `backend/internals/bff/services/reconciliation_service.go`
3. Preserve the existing transport boundary:
   - HTTP views stay in `backend/internals/bff/transport/http/views/`
   - mapping stays in `backend/internals/bff/transport/http/controllers/mappers/`
   - BFF service contracts stay in `backend/internals/bff/services/contracts/`

## 5. Validate implementation and runtime health

Run the baseline verification:

```bash
cd backend && go test ./...
cd backend && go test -tags integration ./tests/integration/bff ./tests/integration/payments
cd backend && go test -tags integration ./tests/integration/cross_service -run 'TestBoundaryLogging_TableDriven_T064|TestUS5_|TestUS6_'
```

Then spot-check startup:

```bash
timeout 12s make svc/run/payments
timeout 12s make HTTP_PORT_bff=18080 svc/run/bff
timeout 12s make svc/run/bills
```

Verified on **2026-04-04**:
- `go test ./...` passed.
- The targeted tagged integration checks above passed for `bff`, `payments`, and the feature-relevant `cross_service` scenarios.
- `payments`, `bff`, and `bills` each reached their startup log and shut down cleanly under the timeout-based smoke check.

Note: the full repository-wide `make test/integration` command still includes unrelated pre-existing failures in file-upload/isolation scenarios outside this feature’s BFF/payments boundary scope.

## 6. Required governance synchronization

Before closing the feature:

1. Update `.specify/memory/architecture-diagram.md`.
2. Update `.specify/memory/bff-flows.md`.
3. Update `.specify/memory/payments-service-flows.md`.
4. Update `/memories/repo/bff-service-boundary-conventions.md`.
5. Finalize any impacted `.github/instructions/*.instructions.md` files.
6. Remove any temporary dependency-error fallback from supported paths once the gRPC contract is fully implemented.

## 7. Completion evidence checklist

- No supported BFF business flow uses direct repository or DB access.
- Payments-owned routes return normal data through the new payments gRPC boundary.
- `backend/internals/payments/transport/grpc/` contains the real server implementation instead of only `.gitkeep`.
- The BFF, payments, and bills services still boot with their normal run commands.
