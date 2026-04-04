# Quickstart: Stabilize Broken BFF Page Flows

## 1. Preconditions

- Work on branch `013-stabilize-bff-page-flows`.
- Confirm the planning artifacts exist:
  - `specs/013-stabilize-bff-page-flows/spec.md`
  - `specs/013-stabilize-bff-page-flows/plan.md`
  - `specs/013-stabilize-bff-page-flows/research.md`
  - `specs/013-stabilize-bff-page-flows/data-model.md`
- Run Go validation commands from the backend module root (`backend/`).

## 2. Bring up the local stack

Recommended baseline startup:

```bash
docker compose --profile dev up -d
make migrate/up local
```

For a full dev loop, either run everything together:

```bash
make dev-up
```

Or run the affected services individually while debugging:

```bash
make svc/run/identity
make svc/run/files
make svc/run/bills
make svc/run/payments
make svc/run/bff
make frontend/dev
```

## 3. Reproduce the user-visible problem

1. Sign in as the default dev/test user.
2. Open the in-scope screens in this order:
   - `/documents`
   - `/payments`
   - analyses flow (`/history` and `/reconciliation` if split in the current router)
   - `/settings`
3. Capture the exact failed requests and map them to `contracts/page-flow-validation-matrix.md`.
4. Confirm whether the failure is caused by auth/session, project scope, missing seed data, downstream contract behavior, or mapper/error translation.

## 4. Recommended implementation order

1. Verify auth/session and project-membership prerequisites for the default user.
2. Align or create the populated default-user seed data required by `contracts/default-user-seed-contract.md`.
3. Stabilize `documents` and `settings` flows backed by the files service.
4. Stabilize `payments` and `analyses` (`history` + `reconciliation`) flows backed by bills/payments.
5. Add canonical integration coverage for the page bootstrap paths and failure scenarios.
6. Re-run the frontend smoke path to confirm the four screens render populated content without request errors.

## 5. Validation commands

Primary verification commands:

```bash
cd backend && go test ./...
make test/integration
make frontend/test
```

Optional focused integration runs while iterating:

```bash
cd backend && go test -v -tags integration ./tests/integration/bff/...
cd backend && go test -v -tags integration ./tests/integration/payments/...
cd backend && go test -v -tags integration ./tests/integration/cross_service/...
```

Optional startup smoke checks:

```bash
timeout 12s make svc/run/files
timeout 12s make svc/run/payments
timeout 12s make svc/run/bff
```

> Do not claim the feature is complete until fresh command output confirms the relevant backend and integration checks pass.

## 6. Acceptance checklist

- The default user can open `documents`, `payments`, `analyses`, and `settings` without blocker-level backend request errors.
- All four pages show representative populated content in the canonical dev/test seed path.
- Expected auth/access/empty-state scenarios return supported responses instead of generic internal failures.
- Regression coverage exists in canonical backend integration-test locations and follows behavior-based snake_case naming with BDD + AAA structure.
- Required memory-sync updates are completed before the feature is closed.
