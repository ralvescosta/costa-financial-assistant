---
applyTo: "**/*"
---

# Architecture Instructions

## Rule: Backend Service Boundaries

**Description**: Each domain module has its own service entry point, wiring, and clean architecture layers.

**When it applies**: Adding or modifying backend service code.

**Copilot MUST**:
- Implement one binary per domain (`bff`, `bills`, `files`, `payments`, `onboarding`, `identity`) under `backend/cmd/<service>/`.
- Keep each service CLI entrypoint in `backend/cmd/<service>/cmd.go` (cobra command).
- Keep dependency wiring in `backend/cmd/<service>/container.go` (dig container).
- Follow clean architecture layers: `transport/` → `controllers/` (BFF) or `grpc/` handlers → `services/` → `repositories/`.
- Place shared/reusable packages only in `backend/pkgs/` (configs, otel, secrets).

**Copilot MUST NOT**:
- Mix service wiring across domain boundaries.
- Call repositories directly from transport/controller layer — always go through the service layer.
- Add business logic inside transport handler functions.
- Create ad-hoc utility packages outside `backend/pkgs/`.

**Correct layer order**:
```
backend/internals/bff/financial/controllers/documents_controller.go
  → backend/internals/files/services/document_service.go
    → backend/internals/files/repositories/document_repository.go
```

---

## Rule: Graceful Shutdown Strategy

**Description**: All backend service commands MUST implement graceful shutdown on `SIGINT` (Ctrl+C) and `SIGTERM` signals. Shutdown MUST drain in-flight work and close network listeners cleanly without data loss.

**When it applies**: Implementing any backend service in `backend/cmd/<service>/cmd.go` and `backend/cmd/<service>/container.go`.

**Copilot MUST**:
- Use `signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)` in `main.go` to create a cancellable context passed to all service commands via `cmd.Context()`.
- In each service's `container.go`, spawn a goroutine that waits on `<-ctx.Done()` then calls the transport layer's stop method:
  - **HTTP servers (BFF/Echo)**: Call `srv.Shutdown(shutdownCtx)` with a 30-second timeout context—never use `context.Background()`. Log before shutdown with `logger.Info("bff: shutting down HTTP server")` and call `logger.Sync()` after to flush buffered logs.
  - **gRPC servers**: Call `srv.GracefulStop()` (blocks until in-flight RPCs drain). Log before with `logger.Info("<service>: shutting down gRPC server")` and call `logger.Sync()` after.
  - **RabbitMQ consumers**: Pass the root context directly to `consumer.Start(ctx)`. No explicit stop method needed—the consumer's event loop will naturally exit when `ctx.Done()` triggers.
- Always log shutdown initiation with the service name for observability and call `logger.Sync()` to ensure logs flush before process exit.

**Copilot MUST NOT**:
- Use `context.Background()` in HTTP server shutdown—it will never timeout and may block indefinitely.
- Omit shutdown logging—shutdown events are critical for operational observability.
- Forget to call `logger.Sync()` after stopping a server—buffered logs may be lost on exit.
- Spawn goroutines that perform cleanup without respecting context cancellation.
- Use `panic()` or `os.Exit()` to force-stop servers—always use the graceful APIs.

**Reference pattern**:
```go
// main.go
ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
defer cancel()
root.ExecuteContext(ctx)

// container.go — HTTP example (BFF)
go func() {
    <-ctx.Done()
    logger.Info("bff: shutting down HTTP server")
    shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    _ = srv.Shutdown(shutdownCtx)
    logger.Sync()
}()

// container.go — gRPC example (bills)
go func() {
    <-ctx.Done()
    logger.Info("bills: shutting down gRPC server")
    srv.GracefulStop()
    logger.Sync()
}()
```

**Exemptions**:
- The `migrations` service (run-and-exit) is exempt from this rule.

---

## Rule: BFF with Echo + Huma (OpenAPI-First)

**Description**: The BFF service MUST use Echo as the HTTP server and Huma for OpenAPI-first route registration. All `huma.Register(...)` calls MUST live exclusively in route module files (`*_routes.go`); controllers are pure behaviour structs.

**When it applies**: Adding or modifying BFF routes, handlers, or middleware.

