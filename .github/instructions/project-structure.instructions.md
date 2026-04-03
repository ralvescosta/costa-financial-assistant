---
applyTo: "**/*"
---

# Project Structure Instructions

## Rule: Backend — Service Entry Points

**Description**: Each backend service has its own `cmd/<service>/` directory with exactly two files.

**When it applies**: Adding or modifying backend service entry points.

**Copilot MUST**:
- Place the cobra `Command` definition in `backend/cmd/<service>/cmd.go`.
- Place the `go.uber.org/dig` container build in `backend/cmd/<service>/container.go`.
- Keep entry-point files free of business logic — only CLI flag definitions and DI wiring.

**Copilot MUST NOT**:
- Add business functions to `cmd.go` or `container.go`.
- Create additional top-level `main` packages outside of `backend/cmd/`.

**Valid service names**: `bff`, `bills`, `files`, `identity`, `onboarding`, `payments`.

---

## Rule: Backend — Domain Module Layout

**Description**: Each domain module under `backend/internals/<service>/` follows a fixed sub-directory convention.

**When it applies**: Adding files to any backend domain module.

**Copilot MUST**:
- Place database/storage access in `backend/internals/<service>/repositories/`.
- Place business rules and orchestration in `backend/internals/<service>/services/`.
- Place RabbitMQ consumers and publishers in `backend/internals/<service>/transport/rmq/`.
- Place gRPC server handler implementations in `backend/internals/<service>/transport/grpc/server.go`.
- Place BFF HTTP controllers in `backend/internals/bff/financial/controllers/`.
- Place BFF HTTP middleware in `backend/internals/bff/financial/transport/http/middleware/`.

**Copilot MUST NOT**:
- Create service logic files inside `transport/` directories.
- Create HTTP handlers inside `internals/<non-bff-service>/` directories.
- Place reusable cross-service utilities anywhere except `backend/pkgs/`.

---

## Rule: Backend — Shared Packages

**Description**: Cross-service utilities live exclusively in `backend/pkgs/`.

**When it applies**: Identifying where to place code used by more than one service.

**Copilot MUST**:
- Use `backend/pkgs/configs/` for Viper-based configuration loading and struct definitions.
- Use `backend/pkgs/otel/` for OpenTelemetry bootstrap helpers (tracer, meter, logger providers).
- Use `backend/pkgs/secrets/` for secret resolution adapters (Vault, AWS Secrets Manager).

**Copilot MUST NOT**:
- Duplicate cross-service utilities in a specific domain module.
- Add domain-specific business logic to `backend/pkgs/`.

---

## Rule: Backend — Proto Layout

**Description**: gRPC contracts are versioned under `backend/protos/<module>/v1/` with a strict file split.

**When it applies**: Adding or modifying protobuf contracts.

**Copilot MUST**:
- Define domain message types in `messages.proto`.
- Define service declarations and RPC I/O wrappers in `grpc.proto`.
- Keep all shared cross-module types in `backend/protos/common/v1/messages.proto`.
- Regenerate artifacts to `backend/protos/generated/` using `make proto/generate`.

**Copilot MUST NOT**:
- Mix domain messages and service declarations in the same file.
- Import types from `backend/protos/generated/` in files outside the designated generated path.
- Modify generated files manually.

---

## Rule: Backend — Migration Files

**Description**: Database migrations live under `backend/<service>/migrations/` with sequential numeric prefixes.

**When it applies**: Adding schema changes for any service.

**Copilot MUST**:
- Name files as `<NNNNNN>_<slug>.up.sql` and `<NNNNNN>_<slug>.down.sql` (6-digit zero-padded number).
- Place onboarding migrations in `backend/onboarding/migrations/`.
- Place files-service migrations in `backend/files/migrations/`.
- Place bills-service migrations in `backend/bills/migrations/`.
- Place payments-service migrations in `backend/payments/migrations/`.

**Copilot MUST NOT**:
- Add migration files for one service into another service's migrations folder.
- Create `.up.sql` files without the corresponding `.down.sql`.

---

## Rule: Backend — Integration Tests

**Description**: Cross-service and transport-level integration tests live under `backend/tests/integration/`.

**When it applies**: Adding backend integration tests.

**Copilot MUST**:
- Place all integration test files in `backend/tests/integration/`.
- Name integration test files after the user story or flow they cover (e.g., `us1_upload_classify_test.go`).
- Implement the ephemeral DB lifecycle in `backend/tests/integration/testmain_test.go`.

**Copilot MUST NOT**:
- Place integration tests inside domain module directories.
- Rely on a pre-existing shared database.

---

## Rule: Backend — BFF Contract and Mapper Placement

**Description**: BFF transport/service boundary files must follow deterministic placement for ownership clarity.

**When it applies**: Adding or modifying BFF service contracts, mappers, and boundary tests.

**Copilot MUST**:
- Place transport-agnostic BFF service contracts in `backend/internals/bff/services/contracts/`.
- Place HTTP mapper implementations in `backend/internals/bff/transport/http/controllers/mappers/`.
- Place BFF service boundary tests in `backend/internals/bff/services/*_test.go`.
- Place BFF route registration/reachability integration tests in `backend/tests/integration/bff/*_routes_registration_test.go`.

**Copilot MUST NOT**:
- Place service contracts under `backend/internals/bff/transport/http/views/`.
- Place mapper files under `backend/internals/bff/services/`.
- Place BFF route integration tests outside `backend/tests/integration/bff/`.

---

## Rule: Frontend — Directory Conventions

**Description**: Frontend source follows a fixed directory layout under `frontend/src/`.

**When it applies**: Adding or modifying frontend source files.

**Copilot MUST**:
- Place all `@tanstack/react-query` hooks and custom data-fetching logic in `frontend/src/hooks/`.
- Place page-level composition components (route targets) in `frontend/src/pages/`.
- Place reusable UI primitives and layout components in `frontend/src/components/`.
- Place raw API client functions (fetch wrappers) in `frontend/src/services/`.
- Define all design tokens in `frontend/src/styles/tokens.ts`.
- Place router and provider setup in `frontend/src/app/`.
- Place TypeScript type definitions shared across the frontend in `frontend/src/types/`.

**Copilot MUST NOT**:
- Add server-state query logic directly in `pages/` or `components/`.
- Define color or typography values outside `tokens.ts`.
- Create nested `hooks/` directories inside `pages/` or `components/`.

---

## Rule: Frontend — Test File Placement

**Description**: Frontend test files must live alongside their corresponding hook files.

**When it applies**: Adding frontend tests.

**Copilot MUST**:
- Place hook test files in `frontend/src/hooks/<hookName>.test.ts`.
- Use the `.test.ts` extension for all hook test files.

**Copilot MUST NOT**:
- Place test files in `pages/` or `components/` directories.
- Create a separate top-level `tests/` directory for frontend hook tests.

---

## Rule: Spec and Design Artifacts

**Description**: Feature specifications and design artifacts live under `specs/<feature-slug>/`.

**When it applies**: Adding new feature documentation.

**Copilot MUST**:
- Place specs in `specs/<NNN>-<feature-slug>/` with the standard artifact set: `spec.md`, `plan.md`, `research.md`, `data-model.md`, `quickstart.md`, `tasks.md`.
- Place API contracts in `specs/<NNN>-<feature-slug>/contracts/`.
- Place validation checklists in `specs/<NNN>-<feature-slug>/checklists/`.

**Copilot MUST NOT**:
- Mix source code and documentation in the same directory.
- Store generated protobuf artifacts or compiled outputs in `specs/`.
