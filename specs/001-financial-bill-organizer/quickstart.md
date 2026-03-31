# Quickstart - Financial Bill Organizer

## Prerequisites
- Docker + Docker Compose
- Go toolchain (latest stable)
- Node.js LTS + npm
- GNU Make

## 1) Start local infrastructure

```bash
make dev-up
```

Expected core services: PostgreSQL, RabbitMQ, MinIO, grafana/otel-lgtm.

## 2) Configure environment

Create or update backend env file (example: `.env.dev`) with at least:
- `APP_ENV=dev`
- `SERVICE_NAME` and `VERSION` per service
- DB/RMQ/Redis/MinIO connection values
- `SECRETS_PROVIDER` (`vault` or `aws`)
- identity bootstrap keys/claims config for Phase-1

Use `${SECRET_KEY}` sentinel values for secrets resolved by `pkgs/secrets`.

## 3) Run migrations

```bash
make migrate/up/onboarding
make migrate/up/identity
make migrate/up/files
make migrate/up/bills
make migrate/up/payments
```

Seed migrations must create at least one bootstrap user + project for Phase-1.

## 4) Generate proto artifacts

```bash
make proto/generate
```

This must update `backend/protos/generated/` deterministically.

## 5) Run backend services

```bash
make run/onboarding
make run/identity
make run/files
make run/bills
make run/payments
make run/bff
```

BFF must expose:
- feature routes under `/api/v1/*`
- OpenAPI contract at `/openapi.json`

## 6) Bootstrap frontend (React + Vite + Tailwind)

```bash
# from repository root
npm create vite@latest frontend -- --template react-ts
cd frontend
npm install
npm install -D tailwindcss @tailwindcss/vite
npm install react-router-dom @tanstack/react-query zod
```

Configure Tailwind in Vite and implement tokenized theme system from spec.

## 7) Run frontend

```bash
cd frontend
npm run dev
```

## 8) Testing

## Frontend hook tests (BDD + Triple-A only)

```bash
cd frontend
npm run test
```

## Backend unit tests

```bash
make test/bff
make test/files
make test/bills
```

## Backend integration tests (ephemeral DB)

```bash
make test/integration/bff
make test/integration/files
make test/integration/bills
```

Integration tests must provision an isolated test DB, apply migrations, run tests,
then destroy/rollback DB state in `TestMain`.

## 9) Manual validation checklist
- Upload PDF and classify bill/statement.
- Confirm async status transitions to analysed/failed.
- Validate payment dashboard list, mark-paid idempotency, and overdue flags.
- Validate reconciliation summary and manual link flow.
- Validate project switch isolation (no cross-project data leakage).
- Validate role permissions (`read_only`, `update`, `write`).
- Validate `/openapi.json` includes operation metadata and schemas.
- Validate history dashboard timeline, category breakdown, and compliance panels.

## 10) Automated validation script

Run the feature validation script from repository root:

```bash
./scripts/validate-financial-bill-organizer.sh
```

Optional: skip integration tests (useful when local DB/infra is not running):

```bash
SKIP_INTEGRATION_TESTS=1 ./scripts/validate-financial-bill-organizer.sh
```
