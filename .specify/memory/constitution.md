<!--
SYNC IMPACT REPORT
==================
Version change: (initial template) → 1.0.0
Modified principles: N/A — first ratification, all principles newly defined.
Added sections:
  - Core Principles (7 principles)
  - Technology Stack
  - Development Workflow
  - Governance
Removed sections: N/A
Templates requiring updates:
  - .specify/templates/plan-template.md ✅ — path conventions and constitution check gates
    aligned with monorepo structure and 7 principles; no structural edits needed,
    constitution check guidance is principle-driven.
  - .specify/templates/spec-template.md ✅ — no structural changes required; existing
    scaffold is compatible with this constitution.
  - .specify/templates/tasks-template.md ✅ — path conventions updated mentally to reflect
    monorepo layout (backend/, frontend/); no file edits required as template uses
    advisory comments.
  - .github/agents/*.md ✅ — no outdated CLAUDE-only references found in template files.
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
dependencies: PostgreSQL, RabbitMQ, and a file-storage service (e.g., MinIO).
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
Mocks MUST be auto-generated using `uber/mock` (`mockgen`) and placed in `mocks/`;
hand-written mocks are forbidden.
Integration tests MUST cover inter-service gRPC contracts and any RabbitMQ event contracts.
Frontend component tests are REQUIRED for all reusable components.
Tests MUST live alongside source (`*_test.go` for Go; `*.test.tsx` for React).
No production code may be merged that reduces existing test coverage without explicit
documented justification.

### VI. Observability & Structured Logging

All backend services MUST emit structured, leveled logs (e.g., using `slog` or `zap`).
Errors MUST be wrapped with contextual information at every layer boundary
(use `fmt.Errorf("...: %w", err)` or equivalent).
Correlation/trace IDs MUST be propagated across gRPC calls and RabbitMQ messages.
Services MUST expose Prometheus-compatible metrics endpoints where runtime telemetry
is relevant to the domain.

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
- **Mock generation**: `uber/mock` (`mockgen`)
- **Testing**: standard `testing` package + `testify` for assertions

### Frontend

- **Framework**: React (current LTS)
- **Language**: TypeScript (strict mode enabled)
- **Pattern**: Custom hooks for logic, plain TSX for presentation
- **Testing**: Vitest + React Testing Library (or Jest if project tooling requires)

### Infrastructure

- **Containerization**: Docker + `docker buildx` for multi-platform images
- **Local dev orchestration**: Docker Compose
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

**Version**: 1.0.0 | **Ratified**: 2026-03-30 | **Last Amended**: 2026-03-30
