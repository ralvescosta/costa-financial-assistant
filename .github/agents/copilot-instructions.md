# costa-financial-assistant — Development Guidelines

## Project Summary

Monorepo web application: a multi-tenant financial bill organizer.
- Ingests and classifies PDF bills and bank statements.
- Processes PDFs asynchronously to extract structured financial data.
- Supports payment tracking, statement reconciliation, and history analytics.
- Enforces strict project-scoped multi-tenancy with role-based access control.

## Active Technologies
- TypeScript 5.8.x, React 18.3.x + react-router-dom 6.30.x, @tanstack/react-query 5.76.x, zod 3.24.x, TailwindCSS 3.4.x (002-frontend-auth-navigation)
- Browser HTTP-only cookies for auth/session (set by BFF), local client storage only for UI preferences and short-lived draft restore metadata (002-frontend-auth-navigation)
- Go 1.25.6; Markdown governance artifacts + `testing` (stdlib), `github.com/stretchr/testify`, `github.com/testcontainers/testcontainers-go` (to be standardized in scope), `github.com/golang-migrate/migrate/v4` (005-standardize-integration-tests)
- PostgreSQL ephemeral test database for integration runs (005-standardize-integration-tests)
- Go 1.25.6 + Echo v4.15.1, Huma v2.37.3, Dig v1.19.0, Zap v1.27.1, gRPC v1.80.0 generated clients, Testify v1.11.1, go-playground/validator for controller-side HTTP contract validation (006-bff-http-separation)
- Existing downstream PostgreSQL-backed services and payments repositories; no new persistent store introduced (006-bff-http-separation)

### Backend (Go — latest stable)
- HTTP framework: `github.com/labstack/echo/v4` + `github.com/danielgtaylor/huma/v2` (OpenAPI-first BFF)
- Middleware: `github.com/labstack/echo-contrib/echoprometheus`, `github.com/labstack/echo-contrib/otelecho`
- Dependency injection: `go.uber.org/dig`
- CLI: `github.com/spf13/cobra`
- Configuration: `github.com/spf13/viper`
- Logging: `go.uber.org/zap`
- OTel: OpenTelemetry Go SDK (`go.opentelemetry.io/otel`)
- Database migrations: `github.com/golang-migrate/migrate/v4`
- JWT: `github.com/golang-jwt/jwt/v5`
- Testing: `github.com/stretchr/testify` + `go.uber.org/mock`

### Frontend (TypeScript strict + React LTS on Vite)
- Build: `vite` + `@vitejs/plugin-react`
- Styling: `tailwindcss` (Tailwind v4 with CSS variables from design tokens)
- Routing: `react-router-dom`
- Server state: `@tanstack/react-query`
- Validation: `zod`
- Testing: `vitest` (hook tests only, BDD + Triple-A)

### Infrastructure
- Database: PostgreSQL
- Cache: Redis
- Object storage: S3-compatible (MinIO locally)
- Messaging: RabbitMQ
- Observability stack: `grafana/otel-lgtm` (local)

## Project Structure

```
backend/
├── cmd/
│   ├── bff/           # BFF CLI + DI wiring (container.go + cmd.go)
│   ├── bills/
│   ├── files/
│   ├── identity/
│   ├── onboarding/
│   └── payments/
├── internals/
│   ├── bff/financial/ # controllers/, services/, transport/http/middleware/
│   ├── bills/         # repositories/, services/, transport/grpc| rmq/
│   ├── files/         # repositories/, services/, transport/grpc|rmq/
│   ├── identity/
│   ├── onboarding/
│   └── payments/
├── pkgs/
│   ├── configs/       # Viper config loading
│   ├── otel/          # OTel bootstrap helpers
│   └── secrets/       # Vault / AWS Secrets Manager adapters
└── protos/
    ├── bills/v1/      # messages.proto + grpc.proto
    ├── files/v1/
    ├── identity/v1/
    ├── onboarding/v1/
    ├── common/v1/     # shared Pagination, Money, ProjectContext, ...
    └── generated/     # auto-generated Go protobuf artifacts

frontend/
└── src/
    ├── app/           # providers, router, theme
    ├── pages/         # composition roots only
    ├── components/    # UI primitives
    ├── hooks/         # all server-state and logic (React Query)
    ├── services/      # raw API client functions
    ├── styles/
    │   └── tokens.ts  # semantic + primitive design tokens
    └── types/

specs/
└── 001-financial-bill-organizer/
    ├── spec.md        # user stories + acceptance criteria + requirements
    ├── plan.md        # implementation plan + constitution gates
    ├── research.md    # technical decisions
    ├── data-model.md  # entity definitions + state machines
    ├── quickstart.md  # local setup + test commands
    ├── tasks.md       # implementation task backlog
    └── contracts/
        ├── bff-openapi-contract.md
        └── grpc-service-contracts.md
```

