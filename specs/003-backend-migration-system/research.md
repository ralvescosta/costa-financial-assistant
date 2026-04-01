# Research: Backend Migration System Overhaul

**Feature**: 003-backend-migration-system  
**Created**: 2026-04-01

## Feature Overview

This research documents the key decisions and patterns used in creating the migration system specification.

## Migration Tools Evaluation

### golang-migrate/migrate

- **Decision**: Use golang-migrate/migrate as the migration engine
- **Rationale**:
  - Industry standard for Go projects (used by Cloud Native Computing Foundation projects)
  - Supports multiple databases (PostgreSQL, MySQL, SQLite, etc.)
  - Simple CLI and programmatic API
  - Active maintenance and large community
  - Battle-tested in production systems
- **Alternatives Considered**:
  - sql-migrate: Simpler but less flexible
  - Flyway (Java-based): Too heavyweight for Go-only projects
  - Custom migration service: High maintenance burden
- **References**: 
  - GitHub: golang-migrate/migrate
  - Documentation: https://github.com/golang-migrate/migrate

### Migration File Format

- **Decision**: Use plain SQL files (.up.sql and .down.sql)
- **Rationale**:
  - Portable across tools and languages
  - Easy to review and understand
  - Direct execution on target database
  - Version control friendly (diff-able)
- **Alternative Considered**: Embedded migrations in Go code
  - **Rejected**: Requires recompilation for schema changes

## DDL vs DML Separation

### Design Rationale

- **Decision**: Separate DDL (schema) and DML (data) into different folders and tracking tables
- **Rationale**:
  - DDL represents infrastructure; DML represents configuration/seed data
  - Different execution guarantees: DDL is versioned and atomic; DML is environment-specific
  - Enables parallel development of schema and data changes
  - Supports zero-downtime migrations (blue-green deployment patterns)
- **Implementation**:
  - migrations/ddl/ for schema changes (shared across all environments)
  - migrations/dml/<env>/ for environment-specific data (local, dev, stg, prd)
  - Separate tracking tables (migrations_ddl, migrations_dml)

### Execution Order

- **DDL First, Then DML**: Ensures schema prerequisites exist before data operations
- **Within DDL**: Numeric order (000001, 000002, etc.)
- **Within DML**: Numeric order; sorted by environment

### Why Separate Tables?

- **migrations_ddl**: Tracks schema version across all environments (shared)
- **migrations_dml**: Tracks data changes per environment (environment-specific)
- **Trade-offs**:
  - Pro: Clear separation of concerns; easier troubleshooting
  - Con: Requires checking both tables for full history
- **Mitigation**: Status command queries both tables and presents unified view

## Environment Strategy

### Environment Identifier

- **Decision**: Support four environments: local, dev, stg, prd
- **Rationale**:
  - local: Developer's machine (rapid iteration, realistic data)
  - dev: Shared development environment (unreliable, for integration testing)
  - stg: Staging (production-like, pre-production validation)
  - prd: Production (stable, verified migrations)
- **Precedent**: Standard DevOps practice (see CALMS framework)

### DML Strategy Per Environment

- **local**: Seed default user, bootstrap projects, test data
- **dev**: Same as local, plus shared test fixtures
- **stg**: Minimal seed data matching production structure
- **prd**: No seed data (only critical reference data); populated via separate processes
- **Safety**: System prevents prd DML execution unless APP_ENV=prd explicitly set

### Version Independence

- **Decision**: DML migrations share the same version counter as DDL within an environment
- **Alternative Considered**: Separate version counters for DML per environment
  - **Rejected**: Adds complexity; the global counter with environment tagging is simpler

## Folder Structure Design

### Standardization Benefits

- **Decision**: Enforce consistent folder structure across all services
- **Format**: `<service>/migrations/ddl/` and `<service>/migrations/dml/<env>/`
- **Rationale**:
  - Auto-discovery simplifies configuration
  - Enables standardized tooling and scripts
  - Scales to new services without tooling changes
  - Reduces onboarding time for new developers
