# Feature Plan: Backend Migration System Overhaul

**Feature**: 003-backend-migration-system  
**Status**: Design Phase (Post-Clarification)  
**Priority**: P0 (Blocker for database operations)  
**Created**: 2026-04-01  
**Last Updated**: 2026-04-01 (Clarifications resolved)

## Executive Summary

This feature implements a production-grade database migration system for Costa Financial Assistant using `golang-migrate/migrate`. It provides reliable, reproducible DDL schema management and environment-aware DML data seeding across six backend services, each in its own PostgreSQL schema. The system enforces multi-tenant data isolation, implements two-factor production safety controls, and provides comprehensive operator tooling for safe database evolution.

**Key Design Decisions (from Clarification Session)**:
- ✅ Each service owns isolated PostgreSQL schema (bills, files, identity, onboarding, payments)
- ✅ Two-factor production safety: `APP_ENV=prd` + `--approve-production` flag required
- ✅ Multi-service execution order: lexicographic (bills → files → identity → onboarding → payments)
- ✅ Rollback failure recovery: manual intervention + `--force-rollback` flag
- ✅ Environment variable priority: `APP_ENV` → `ENVIRONMENT` → default `local`

## Architecture Context

### System Topology

Costa Financial Assistant is a modular monorepo with:
- **BFF Service** (Echo HTTP + Huma OpenAPI): User-facing REST API gateway
- **6 gRPC Services**: bills, files, identity, onboarding, payments, + migrations (CLI-only)
- **PostgreSQL**: Multi-schema database (one schema per service)
- **RabbitMQ**: Async event processing (document upload, reconciliation)
- **Redis**: Cache-aside for performance (JWT validation, dashboards)
- **OpenTelemetry**: Distributed tracing and metrics

Migration system sits at the **infrastructure layer**, providing schema versioning for all services.

### Per-Service Schema Model

```
PostgreSQL (Multi-Schema)
├── bills (schema)
│   ├── migrations_ddl (tracks DDL migrations)
│   ├── migrations_dml (tracks DML migrations)
│   ├── bills, bill_records, bill_types, ...
│   └── (all bill-related tables)
├── files (schema)
│   ├── migrations_ddl
│   ├── migrations_dml
│   ├── documents, document_metadata, ...
│   └── (all file-related tables)
├── identity (schema)
│   ├── migrations_ddl
│   ├── migrations_dml
│   ├── users, roles, permissions, ...
│   └── (all identity-related tables)
├── onboarding (schema)
│   ├── migrations_ddl
│   ├── migrations_dml
│   ├── registrations, projects, ...
│   └── (all onboarding-related tables)
├── payments (schema)
│   ├── migrations_ddl
│   ├── migrations_dml
│   ├── transactions, reconciliations, ...
│   └── (all payment-related tables)
└── public (schema) — RESERVED for cluster-wide objects
    ├── (no application tables permitted here)
```

**Rationale**: Per-schema isolation enforces service boundaries at the database layer, enables role-based access control (RBAC), simplifies catastrophic rollback (drop schema), and prevents accidental cross-service queries.

## Technical Design

### 1. Migration System Architecture

#### Component: Migration Service (Go)

Located in: `backend/internals/migrations/services/migration_service.go`

**Responsibilities**:
- Wrap `golang-migrate/migrate` library with Costa-specific logic
- Implement discovery algorithm for folder structure
- Orchestrate DDL → DML execution with per-service schemas
- Track executed migrations in dedicated tables
- Implement rollback with `--force-rollback` recovery path
- Enforce production safety gates

**Key Methods**:
```go
type MigrationService interface {
    // Core operations
    MigrateUp(ctx context.Context, opts MigrateOptions) error       // DDL + DML
    MigrateDown(ctx context.Context, opts MigrateOptions) error     // Rollback
    GetStatus(ctx context.Context) (*MigrationStatus, error)        // Query state
    
    // Internal helpers
    discoverMigrations(serviceName string) (*MigrationSet, error)
    executeDDL(ctx context.Context, service string, migrations []Migration) error
    executeDML(ctx context.Context, service, env string, migrations []Migration) error
    validateProductionSafety(ctx context.Context, opts MigrateOptions) error
}

type MigrateOptions struct {
    Service              string   // "bills", "files", or "" for all
    Environment          string   // "local", "dev", "stg", "prd"
    ApproveProduction    bool     // two-factor: explicit flag
    ForceRollback        bool     // recovery flag for failed rollbacks
    TargetVersion        *int     // optional: migrate to specific version
}

type MigrationStatus struct {
    ServiceStatuses map[string]ServiceMigrationStatus
    // ServiceMigrationStatus: {PendingDDL: 3, AppliedDDL: 5, PendingDML: 2, AppliedDML: 4, LastDDLVersion: 5, ...}
}
```