**Copilot MUST**:
- Register every route via `huma.Register` with `OperationID`, `Summary`, `Description`, and `Tags`.
- Place all `huma.Register(...)` calls in dedicated route module files under `backend/internals/bff/transport/http/routes/`.
- Have each route module accept a narrow capability interface (not a concrete controller type) via its constructor.
- Provide controllers to the dig container using `dig.As(new(routes.XxxCapability))` so they are resolved as capability interfaces.
- Wire `otelecho` as the first middleware for distributed trace propagation.
- Keep all business decisions in `backend/internals/bff/services/` away from controllers.
- Place JWT validation and JWKS cache logic in `backend/internals/bff/transport/http/middleware/`.
- Place project-membership and role guard logic in `backend/internals/bff/transport/http/middleware/project_guard.go`.

**Copilot MUST NOT**:
- Call `huma.Register(...)` inside a controller method — route registration belongs to route modules.
- Add a `Register(api, auth)` method to any controller struct.
- Register raw Echo routes that skip Huma operation metadata.
- Embed business rules in controller-level handler closures.
- Use any HTTP framework other than Echo (no Gin, Fiber, Chi).

**Reference files**: `backend/cmd/bff/container.go` (bootstrap), `backend/internals/bff/transport/http/routes/` (route modules), `backend/internals/bff/transport/http/controllers/` (controllers).

---

## Rule: BFF Controller/Service Boundary

**Description**: BFF controllers MUST be pure HTTP adapters. All downstream orchestration (gRPC calls, response aggregation, and multi-step logic) MUST live in `backend/internals/bff/services/`. The BFF MUST NOT access domain databases or domain repositories directly.

**When it applies**: Adding or modifying any BFF controller or service.

**Copilot MUST**:
- Keep controllers as thin HTTP adapters: extract context claims, call one BFF service method, map the result to a view type, return.
- Place all gRPC client calls and cross-service orchestration in `backend/internals/bff/services/<resource>_service.go`.
- Treat the BFF as a gateway responsible for authentication, authorization, validation, and frontend response composition — not as a domain-data owner.
- Obtain business reads/writes through the owning downstream service boundary only, using gRPC clients and service-owned contracts.
- Define BFF service contracts in `backend/internals/bff/interfaces/services.go` as narrow interfaces consumed by controllers.
- Inject BFF service interfaces (not concrete types) into controller constructors via the Dig container.
- Wire `*validator.Validate` via `validator.New` in `container.go` and pass it to every controller constructor.
- Call `b.validateInput(input)` on request body structs before delegating to the BFF service.

**Copilot MUST NOT**:
- Call a gRPC client directly from an HTTP controller handler.
- Place multi-step orchestration logic (fan-out, aggregation, field mapping from proto types) in a controller.
- Inject or call domain repositories, SQL clients, or persistence-backed domain services from BFF packages for business data access.
- Instantiate `*validator.Validate` inside a controller — always inject it via the Dig container.

**Correct layer order**:
```
[HTTP request]
  → controller.HandleXxx (extract claims, validate, call service)
    → bffinterfaces.XxxService.XxxMethod (authenticate/authorize context, orchestrate downstream)
      → billsv1.BillsServiceClient / filesv1.FilesServiceClient / paymentsv1.PaymentsServiceClient / etc.
        → downstream service business logic + repositories + database
```

**Reference files**: `backend/internals/bff/interfaces/services.go`, `backend/internals/bff/services/`, `backend/internals/bff/transport/http/controllers/base_controller.go`.

---

## Rule: BFF Service Contract Ownership and Mapper Boundary

**Description**: BFF service contracts and HTTP mappers must preserve a strict ownership boundary between transport and service layers.

**When it applies**: Adding or modifying BFF controllers, mappers, or service contracts.

**Copilot MUST**:
- Keep service-owned request/response contracts in `backend/internals/bff/services/contracts/`.
- Keep HTTP-to-service and service-to-HTTP transformations in `backend/internals/bff/transport/http/controllers/mappers/`.
- Ensure controllers call mapper helpers before and after service calls, preserving controller thinness.
- Keep `backend/internals/bff/interfaces/services.go` dependent on `services/contracts` types, not `transport/http/views` types.

**Copilot MUST NOT**:
- Import `transport/http/views` in BFF service implementation files.
- Return transport view types directly from BFF service methods.
- Move mapper logic into route modules or service implementations.

**Reference files**: `backend/internals/bff/services/contracts/`, `backend/internals/bff/transport/http/controllers/mappers/`, `backend/internals/bff/interfaces/services.go`.