- **Naming Convention**: `<NNNNNN>_<descriptive_slug>.{up,down}.sql`
  - NNNNNN: 6-digit zero-padded sequential number (000001, 000002, etc.)
  - Slug: Human-readable description (create_documents, add_index_status)
  - Enables easy parsing and sorting

## Idempotency Strategy

- **Decision**: Migrations must be idempotent (safe to run multiple times)
- **Techniques**:
  - DDL: Use CREATE IF NOT EXISTS, never DROP without condition
  - DML: Handle duplicate inserts via UPSERT or unique constraints
  - Transactions: Wrap each migration in a transaction for atomicity
- **Rationale**: Enables recovery from partial execution; simplifies testing

### Rollback Strategy

- **Decision**: Provide reversible migrations (.down.sql for every .up.sql)
- **Approaches**:
  1. Automatic generation (not reliable for complex changes)
  2. Manual writing (required; paired with .up.sql)
- **Mandate**: Never skip .down.sql; incomplete rollback capability is a deployment risk

## Tracking Table Design

### migrations_ddl Table

```sql
CREATE TABLE migrations_ddl (
  version BIGINT PRIMARY KEY,
  dirty BOOLEAN,
  name VARCHAR(255),
  executed_at TIMESTAMP,
  execution_time_ms BIGINT,
  success BOOLEAN
)
```

- Follows golang-migrate/migrate convention
- Prevents running migrations twice
- Enables status queries and audit trail

### migrations_dml Table

```sql
CREATE TABLE migrations_dml (
  version BIGINT,
  environment VARCHAR(50),
  dirty BOOLEAN,
  name VARCHAR(255),
  executed_at TIMESTAMP,
  execution_time_ms BIGINT,
  success BOOLEAN,
  PRIMARY KEY (version, environment)
)
```

- Composite key: (version, environment) allows same version in different environments
- Environment field filters which DML applies per deployment
- Separate from DDL for clarity

## CLI Design

### Command Structure

```bash
migrate up [--service <name>] [--env <env>]
migrate down [--service <name>]
migrate status [--service <name>]
migrate validate [--service <name>]
```

- Wrapped as make targets: `make migrate/up`, `make migrate/down`
- Integrates with existing go.mod dependency management
- Leverages existing Cobra CLI framework

### Error Handling

- **On Failure**: Exit with code 1; emit structured log; halt subsequent migrations
- **On Success**: Exit with code 0; summarize applied migrations
- **Logging**: Use structured logging (zap or slog) for JSON output

## Observability Integration

### Logging

- **What to log**:
  - Migration start/end with timestamp
  - Execution duration
  - SQL executed (sanitized)
  - Error messages with context
- **Format**: Structured JSON via project's logger

### Metrics

- Migration duration (histogram)
- Migration success rate (counter)
- Pending migration count (gauge)

### Tracing

- Each migration wrapped in a trace span
- Parent span: entire migration run
- Child spans: per-migration execution

## Related Components

- Database connection pool (must support transaction isolation)
- PostgreSQL driver (pq or pgx)
- CLI framework (Cobra)
- Logging framework (zap or slog)
- Config management (Viper)

## Deployment Considerations

### Zero-Downtime Migrations

- **Strategy**: DDL changes backward-compatible with old code
- **Pattern**: Add column → Populate → Remove OLD column (after code deployed)
- **Constraints**: No blocking locks on large tables

### Rollback Strategy

- **Scenario 1**: Bug found before production
  - Simply run `migrate down` to revert
- **Scenario 2**: Bug found after production deployment
  - Run `migrate down` in production (if safe)
  - OR write hotfix migration to undo damage
- **Best Practice**: Test rollbacks in staging before production

## Future Enhancements (Out of Scope)

- Automatic .down.sql generation
- Schema diffing / migration auto-generation
- Tenancy-aware migrations (per-tenant schema)
- Parallelized migration execution
- Migration validation before execution
- Integration with database CI tools (dbmate, atlas)

## Technology Stack Used

- **golang-migrate/migrate**: Core migration engine
- **PostgreSQL**: Target database
- **Cobra**: CLI framework (already in project)
- **go.uber.org/zap**: Structured logging
- **database/sql**: Go standard library for DB access
