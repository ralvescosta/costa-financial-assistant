<!--
SYNC IMPACT REPORT
==================
Version change: 1.0.0 → 1.1.0
Bump rationale: MINOR — three existing principles materially expanded with new mandatory
  tooling and testing discipline rules; no principles removed or redefined.
Modified principles:
  - III. Cloud Native & Containerization — added grafana/otel-lgtm to required docker-compose
    services.
  - V. Test Discipline — added mandatory BDD + Triple-A (Arrange-Act-Assert) structure for
    all unit tests.
  - VI. Observability & Structured Logging — replaced ambiguous logger choice with zap
    (go.uber.org/zap) as the single mandatory logger; added full OpenTelemetry instrumentation
    requirement (logs via zap OTel hook, metrics, traces) for all backend services.
Added sections: None
Removed sections: None
Templates requiring updates:
  - .specify/templates/plan-template.md ✅ — constitution check gates still valid; Technical
    Context section for new features should now list OTel SDK and zap as primary dependencies.
  - .specify/templates/spec-template.md ✅ — no structural changes required.
  - .specify/templates/tasks-template.md ✅ — Phase 2 (Foundational) tasks for new backend
    services MUST now include OTel SDK bootstrap and zap logger setup tasks.
  - .github/agents/*.md ✅ — no outdated agent-specific references found.
Deferred TODOs: None — all fields resolved.
-->

# Costa Financial Assistant Constitution

## Core Principles

### I. Modular Monorepo Architecture (NON-NEGOTIABLE)

The entire project MUST live in a single monorepo.
Frontend and backend are top-level modules with no shared source coupling.
Backend services (BFF API, files-grpc, bills-grpc, scheduler-grpc, and future services)
MUST be independently deployable units with no cross-service direct imports.
Circular dependencies between modules are forbidden.
Each microservice owns its own domain; shared utilities MUST live in `pkgs/` and MUST NOT
encode domain logic.

### II. SOLID Principles & Clean Architecture (NON-NEGOTIABLE)

Every backend service MUST follow SOLID principles at all layers.
The canonical Go project layout for each service is:

```
cmd/
  container.go      # dependency injection wiring via uber/dig
  <service>_cmd.go  # cobra CLI command entrypoint
internals/
  <domain>/
    services/       # business logic — depends only on interfaces
    interfaces/     # Go interfaces (*.go) — no implementations here
    *.go            # domain types, errors
    *_test.go       # unit tests alongside source
  repositories/     # data access implementations
    *.go
    *_test.go
  transport/
    grpc/           # gRPC handlers
    rmq/            # RabbitMQ consumers/producers (.gitkeep if unused)
pkgs/               # shared, domain-agnostic utilities
mocks/              # auto-generated mocks (uber/mock)
```

Services MUST depend on interfaces, never on concrete implementations.
The `internals/` tree MUST NOT import from `cmd/`.
Dependency injection MUST use `uber/dig`; manual wiring in `cmd/container.go` is forbidden
beyond registration calls.
CLI entrypoints MUST use `cobra`.

### III. Cloud Native & Containerization (NON-NEGOTIABLE)

Every service and the frontend MUST be containerized via Docker.
Backend container images MUST support multi-platform builds targeting at minimum
`linux/amd64` and `linux/arm64` (use `docker buildx` with `--platform` flag).
All services MUST follow the 12-factor app methodology:
configuration via environment variables, stateless processes, explicit dependency
declaration, disposability.
A `docker-compose.yml` MUST exist at the repository root providing all local-dev
dependencies: PostgreSQL, RabbitMQ, a file-storage service (e.g., MinIO), and the
`grafana/otel-lgtm` image which bundles the full Grafana observability stack
(Loki, Grafana, Tempo, Mimir/Prometheus) for local telemetry validation.
Backend services MUST expose health-check endpoints consumable by container orchestrators.

### IV. Frontend Component-First & Hook Isolation (NON-NEGOTIABLE)

The frontend MUST use React at the current LTS version.
Every screen MUST be a pure composition of reusable components; no business logic is
permitted inside a screen presentation file.
Before creating a new component, an existing component MUST be evaluated for reuse or
extension.
Business logic, state management, and data-fetching MUST be encapsulated in custom hooks
(`use*.ts` files); the corresponding screen file contains only JSX/TSX and style bindings.
Each feature MUST ship at minimum two artifacts: a presentation file (screen) and one or
more custom hook files. Co-locating them in the same file is forbidden.

### V. Test Discipline

Unit tests are MANDATORY for all service and repository implementations in the backend.
All unit tests MUST follow BDD (Behaviour-Driven Development) style and the
Triple-A structure — every test body MUST contain explicit Arrange, Act, and Assert
phases, separated by blank lines and optionally labeled with inline comments
(`// Arrange`, `// Act`, `// Assert`) to make intent unambiguous.
Test names MUST describe behaviour in the form `TestSubject_WhenCondition_ThenOutcome`
or an equivalent BDD-readable sentence.
Mocks MUST be auto-generated using `uber/mock` (`mockgen`) and placed in `mocks/`;
hand-written mocks are forbidden.
Integration tests MUST cover inter-service gRPC contracts and any RabbitMQ event contracts.
Frontend component tests are REQUIRED for all reusable components.
Tests MUST live alongside source (`*_test.go` for Go; `*.test.tsx` for React).
No production code may be merged that reduces existing test coverage without explicit
documented justification.

### VI. Observability & Structured Logging (NON-NEGOTIABLE)

All backend services MUST use `go.uber.org/zap` as the sole structured logger.
No other logging library (`slog`, `logrus`, `log`, etc.) is permitted in service code;
shared `pkgs/` utilities MUST accept a `*zap.Logger` parameter rather than hard-coding
a logger.
The zap logger MUST be configured with an OpenTelemetry log hook so that all log
emissions are forwarded to the OTel log pipeline
(use `go.opentelemetry.io/contrib/bridges/otelzap` or equivalent bridge).
Every backend service MUST instrument the full OpenTelemetry signal triad:
- **Logs**: via the zap→OTel bridge, exported to the OTLP endpoint.
- **Metrics**: via the OTel Metrics SDK (`go.opentelemetry.io/otel/metric`),
  exported to the OTLP endpoint; expose a Prometheus scrape endpoint as a secondary
  exporter where operationally justified.
- **Traces**: via the OTel Trace SDK (`go.opentelemetry.io/otel/trace`), exported
  to the OTLP endpoint; spans MUST be created at service method boundaries and
  at transport layer entry/exit points.
OTel SDK bootstrap (resource attributes, exporters, batch processors) MUST be
initialised in `cmd/container.go` via `dig` and shut down gracefully on process exit.
Errors MUST be wrapped with contextual information at every layer boundary
(use `fmt.Errorf("...: %w", err)` or equivalent).
Trace/span context MUST be propagated across gRPC calls (via gRPC interceptors) and
RabbitMQ messages (via message headers using the W3C TraceContext propagation format).

### VII. Infrastructure-as-Code & Makefile Discipline

A `Makefile` MUST exist at the repository root with targets for each service:
`run/<service>`, `build/<service>`, `docker/build/<service>`, `test/<service>`,
and `lint/<service>`.
All infrastructure provisioning for local development MUST be reproducible via
`make dev-up` (wrapping `docker-compose up`) and `make dev-down`.
No manual steps outside documented `make` targets are acceptable for onboarding.

## Technology Stack

### Backend

- **Language**: Go (latest stable release at time of development)
- **CLI framework**: `cobra` (github.com/spf13/cobra)
- **Dependency injection**: `uber/dig` (go.uber.org/dig)
- **gRPC**: `google.golang.org/grpc` + Protocol Buffers
- **Database**: PostgreSQL (primary persistent store)
- **Cache**: Redis (optional, use where cache benefit is measurable)
- **Async messaging**: RabbitMQ (use only when async decoupling is architecturally justified)
- **Logging**: `go.uber.org/zap` (sole logger; no alternatives permitted)
- **Observability**: OpenTelemetry Go SDK — `go.opentelemetry.io/otel` (traces + metrics)
  + `go.opentelemetry.io/contrib/bridges/otelzap` (log bridge)
  + OTLP exporter (`go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc`,
  `go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc`,
  `go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc`)
- **Mock generation**: `uber/mock` (`mockgen`)
- **Testing**: standard `testing` package + `testify` for assertions;
  all unit tests MUST follow BDD + Triple-A structure

### Frontend

- **Framework**: React (current LTS)
- **Language**: TypeScript (strict mode enabled)
- **Pattern**: Custom hooks for logic, plain TSX for presentation
- **Testing**: Vitest + React Testing Library (or Jest if project tooling requires)

### Infrastructure

- **Containerization**: Docker + `docker buildx` for multi-platform images
- **Local dev orchestration**: Docker Compose
  - Required services: PostgreSQL, RabbitMQ, MinIO, `grafana/otel-lgtm`
    (Loki + Grafana + Tempo + Mimir/Prometheus bundled)
- **Build automation**: GNU Make

## Development Workflow

Every feature branch MUST be reviewed against all seven core principles before merge.
A PR is blocked if:

1. A backend service violates the `cmd/internals/pkgs/mocks` directory contract.
2. A new concrete dependency is injected manually instead of via `dig`.
3. Frontend screen files contain business logic or inline state management outside hooks.
4. Mocks are hand-written rather than generated.
5. A new container image lacks multi-platform build support.
6. The `docker-compose.yml` is not updated when a new infrastructure dependency is added.
7. The `Makefile` is not updated when a new service is added.

Code review MUST verify that each gRPC service definition is accompanied by updated
`.proto` files committed to the repository (proto-first design).

All secrets MUST be injected via environment variables; no credentials may be committed
to the repository.

## Governance

This constitution supersedes all other architectural guidance documents.
Any amendment requires:
1. A documented rationale referencing the principle(s) being changed.
2. A version bump following semantic versioning (MAJOR for principle removal/redefinition,
   MINOR for new principle or material expansion, PATCH for clarifications or wording).
3. Update of this file and re-validation of dependent templates.

All PRs and reviews MUST verify compliance with this constitution.
Complexity violations MUST be documented in the plan's Complexity Tracking table with
justification.
Refer to `.specify/memory/constitution.md` as the authoritative governance reference
during feature planning and implementation.

**Version**: 1.1.0 | **Ratified**: 2026-03-30 | **Last Amended**: 2026-03-30