#### Component: Migration Tracking Tables (SQL)

Two tables created in **each service's schema** during U01 (migrations service DDL):

**`migrations_ddl` Table** (tracks schema changes):
```sql
CREATE TABLE migrations_ddl (
    version          BIGINT PRIMARY KEY,
    name             TEXT NOT NULL UNIQUE,
    description      TEXT,
    executed_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    execution_time_ms BIGINT NOT NULL,
    success          BOOLEAN NOT NULL DEFAULT TRUE,
    error_message    TEXT,
    executed_by      TEXT,  -- user/CI system that applied migration
    checksum         TEXT   -- for integrity verification
);
CREATE INDEX idx_migrations_ddl_executed_at ON migrations_ddl(executed_at DESC);
```

**`migrations_dml` Table** (tracks data changes):
```sql
CREATE TABLE migrations_dml (
    version          BIGINT NOT NULL,
    name             TEXT NOT NULL,
    environment      TEXT NOT NULL,  -- "local", "dev", "stg", "prd"
    executed_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    execution_time_ms BIGINT NOT NULL,
    success          BOOLEAN NOT NULL DEFAULT TRUE,
    error_message    TEXT,
    executed_by      TEXT,
    checksum         TEXT,
    PRIMARY KEY (version, environment)
);
CREATE INDEX idx_migrations_dml_env_executed ON migrations_dml(environment, executed_at DESC);
```

### 2. Folder Structure & File Organization

```
backend/
├── internals/
│   ├── bills/migrations/
│   │   ├── ddl/
│   │   │   ├── 000001_create_bills_table.up.sql
│   │   │   ├── 000001_create_bills_table.down.sql
│   │   │   ├── 000002_add_status_index.up.sql
│   │   │   └── 000002_add_status_index.down.sql
│   │   └── dml/
│   │       ├── local/
│   │       │   ├── 000001_seed_bill_types.up.sql
│   │       │   └── 000001_seed_bill_types.down.sql
│   │       ├── dev/
│   │       │   ├── 000001_seed_test_data.up.sql
│   │       │   └── 000001_seed_test_data.down.sql
│   │       ├── stg/
│   │       │   ├── 000001_seed_staging_fixtures.up.sql
│   │       │   └── 000001_seed_staging_fixtures.down.sql
│   │       └── prd/
│   │           ├── 000001_seed_production_defaults.up.sql
│   │           └── 000001_seed_production_defaults.down.sql
│   │
│   ├── files/migrations/ (similar structure)
│   ├── identity/migrations/ (similar structure)
│   ├── onboarding/migrations/ (similar structure)
│   ├── payments/migrations/ (similar structure)
│   │
│   └── migrations/
│       └── services/
│           ├── migration_service.go
│           ├── discovery.go (folder scanning logic)
│           ├── executor.go (golang-migrate/migrate wrapper)
│           ├── production_safety.go (two-factor checks)
│           └── logger.go (structured logging)
│
├── cmd/
│   └── migrations/
│       ├── cmd.go (cobra command definitions)
│       └── container.go (dig DI wiring)
│
└── protos/
    └── common/v1/
        └── migration_status.proto (gRPC message for status API)
```

### 3. CLI Command Design

Commands executed via `make` or direct CLI:

```bash
# Migrate all services (lexicographic order: bills, files, identity, onboarding, payments)
make migrate/up

# Migrate single service
make migrate/up/bills
make migrate/up/files

# Migrate specific environment (must be set)
make migrate/up env=dev
make migrate/up/bills env=prd --approve-production

# Rollback
make migrate/down
make migrate/down/bills version=5

# Query status
make migrate/status

# Force-rollback after failure (recovery mode)
make migrate/down/bills version=5 --force-rollback
```

