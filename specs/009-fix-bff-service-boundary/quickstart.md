# Quickstart: Enforce Service Boundary Contracts

## 1. Preconditions

- Use branch `009-fix-bff-service-boundary`.
- Ensure the baseline BFF layering structure exists:
  - `backend/internals/bff/transport/http/views/`
  - `backend/internals/bff/transport/http/controllers/`
  - `backend/internals/bff/transport/http/routes/`
  - `backend/internals/bff/services/`
- Confirm active spec and plan:
  - `specs/009-fix-bff-service-boundary/spec.md`
  - `specs/009-fix-bff-service-boundary/plan.md`

## 2. Implement BFF Contract Separation

1. Inspect all active BFF services and remove imports/dependencies on HTTP view contracts.
2. Ensure each endpoint path maps transport view contracts to service contracts in HTTP layer.
3. For service input/output shapes not covered by proto domain messages, create service-owned contracts in BFF service boundary packages.
4. Keep route modules responsible for registration metadata; keep controllers as HTTP adapters.

## 3. Apply Backend Pointer Policy

1. For modified backend boundaries, apply pointer signatures when structs have reference-like fields or size > 3 machine words.
2. Record explicit exceptions where value semantics remain intentional.
3. Verify no nil-safety regressions in mapper and service boundaries.

## 4. Validate Behavior and Policy Compliance

1. Run backend tests from module root:

```bash
cd backend
go test ./...
```

2. Run focused suites for touched areas as needed:

```bash
make svc/test/bff
make svc/test/files
make svc/test/bills
make svc/test/identity
make svc/test/onboarding
make svc/test/payments
```

3. Run integration verification for route behavior and canonical placement compliance:

```bash
make test/integration
```

## 5. Governance Synchronization (Required for Completion)

1. Update impacted `.specify/memory/*` files:
   - `.specify/memory/architecture-diagram.md`
   - `.specify/memory/bff-flows.md`
   - `.specify/memory/files-service-flows.md`
   - `.specify/memory/bills-service-flows.md`
   - `.specify/memory/identity-service-flows.md`
   - `.specify/memory/onboarding-service-flows.md`
   - `.specify/memory/payments-service-flows.md` (create if absent)
2. Update repository memory notes in `/memories/repo/*` for boundary and pointer-policy conventions.
3. Update impacted instruction files listed in the plan.
4. If reusable planning workflow behavior changed, update `.specify/templates/spec-template.md` and/or `.specify/templates/plan-template.md`.

## 6. Completion Checks

- No active BFF service depends on HTTP view contracts.
- All active BFF routes/services follow explicit transport-to-service mapping.
- Pointer threshold policy is applied consistently on modified backend boundaries.
- Instruction and memory sync obligations are completed in the same feature cycle.

## 7. Validation Evidence

- `go test ./...` from `backend/`: PASS
- `backend/scripts/validate_integration_test_conventions.sh`: PASS
- Integration smoke convention verification (`backend/tests/integration/bff/bff_route_registration_smoke_test.go`): PASS
  - Canonical placement: `backend/tests/integration/bff/`
  - Behavior-based snake_case naming: `bff_route_registration_smoke_test.go`
  - Given/When/Then + Arrange/Act/Assert sections present
