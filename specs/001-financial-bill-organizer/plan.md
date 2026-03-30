# Implementation Plan: Financial Bill Organizer

**Branch**: `001-financial-bill-organizer` | **Date**: 2026-03-30 | **Spec**: `/specs/001-financial-bill-organizer/spec.md`
**Input**: Feature specification from `/specs/001-financial-bill-organizer/spec.md`

## Summary

Build a multi-tenant financial bill organizer that ingests and classifies PDFs,
processes extraction asynchronously, supports payment and reconciliation workflows,
and serves history analytics under strict project isolation.

Frontend will be built with React + Vite + Tailwind CSS using minimal libraries,
mobile-first tokenized design, and hook-centric logic testing.
Backend will implement BFF + gRPC modules per constitution, with Echo + Huma MVC,
OpenAPI-first contracts, OTel observability, idempotent mutating operations,
cache-aside reads, migration-only schema changes, and seeded Phase-1 identity bootstrap.

## Technical Context

**Language/Version**: Frontend TypeScript (strict) + React (LTS) on Vite; Backend Go (latest stable)

**Primary Dependencies**: Frontend `react`, `react-dom`, `react-router-dom`, `tailwindcss`, `@tanstack/react-query`, `zod`; Backend `echo/v4`, `huma/v2`, `humaecho`, `otelecho`, `dig`, `cobra`, `zap`, OpenTelemetry SDK, `golang-migrate`, `viper`, `jwt/v5`

**Storage**: PostgreSQL, Redis, S3-compatible object storage, RabbitMQ

**Testing**: Frontend Vitest (hook tests only, BDD + Triple-A); Backend `testing` + `testify` + `uber/mock`; transport integration tests with ephemeral DB lifecycle in `TestMain`

**Target Platform**: Web frontend (mobile-first) + Linux containerized backend (`linux/amd64`, `linux/arm64`)

**Project Type**: Monorepo web application + microservices backend

**Performance Goals**: SC-001 <10s upload acknowledgment; SC-002 <=60s extraction; SC-003 <2s dashboard; SC-007 <3s 12-month history render

**Constraints**: Echo+Huma MVC BFF with OpenAPI metadata; hook-only frontend tests; strict `project_id` tenant isolation; JWT signing only in `identity-grpc` with JWKS validation; migration-only DB changes

**Scale/Scope**: Phase-1 seeded bootstrap tenant/user/project with no interactive login; multi-project collaboration roles; upload/analyze/pay/reconcile/history workflows

Additional technical notes:
- Frontend keeps minimal runtime dependencies and uses tokens via Tailwind-compatible CSS variables.
- Integration tests must provision and destroy isolated DB environments per suite.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Pre-Research Gate Check

1. **Modular monorepo and service boundaries**: PASS
  - Uses existing backend service structure and adds feature logic within boundaries.
2. **SOLID + clean architecture + canonical folders**: PASS
  - Backend implementation plan keeps service/repository/interfaces split.
3. **Cloud-native/containerization**: PASS
  - Keeps Docker/compose-first local runtime and OTel stack.
4. **Frontend component/hook/design system rules**: PASS
  - Hook-only logic in frontend; semantic tokens only; mobile-first enforced.
5. **Testing discipline**: PASS
  - Hook BDD tests only in frontend; backend unit + integration with ephemeral DB.
6. **Observability/logging**: PASS
  - OTel logs/metrics/traces and Echo OTel middleware explicitly included.
7. **Makefile/IaC discipline**: PASS
  - Plan requires Make targets for run/test/migrate/proto generation.
8. **Config/secrets discipline**: PASS
  - `.env` via Viper + `${}` sentinel + `pkgs/secrets` resolution.
9. **Data access discipline**: PASS
  - Cache-aside, index review, idempotency, and Unit of Work included.
10. **Multi-tenancy and identity/access**: PASS
   - Project-scoped model + role enforcement + seeded bootstrap + JWKS validation.

No constitution violations identified at planning stage.

## Project Structure

### Documentation (this feature)

```text
specs/001-financial-bill-organizer/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── bff-openapi-contract.md
│   └── grpc-service-contracts.md
└── tasks.md
```

### Source Code (repository root)

```text
backend/
├── cmd/
│   ├── bff/
│   ├── bills/
│   ├── files/
│   ├── identity/
│   ├── onboarding/
│   └── payments/
├── internals/
│   ├── bff/
│   │   └── financial/
│   │       ├── controllers/
│   │       ├── services/
│   │       ├── interfaces/
│   │       └── transport/http/
│   ├── bills/
│   ├── files/
│   ├── onboarding/
│   ├── identity/
│   ├── payments/
│   └── migrations/
├── pkgs/
│   ├── configs/
│   ├── otel/
│   └── secrets/
└── protos/
   ├── bills/v1/{messages.proto,grpc.proto}
   ├── files/v1/{messages.proto,grpc.proto}
   ├── onboarding/v1/{messages.proto,grpc.proto}
   ├── identity/v1/{messages.proto,grpc.proto}
   ├── common/v1/
   └── generated/

frontend/
├── src/
│   ├── app/
│   ├── pages/
│   ├── components/
│   ├── hooks/
│   ├── services/
│   ├── styles/
│   │   └── tokens.ts
│   └── types/
└── tests/
   └── hooks/
```

**Structure Decision**:
Web application + microservices structure selected. Existing `backend/` is extended
with domain modules and transport folders per constitution, and `frontend/` is
introduced as a Vite React app with Tailwind CSS and hook-centric architecture.

## Phase 0: Research Outcome

All prior NEEDS CLARIFICATION items are resolved in `research.md`:
- Minimal frontend dependency set defined
- UI layout direction selected
- BFF contract approach finalized (Echo + Huma + OpenAPI)
- Multi-tenant bootstrap strategy and authorization boundaries defined
- Integration test database strategy defined

## Phase 1: Design Outcome

Artifacts generated:
- `data-model.md`
- `contracts/bff-openapi-contract.md`
- `contracts/grpc-service-contracts.md`
- `quickstart.md`

## Post-Design Constitution Check

Re-evaluation status: PASS

- Principle I (BFF Echo/Huma MVC): Covered in contracts + structure
- Principle II (clean architecture + proto layout): Covered in structure + gRPC contract
- Principle IV (tokenized design + responsive typography): Covered in frontend constraints
- Principle V (test strategy): Covered in quickstart + research decisions
- Principle VI (OTel + otelecho): Covered in BFF contract/quickstart
- Principle IX (cache/index/migrate/idempotency/UoW): Covered in data model and service rules
- Principle X (project_id isolation + roles + JWKS): Covered in data model and contracts

No complexity exceptions required.

## Complexity Tracking

No constitution violations requiring justification.