**Cobra Command Structure**:
```go
// Root command: migrate root [--service SERVICE] [--env ENV]
rootCmd := &cobra.Command{
    Use:   "migrate",
    Short: "Manages database migrations",
}

// Subcommands
upCmd := &cobra.Command{
    Use:   "up",
    Short: "Apply pending migrations",
    // Flags: --service, --env, --approve-production
}

downCmd := &cobra.Command{
    Use:   "down",
    Short: "Rollback recent migration",
    // Flags: --service, --env, --version, --force-rollback
}

statusCmd := &cobra.Command{
    Use:   "status",
    Short: "Display migration status for all services",
    // Flags: --service, --format (json|table)
}
```

### 4. Execution Flow Diagrams

#### DDL Migration Flow

```
@speckit.plan Phase 1: Generate DDL flow diagram
```

#### DML Migration Flow with Environment Filtering

```
@speckit.plan Phase 1: Generate DML+environment flow diagram
```

#### Production Safety (Two-Factor) Flow

```
@speckit.plan Phase 1: Generate production approval flow diagram
```

### 5. Error Handling & Recovery

**Scenario 1: DDL Execution Fails**
- Transaction rolls back automatically (PostgreSQL)
- Migration record NOT inserted into `migrations_ddl`
- Error logged with context (migration name, SQL line, error message)
- System exits with code 1
- User re-runs command after fixing SQL

**Scenario 2: DML Execution Fails**
- Transaction rolls back automatically
- Migration record NOT inserted into `migrations_dml`
- Error logged with context
- System exits with code 1
- User re-runs or investigates data constraints

**Scenario 3: Rollback Fails (Constraint Violation)**
- Migration record LEFT UNCHANGED in tracking table (audit trail preserved)
- Error logged extensively (SQL error, context, suggestions)
- System exits with code 1
- User investigates data integrity issue
- User runs: `migrate down --version 5 --force-rollback` (requires explicit flag)
- On `--force-rollback`: System attempts rollback again, logs with "FORCED" marker, records force flag in audit

### 6. Observability Integration

#### Structured Logging

Using `go.uber.org/zap` (same as existing Costa services):

```go
logger.Info("migration: starting DDL execution",
    zap.String("service", "bills"),
    zap.Int64("migrations_count", 3),
)

logger.Info("migration: DDL applied successfully",
    zap.String("service", "bills"),
    zap.Int64("version", 1),
    zap.String("migration_name", "create_bills_table"),
    zap.Duration("execution_time", 245*time.Millisecond),
)

logger.Error("migration: DDL execution failed",
    zap.String("service", "bills"),
    zap.String("migration_name", "create_bills_table"),
    zap.Error(err),
    zap.String("sql_context", "CREATE TABLE bills ..."),
)
```

#### Tracing via OpenTelemetry

Each migration operation creates a span:

```go
ctx, span := tracer.Start(ctx, "migrate.up",
    trace.WithAttributes(
        attribute.String("service", opts.Service),
        attribute.String("environment", opts.Environment),
        attribute.Int64("migration_count", len(migrations)),
    ),
)
defer span.End()
```

For each migration:
```go
ctx, span := tracer.Start(ctx, "migrate.execute_migration",
    trace.WithAttributes(
        attribute.String("migration.name", migration.Name),
        attribute.Int64("migration.version", migration.Version),
        attribute.String("migration.type", "ddl"), // or "dml"
    ),
)
defer span.End()
```

### 7. Database Connection Management

**Connection pooling strategy**:
- Use existing `backend/pkgs/configs` connection string resolution
- Reuse database connections from service container if available
- Standalone migrations command gets its own connection pool (10 connections, 2-minute idle timeout)

**Transaction handling**:
- Each migration is wrapped in its own transaction (provided by `golang-migrate/migrate`)
- DDL and DML migrations are independent transaction scopes
- Rollback happens at transaction level (no manual transaction management needed)

### 8. CI/CD Integration

#### Makefile Targets

```makefile
# Run all migrations (local environment by default)
migrate/up:
    @cd backend && go run ./cmd/migrations up

# Run migrations for specific service
migrate/up/%:
    @cd backend && go run ./cmd/migrations up --service $*

# Rollback
migrate/down:
    @cd backend && go run ./cmd/migrations down

migrate/down/%:
    @cd backend && go run ./cmd/migrations down --service $*

# Status
migrate/status:
    @cd backend && go run ./cmd/migrations status --format table

# Production migration (explicit approval required)
migrate/up/prd:
    @cd backend && go run ./cmd/migrations up --env prd --approve-production

# Development CI target
migrate/ci:
    @cd backend && go run ./cmd/migrations up --env dev
```