---

## Rule: BFF HTTP Views Layer

**Description**: ALL BFF HTTP request and response contract types MUST be defined in `backend/internals/bff/transport/http/views/`. No HTTP-facing struct may live in a controller file.

**When it applies**: Adding or modifying any BFF HTTP request or response type.

**Copilot MUST**:
- Define every request input struct (path params, query params, request bodies) and every response output struct under `backend/internals/bff/transport/http/views/<resource>_views.go`.
- Apply `validate:"..."` struct tags to all input fields that require non-trivial validation (uuid4, required, oneof, etc.).
- Reference only `views.*` types in route capability interfaces defined in `backend/internals/bff/transport/http/routes/contracts.go`.
- Reference only `views.*` types in BFF service interface methods defined in `backend/internals/bff/interfaces/services.go`.
- Return `*views.XxxResponse` from every BFF service method; never return raw proto or repository types.

**Copilot MUST NOT**:
- Declare HTTP request or response structs inside controller files.
- Import controller-package types into route modules or route tests — always import from `views`.
- Return raw proto-generated types (e.g. `*billsv1.BillRecord`) directly from BFF service methods.

**Reference files**: `backend/internals/bff/transport/http/views/`, `backend/internals/bff/transport/http/routes/contracts.go`.

---

## Rule: AppError Translation Boundaries (Backend)

**Description**: Backend layer transitions MUST enforce `AppError` as the only cross-layer error contract.

**When it applies**: Repository, service, transport, and async boundary implementations.

**Copilot MUST**:
- Translate native dependency errors at repository/service/transport/async boundaries using `backend/pkgs/errors`.
- Propagate existing `AppError` values unchanged through intermediate layers.
- Map `AppError` category/retryability to transport-safe protocol status/messages.
- Ensure unknown/unmapped translation contexts deterministically fall back to the unknown catalog entry.

**Copilot MUST NOT**:
- Return raw dependency errors across boundaries.
- Delay translation until outer layers when boundary-local translation is available.
- Leak native dependency details in gRPC/HTTP response messages.

---

**Description**: Inter-service communication MUST use gRPC with proto-first contracts versioned under `backend/protos/`.

**When it applies**: Adding or modifying inter-service contracts.

**Copilot MUST**:
- Define domain message types in `backend/protos/<module>/v1/messages.proto`.
- Define service RPC declarations in `backend/protos/<module>/v1/grpc.proto`.
- Place shared types (`Pagination`, `Money`, `ProjectContext`, `ErrorEnvelope`, `AuditMetadata`) in `backend/protos/common/v1/`.
- Regenerate artifacts in `backend/protos/generated/` via `make proto/generate` before consuming new types.
- Implement gRPC server handlers in `backend/internals/<service>/transport/grpc/server.go`.

**Copilot MUST NOT**:
- Reuse or remove field numbers in v1 protos — introduce a new version folder for breaking changes.
- Import generated types from outside `backend/protos/generated/`.
- Directly couple BFF handlers to gRPC server implementations — call via gRPC client interfaces.

---

## Rule: Dependency Injection via dig

**Description**: All service wiring MUST use `go.uber.org/dig` in each service's `container.go`.

**When it applies**: Introducing new services, repositories, clients, or middleware.

**Copilot MUST**:
- Register all providers as constructor functions in the `dig.Container`.
- Accept interfaces, not concrete types, in constructor signatures.
- Confine container construction entirely to `backend/cmd/<service>/container.go`.

**Copilot MUST NOT**:
- Instantiate concrete structs directly in business logic files.
- Use package-level global variables as a DI substitute.
- Call constructors manually in non-test code outside container.go.

---

## Rule: Repository Contracts Placement

**Description**: Repository interfaces are centralized in each service `interfaces/` package, and repository packages contain implementations only.

**When it applies**: Adding or modifying repository contracts or repository implementations.

**Copilot MUST**:
- Declare repository interfaces in `backend/internals/<service>/interfaces/`.
- Keep `backend/internals/<service>/repositories/` focused on concrete implementations, SQL, and repository-scoped errors.
- Return `interfaces.<RepositoryContract>` from repository constructors.
- Inject repository contracts as interfaces into services via constructor parameters.

