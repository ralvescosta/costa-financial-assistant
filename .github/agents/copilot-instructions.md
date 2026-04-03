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
- Markdown documentation artifacts in monorepo workflow (Speckit v0.4.3) + `.specify/templates/spec-template.md`, `.specify/memory/*.md`, Speckit scripts (`setup-plan.sh`, `update-agent-context.sh`) (007-review-bff-spec)
- Git-tracked repository files only (no runtime DB changes) (007-review-bff-spec)
- Go 1.25.6 + `go.uber.org/zap`, `google.golang.org/grpc`, `github.com/lib/pq`, `github.com/labstack/echo/v4`, `github.com/danielgtaylor/huma/v2`, `go.uber.org/dig` (008-standardize-app-errors)
- PostgreSQL plus object/file storage paths already in backend services; no schema change in this feature (008-standardize-app-errors)

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
│   ├── bff/           # clients/, interfaces/, services/, transport/http/{routes/,views/,controllers/,middleware/}
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
3. **Clean architecture**: transport → controller/handler → BFF service → (gRPC client | repository). No layer may skip down. For BFF: all HTTP contracts live in `transport/http/views/`; controllers are pure HTTP adapters that validate, call one BFF service method, and map to a view response; BFF services own all downstream gRPC orchestration.
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


## BFF Boundary Model (006-bff-http-separation)

The BFF service now enforces a strict three-layer contract:

1. **`transport/http/views/`** — sole owner of all HTTP request/response structs. Controllers and route contracts reference only `views.*` types. Every input field that needs runtime validation carries a `validate:` tag.
2. **`transport/http/controllers/`** — pure HTTP adapters: extract JWT claims, call `b.validateInput(input)`, delegate to one BFF service method, map the result to a view type, return. No gRPC imports, no repository imports.
3. **`services/`** — owns all downstream gRPC orchestration and transport-neutral application workflows. Service methods accept and return transport-agnostic types.

Route registration lives exclusively in `transport/http/routes/*_routes.go` via `huma.Register(...)`. Route capability interfaces in `routes/contracts.go` are narrow and depend only on `views.*` types.

DI wiring in `cmd/bff/container.go`: `validator.New()` → injected into all controllers; each controller provided as its route capability interface via `dig.As(new(routes.XxxCapability))`.

**Layer call order**:
```
[HTTP request]
  → routes/*_routes.go (huma.Register handler closure)
    → controller.HandleXxx (validate views.XxxInput, call service)
      → bffinterfaces.XxxService.XxxMethod (orchestrate gRPC + repos)
        → billsv1.BillsServiceClient / filesv1.FilesServiceClient / ...
```

**Integration test layout**: `backend/tests/integration/bff/` — named `<resource>_routes_registration_test.go`, `bff_route_registration_smoke_test.go`, `validate_openapi_metadata_test.go`. Cross-service tests live in `backend/tests/integration/cross_service/`.

## Recent Changes
- 008-standardize-app-errors: Added Go 1.25.6 + `go.uber.org/zap`, `google.golang.org/grpc`, `github.com/lib/pq`, `github.com/labstack/echo/v4`, `github.com/danielgtaylor/huma/v2`, `go.uber.org/dig`
- 007-review-bff-spec: Added Markdown documentation artifacts in monorepo workflow (Speckit v0.4.3) + `.specify/templates/spec-template.md`, `.specify/memory/*.md`, Speckit scripts (`setup-plan.sh`, `update-agent-context.sh`)
- 006-bff-http-separation: Added Go 1.25.6 + Echo v4.15.1, Huma v2.37.3, Dig v1.19.0, Zap v1.27.1, gRPC v1.80.0 generated clients, Testify v1.11.1, go-playground/validator for controller-side HTTP contract validation
