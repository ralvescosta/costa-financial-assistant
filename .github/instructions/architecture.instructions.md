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

## Rule: BFF with Echo + Huma (OpenAPI-First)

**Description**: The BFF service MUST use Echo as the HTTP server and Huma for OpenAPI-first route registration.

**When it applies**: Adding or modifying BFF routes, handlers, or middleware.

**Copilot MUST**:
- Register every route via `huma.Register` with `OperationID`, `Summary`, `Description`, and `Tags`.
- Wire `otelecho` as the first middleware for distributed trace propagation.
- Keep all business decisions in `backend/internals/bff/financial/services/` away from controllers.
- Place JWT validation and JWKS cache logic in `backend/internals/bff/financial/transport/http/middleware/`.
- Place project-membership and role guard logic in `backend/internals/bff/financial/transport/http/middleware/project_guard.go`.

**Copilot MUST NOT**:
- Register raw Echo routes that skip Huma operation metadata.
- Embed business rules in controller-level handler closures.
- Use any HTTP framework other than Echo (no Gin, Fiber, Chi).

**Reference files**: `backend/cmd/bff/container.go` (bootstrap), `backend/internals/bff/financial/controllers/` (controllers).

---

## Rule: gRPC Service Contracts (Proto-First)

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
- Create `.up.sql` and `.down.sql` pairs under `backend/<service>/migrations/`.
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