**Copilot MUST NOT**:
- Declare exported repository contracts in `backend/internals/<service>/repositories/*.go`.
- Duplicate the same repository contract across `interfaces/` and `repositories/` packages.
- Make service constructors depend on concrete repository structs unless explicitly required for a legacy exception.

**Reference files**: `backend/internals/files/interfaces/`, `backend/internals/files/repositories/`, `backend/cmd/files/container.go`.

---

## Rule: Multi-Tenant Project Isolation

**Description**: ALL domain data is scoped to `project_id`. No cross-project data access is allowed.

**When it applies**: Writing any repository query, service, or BFF handler.

**Copilot MUST**:
- Include `project_id` as a required filter in every domain table query.
- Extract and verify `project_id` from the validated JWT claim before any data access.
- Return 403 when the authenticated user is not a member of the requested project.

**Copilot MUST NOT**:
- Execute queries on tenant-scoped tables without a `project_id` predicate.
- Trust a client-supplied `project_id` value that has not been verified against the JWT claim.
- Allow any operation to read or write records belonging to a different project.

---

## Rule: Unit of Work for Transactional Repositories

**Description**: Multi-step DB writes MUST use the Unit of Work pattern.

**When it applies**: Writing repository implementations that span multiple writes.

**Copilot MUST**:
- Implement `UnitOfWork` in `backend/internals/<service>/repositories/unit_of_work.go`.
- Begin, commit, and rollback transactions explicitly through the `UnitOfWork` interface.
- Inject `UnitOfWork` via constructor into services that require atomic writes.

**Copilot MUST NOT**:
- Spread transaction lifecycle across unrelated methods.
- Use auto-commit queries for operations that must be atomic.

---

## Rule: Cache-Aside for Read-Heavy Queries

**Description**: Frequently read, project-scoped data MUST follow cache-aside through Redis.

**When it applies**: Adding read endpoints for lists or dashboard projections.

**Copilot MUST**:
- Check the Redis cache before querying the database.
- Populate the cache on miss and set a TTL.
- Invalidate the relevant cache keys on write/delete operations.

**Copilot MUST NOT**:
- Read directly from the database when a cache-aside path is already established for that query.
- Forget to invalidate cache entries after mutations.

---

## Rule: Migration-Only Schema Changes

**Description**: All database schema changes MUST be delivered as reversible SQL migration files.

**When it applies**: Adding or modifying tables, columns, indexes, or enums.

**Copilot MUST**:
- Create `.up.sql` and `.down.sql` pairs under `backend/internals/<service>/migrations/`.
- Use a sequential numeric prefix and descriptive slug: `000001_create_documents.up.sql`.
- Write all indexes with `CREATE INDEX IF NOT EXISTS`.
- Apply migrations via `make migrate/up/<service>`.

**Copilot MUST NOT**:
- Edit migration files that have already been applied.
- Perform schema changes outside the migration mechanism.
- Create an `.up.sql` without the corresponding `.down.sql`.

---

## Rule: Frontend Hook-Centric Architecture

**Description**: All server-state and business logic lives in React hooks, not in page or component bodies.

**When it applies**: Adding or modifying frontend features.

**Copilot MUST**:
- Place all `@tanstack/react-query` queries and mutations in `frontend/src/hooks/`.
- Keep page components as composition roots only — they wire hooks to UI primitives.
- Validate API request/response shapes with `zod` schemas before mutations.
- Use `react-router-dom` for all navigation and route declarations.

**Copilot MUST NOT**:
- Call `fetch` directly inside page or component bodies.
- Inline React Query logic inside JSX component files.
- Import server-state logic from one page component into another.

---

## Rule: Design Token System (Frontend)

**Description**: All visual values MUST flow through the semantic token system in `frontend/src/styles/tokens.ts`.

**When it applies**: Adding or modifying any UI component, page, or style.

**Copilot MUST**:
- Define primitive palette tokens and semantic tokens (`colorPrimary`, `colorSurface`, `colorTextPrimary`, `colorDanger`, etc.) in `frontend/src/styles/tokens.ts`.
- Map semantic tokens to Tailwind-compatible CSS variables for both `light` and `dark` themes.
- Persist the active theme preference and fall back to `prefers-color-scheme` when no preference is stored.
- Apply theme switching without a page reload.

**Copilot MUST NOT**:
- Hardcode hex, rgb, or hsl values in component or page files.
- Reference primitive palette tokens directly in components — use semantic tokens only.
- Define a semantic token without providing both light and dark values.