#### GitHub Actions Workflow

```yaml
# .github/workflows/migrations.yml
name: Database Migrations

on:
  push:
    branches: [main]
    paths: ['backend/internals/*/migrations/']

jobs:
  migrate-staging:
    runs-on: ubuntu-latest
    environment: staging
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v4
      - run: make migrate/up/dev
        env:
          DB_HOST: ${{ secrets.STG_DB_HOST }}
          APP_ENV: stg

  migrate-production:
    runs-on: ubuntu-latest
    environment: production
    needs: migrate-staging
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v4
      - run: make migrate/up/prd --approve-production
        env:
          DB_HOST: ${{ secrets.PRD_DB_HOST }}
          APP_ENV: prd
```

---

## Design Phase Deliverables (To be Created)

### Phase 0: Research (If Needed)
- ✅ golang-migrate/migrate API and CLI modes (RESOLVED)
- ✅ PostgreSQL schema isolation best practices (RESOLVED)
- ✅ Transaction semantics for DDL vs DML (RESOLVED)

### Phase 1: Design Artifacts
- **data-model.md**: Migration tracking table schemas, indexing strategy, query patterns
- **contracts/migration_status.proto** (optional if exposing status via gRPC)
- **contracts/migration_cli_spec.md**: Detailed CLI flag and output specifications
- **quickstart.md**: Developer guide for adding migrations to a service

---

## Implementation Phases (Priority Order)

**Phase 1: Infrastructure Setup** (1-2 days)
1. Create migration folder structure for all 6 services
2. Create `migrations_ddl` and `migrations_dml` tables in each schema
3. Define initial migration files (empty DDL placeholders)
4. Test folder auto-discovery algorithm

**Phase 2: Migration Service Core** (2-3 days)
1. Implement MigrationService wrapping `golang-migrate/migrate`
2. Implement DDL execution pipeline
3. Implement DML execution pipeline with environment filtering
4. Implement rollback logic
5. Implement two-factor production safety checks

**Phase 3: CLI & Tooling** (1-2 days)
1. Implement Cobra commands (up, down, status)
2. Integrate with existing Makefile
3. Test CLI flag combinations
4. Document troubleshooting commands

**Phase 4: Observability & Testing** (2-3 days)
1. Add structured logging throughout
2. Add OpenTelemetry tracing
3. Write integration tests for all user stories
4. Test production safety gates

**Phase 5: Documentation & Cleanup** (1 day)
1. Write developer quickstart
2. Document migration file format
3. Troubleshooting guide
4. Code review and cleanup

---

## Success Metrics

| Metric | Target | Rationale |
|--------|--------|-----------|
| SC-001: DDL execution order | All migrations run in numeric order | Prevents schema conflicts |
| SC-002: Idempotent execution | No duplicates on re-run | Stability in CI/CD |
| SC-003: DDL before DML | DDL always completes first | Data operations depend on schema |
| SC-004: Environment filtering | Only specified env migrations run | Prevents cross-environment data leaks |
| SC-005: Atomic rollback | No partial changes on failure | Data integrity protection |
| SC-006: Rollback execution | Down migrations roll back in reverse | Ensures consistent undo semantics |
| SC-007: Status accuracy | Status reflects actual DB state | Operator confidence |
| SC-008: Performance SLA | All migrations < 30 seconds | Operability in CI/CD |
| SC-009: Auto-discovery | No per-service config needed | Maintainability at scale |
| SC-010: Environment seeding | Correct data per environment | Reduces manual setup |

---

## Known Constraints

- PostgreSQL only (project standard)
- 6 backend services (fixed scope)
- golang-migrate/migrate is required (dependency)
- Per-service schemas (architectural decision Q3)
- Two-factor production approval (security decision Q1)

Create initial DML migrations for local/dev environments:
- identity: `000001_seed_default_user.up.sql` (in `dml/local/` and `dml/dev/`)

## Key Dependencies

- golang-migrate/migrate library (add to go.mod)
- PostgreSQL database connectivity (already present)
- Cobra CLI framework (already available in project)
- Structured logging via zap (already in project)

## Success Metrics

- All DDL migrations execute in correct order without re-execution
- Environment-specific DML migrations execute correctly
- Migration tracking tables prevent duplicate execution
- Migrations complete for all 6 services in under 30 seconds
- Clear error logging on failures with context
- Status command accurately reports pending migrations
