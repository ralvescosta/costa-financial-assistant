# Quickstart: Restore Seeded Login & Session Propagation

## 1. Preconditions

- Work on branch `012-restore-login-session`.
- Confirm the planning artifacts exist:
  - `specs/012-restore-login-session/spec.md`
  - `specs/012-restore-login-session/plan.md`
  - `specs/012-restore-login-session/research.md`
  - `specs/012-restore-login-session/data-model.md`
- Run Go validation commands from the backend module root (`backend/`).

## 2. Restore persistent bootstrap authentication

1. Add or verify an idempotent seed path for the default owner user and password hash.
2. Ensure the same seed path also creates the owner’s project membership and full route permissions.
3. Keep the experience login-only — do **not** add a registration screen or register endpoint in this cycle.

## 3. Standardize the gRPC contracts

1. Verify and preserve `common.v1.Session` in `backend/protos/common/v1/messages.proto` as the canonical authenticated caller envelope.
2. Update or confirm authenticated requests in `onboarding`, `files`, `bills`, and `payments` so the caller identity travels as `Session` while `ProjectContext` remains the tenant boundary.
3. Audit every list/select request for a populated `common.v1.Pagination` contract and ensure the BFF always forwards defaults when query params are omitted.
4. Re-generate protobuf artifacts:

```bash
make proto/generate
```

## 4. Restore the BFF login gateway

1. Add auth route/controller/view/service modules under `backend/internals/bff/transport/http/` and `backend/internals/bff/services/`.
2. Wire any missing identity/onboarding gRPC client dependencies in `backend/cmd/bff/container.go`.
3. Preserve the enforced boundary:
   - service contracts in `backend/internals/bff/services/contracts/`
   - HTTP views in `backend/internals/bff/transport/http/views/`
   - mapping in `backend/internals/bff/transport/http/controllers/mappers/`

## 5. Apply deterministic pagination defaults

1. Normalize the verified-scope BFF defaults to `page_size = 20` and `page_token = ""` when the frontend omits query params.
2. Update any route that still defaults to `25` unless a smaller documented clamp is intentionally preserved.
3. Always forward a non-nil `common.v1.Pagination` on authenticated list/select requests.

## 6. Validate the feature

Run the baseline verification commands:

```bash
make proto/generate
cd backend && go test ./...
make test/integration
make frontend/test
```

Optional targeted startup smoke checks:

```bash
timeout 12s make svc/run/identity
timeout 12s make svc/run/onboarding
timeout 12s make HTTP_PORT_bff=18080 svc/run/bff
```

> Record fresh command output before claiming the feature is complete.

### Current verified evidence

- `cd backend && go test ./...` → full backend suite passing after the auth/session changes.
- `cd frontend && npm test -- src/hooks/useAuthSession.test.tsx src/hooks/useAuthRefresh.test.tsx src/hooks/useTokenRefreshInterceptor.test.tsx src/hooks/usePersistentSession.test.ts` → focused frontend auth suite passing (`4` files, `11` tests).

## 7. Required governance sync

Before closing the feature:

1. Update `.specify/memory/architecture-diagram.md`.
2. Update `.specify/memory/bff-flows.md`.
3. Update `.specify/memory/identity-service-flows.md`.
4. Update `.specify/memory/onboarding-service-flows.md`.
5. Update any impacted `files`, `bills`, and `payments` service-flow memory files.
6. Re-check `/memories/repo/bff-service-boundary-conventions.md` and any impacted `.github/instructions/*.instructions.md` files.

## 8. Completion evidence checklist

- The seeded owner login works on a fresh environment without manual DB repair.
- No registration route or registration UI is exposed.
- Protected routes accept the same authenticated session end-to-end.
- Authenticated gRPC requests include `Session`, and list/select RPCs include populated `Pagination` with the verified-scope default of `page_size = 20` and `page_token = ""` when omitted by the UI.