## Make Commands

```bash
# Infrastructure
make dev-up                        # Start PostgreSQL, RabbitMQ, MinIO, OTel stack

# Migrations
make migrate/up/<service>          # Apply migrations for a service
make migrate/down/<service>        # Roll back migrations for a service

# Protobuf
make proto/generate                # Regenerate backend/protos/generated/

# Run services
make run/bff
make run/files
make run/bills
make run/payments
make run/onboarding
make run/identity

# Unit tests (per service)
make test/bff
make test/files
make test/bills

# Integration tests (ephemeral DB lifecycle)
make test/integration/bff
make test/integration/files
make test/integration/bills
```

## Core Architectural Rules

1. **BFF only** uses Echo + Huma with `otelecho` middleware. Every route must have `OperationID`, `Summary`, `Description`, and `Tags` in the Huma operation declaration.
2. **gRPC proto-first**: domain messages in `messages.proto`, service methods in `grpc.proto`, shared types in `common/v1`.
3. **Clean architecture**: transport → controller/handler → service → repository. No layer may skip down.
4. **DI via dig**: all wiring in `backend/cmd/<service>/container.go`. Constructors take interfaces.
5. **Project isolation**: every domain query filters by `project_id` extracted from the verified JWT claim.
6. **JWT/JWKS**: JWT signing is exclusive to `identity-grpc`. All other services validate via the JWKS endpoint cached in `backend/internals/bff/financial/transport/http/middleware/jwks_cache.go`.
7. **Logging via zap**: `go.uber.org/zap` everywhere. Never `logrus`, `fmt.Printf`, or standard `log`.
8. **OTel tracing**: spans at all I/O boundaries; `otelecho` propagates trace context into all BFF requests.
9. **Migrations only**: schema changes exclusively through `.up.sql`/`.down.sql` pairs under `backend/<service>/migrations/`.
10. **Frontend hook-only tests**: Vitest tests exist only in `frontend/src/hooks/`, BDD + Triple-A pattern, no component render tests.

## Instruction Files

| File | Coverage |
|---|---|
| `.github/instructions/architecture.instructions.md` | Service boundaries, BFF, gRPC, DI, tenancy, UoW, caching, migrations, frontend architecture |
| `.github/instructions/golang.instructions.md` | Quality gates, error handling, interfaces, imports, toolchain |
| `.github/instructions/coding-conventions.instructions.md` | Naming, function size, constants, error messages, style |
| `.github/instructions/observability.instructions.md` | zap logging, OTel spans, log levels, sensitive data redaction |
| `.github/instructions/security.instructions.md` | Dependencies, secrets, input validation, JWT/JWKS, project isolation, RBAC, file upload |
| `.github/instructions/testing.instructions.md` | BDD+AAA, uber/mock, fixtures, ephemeral DB lifecycle, frontend Vitest hooks |
| `.github/instructions/commit-message.instructions.md` | Conventional Commits format and rules |
| `.github/instructions/ai-behavior.instructions.md` | AI code generation rules and precedence |


## Recent Changes
- 006-bff-http-separation: Added Go 1.25.6 + Echo v4.15.1, Huma v2.37.3, Dig v1.19.0, Zap v1.27.1, gRPC v1.80.0 generated clients, Testify v1.11.1, go-playground/validator for controller-side HTTP contract validation
- 005-standardize-integration-tests: Added Go 1.25.6; Markdown governance artifacts + `testing` (stdlib), `github.com/stretchr/testify`, `github.com/testcontainers/testcontainers-go` (to be standardized in scope), `github.com/golang-migrate/migrate/v4`
- 002-frontend-auth-navigation: Added TypeScript 5.8.x, React 18.3.x + react-router-dom 6.30.x, @tanstack/react-query 5.76.x, zod 3.24.x, TailwindCSS 3.4.x
