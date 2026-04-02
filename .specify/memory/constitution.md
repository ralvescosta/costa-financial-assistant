<!--
SYNC IMPACT REPORT
==================
Version change: 1.10.0 → 1.11.0
Bump rationale: MINOR — Material governance expansion that makes end-of-execution
  memory-flow updates mandatory in every spec/plan workflow and requires instruction
  updates for refactor/reorganization work to preserve implementation patterns.
Modified principles:
  - (Development Workflow): Added PR-blocking checks that reject specs/plans/tasks
    missing mandatory memory-flow synchronization and instruction updates for
    refactor/reorganization work.
  - (Governance): Added mandatory spec and plan completion obligations for memory
    diagram sync and instruction sync.
  - (Memory References): Added explicit service-flow mapping and required end-of-
    execution update workflow for BFF, files, bills, and migration-related changes.
Added sections:
  - Governance → "Spec & Plan Completion Enforcement"
  - Memory References → "Mandatory Service-Flow Mapping"
Removed sections: None
Templates requiring updates:
  - .specify/templates/plan-template.md ✅ updated
  - .specify/templates/spec-template.md ✅ updated
  - .specify/templates/tasks-template.md ✅ updated
  - .specify/memory/architecture-diagram-maintenance.md ✅ updated
  - README.md ✅ updated
  - .specify/templates/commands/*.md ⚠ pending (directory not present in repository;
    equivalent command prompts in .github/prompts/*.prompt.md were reviewed)
Deferred TODOs: None.
-->


# Costa Financial Assistant Constitution

## Core Principles

### I. Modular Monorepo Architecture (NON-NEGOTIABLE)

The entire project MUST live in a single monorepo.
Frontend and backend are top-level modules with no shared source coupling.
Backend services (BFF API, files-grpc, bills-grpc, scheduler-grpc, onboarding-grpc,
identity-grpc, and future services) MUST be independently deployable units with no
cross-service direct imports.
Circular dependencies between modules are forbidden.
Each microservice owns its own domain; shared utilities MUST live in `pkgs/` and MUST NOT
encode domain logic.

**BFF HTTP Framework & MVC**:
The BFF API service MUST be built with `github.com/labstack/echo/v4` as the HTTP router
and `github.com/danielgtaylor/huma/v2` as the OpenAPI-first handler layer, integrated
via `github.com/danielgtaylor/huma/v2/adapters/humaecho`.
No alternative HTTP framework (Gin, Chi, Fiber, net/http ServeMux, etc.) may be used
in the BFF service.

Server bootstrap MUST follow this structure, wired through `cmd/container.go` via `dig`:
```go
e := echo.New()
e.HideBanner = true
e.Use(middleware.Recover())
e.Use(otelecho.Middleware(serviceName))   // OTel trace propagation
apiServer := humaecho.New(e, huma.DefaultConfig(serviceName, version))
```
`serviceName` and `version` MUST be supplied from the typed config struct
(loaded by `pkgs/configs`); hardcoding them is forbidden.

The BFF MUST follow the **MVC pattern** with these mandatory layers:
```
internals/
  <domain>/
    controllers/     # HTTP layer: parse/validate HTTP input, call service,
                     # and return HTTP output structs; NO business logic or route registration here
    services/        # business logic — domain operations, orchestration
    interfaces/      # Go interfaces for services and repositories
    *.go             # domain types, errors
  repositories/      # data access implementations
```
- A **Controller** file (`*_controller.go`) in `controllers/` MUST be responsible for
  one domain resource. Each handler function MUST be a method on a controller struct.
- A **Route** file (`*_routes.go`) in `transport/http/routes/` MUST be responsible for
  registering Huma operations for one resource area. Route modules MUST receive
  controller dependencies via dependency injection and MUST register themselves from
  `cmd/container.go`.
- Controllers MUST NOT contain business logic and MUST NOT own `huma.Register(...)`
  declarations. They translate HTTP Input structs to domain calls on the service
  interface and translate the result to an Output struct.
- Services MUST NOT import `huma` or `echo` packages. The service layer is
  transport-agnostic.

**OpenAPI documentation**:
Every Huma route registration MUST supply a fully populated `huma.Operation` with:
- `OperationID` — unique, kebab-case, descriptive (e.g., `create-bill`, `get-statement`).
- `Summary` — one-sentence human description.
- `Tags` — at minimum one tag grouping the resource (e.g., `["bills"]`).
- `Description` — expanded behaviour description including side effects.
Input structs MUST use `huma` validation tags (`minimum`, `maximum`, `pattern`,
`required`, `doc`) so the generated OpenAPI schema is accurate and complete.
The OpenAPI document MUST be accessible at `/openapi.json` (provided automatically
by Huma); no additional documentation format is required.

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
    unit_of_work.go # Unit of Work: transaction boundary coordinator
    *.go
    *_test.go
  transport/
    grpc/           # gRPC handlers
    rmq/            # RabbitMQ consumers/producers (.gitkeep if unused)
migrations/         # SQL migration files managed by golang-migrate
  <NNN>_<description>.up.sql
  <NNN>_<description>.down.sql
pkgs/               # shared, domain-agnostic utilities
mocks/              # auto-generated mocks (uber/mock)
```

Services MUST depend on interfaces, never on concrete implementations.
The `internals/` tree MUST NOT import from `cmd/`.
Dependency injection MUST use `uber/dig`; manual wiring in `cmd/container.go` is forbidden
beyond registration calls.
CLI entrypoints MUST use `cobra`.

Every domain concept that exposes behaviour MUST be modelled as a Go interface defined in
`internals/<domain>/interfaces/` paired with a concrete struct that implements it in
`internals/<domain>/`. All methods that belong to a type MUST be receiver methods on that
struct — standalone (orphan) functions are forbidden inside `internals/`. Only `pkgs/`
may contain package-level helper functions, because those are shared, domain-agnostic
utilities with no natural owning struct.

**Protobuf repository layout**:
All `.proto` files MUST live under `backend/protos/` following this canonical structure:

```
backend/protos/
  <module>/
    v<N>/
      messages.proto   # domain message types for this module
      grpc.proto       # gRPC service definition (Request / Response wrappers + service)
  common/
    v<N>/
      *.proto          # shared message types used by more than one module
  generated/           # Go code generated from all .proto files (committed)
```

Rules:
- **One versioned folder per gRPC module** — e.g. `bills/v1/`, `files/v1/`,
  `identity/v1/`. The folder name uses the major version (`v1`, `v2`, ...); a new
  folder MUST be created for each breaking API change; old versions are kept until
  all consumers are migrated.
- **`messages.proto`** — MUST contain only domain-level message definitions for the
  module (entities, value objects, enums). These messages are intentionally decoupled
  from transport so that the generated Go structs can be used across the domain layer
  without importing gRPC-specific types. `messages.proto` MUST NOT define any `service`
  or `rpc` blocks.
- **`grpc.proto`** — MUST contain only the gRPC service definition: `Request` and
  `Response` wrapper messages (which may embed or compose types from `messages.proto`
  via import) and the `service` / `rpc` declarations. No standalone domain messages
  MUST be defined here.
- **`common/v<N>/`** — MUST contain cross-cutting message types that are referenced
  by more than one module (e.g., pagination cursors, error envelopes, audit fields).
  No `service` blocks are allowed in `common/`.
- **`generated/`** — contains the Go source files produced by `protoc` + `protoc-gen-go`
  + `protoc-gen-go-grpc`. This directory MUST be committed to the repository so that
  consumers do not require a local `protoc` installation to build.
  A `Makefile` target (`proto/generate`) MUST regenerate all files deterministically.
- Proto files MUST use `proto3` syntax. Packages MUST follow the convention
  `<module>.v<N>` (e.g., `package bills.v1`).
- Field names in `.proto` files MUST use `snake_case`; the generated Go field names
  will be automatically converted to `PascalCase` by `protoc-gen-go`.
- Breaking changes to an existing versioned proto (renaming or removing fields,
  changing field numbers) are forbidden; introduce a new version (`v2`, etc.) instead.

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

### IV. Frontend Component-First, Hook Isolation & Design System (NON-NEGOTIABLE)

The frontend MUST use React at the current LTS version.
Every screen MUST be a pure composition of reusable components; no business logic is
permitted inside a screen presentation file.
Before creating a new component, an existing component MUST be evaluated for reuse or
extension.
Business logic, state management, and data-fetching MUST be encapsulated in custom hooks
(`use*.ts` files); the corresponding screen file contains only JSX/TSX and style bindings.
Each feature MUST ship at minimum two artifacts: a presentation file (screen) and one or
more custom hook files. Co-locating them in the same file is forbidden.

**Design Token System & Color Palette**:
The frontend MUST maintain a single, centralised design token file
(e.g., `src/styles/tokens.ts` or `src/theme/tokens.ts`) that is the authoritative
source for every color, spacing, typography scale, border radius, shadow, and z-index
value used anywhere in the application. No color or visual constant may be hardcoded
outside this file.

The color palette MUST be defined following best-practice token layering:
1. **Primitive tokens** — raw named values that define the complete palette
   (e.g., `blue500: '#3B82F6'`, `neutral900: '#111827'`). These MUST NOT be referenced
   directly in components.
2. **Semantic tokens** — purpose-driven aliases that map to primitives and carry meaning
   (e.g., `colorPrimary`, `colorSurface`, `colorTextPrimary`, `colorBorder`,
   `colorDanger`, `colorSuccess`). Components MUST reference only semantic tokens.
3. **Component tokens** — optional, component-scoped tokens that map to semantic tokens
   (e.g., `buttonPrimaryBackground`) for components that need fine-grained control.

The palette MUST include at minimum the following semantic token categories:
- Background surfaces (`colorBackground`, `colorSurface`, `colorSurfaceElevated`)
- Text (`colorTextPrimary`, `colorTextSecondary`, `colorTextDisabled`,
  `colorTextInverse`)
- Brand / interactive (`colorPrimary`, `colorPrimaryHover`, `colorPrimaryActive`)
- Status (`colorSuccess`, `colorWarning`, `colorDanger`, `colorInfo`)
- Border & divider (`colorBorder`, `colorBorderFocus`, `colorDivider`)
- Overlay / shadow (`colorOverlay`)

**Dark mode and light mode are both MANDATORY**.
The theme system MUST expose a `light` and a `dark` variant; each variant MUST define
values for every semantic token. No semantic token may be left undefined in either theme.
The active theme MUST be toggled by the user and the preference MUST be persisted
(e.g., in `localStorage`). The system MUST also respect the OS-level
`prefers-color-scheme` preference on first load when no stored preference exists.
Theme switching MUST NOT require a page reload.

The token file and theme definitions MUST be version-controlled alongside source;
design changes that alter semantic token names are breaking changes and require a
constitution patch at minimum.

**Typography Scale**:
The centralised design token file MUST define a complete typography scale.
No font size, font weight, or line-height value may be hardcoded in a component or
style file outside the token file.

The scale MUST follow a two-layer structure matching the color token approach:
1. **Primitive font-size tokens** — raw rem values
   (e.g., `fontSizeXs: '0.75rem'`, `fontSizeBase: '1rem'`). These MUST NOT be
   referenced directly in components.
2. **Semantic typography tokens** — role-named aliases mapped to primitives
   (e.g., `fontSizeBody`, `fontSizeHeading1`, `fontSizeCaption`). Components MUST
   reference only these semantic tokens.

The scale MUST include at minimum the following primitive tokens:

| Token | Value | px equiv | Use |
|---|---|---|---|
| `fontSizeXs` | `0.75rem` | 12px | Captions, badges, fine print |
| `fontSizeSm` | `0.875rem` | 14px | Secondary labels, metadata |
| `fontSizeBase` | `1rem` | 16px | Body text (default) |
| `fontSizeLg` | `1.125rem` | 18px | Large body, card content |
| `fontSizeXl` | `1.25rem` | 20px | Section sub-headings |
| `fontSize2xl` | `1.5rem` | 24px | Page sub-headings |
| `fontSize3xl` | `1.875rem` | 30px | Page primary headings |
| `fontSize4xl` | `2.25rem` | 36px | Hero / display numbers |

And at minimum the following semantic tokens:

| Semantic token | → Primitive |
|---|---|
| `fontSizeCaption` | `fontSizeXs` |
| `fontSizeBodySmall` | `fontSizeSm` |
| `fontSizeBody` | `fontSizeBase` |
| `fontSizeLabel` | `fontSizeSm` |
| `fontSizeHeading4` | `fontSizeLg` |
| `fontSizeHeading3` | `fontSizeXl` |
| `fontSizeHeading2` | `fontSize2xl` |
| `fontSizeHeading1` | `fontSize3xl` |
| `fontSizeDisplay` | `fontSize4xl` |

Font weight tokens MUST also be defined:
`fontWeightRegular` (400), `fontWeightMedium` (500), `fontWeightSemibold` (600),
`fontWeightBold` (700).

Line-height tokens MUST also be defined:
`lineHeightTight` (1.25), `lineHeightSnug` (1.375), `lineHeightNormal` (1.5),
`lineHeightRelaxed` (1.625).

All font-size values in the token file MUST use `rem` units; `px` is forbidden for
font sizes to ensure accessibility and browser zoom compatibility.

**Mobile-First Responsive Layout**:
All frontend screens and components MUST be designed and implemented mobile-first:
base styles target the smallest viewport (320px minimum) and breakpoints are applied
using `min-width` media queries exclusively. `max-width` media queries for layout
breakpoints are forbidden.

The following breakpoints MUST be defined as tokens and used consistently:

| Token | Min-width | Target |
|---|---|---|
| `breakpointSm` | `480px` | Large phones |
| `breakpointMd` | `768px` | Tablets |
| `breakpointLg` | `1024px` | Laptops / small desktops |
| `breakpointXl` | `1280px` | Desktops |
| `breakpoint2xl` | `1536px` | Large / wide desktops |

Responsive rules:
- Every screen MUST be functional and usable at 320px viewport width without horizontal
  scrolling or content clipping.
- Touch targets (buttons, links, interactive elements) MUST be at minimum 44×44 CSS
  pixels to comply with WCAG 2.5.5 and platform HIG guidelines.
- Images and media MUST use relative sizing (e.g., `max-width: 100%`, `width: 100%`)
  and MUST NOT have fixed pixel dimensions that break at narrow viewports.
- Layouts MUST be tested at a minimum of these widths: 320px, 375px, 768px, 1024px,
  1280px before merge.
- Screen density: all icons and image assets MUST be provided in SVG or at 1×/2×/3×
  resolutions to support standard, Retina, and high-DPI displays.

### V. Test Discipline

**Backend unit tests**:
Unit tests are MANDATORY for all service and repository implementations where logic is
non-trivial. All unit tests MUST follow BDD (Behaviour-Driven Development) style and
the Triple-A structure — every test body MUST contain explicit Arrange, Act, and Assert
phases, separated by blank lines and labeled with inline comments
(`// Arrange`, `// Act`, `// Assert`) to make intent unambiguous.
Test names MUST describe behaviour in the form `TestSubject_WhenCondition_ThenOutcome`
or an equivalent BDD-readable sentence.
Mocks MUST be auto-generated using `uber/mock` (`mockgen`) and placed in `mocks/`;
hand-written mocks are forbidden.

**Backend integration tests**:
Integration tests MUST stimulate the transport layer (gRPC handler or HTTP handler)
end-to-end and verify that the full request/response cycle works correctly, including
middleware, validation, persistence, and event publishing.
Integration tests MUST be placed only in canonical ownership folders under
`backend/tests/integration/`:
- `bff/`, `bills/`, `files/`, `identity/`, `onboarding/`, `payments/`, `cross_service/`.
Integration filenames MUST be behavior-based snake_case and end with `_test.go`.
Legacy story-prefixed filenames (for example `us1_*`) are forbidden.
Integration tests that require database access MUST use an **ephemeral test database**:
1. Provision a dedicated database instance (via Docker Compose test profile or a test
   environment variable pointing to an isolated schema).
2. Run all `golang-migrate` migrations (`migrate up`) against it before any test runs.
3. Execute the integration test suite.
4. Tear down the database after all tests complete (`migrate down` / container destroy).
The ephemeral database lifecycle MUST be managed in `TestMain`; individual tests MUST
NOT assume pre-existing database state — each test MUST set up and clean up its own
fixtures.
All integration scenarios MUST be table-driven where applicable and include explicit
Given/When/Then semantics plus visible Arrange/Act/Assert structure.
Integration tests MUST cover inter-service gRPC contracts and any RabbitMQ event
contracts.

**Frontend tests**:
Frontend testing is **scoped to custom hooks only** — UI component rendering tests are
explicitly out of scope and MUST NOT be written.
Every custom hook (`use*.ts`) that contains non-trivial logic MUST have a corresponding
test file (`use*.test.ts`).
Hook tests MUST follow BDD style using `describe` / `it` blocks:
- Outer `describe` = subject (hook name).
- Inner `describe` = scenario / condition.
- `it` = expected outcome in plain language.
All hook tests MUST follow the Triple-A structure (Arrange / Act / Assert) within each
`it` block.
Hook tests MUST mock all external calls (API, browser APIs, third-party libraries) at
the hook boundary; no test may make real network requests.

Tests MUST live alongside source (`*_test.go` for Go; `use*.test.ts` for React hooks).
No production code may be merged that removes an existing test without explicit
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

**Error logging rules** — whenever an error occurs, the log call MUST include:
- `zap.Error(err)` as the primary error field.
- At minimum one contextual field identifying the operation (e.g., `zap.String("op", ...)`).
- Any relevant entity IDs or input parameters that aid reproduction
  (e.g., `zap.String("user_id", ...)`, `zap.String("file_id", ...)`).
Logs MUST NOT be swallowed; every returned error MUST be logged at the point where it
cannot be propagated further upward.

**OTel context correlation** — every log call MUST attach the active OpenTelemetry
trace context so logs are correlated with traces. The `otelzap` bridge handles this
automatically when the logger is initialised with it; callers MUST pass the
`context.Context` carrying the active span to logger methods (`logger.Ctx(ctx).Error(...)`).
This ensures `trace_id` and `span_id` fields appear on every log record.

**Trace propagation**:
- gRPC: both server and client MUST use `go.opentelemetry.io/contrib/instrumentation/
  google.golang.org/grpc/otelgrpc` interceptors (unary + streaming).
- HTTP (BFF): MUST use `github.com/labstack/echo-contrib/otelecho` (the Echo-native
  OTel middleware, `otelecho.Middleware(serviceName)`) registered on the Echo instance
  before all route handlers so inbound W3C `traceparent` headers are extracted and
  outbound calls carry propagated context. The generic `otelhttp` wrapper MUST NOT be
  used on an Echo server.
- RabbitMQ: trace context MUST be injected into and extracted from message headers using
  the W3C TraceContext propagation format.

Errors MUST be wrapped with contextual information at every layer boundary
(use `fmt.Errorf("...: %w", err)` or equivalent).

**Custom metrics — per-module health**:
Each independently-deployed service MUST define at minimum:
- An `up` gauge (`<service>_up`) set to `1` while the service is healthy, `0` otherwise.
- A `build_info` gauge carrying `version`, `go_version`, and `service` labels.
These metrics MUST be registered during OTel SDK bootstrap.

**Custom metrics — BFF endpoint tracking**:
The BFF MUST expose per-endpoint HTTP metrics following RED (Rate, Errors, Duration)
best practices:
- `bff_http_requests_total` counter — labels: `method`, `route`, `status_code`.
- `bff_http_request_duration_seconds` histogram — labels: `method`, `route`;
  buckets: `[.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5]`.
- `bff_http_errors_total` counter — labels: `method`, `route`, `status_code`;
  incremented only on 4xx/5xx responses.
These metrics MUST be populated by a centralised OTel HTTP middleware; duplicating metric
instrumentation per handler is forbidden.

### VII. Infrastructure-as-Code & Makefile Discipline

A `Makefile` MUST exist at the repository root with targets for each service:
`run/<service>`, `build/<service>`, `docker/build/<service>`, `test/<service>`,
and `lint/<service>`.
All infrastructure provisioning for local development MUST be reproducible via
`make dev-up` (wrapping `docker-compose up`) and `make dev-down`.
No manual steps outside documented `make` targets are acceptable for onboarding.

### VIII. Configuration & Secrets Management (NON-NEGOTIABLE)

All service configuration MUST be stored in `.env` files, one per environment
(e.g., `.env.dev`, `.env.staging`, `.env.prod`). No configuration value may be
hardcoded in source code.

`pkgs/configs` MUST implement a typed configuration loader that:
1. Reads the `APP_ENV` environment variable to select the target `.env` file.
2. Loads the file using `github.com/spf13/viper` into a well-typed Go struct
   that is shared across the service via dependency injection.
3. After loading, scans all string values for the sentinel pattern `${<KEY>}`.
   A value matching this pattern signals that the real secret lives in an external
   vault and MUST NOT appear in the `.env` file in plaintext.

`pkgs/secrets` MUST define a `SecretsProvider` interface:
```go
type SecretsProvider interface {
    GetSecret(ctx context.Context, key string) (string, error)
}
```
Two concrete implementations MUST exist:
- **HashiCorpVaultProvider** — using `github.com/hashicorp/vault/api`.
- **AWSSecretsManagerProvider** — using
  `github.com/aws/aws-sdk-go-v2/service/secretsmanager`.

The active provider is selected by the `SECRETS_PROVIDER` config value
(`vault` | `aws`); switching providers requires no code change.

At application startup, before any handler begins accepting traffic, the config
loader MUST iterate over all `${}` sentinel fields, call `SecretsProvider.GetSecret`
for each, and replace the sentinel with the resolved value in the config struct.
If any secret resolution fails, the application MUST abort startup with a fatal log.

Security constraints:
- `.env` files containing real secret values MUST NOT be committed to the repository.
- `.gitignore` MUST exclude all `.env.*` files except `.env.example`.
- `.env.example` MUST list every required key with a placeholder value, serving as
  the canonical documentation of the service's configuration surface.

### IX. Data Access & Database Discipline (NON-NEGOTIABLE)

**Cache-aside pattern**:
Every repository read operation that is eligible for caching MUST implement the
cache-aside pattern:
1. Check Redis for a cached result using a deterministic cache key.
2. On cache hit: deserialise and return the cached value.
3. On cache miss: query PostgreSQL, store the result in Redis with an appropriate TTL,
   then return the result.
Write operations (INSERT / UPDATE / DELETE) MUST invalidate or update affected cache
entries atomically with the database write where consistency requires it.
Cache keys MUST follow the convention `<service>:<entity>:<identifier>` to prevent
collisions across services.

**Index discipline**:
Before writing any repository query, the developer MUST:
1. Identify every column referenced in `WHERE`, `ORDER BY`, `JOIN ON`, or `GROUP BY`
   clauses of that query.
2. Verify that a suitable index exists in the migration history for those columns.
3. If no adequate index exists, a new migration file MUST be created to add it
   **before or alongside** the feature that introduces the query.
Index creation migrations MUST be idempotent (`CREATE INDEX IF NOT EXISTS`).
Removing an index requires a corresponding `.down.sql` entry.

**Migration strategy**:
The sole migration runner is `github.com/golang-migrate/migrate/v4`.
Each backend service that owns a database schema MUST have a `migrations/` directory
at the service root (part of the canonical layout defined in Principle II) containing
sequentially numbered SQL files:
```
migrations/
  000001_create_<table>.up.sql
  000001_create_<table>.down.sql
  000002_add_<column_or_index>.up.sql
  000002_add_<column_or_index>.down.sql
```
Rules:
- Both DDL (schema changes: `CREATE TABLE`, `ALTER TABLE`, `CREATE INDEX`, etc.) and
  DML seed/reference data changes (`INSERT`, `UPDATE` on reference tables) MUST be
  tracked as versioned migration files — never applied manually.
- Every `.up.sql` MUST have a corresponding `.down.sql` that fully reverses the change.
- Migration files are immutable once merged; existing files MUST NOT be edited.
  Corrections require a new migration.
- Migrations MUST be applied automatically at service startup via the
  `golang-migrate` programmatic API, before the service begins accepting traffic.
- The `Makefile` MUST expose `migrate/up/<service>` and `migrate/down/<service>`
  targets for manual control during development.

**Idempotency**:
Every mutating resource endpoint or message handler — whether HTTP, gRPC, or RabbitMQ
consumer — MUST be idempotent: processing the same request or message more than once
MUST produce the same observable state as processing it exactly once.
Implementation rules:
- Each mutating operation MUST accept and persist a client-supplied or
  broker-assigned idempotency key (e.g., `X-Idempotency-Key` header for HTTP,
  a `idempotency_key` field in the Protobuf message, a `message_id` property for
  RabbitMQ AMQP messages).
- The idempotency key MUST be stored in a dedicated column or table alongside the
  resource; a unique database index MUST be created on this column to enforce
  deduplication at the persistence layer.
- On receipt of a duplicate key, the service MUST return the original response
  (or acknowledgement) without re-executing the operation.
- Idempotency key expiry policy (TTL or retention window) MUST be defined per resource
  and documented in the corresponding migration or spec.
- RabbitMQ consumers MUST check the idempotency key before processing and acknowledge
  duplicate messages without re-processing.

**Unit of Work / Atomicity**:
Whenever a business operation must modify more than one aggregate, table, or external
system (e.g., write to PostgreSQL + invalidate Redis + publish a RabbitMQ event), the
operation MUST be wrapped in a transaction boundary coordinated by the Unit of Work
pattern.
Each service MUST implement a `UnitOfWork` interface and its concrete implementation
in `internals/repositories/unit_of_work.go`:
```go
type UnitOfWork interface {
    Begin(ctx context.Context) (UnitOfWork, error)
    Commit(ctx context.Context) error
    Rollback(ctx context.Context) error
    // Repository accessors scoped to this transaction, e.g.:
    BillRepository() BillRepository
    FileRepository() FileRepository
}
```
Rules:
- The `UnitOfWork` struct MUST hold the active `*sql.Tx` (or equivalent) and expose
  transactional repository instances bound to that transaction.
- Service methods that require atomicity MUST accept a `UnitOfWork` factory via
  dependency injection (`uber/dig`) and call `Begin` → business logic → `Commit`;
  `Rollback` MUST be deferred immediately after `Begin`.
- Direct use of `*sql.DB` or `*sql.Tx` outside of repository implementations and
  `unit_of_work.go` is forbidden in `internals/`.
- Cache invalidation and outbox/event publishing that must be atomic with a database
  write MUST be coordinated inside the same Unit of Work scope using the
  transactional outbox pattern or equivalent.

### X. Multi-Tenancy, Identity & Access (NON-NEGOTIABLE)

The system is designed as a **multi-tenant** platform from inception.
Even though the initial deployment serves a single user, every data model MUST be
tenant-aware from day one so that onboarding additional users requires zero schema
changes.

**Tenant data model**:
- A `users` table MUST exist as the authoritative user registry
  (UUID PK, email, display name, status, timestamps).
- A `projects` table MUST exist. Every registered user owns one or more projects.
  Each project carries a `type` column (e.g., `personal`, `conjugal`, `shared`) and
  an `owner_id` FK to `users`.
- A `project_members` table MUST exist to model collaboration: the project owner may
  invite other users; each row carries a `role` column with exactly three values:
  - `read_only` — may query data; MUST NOT mutate.
  - `update` — may mutate existing records; MUST NOT create new top-level resources.
  - `write` — full create / update / delete access within the project scope.
- **Every domain table** (bills, statements, transactions, bank accounts, file uploads,
  etc.) MUST carry a `project_id` FK to `projects`. No table that holds user-scoped
  data may exist without tracing back to a tenant owner. Pure reference or lookup
  tables (e.g., currency codes, country codes) are exempt with explicit justification.
- All repository queries that touch tenant-owned data MUST include the caller's
  `project_id` (or a validated project membership check) in the `WHERE` clause.
  Cross-tenant data access is forbidden and MUST be enforced at the repository layer.

**Service responsibilities**:
- `onboarding-grpc`: owns the user and project lifecycle — create account, create
  project, invite collaborator, update member role, remove member, deactivate account.
- `identity-grpc`: owns token issuance and identity verification — signs JWTs with
  claims (`sub`, `project_id`, `role`, `iat`, `exp`), exposes the public JWKS endpoint,
  handles token refresh. No other service may issue tokens or embed issuer-specific
  signing logic.

**Phase-1 bootstrap** (mandatory interim strategy until full auth flow is implemented):
1. Seed migration files in `migrations/` MUST insert at least one `users` row and one
   `projects` row with stable, well-known UUIDs. These records serve as the default
   development tenant and are the fixture baseline for all integration tests.
2. `identity-grpc` MUST issue JWTs in Phase-1 whose claims (`sub`, `project_id`,
   `role`) are sourced from service configuration rather than a real auth challenge.
   No login screen or credential verification is required in Phase-1.
3. `identity-grpc` MUST expose `GET /jwks` returning a valid JWKS document containing
   the public key whose corresponding private key signs all Phase-1 JWTs, so the
   validation path across all services is identical to production from day one.
4. Every service MUST validate incoming JWTs by fetching and caching the JWKS from
   `identity-grpc`. Accepting unvalidated tokens or bypassing JWT validation in
   non-test code paths is forbidden.
5. When the full identity flow (registration, login, Keycloak / external IdP) is
   implemented, only `identity-grpc` changes internally; all consuming services
   continue JWKS-based validation with zero modifications.

**JWT signing key management**:
The private key used by `identity-grpc` to sign tokens MUST be loaded via
`pkgs/secrets` at startup (never hardcoded in source). The `GET /jwks` endpoint
MUST serve only the public-key portion of the signing key pair.

## Technology Stack

### Backend

- **Language**: Go (latest stable release at time of development)
- **CLI framework**: `cobra` (github.com/spf13/cobra)
- **Dependency injection**: `uber/dig` (go.uber.org/dig)
- **HTTP framework (BFF only)**: `github.com/labstack/echo/v4` (router) +
  `github.com/danielgtaylor/huma/v2` + `humaecho` adapter (OpenAPI-first handlers) +
  `github.com/labstack/echo-contrib/otelecho` (OTel trace middleware)
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
- **Database migrations**: `github.com/golang-migrate/migrate/v4` (sole migration runner;
  applied automatically at service startup)
- **Configuration**: `github.com/spf13/viper` (env file loading into typed struct)
- **Secrets**: `pkgs/secrets` interface; providers: `github.com/hashicorp/vault/api`
  (HashiCorp Vault) and `github.com/aws/aws-sdk-go-v2/service/secretsmanager`
  (AWS Secrets Manager)
- **JWT** (`identity-grpc` only): `github.com/golang-jwt/jwt/v5` for token issuance
  and local JWKS-validated parsing; no other service may use this for token signing
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

Every feature branch MUST be reviewed against all ten core principles before merge.
A PR is blocked if:

1. A backend service violates the `cmd/internals/pkgs/mocks/migrations` directory contract.
2. A new concrete dependency is injected manually instead of via `dig`.
3. Frontend screen files contain business logic or inline state management outside hooks.
4. Mocks are hand-written rather than generated.
5. A new container image lacks multi-platform build support.
6. The `docker-compose.yml` is not updated when a new infrastructure dependency is added.
7. The `Makefile` is not updated when a new service is added.
8. An orphan (non-receiver) function appears inside `internals/`.
9. An error is returned without being logged at its final propagation boundary.
10. A log call omits `zap.Error(err)` or contextual fields when logging an error event.
11. A log call does not pass a `context.Context` carrying the active OTel span
    (breaking log-trace correlation).
12. A new gRPC service is introduced without `otelgrpc` interceptors on both sides.
13. A new HTTP route in the BFF is introduced without the `otelecho.Middleware`
    registered on the Echo instance, or the route is registered outside of a
    controller struct's `Register(api huma.API)` method, or the Huma operation
    definition omits `OperationID`, `Summary`, `Tags`, or `Description`.
14. A new independently-deployed service ships without `up` and `build_info` health metrics.
15. A configuration value is hardcoded instead of loaded from the `.env` / config struct.
16. A secret value appears in plaintext in any committed `.env` file
    (must use `${}` sentinel and be resolved via `pkgs/secrets` at startup).
17. A new repository read operation skips the cache-aside check for Redis before
    querying PostgreSQL (unless the entity is explicitly documented as non-cacheable).
18. A new query is introduced without a corresponding index review; if no suitable index
    exists, a migration adding one MUST be part of the same PR.
19. A schema or data change is applied without a versioned migration file in `migrations/`;
    direct DDL/DML executed outside the migration runner is forbidden.
20. An existing migration file is edited after being merged (corrections require a new
    migration file).
21. A service startup sequence does not run `golang-migrate` before accepting traffic.
22. A mutating HTTP/gRPC endpoint or RabbitMQ consumer is implemented without an
    idempotency key check and a unique index on the idempotency key column.
23. A business operation that modifies more than one aggregate or interacts with more
    than one persistence/messaging system does not use `UnitOfWork` from
    `internals/repositories/unit_of_work.go`.
24. A service method directly holds or passes a `*sql.Tx` outside of repository
    implementations and `unit_of_work.go`.
25. A frontend component or style file references a hardcoded color, spacing, or other
    visual constant instead of a semantic token from the centralised design token file.
26. A new semantic token is introduced without a corresponding value defined in both
    the `light` and `dark` theme variants.
27. A frontend component or style file uses a hardcoded font size, font weight, or
    line-height value instead of a semantic typography token from the token file.
28. A screen or component layout is not implemented mobile-first (base styles for
    320px minimum; breakpoints via `min-width` only; touch targets ≥ 44×44px).
29. A new domain table is introduced without a `project_id` FK to `projects`
    establishing tenant isolation; pure reference / lookup tables require explicit
    documented justification to be exempt.
30. A service other than `identity-grpc` issues JWTs or contains token-signing logic;
    all services MUST validate tokens exclusively via the JWKS endpoint of
    `identity-grpc`.
31. A backend integration test that requires database access does not use an ephemeral
    test database (provision → `migrate up` → test → destroy), OR individual tests
    assert against state they did not set up themselves.
32. The feature spec or plan omits mandatory memory-flow impact and end-of-execution
  memory diagram update tasks for all impacted services.
33. A feature touching `bff`, `files`, `bills`, or migration-driven data/service flows
  is merged without updating the corresponding files under `.specify/memory/`
  in the same PR.
34. A refactor or structural reorganization is merged without updating affected
  instruction files under `.github/instructions/` (and `.specify/templates/` when
  workflow patterns change) to preserve implementation consistency.

Code review MUST verify that each gRPC service definition is accompanied by updated
`.proto` files committed under `backend/protos/<module>/v<N>/` following the
`messages.proto` / `grpc.proto` layout convention defined in Principle II.

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

### Spec & Plan Completion Enforcement

Every feature specification and implementation plan is governed by a mandatory
"definition of done" that includes documentation synchronization.

Mandatory rules:
1. Every `spec.md` MUST declare a `Memory Diagram Flow Impact` section identifying
  whether the feature affects memory diagrams and which files are impacted.
2. Every `plan.md` MUST include end-of-execution tasks that update all impacted memory
  diagram flow files in `.specify/memory/` within the same execution cycle.
3. If the feature changes service structure, communication paths, migrations that alter
  flow semantics, or BFF/files/bills responsibilities, the plan MUST include explicit
  updates to the corresponding memory flow documents and
  `.specify/memory/architecture-diagram.md` when cross-service flow changes occur.
4. If the feature is a refactor or structural reorganization, `spec.md` and `plan.md`
  MUST include an `Instruction Impact` section and explicit tasks to update all
  impacted instruction files under `.github/instructions/` so future implementations
  continue to follow the updated pattern.
5. A feature implementation is incomplete until these synchronization tasks are done;
  they MUST NOT be deferred to a separate PR.

## Memory References

The project maintains a curated set of **architecture and design artifacts** stored in
`.specify/memory/` that serve as the living reference documentation for system design,
service relationships, and technology decisions.

### Architecture Diagrams

**Location**: `.specify/memory/architecture-diagram.md`

The architecture diagram is the **SSOT (single source of truth)** for system design
and communication patterns. It visualizes:
- All microservices and their roles (BFF, files, bills, payments, identity, onboarding, migrations)
- Service communication protocols (gRPC, HTTP, RabbitMQ)
- Data layer (PostgreSQL, Redis, S3, RabbitMQ)
- Observability infrastructure (OpenTelemetry)
- Key data flows and integration patterns

**Update obligation**: The architecture diagram MUST be updated during sprint execution
whenever:
- A new microservice is created or removed
- Inter-service communication patterns are added or changed (gRPC, HTTP, events)
- An external dependency is introduced (new database, cache, message queue, storage)
- A service communication protocol is modified
- A critical new data flow or architectural pattern is implemented

See `.specify/memory/architecture-diagram-maintenance.md` for the complete maintenance
process, including trigger checklists, step-by-step update instructions, versioning
rules (MAJOR/MINOR/PATCH), and integration into sprint planning and PR reviews.

### Mandatory Service-Flow Mapping

For every feature, memory-flow documentation MUST be updated by impacted service scope.

Required mapping:
- `bff` changes MUST update `.specify/memory/bff-flows.md`.
- `files` changes MUST update `.specify/memory/files-service-flows.md`.
- `bills` changes MUST update `.specify/memory/bills-service-flows.md`.
- `identity` changes MUST update `.specify/memory/identity-service-flows.md`.
- `onboarding` changes MUST update `.specify/memory/onboarding-service-flows.md`.
- `migrations` changes MUST update all affected service-flow files above and MUST
  update `.specify/memory/architecture-diagram.md` whenever migration work changes
  cross-service data flow, ownership boundaries, or integration behavior.

Execution-close obligation:
- The final execution step for every spec/plan MUST include memory-flow diagram updates
  for every impacted service.
- If no memory files change, the plan MUST document an explicit no-impact rationale.
- A reviewer MUST reject any implementation where changed flows are not reflected in
  `.specify/memory/` before merge.

### Feature Diagrams (Future)

As the project grows, feature-specific diagrams may be stored in `.specify/memory/`
to document:
- Domain-specific data models and relationships (e.g., `diagram-bills-domain.md`)
- Complex workflow or state machine diagrams (e.g., `diagram-payment-reconciliation.md`)
- Integration patterns for new services (e.g., `diagram-onboarding-flow.md`)

All such diagrams follow the same usage and maintenance principle: they are **living
documentation** updated concurrently with code changes affecting the documented system.
Each feature diagram SHOULD include:
- A Mermaid graph (or ASCII diagram) visualizing the domain or flow.
- A metadata section listing the diagram's purpose, last update date, and dependencies.
- Cross-references to related architecture diagrams and specification files.

### Memory Reference Usage in Development

**During sprint planning**:
- Tech lead uses the architecture diagram to identify services affected by the current
  sprint's features.
- Assign explicit tasks to update memory diagrams if the feature introduces new services,
  connections, or data flows.

**During feature implementation**:
- Developer consults `.specify/memory/architecture-diagram.md` when designing service
  interactions to verify correct communication protocols and data flows.
- When code changes architecture (new service, new RabbitMQ consumer, new Redis cache),
  the developer creates or updates a corresponding memory diagram.

**During code review**:
- Reviewer checks the "Before Merge" diagram update checklist in
  `architecture-diagram-maintenance.md` to verify that architectural changes are
  reflected in diagrams.
- Diagrams MUST be updated in the same PR that introduces the architectural change.

**During integration test development**:
- Test author cross-references the architecture diagram to validate that the test
  exercises the documented data flow end-to-end.

**During onboarding**:
- New team members are directed to `.specify/memory/architecture-diagram.md` as the
  first reference for understanding the system architecture and service topology.

### Diagram Versioning & Changelog

All memory diagrams MUST include a metadata footer stating:
- **Version**: Semantic version (e.g., 1.2.0) incremented on each meaningful update.
- **Last Updated**: ISO date (YYYY-MM-DD).
- **Update Trigger**: The spec, feature, or issue that prompted the update.

This enables team members to quickly assess:
- Whether a diagram reflects the latest deployed state.
- What architectural changes occurred between versions.
- Which feature/spec introduced a change for traceability.

Versioning rules (apply to all memory diagrams):
- **MAJOR**: Architecture paradigm shift (e.g., add microservices, switch protocols,
  remove critical layer).
- **MINOR**: New service added, new connection added, significant new data flow pattern.
- **PATCH**: Clarifications, wording improvements, non-semantic corrections.

### Memory Diagram Consistency Obligation

If a memory diagram becomes out-of-sync with the codebase (e.g., a documented service
no longer exists, a communication pattern is not reflected in code), the diagram MUST
be updated within the sprint or explicitly marked as deprecated with a clear explanation
of why the documented pattern is no longer in use.

**Version**: 1.11.0 | **Ratified**: 2026-03-30 | **Last Amended**: 2026-04-02
