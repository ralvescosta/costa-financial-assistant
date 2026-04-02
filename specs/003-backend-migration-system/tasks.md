# Tasks: Backend Migration System Overhaul

**Feature**: 003-backend-migration-system  
**Status**: Ready for Implementation  
**Created**: 2026-04-01  
**Total Estimated Tasks**: 42 tasks across 7 phases  
**Critical Path**: Setup → Foundational → US1-DDL → US3-CLI → US2-DML → US4-Folder → Polish  
**MVP Checkpoint**: After US1 (DDL migrations fully functional)

---

## Phase Overview & Parallelization Strategy

### Parallel Tracks

| Phase | Tasks | Duration Est. | Can Run Parallel With | Dependencies |
|-------|-------|----------------|----------------------|--------------|
| Setup | T001-T003 | 2 hours | None | None |
| Foundational | T004-T009 | 4 hours | None | Setup |
| US1-DDL | T010-T018 | 6 hours | None | Foundational |
| **US3-CLI** | T019-T027 | 5 hours | ✅ After US1 (parallel with US2) | US1 |
| **US2-DML** | T028-T035 | 5 hours | ✅ After US1 (parallel with US3) | US1 |
| US4-Folder | T036-T039 | 2 hours | ✅ With US2/US3 | Foundational |
| Polish | T040-T042 | 3 hours | None | US2, US3, US4 |

**Recommended Execution Strategy**:
1. Run Setup sequentially (T001-T003)
2. Run Foundational sequentially (T004-T009)
3. Run US1-DDL sequentially (T010-T018)
4. **[P] Fork here**: Run US3-CLI (T019-T027) **AND** US2-DML (T028-T035) **AND** US4-Folder (T036-T039) in parallel
5. Run Polish sequentially once all complete (T040-T042)

---

## Phase 1: Setup

### [X] T001: Add golang-migrate/migrate to Dependencies [Setup]

**Objective**: Integrate golang-migrate/migrate library into the backend project.

**File Paths**:
- `backend/go.mod`
- `backend/go.sum`

**Acceptance Criteria**:
- ✅ `github.com/golang-migrate/migrate/v4` is added to `go.mod` with the latest stable version
- ✅ `go mod download` and `go mod tidy` complete successfully
- ✅ All transitive dependencies are resolved
- ✅ CI pipeline passes with new dependency

**Test Task**: No unit test for this task; validation is via `go mod verify`

**Command**:
```bash
cd backend
go get -u github.com/golang-migrate/migrate/v4
go get -u github.com/golang-migrate/migrate/v4/database/postgres
go get -u github.com/golang-migrate/migrate/v4/source/file
go mod tidy
```

**Responsible For**: FR-001

---

### [X] T002: Create Migration Tracking Tables (DDL) [Setup]

**Objective**: Design and create the two foundational migration tracking tables (`migrations_ddl` and `migrations_dml`) in each service schema.

**File Paths**:
- `backend/internals/migrations/migrations/000001_create_migration_tables.up.sql`
- `backend/internals/migrations/migrations/000001_create_migration_tables.down.sql`

**Acceptance Criteria**:
- ✅ Both SQL files exist and are syntactically valid PostgrSQL
- ✅ `.up.sql` creates `migrations_ddl` and `migrations_dml` tables with all required columns (version, name, executed_at, execution_time_ms, success, error_message, executed_by, checksum)
- ✅ `.up.sql` creates appropriate indexes on `executed_at` and `environment` columns
- ✅ `.down.sql` drops both tables cleanly without errors
- ✅ Idempotent execution: running `.up.sql` twice on same schema works (uses IF NOT EXISTS)

**Test Task** [Test-First]:
- Write integration test: `backend/tests/integration/t002_migration_tables_test.go`
  - Test 1.1: Verify `migrations_ddl` table exists after `.up.sql` with correct schema
  - Test 1.2: Verify `migrations_dml` table exists after `.up.sql` with correct schema
  - Test 1.3: Verify idempotent execution (run twice, no errors)
  - Test 1.4: Verify `.down.sql` removes tables completely
  - Test 1.5: Verify composite primary key on `migrations_dml` (version, environment)

**Responsible For**: FR-002, FR-003, FR-004

---

### [X] T003: Setup Migration Service Go Module [Setup]

**Objective**: Create the foundational Go module structure for the migration service with stub interfaces.

**File Paths**:
- `backend/internals/migrations/services/migration_service.go` (interface definition with stubs)
- `backend/internals/migrations/services/discovery.go` (stub for discovery logic)
- `backend/internals/migrations/services/executor.go` (stub for golang-migrate integration)
- `backend/internals/migrations/services/production_safety.go` (stub for two-factor checks)

**Acceptance Criteria**:
- ✅ `migration_service.go` defines `MigrationService` interface with methods:
  - `MigrateUp(ctx context.Context, opts MigrateOptions) error`
  - `MigrateDown(ctx context.Context, opts MigrateOptions) error`
  - `GetStatus(ctx context.Context) (*MigrationStatus, error)`
- ✅ `MigrateOptions` struct contains: `Service`, `Environment`, `ApproveProduction`, `ForceRollback`, `TargetVersion`
- ✅ `MigrationStatus` struct defined with service-level summaries
- ✅ All functions have doc comments explaining purpose and parameters
- ✅ Code compiles (stubs may panic with `TODO: implement`)

**Test Task**: No unit test; validation is via `go build ./...`

**Responsible For**: FR-001 (partial - foundational setup)

---

## Phase 2: Foundational

### [X] T004: Implement golang-migrate Wrapper (Low-Level) [Foundational]

**Objective**: Create a thin wrapper around golang-migrate that discovers migrations from the standardized folder structure and executes them.

**File Paths**:
- `backend/internals/migrations/services/executor.go`
- `backend/internals/migrations/services/executor_test.go`

**Acceptance Criteria**:
- ✅ `Executor` struct wraps `*migrate.Migrate` instance
- ✅ `NewExecutor(db *sql.DB, sourceURL string) (*Executor, error)` creates instance from postgres DSN
- ✅ `UpAll(ctx context.Context) error` runs all pending migrations
- ✅ `Down(ctx context.Context) error` rolls back the most recent migration
- ✅ `DownN(ctx context.Context, n int) error` rolls back N migrations
- ✅ `Version(ctx context.Context) (uint, bool, error)` returns current version
- ✅ Error handling: returns non-nil error if any operation fails

**Test Task** [Test-First]:
- Write unit test: `backend/internals/migrations/services/executor_test.go`
  - Test 2.1: Create executor with valid PostgreSQL connection
  - Test 2.2: UpAll applies all pending migrations without errors
  - Test 2.3: Version returns correct current version after UpAll
  - Test 2.4: Down rolls back most recent migration
  - Test 2.5: DownN(2) rolls back exactly 2 migrations
  - Test 2.6: Error handling: invalid DSN returns error
  - Test 2.7: Idempotency: UpAll twice on same DB = same final state

**Responsible For**: FR-001

---

### [X] T005: Implement Migration Discovery Algorithm [Foundational]

**Objective**: Implement folder scanning logic to auto-discover migration files from the standardized structure.

**File Paths**:
- `backend/internals/migrations/services/discovery.go`
- `backend/internals/migrations/services/discovery_test.go`

**Key Functions**:
- `DiscoverMigrations(basePath string) (map[string]*MigrationSet, error)` - scans all service folders
- `DiscoverServiceMigrations(servicePath string) (*MigrationSet, error)` - scans one service
- `ScanDDL(servicePath string) ([]Migration, error)` - finds ddl/ folder
- `ScanDML(servicePath string, env string) ([]Migration, error)` - finds dml/<env>/ folder

**Acceptance Criteria**:
- ✅ Scans `backend/internals/<service>/migrations/ddl/` and discovers all `.up.sql` and `.down.sql` pairs
- ✅ Scans `backend/internals/<service>/migrations/dml/<environment>/` and discovers environment-specific migrations
- ✅ Returns `[]Migration` sorted by numeric version (000001 before 000002)
- ✅ Validates that every `.up.sql` has a matching `.down.sql` (or vice versa); returns error if mismatch
- ✅ Extracts version and name from filename using regex: `^(\d{6})_(.+)\.(up|down)\.sql$`
- ✅ Handles missing folders gracefully (returns empty list, not error)
- ✅ Supports all six services: bills, files, identity, onboarding, payments, migrations

**Test Task** [Test-First]:
- Write unit test: `backend/internals/migrations/services/discovery_test.go`
  - Test 3.1: Scan valid DDL folder with 3 migrations, returns 3 sorted by version
  - Test 3.2: Scan valid DML folder for 'local' environment, returns environment-specific migrations
  - Test 3.3: Error on unpaired migration (up without down)
  - Test 3.4: Extract version and name correctly from filename `000001_create_table.up.sql`
  - Test 3.5: Handle missing folder (returns empty, no error)
  - Test 3.6: Discover all six services, each with its own migrations
  - Test 3.7: Sorting: migrations returned in numeric order (000001, 000002, ..., 000010)

**Responsible For**: FR-006, FR-007, FR-008, FR-009

---

### [X] T006: Implement Production Safety Check [Foundational]

**Objective**: Implement two-factor safety validation to prevent accidental production migration execution.

**File Paths**:
- `backend/internals/migrations/services/production_safety.go`
- `backend/internals/migrations/services/production_safety_test.go`

**Key Functions**:
- `ValidateProductionAccess(ctx context.Context, env string, approveProduction bool) error`
- `GetEnvironment() string` - checks APP_ENV, ENVIRONMENT, defaults to local

**Acceptance Criteria**:
- ✅ `GetEnvironment()` checks `APP_ENV` first, then `ENVIRONMENT`, defaults to `local`; logs which source was used
- ✅ Rejects invalid environment values (only allows: local, dev, stg, prd)
- ✅ For non-production (local, dev, stg): returns nil (no extra approval needed)
- ✅ For production (prd):
  - MUST have `APP_ENV=prd` explicitly set (checked by GetEnvironment)
  - MUST have `approveProduction=true` flag set
  - If either fails: return error with message "Production migration requires APP_ENV=prd and --approve-production flag"
- ✅ Logs all checks for audit trail: "Environment approved: prd with explicit approval flag"

**Test Task** [Test-First]:
- Write unit test: `backend/internals/migrations/services/production_safety_test.go`
  - Test 4.1: GetEnvironment from APP_ENV takes priority
  - Test 4.2: GetEnvironment falls back to ENVIRONMENT if APP_ENV not set
  - Test 4.3: GetEnvironment defaults to local if neither set
  - Test 4.4: Rejects invalid environment (e.g., "prod" instead of "prd")
  - Test 4.5: Local/dev/stg approved without ApproveProduction flag
  - Test 4.6: Prd requires both APP_ENV=prd and ApproveProduction=true
  - Test 4.7: Prd rejected if APP_ENV≠prd (even with flag)
  - Test 4.8: Prd rejected if ApproveProduction=false (even with APP_ENV=prd)

**Responsible For**: FR-027, FR-028, FR-029

---

### [X] T007: Implement Migration Record Tracking [Foundational]

**Objective**: Create functions to insert/query migration records into tracking tables with proper error handling.

**File Paths**:
- `backend/internals/migrations/services/tracker.go`
- `backend/internals/migrations/services/tracker_test.go`

**Key Functions**:
- `RecordDDLMigration(ctx context.Context, tx *sql.Tx, version int, name string, duration time.Duration, success bool, errorMsg string) error`
- `RecordDMLMigration(ctx context.Context, tx *sql.Tx, version int, name string, env string, duration time.Duration, success bool, errorMsg string) error`
- `IsDDLMigrationApplied(ctx context.Context, tx *sql.Tx, version int) (bool, error)`
- `IsDMLMigrationApplied(ctx context.Context, tx *sql.Tx, version int, env string) (bool, error)`

**Acceptance Criteria**:
- ✅ `RecordDDLMigration` inserts into `migrations_ddl` table with all fields
- ✅ `RecordDMLMigration` inserts into `migrations_dml` table with (version, environment) scope
- ✅ Both functions set `executed_at=NOW()`, `executed_by=CURRENT_USER` automatically
- ✅ `IsDDLMigrationApplied` returns true/false based on `WHERE version = ?`
- ✅ `IsDMLMigrationApplied` returns true/false based on `WHERE version = ? AND environment = ?`
- ✅ All functions accept tx (transaction) for atomic operations (no auto-commit)
- ✅ Error handling: returns error if insert fails (e.g., unique constraint violation)

**Test Task** [Test-First]:
- Write unit test: `backend/internals/migrations/services/tracker_test.go`
  - Test 5.1: RecordDDLMigration inserts record with correct fields
  - Test 5.2: RecordDMLMigration inserts with (version, environment) composite key
  - Test 5.3: IsDDLMigrationApplied(1) returns true after recording version 1
  - Test 5.4: IsDMLMigrationApplied(1, 'local') returns true after recording
  - Test 5.5: IsDMLMigrationApplied(1, 'dev') returns false if only 'local' recorded
  - Test 5.6: Error on duplicate unique key (e.g., recording same version twice)
  - Test 5.7: Records in transaction context (not auto-committed)

**Responsible For**: FR-002, FR-003, FR-004, FR-005

---

### [X] T008: Implement Environment Variable Reading with Logging [Foundational]

**Objective**: Centralize environment variable reading with comprehensive logging and validation.

**File Paths**:
- `backend/pkgs/configs/environment.go`

**Key Functions**:
- `GetAppEnvironment(logger Logger) string`

**Acceptance Criteria**:
- ✅ Reads `APP_ENV`, falls back to `ENVIRONMENT`, defaults to `local`
- ✅ Logs decision with structured fields: `logger.Info("environment_determined", zap.String("env", env), zap.String("source", source))`
- ✅ Validates environment is one of: local, dev, stg, prd (returns error if invalid)
- ✅ Returns error message: "Invalid environment: %s. Must be one of: local, dev, stg, prd"
- ✅ Suitable for reuse across all services (general-purpose package)

**Test Task**: No dedicated test; validation via logging output inspection

**Responsible For**: FR-027

---

### [X] T009: Setup Cobra CLI Command Structure [Foundational]

**Objective**: Create the Cobra command scaffold for all migration CLI commands.

**File Paths**:
- `backend/cmd/migrations/cmd.go`
- `backend/cmd/migrations/container.go`

**Acceptance Criteria**:
- ✅ Root command `migrate` defined with description "Manages database migrations"
- ✅ Subcommands defined (stubs only):
  - `migrate up` with flags: --service, --env, --approve-production
  - `migrate down` with flags: --service, --env, --version, --force-rollback
  - `migrate status` with flags: --service, --format
- ✅ All flags have help text and appropriate types (string, bool, int)
- ✅ Commands parse flags correctly without errors
- ✅ `container.go` scaffolded to wire MigrationService via dig (empty implementation)

**Test Task**: No unit test; validation via `go build ./...` and `./backend up --help`

**Responsible For**: FR-016, FR-017, FR-018, FR-019, FR-020

---

## Phase 3: US1 — Migrate DDL Schema Changes (Priority: P0)

### [X] T010: Implement DDL Migration Orchestration Logic [US1-DDL]

**Objective**: Implement the core DDL execution flow: discover DDL files, validate prerequisites, execute in order, track records.

**File Paths**:
- `backend/internals/migrations/services/migration_service.go` (implement MigrateUp for DDL phase)
- `backend/internals/migrations/services/migration_service_test.go` (unit tests)

**Key Function**:
- `(ms *MigrationService) MigrateUpDDL(ctx context.Context, serviceName string) error`

**Acceptance Criteria**:
- ✅ Accepts serviceName (e.g., "files") or empty string (all services)
- ✅ If empty service name: iterates all services in lexicographic order: bills, files, identity, onboarding, payments
- ✅ For each service:
  1. Creates service schema if missing: `CREATE SCHEMA IF NOT EXISTS <service>;`
  2. Creates `migrations_ddl` and `migrations_dml` tables if missing
  3. Discovers `.up.sql` files from `<service>/migrations/ddl/` in sort order
  4. For each migration:
     a. Checks if already applied via `IsDDLMigrationApplied`; skips if yes
     b. Begins transaction
     c. Executes `.up.sql` in service schema context
     d. Records in `migrations_ddl` tracking table (success=true)
     e. Commits transaction
   5. Logs summary: "Applied 3 DDL migrations to bills schema"
- ✅ On any error:
  1. Rolls back current transaction
  2. Does NOT record failed migration in tracking table
  3. Logs detailed error with migration name and SQL line number
  4. Returns error and stops execution (no further migrations attempted)
- ✅ DDL execution completes before returning (blocking, not async)

**Test Task** [Test-First]:
- Write integration test: `backend/tests/integration/t010_ddl_orchestration_test.go`
  - Test 10.1: Execute DDL for 'files' service on clean database
    - Verify: schema created, tables created, migrations_ddl has 1 record
  - Test 10.2: Execute DDL for all services (no service param)
    - Verify: all 6 services processed in order (bills, files, identity, onboarding, payments)
  - Test 10.3: Idempotency - execute DDL twice, second run skips already-applied migrations
    - Verify: migrations_ddl has same records, no duplicates
  - Test 10.4: Failure handling - malformed SQL in migration
    - Verify: transaction rolled back, no record in migrations_ddl, error logged
  - Test 10.5: Multiple DDL migrations executed in numeric order
    - Verify: 000001 completes before 000002 (verify via timestamps)

**Responsible For**: FR-001, FR-005, FR-005a, FR-005b, FR-005c, FR-011, FR-014, FR-030

---

### [X] T011: Create Initial DDL Migrations for Bills Service [US1-DDL]

**Objective**: Create foundational DDL migrations for the bills service to establish its schema.

**File Paths**:
- `backend/internals/bills/migrations/ddl/000001_create_bills_table.up.sql`
- `backend/internals/bills/migrations/ddl/000001_create_bills_table.down.sql`
- `backend/internals/bills/migrations/ddl/000002_add_bill_indexes.up.sql`
- `backend/internals/bills/migrations/ddl/000002_add_bill_indexes.down.sql`

**Acceptance Criteria**:
- ✅ `.up` files create required tables with all columns (id, project_id, bill_type, status, created_at, updated_at, etc.)
- ✅ `.up` files include FOREIGN KEY constraints to projects table
- ✅ `.up` files create indexes on frequently-queried columns (project_id, status, created_at)
- ✅ `.down` files drop tables in reverse dependency order (indexes drop automatically with tables)
- ✅ SQL uses `CREATE TABLE IF NOT EXISTS` for idempotency
- ✅ All SQL is valid PostgreSQL syntax

**Test Task**: User Story 1 Integration Test (T010) covers this; additionally inspect schema after T010 execution

**Responsible For**: FR-006, FR-007, FR-010

---

### [X] T012: Create DDL Migrations for Files Service [US1-DDL]

**Objective**: Create DDL migrations for files service (documents, metadata, etc.).

**File Paths**:
- `backend/internals/files/migrations/ddl/000001_create_documents_table.up.sql`
- `backend/internals/files/migrations/ddl/000001_create_documents_table.down.sql`
- `backend/internals/files/migrations/ddl/000002_add_indexes.up.sql`
- `backend/internals/files/migrations/ddl/000002_add_indexes.down.sql`

**Acceptance Criteria** (same as T011):
- ✅ Tables with all required columns (id, filename, file_hash, classification, analysis_status, etc.)
- ✅ Indexes on frequently-queried columns
- ✅ Foreign key constraints
- ✅ `.down` files drop tables cleanly

**Test Task**: Validated via Phase 3 integration tests (T010)

**Responsible For**: FR-006, FR-007, FR-010

---

### [X] T013: Create DDL Migrations for Identity Service [US1-DDL]

**Objective**: Create DDL migrations for identity service (users, roles, permissions).

**File Paths**:
- `backend/internals/identity/migrations/ddl/000001_create_users_table.up.sql`
- `backend/internals/identity/migrations/ddl/000001_create_users_table.down.sql`
- `backend/internals/identity/migrations/ddl/000002_create_roles_table.up.sql`
- `backend/internals/identity/migrations/ddl/000002_create_roles_table.down.sql`
- `backend/internals/identity/migrations/ddl/000003_add_timestamps_indexes.up.sql`
- `backend/internals/identity/migrations/ddl/000003_add_timestamps_indexes.down.sql`

**Acceptance Criteria**: Same as T011

**Test Task**: Validated via Phase 3 integration tests (T010)

**Responsible For**: FR-006, FR-007, FR-010

---

### [X] T014: Create DDL Migrations for Onboarding Service [US1-DDL]

**Objective**: Create DDL migrations for onboarding service (registrations, projects, etc.).

**File Paths**:
- `backend/internals/onboarding/migrations/ddl/000001_create_projects_table.up.sql`
- `backend/internals/onboarding/migrations/ddl/000001_create_projects_table.down.sql`
- `backend/internals/onboarding/migrations/ddl/000002_add_ownership_constraints.up.sql`
- `backend/internals/onboarding/migrations/ddl/000002_add_ownership_constraints.down.sql`

**Acceptance Criteria**: Same as T011

**Test Task**: Validated via Phase 3 integration tests (T010)

**Responsible For**: FR-006, FR-007, FR-010

---

### [X] T015: Create DDL Migrations for Payments Service [US1-DDL]

**Objective**: Create DDL migrations for payments service (transactions, reconciliation).

**File Paths**:
- `backend/internals/payments/migrations/ddl/000001_create_transactions_table.up.sql`
- `backend/internals/payments/migrations/ddl/000001_create_transactions_table.down.sql`
- `backend/internals/payments/migrations/ddl/000002_create_reconciliation_table.up.sql`
- `backend/internals/payments/migrations/ddl/000002_create_reconciliation_table.down.sql`

**Acceptance Criteria**: Same as T011

**Test Task**: Validated via Phase 3 integration tests (T010)

**Responsible For**: FR-006, FR-007, FR-010

---

### [X] T016: Implement Rollback Logic for DDL [US1-DDL]

**Objective**: Implement the `.down.sql` execution flow with proper cascade ordering and error recovery.

**File Paths**:
- `backend/internals/migrations/services/migration_service.go` (implement MigrateDown method)
- `backend/internals/migrations/services/migration_service_test.go` (tests)

**Key Function**:
- `(ms *MigrationService) MigrateDown(ctx context.Context, opts MigrateOptions) error`

**Acceptance Criteria**:
- ✅ Accepts serviceName and optional targetVersion
- ✅ If no targetVersion: rolls back most recent migration
- ✅ If targetVersion specified: rolls back all migrations >= targetVersion in reverse order (most recent first)
- ✅ Before rollback:
  1. Checks if rollback is safe (e.g., no data dependencies); logs warnings if risky
  2. If fails due to constraint: returns error, does NOT modify tracking table, logs detailed error message
  3. Exits with code 1
- ✅ On successful rollback:
  1. Executes `.down.sql` for the migration
  2. Deletes the record from `migrations_ddl` (or marks success=false if audit trail required)
  3. Logs: "Rolled back migration: create_bills_table (version 1)"
- ✅ On rollback failure (with --force-rollback flag):
  1. Attempts rollback again
  2. Logs with [FORCED] marker for audit trail
  3. If fails again, stops and returns error

**Test Task** [Test-First]:
- Write integration test: `backend/tests/integration/t016_rollback_test.go`
  - Test 16.1: Apply DDL migration, then rollback
    - Verify: table exists after up; table dropped after down; migrations_ddl records cleaned up
  - Test 16.2: Rollback most recent migration only (leaves previous ones intact)
  - Test 16.3: Rollback with constraint violation
    - Verify: error returned, migrations_ddl NOT modified, detailed error logged
  - Test 16.4: Rollback with --force-rollback flag retries after failure
  - Test 16.5: Multiple rollbacks in sequence (down N migrations)

**Responsible For**: FR-025, FR-026, FR-026a, FR-030

---

### [X] T017: Implement Status Command for DDL Queries [US1-DDL]

**Objective**: Implement migration status reporting showing applied vs pending migrations.

**File Paths**:
- `backend/internals/migrations/services/migration_service.go` (implement GetStatus method)
- `backend/internals/migrations/services/migration_service_test.go`

**Key Function**:
- `(ms *MigrationService) GetStatus(ctx context.Context) (*MigrationStatus, error)`

**MigrationStatus Structure**:
```go
type MigrationStatus struct {
    Active   bool
    Dirty    bool
    Version  uint
    ServiceStatuses map[string]ServiceMigrationStatus
    // ServiceMigrationStatus: {PendingDDL: 3, AppliedDDL: 5, LastDDLVersion: 5, LastDDLTime: time.Time, ...}
}
```

**Acceptance Criteria**:
- ✅ Queries `migrations_ddl` and `migrations_dml` tracking tables for each service
- ✅ Calculates:
  - Total applied DDL migrations per service
  - Total pending DDL migrations (files exist but not yet recorded)
  - Last applied version and timestamp
- ✅ Returns structured data for programmatic use (JSON, table, etc.)
- ✅ Can format as table or JSON for CLI output

**Test Task** [Test-First]:
- Write unit test: `backend/internals/migrations/services/migration_service_test.go::TestGetStatus`
  - Test 17.1: After applying 3 DDL migrations, GetStatus shows AppliedDDL=3, PendingDDL=0
  - Test 17.2: With 5 DDL files but only 3 applied, shows PendingDDL=2
  - Test 17.3: All services included in status
  - Test 17.4: Empty database shows pending=count of all .up.sql files

**Responsible For**: FR-018, FR-030, FR-032

---

### T018: Write Integration Test Suite for US1 (DDL) [US1-DDL]

**Objective**: Comprehensive test suite covering all DDL user story scenarios from spec.md.

**File Paths**:
- `backend/tests/integration/us1_ddl_migrations_test.go`
- `backend/tests/integration/testmain_test.go` (update with US1 test setup)

**Test Cases** (mapping to spec scenarios):
- **Scenario 1.1**: DDL files discovered and executed in order
- **Scenario 1.2**: migrations_ddl table created and first migration recorded
- **Scenario 1.3**: Multiple DDL migrations executed in numeric order
- **Scenario 1.4**: Already-executed migration skipped on re-run
- **Scenario 1.5**: Failed DDL migration rolled back, not recorded
- **Scenario 1.6**: Rollback executes .down.sql and removes record
- **Scenario 1.7**: .down.sql files exist and execute correctly
- **Scenario 1.8**: Multiple services' migrations run independently

**Acceptance Criteria**:
- ✅ All 8 scenarios from spec.md covered by test cases
- ✅ Tests pass consistently (deterministic)
- ✅ Each test is independent and can run in any order
- ✅ Tests verify final state (schema, tracking tables) not just side effects

**Test Task** [Already inherent to this task]:
- Full test coverage for US1 scenarios

**Responsible For**: SC-001, SC-002, SC-005, SC-006, SC-009

---

## Phase 4: US3 — Migration Module CLI Commands (Priority: P1)

**[P] This phase runs parallel with US2-DML after US1 completes. Can be executed by a second developer/team.**

### [X] T019: Implement migrate up CLI Command [US3-CLI]

**Objective**: Implement the Cobra command handler for `migrate up` with all flags and validation.

**File Paths**:
- `backend/cmd/migrations/cmd.go` (implement upCmd)
- `backend/cmd/migrations/cmd_test.go` (unit tests for command)

**Key Function**:
- `upCmd.RunE` executes: parse flags → validate → call MigrationService.MigrateUp

**Acceptance Criteria**:
- ✅ Parses flags: --service (string), --env (string), --approve-production (bool)
- ✅ Validates environment is one of [local, dev, stg, prd]
- ✅ Calls production safety check if env=prd
- ✅ On success: `make migrate/up` exits 0, logs "Applied X DDL and Y DML migrations"
- ✅ On failure: exits non-zero, logs detailed error message
- ✅ Logs execution time (total duration)

**Test Task** [Test-First]:
- Write unit test: `backend/cmd/migrations/cmd_test.go`
  - Test 19.1: Parse --service flag
  - Test 19.2: Parse --env flag with validation
  - Test 19.3: Execute with mock MigrationService, verify success
  - Test 19.4: Production safety check enforced when env=prd
  - Test 19.5: Exit code 0 on success, non-zero on failure

**Responsible For**: FR-016, FR-019, FR-020, FR-021, FR-022

---

### [X] T020: Implement migrate down CLI Command [US3-CLI]

**Objective**: Implement Cobra command handler for `migrate down` with version targeting and force-rollback support.

**File Paths**:
- `backend/cmd/migrations/cmd.go` (implement downCmd)
- `backend/cmd/migrations/cmd_test.go`

**Acceptance Criteria**:
- ✅ Parses flags: --service, --version (int), --force-rollback (bool)
- ✅ If --force-rollback and rollback fails: logs [FORCED] marker, attempt retry
- ✅ On success: exits 0, logs summary
- ✅ On failure: exits non-zero, logs error with suggestion to use --force-rollback

**Test Task** [Test-First]:
- Write unit test: `backend/cmd/migrations/cmd_test.go`
  - Test 20.1: Parse --version flag
  - Test 20.2: Parse --force-rollback flag
  - Test 20.3: Execute rollback with mock service
  - Test 20.4: Force-rollback retries and logs

**Responsible For**: FR-017, FR-023, FR-024a, FR-025, FR-026, FR-026a

---

### [X] T021: Implement migrate status CLI Command [US3-CLI]

**Objective**: Implement Cobra command for `migrate status` displaying migration state.

**File Paths**:
- `backend/cmd/migrations/cmd.go` (implement statusCmd)
- `backend/cmd/migrations/cmd_test.go`

**Acceptance Criteria**:
- ✅ Parses flag: --service (optional), --format (json|table, default=table)
- ✅ Calls MigrationService.GetStatus()
- ✅ Output formats:
  - **Table**: Rows per service showing: Service | Applied DDL | Pending DDL | Applied DML | Pending DML | Last Version
  - **JSON**: Structured output for programmatic use
- ✅ Always exits 0 (even if pending migrations exist)

**Test Task** [Test-First]:
- Write unit test: `backend/cmd/migrations/cmd_test.go`
  - Test 21.1: Status shows correct counts
  - Test 21.2: Table format output is readable
  - Test 21.3: JSON format output is valid JSON

**Responsible For**: FR-018, FR-021, FR-022

---

### [X] T022: Add migrate/up/% Makefile Targets [US3-CLI]

**Objective**: Create Makefile targets for easy invocation of migration commands.

**File Paths**:
- `Makefile` (add targets)

**Acceptance Criteria**:
- ✅ `make migrate/up` → runs `migrate up` with all services
- ✅ `make migrate/up/bills` → runs `migrate up --service bills`
- ✅ `make migrate/up/files` → runs `migrate up --service files`
- ✅ `make migrate/up/identity` → runs `migrate up --service identity`
- ✅ `make migrate/up/onboarding` → runs `migrate up --service onboarding`
- ✅ `make migrate/up/payments` → runs `migrate up --service payments`
- ✅ `make migrate/down` → runs `migrate down`
- ✅ `make migrate/status` → runs `migrate status --format table`
- ✅ `make migrate/validate` → validates folder structure and migration files

**Test Task**: Manual verification: `make migrate/up` works correctly

**Responsible For**: FR-016, FR-017, FR-018, FR-020

---

### [X] T023: Implement Logging and Structured Output [US3-CLI]

**Objective**: Add comprehensive structured logging to all CLI commands using zap/slog.

**File Paths**:
- `backend/cmd/migrations/cmd.go` (update all commands with logging)
- `backend/internals/migrations/services/migration_service.go` (structured logging)

**Acceptance Criteria**:
- ✅ Each migration step logged with timestamp, migration name, status, duration
- ✅ Example: `{"level":"info","ts":"2026-04-01T10:30:00Z","msg":"Applied migration","version":1,"name":"create_bills_table","duration_ms":45,"service":"bills"}`
- ✅ Error logs include: migration name, SQL error details, line number (if available)
- ✅ Success summary logs total count and total duration

**Test Task**: Validate log output format during integration tests

**Responsible For**: FR-021, FR-030

---

### T024: Write Integration Test Suite for US3 (CLI Commands) [US3-CLI]

**Objective**: Comprehensive test suite for CLI commands covering all scenarios.

**File Paths**:
- `backend/tests/integration/us3_cli_commands_test.go`

**Test Cases**:
- **migrate up** command applies migrations and exits 0
- **migrate up --service** specifies single service
- **migrate down** rolls back and exits 0
- **migrate status** displays pending/applied counts
- **migrate status --format=json** outputs valid JSON
- Error cases: invalid environment, missing service, CLI parsing errors

**Test Task** [Already inherent]:
- Full test coverage for CLI scenarios

**Responsible For**: SC-007, SC-008

---

### T025: Integrate OpenTelemetry Tracing [US3-CLI]

**Objective**: Add OTel trace spans to all migration operations for observability.

**File Paths**:
- `backend/internals/migrations/services/migration_service.go` (wrap operations with trace spans)
- `backend/cmd/migrations/cmd.go` (trace CLI commands)

**Acceptance Criteria**:
- ✅ Parent span: entire migrate up/down operation
- ✅ Child spans: per-service execution, per-migration execution, database operations
- ✅ Span attributes: service name, migration version, status, error details
- ✅ Uses existing OTel integration from `backend/pkgs/otel/`

**Test Task**: Validate via trace output inspection during integration tests

**Responsible For**: FR-030 (observability)

---

### T026: Implement Graceful Shutdown for Migration Service [US3-CLI]

**Objective**: Add signal handling (SIGINT, SIGTERM) to allow graceful shutdown during migrations.

**File Paths**:
- `backend/cmd/migrations/container.go` (implement graceful shutdown)

**Acceptance Criteria**:
- ✅ On SIGINT/SIGTERM: wait for current migration to complete, then exit
- ✅ Do not abort migration mid-execution
- ✅ Log shutdown signal: "migrations: received SIGTERM, waiting for in-flight migration to complete"
- ✅ Exit code 0 on normal completion, 1 on interrupt

**Test Task**: Manual verification with Ctrl+C during migration

**Responsible For**: Architecture rule on graceful shutdown

---

### T027: Create CLI Help Documentation [US3-CLI]

**Objective**: Add comprehensive help text to all CLI commands.

**File Paths**:
- `backend/cmd/migrations/cmd.go` (update all `Use`, `Short`, `Long` fields)

**Acceptance Criteria**:
- ✅ `migrate up --help` shows: description, usage, flag explanations, examples
- ✅ `migrate down --help` shows: targets, version numbering, force-rollback usage
- ✅ `migrate status --help` shows: output formats, per-service filtering
- ✅ Help text is complete and user-friendly

**Test Task**: Manual: `./backend migrate up --help` is readable and helpful

**Responsible For**: User experience (documentation)

---

## Phase 5: US2 — Migrate DML Data Changes (Priority: P1)

**[P] This phase runs parallel with US3-CLI after US1 completes. Can be executed by a second developer/team.**

### [X] T028: Implement DML Migration Logic [US2-DML]

**Objective**: Implement DML execution flow with environment filtering.

**File Paths**:
- `backend/internals/migrations/services/migration_service.go` (implement MigrateUpDML method)
- `backend/internals/migrations/services/migration_service_test.go`

**Key Function**:
- `(ms *MigrationService) MigrateUpDML(ctx context.Context, serviceName string, env string) error`

**Acceptance Criteria**:
- ✅ Precondition: ALL DDL migrations must have completed successfully before DML starts
  - If DDL pending: return error "DDL migrations must complete first"
- ✅ For each service, scans `migrations/dml/<env>/` folder
- ✅ Discovers `.up.sql` files sorted by version
- ✅ For each migration:
  1. Checks if applied (IsDMLMigrationApplied with version + environment)
  2. Executes in transaction
  3. Records in `migrations_dml` with (version, environment) tuple
- ✅ Other environment's migrations are SKIPPED (e.g., if env=local, dml/dev/ is ignored)
- ✅ On error: rolls back, does not record, logs detailed message, stops

**Test Task** [Test-First]:
- Write integration test: `backend/tests/integration/t028_dml_orchestration_test.go`
  - Test 28.1: Execute DML for 'local' environment, only dml/local/ migrations run
  - Test 28.2: Execute DML for 'dev' environment, only dml/dev/ migrations run
  - Test 28.3: Precondition check - returns error if DDL not complete
  - Test 28.4: DML migrations execute after DDL (via MigrateUp which orchestrates both)
  - Test 28.5: Same version in different environments are tracked independently

**Responsible For**: FR-012, FR-013, FR-015, FR-024, FR-028

---

### [X] T029: Create DML Migrations for Bills Service [US2-DML]

**Objective**: Create environment-specific DML seed migrations for bills service.

**File Paths**:
- `backend/internals/bills/migrations/dml/local/000001_seed_bill_types.up.sql`
- `backend/internals/bills/migrations/dml/local/000001_seed_bill_types.down.sql`
- `backend/internals/bills/migrations/dml/dev/000001_seed_bill_types.up.sql`
- `backend/internals/bills/migrations/dml/dev/000001_seed_bill_types.down.sql`
- `backend/internals/bills/migrations/dml/stg/000001_seed_production_defaults.up.sql`
- `backend/internals/bills/migrations/dml/stg/000001_seed_production_defaults.down.sql`
- (prd migrations: minimal, only critical reference data)

**Acceptance Criteria**:
- ✅ `local/` migrations seed bill types for development
- ✅ `dev/` migrations seed bill types (can be identical to local or include shared fixtures)
- ✅ `stg/` migrations seed minimal production-like fixtures
- ✅ All INSERT statements are idempotent (use ON CONFLICT DO NOTHING)
- ✅ All .down.sql files use DELETE WHERE to remove seeded data cleanly

**Test Task**: Validated via T028 integration tests

**Responsible For**: FR-007, FR-008, FR-013, FR-028, SC-010

---

### [X] T030: Create DML Migrations for Files Service [US2-DML]

**Objective**: Create environment-specific DML seed migrations for files service.

**File Paths**:
- `backend/internals/files/migrations/dml/local/000001_seed_default_documents.up.sql`
- `backend/internals/files/migrations/dml/local/000001_seed_default_documents.down.sql`
- `backend/internals/files/migrations/dml/dev/000001_seed_test_documents.up.sql`
- `backend/internals/files/migrations/dml/dev/000001_seed_test_documents.down.sql`
- etc.

**Acceptance Criteria**: Same as T029

**Test Task**: Validated via T028 integration tests

**Responsible For**: FR-007, FR-008, FR-013, FR-028, SC-010

---

### [X] T031: Create DML Migrations for Identity Service [US2-DML]

**Objective**: Create environment-specific DML seed migrations for identity service (users, roles).

**File Paths**:
- `backend/internals/identity/migrations/dml/local/000001_seed_default_user.up.sql`
- `backend/internals/identity/migrations/dml/local/000001_seed_default_user.down.sql`
- `backend/internals/identity/migrations/dml/dev/000001_seed_test_users.up.sql`
- `backend/internals/identity/migrations/dml/dev/000001_seed_test_users.down.sql`
- etc.

**Acceptance Criteria**: Same as T029

**Test Task**: Validated via T028 integration tests

**Responsible For**: FR-007, FR-008, FR-013, FR-028, SC-010

---

### [X] T032: Create DML Migrations for Onboarding Service [US2-DML]

**Objective**: Create environment-specific DML seed migrations for onboarding service (projects, registrations).

**File Paths**:
- `backend/internals/onboarding/migrations/dml/local/000001_seed_default_project.up.sql`
- `backend/internals/onboarding/migrations/dml/local/000001_seed_default_project.down.sql`
- etc.

**Acceptance Criteria**: Same as T029

**Test Task**: Validated via T028 integration tests

**Responsible For**: FR-007, FR-008, FR-013, FR-028, SC-010

---

### [X] T033: Create DML Migrations for Payments Service [US2-DML]

**Objective**: Create environment-specific DML seed migrations for payments service.

**File Paths**:
- `backend/internals/payments/migrations/dml/local/000001_seed_transaction_types.up.sql`
- `backend/internals/payments/migrations/dml/local/000001_seed_transaction_types.down.sql`
- etc.

**Acceptance Criteria**: Same as T029

**Test Task**: Validated via T028 integration tests

**Responsible For**: FR-007, FR-008, FR-013, FR-028, SC-010

---

### T034: Write Integration Test Suite for US2 (DML) [US2-DML]

**Objective**: Comprehensive test suite for DML user story.

**File Paths**:
- `backend/tests/integration/us2_dml_migrations_test.go`

**Test Cases** (mapping to spec scenarios):
- **Scenario 2.1**: DML migrations execute after DDL
- **Scenario 2.2**: Environment-specific DML (local seeds from local/ folder)
- **Scenario 2.3**: Different environments have different DML (dev vs stg)
- **Scenario 2.4**: DML records tracked in migrations_dml independently per environment
- **Scenario 2.5**: Failed DML rolls back, no record
- **Scenario 2.6**: Same version in different environments tracked separately
- **Scenario 2.7**: DML rollback via .down.sql
- **Scenario 2.8**: Precondition: DDL must complete before DML

**Acceptance Criteria**:
- ✅ All 8 scenarios covered
- ✅ Tests verify data was seeded correctly (query and verify presence)
- ✅ Tests verify absent in wrong environments (e.g., prd seeds not in local)

**Test Task** [Already inherent]:
- Full test coverage for US2

**Responsible For**: SC-003, SC-004, SC-010

---

### T035: Performance Test: 30-Second SLA [US2-DML]

**Objective**: Validate that migrate up (all services, DDL+DML) completes in under 30 seconds.

**File Paths**:
- `backend/tests/integration/t035_performance_test.go`

**Test Case**:
- Start time, run `MigrateUp` for all services (local environment), measure elapsed time
- Assert elapsed time < 30 seconds
- Log duration breakdown per service if tool allows

**Acceptance Criteria**:
- ✅ Total execution time from start to finish < 30 seconds
- ✅ Test passes consistently (deterministic)

**Test Task** [Already inherent]:

**Responsible For**: SC-008

---

## Phase 6: US4 — Folder Structure Standardization and Discovery (Priority: P0)

**[P] This phase runs parallel with US2 and US3 after Foundational completes. Can be executed concurrently.**

### [X] T036: Validate and Standardize Bills Migrations Folder [US4-Folder]

**Objective**: Ensure bills service migrations folder follows standardized structure.

**File Paths**:
- `backend/internals/bills/migrations/ddl/` (validated as part of T011-T015)
- `backend/internals/bills/migrations/dml/local/`
- `backend/internals/bills/migrations/dml/dev/`
- `backend/internals/bills/migrations/dml/stg/`
- `backend/internals/bills/migrations/dml/prd/`

**Acceptance Criteria**:
- ✅ Folder structure matches standard: `migrations/ddl/` and `migrations/dml/<env>/` for all environments
- ✅ All migration files follow naming convention: `<NNNNNN>_<slug>.{up,down}.sql`
- ✅ Every `.up.sql` has matching `.down.sql`
- ✅ No extraneous files or folders outside standard structure
- ✅ Discovery script can locate all migrations without errors

**Test Task**: Part of T037 discovery validation test

**Responsible For**: FR-006, FR-009

---

### [X] T037: Implement Folder Structure Validation Tool [US4-Folder]

**Objective**: Create a validation command to check that all services follow the standardized folder structure.

**File Paths**:
- `backend/cmd/migrations/validate_folders.go`
- `backend/cmd/migrations/validate_folders_test.go`

**Key Function**:
- `ValidateFolderStructure() error` — scans all services, verifies compliance

**Acceptance Criteria**:
- ✅ Checks all six services exist with appropriate migration folders
- ✅ Validates file naming convention (regex check)
- ✅ Ensures .up.sql and .down.sql pairing
- ✅ Returns error list if any issues found
- ✅ Optionally: auto-creates missing folders with `--auto-create` flag

**Test Task** [Test-First]:
- Write unit test: `backend/cmd/migrations/validate_folders_test.go`
  - Test 37.1: Valid structure passes validation
  - Test 37.2: Errors on unpaired migrations (up without down)
  - Test 37.3: Errors on invalid filename format
  - Test 37.4: Errors on missing expected folders

**Responsible For**: FR-009

---

### [X] T038: Create Makefile Target for Folder Validation [US4-Folder]

**Objective**: Add `make migrate/validate` target to run folder structure check.

**File Paths**:
- `Makefile`

**Acceptance Criteria**:
- ✅ `make migrate/validate` runs validation
- ✅ Exits 0 if all folders valid, 1 if issues found
- ✅ Outputs clear error messages for any problems

**Test Task**: Manual verification

**Responsible For**: FR-009

---

### T039: Write Integration Test Suite for US4 (Folder Structure) [US4-Folder]

**Objective**: Comprehensive test for folder structure discovery and standardization.

**File Paths**:
- `backend/tests/integration/us4_folder_structure_test.go`

**Test Cases** (mapping to spec scenarios):
- **Scenario 4.1**: Migrations discovered from standardized folders
- **Scenario 4.2**: Standard folder structure verified for all services
- **Scenario 4.3**: Files named with sequential numeric prefix
- **Scenario 4.4**: DDL migrations execute before DML (folder order)
- **Scenario 4.5**: Environment subfolders correctly scanned
- **Scenario 4.6**: Missing folders handled gracefully

**Acceptance Criteria**:
- ✅ All 6 scenarios covered
- ✅ Validates discovery logic identifies all services without errors

**Test Task** [Already inherent]:

**Responsible For**: SC-009

---

## Phase 7: Polish & Validation

### T040: End-to-End Integration Test (Full Feature) [Polish]

**Objective**: Complete feature test covering all user stories in a realistic workflow.

**File Paths**:
- `backend/tests/integration/e2e_complete_migration_test.go`

**Workflow**:
1. Start with clean database
2. Run `migrate up` (all services, all environments)
3. Verify all DDL applied to correct schemas
4. Verify all DML applied per environment
5. Run `migrate status` and verify counts
6. Run `migrate down` for one migration
7. Verify schema reverted and tracking records updated
8. Run `migrate up` again to re-apply
9. Verify idempotency (no duplicates)

**Acceptance Criteria**:
- ✅ All steps complete without errors
- ✅ Database state matches expected after each step
- ✅ All 32 FR and 10 SC validated through workflow

**Test Task** [Already inherent]:

**Responsible For**: All FRs and SCs

---

### T041: Documentation: Migration Guide & Examples [Polish]

**Objective**: Create user-facing documentation for using the migration system.

**File Paths**:
- `backend/docs/MIGRATION_SYSTEM.md` (or README update)

**Content**:
- Overview of migration system architecture
- DDL vs DML separation explained
- Folder structure guide
- CLI command reference (up, down, status)
- Environment setup (local, dev, stg, prd)
- Production safety (two-factor approval)
- Troubleshooting guide
- Examples: creating a new migration, rolling back, debugging

**Acceptance Criteria**:
- ✅ Complete and clear
- ✅ No ambiguities
- ✅ Examples are runnable/accurate

**Test Task**: Manual review by team

**Responsible For**: Knowledge transfer

---

### T042: Code Review & Cleanup [Polish]

**Objective**: Review all implementation against requirements, quality gates, and conventions.

**Checklist**:
- ✅ All code passes `golangci-lint` with no warnings
- ✅ All code passes `gosec` security scan
- ✅ All tests pass (unit + integration)
- ✅ Code follows Go idioms and project conventions (per `.github/instructions/`)
- ✅ All exported functions have doc comments
- ✅ Error handling is complete (no ignored errors)
- ✅ Logging is structured and contextual
- ✅ No hardcoded values (use config/env vars)
- ✅ All FRs (FR-001 to FR-032) mapped to specific code/tests
- ✅ All SCs (SC-001 to SC-010) validated by tests

**Acceptance Criteria**:
- ✅ Zero lint errors
- ✅ Zero security issues
- ✅ 100% test pass rate
- ✅ Code style consistent with project

**Test Task**: Automated via CI pipeline

**Responsible For**: Quality assurance

---

## Coverage Matrix: Functional Requirements → Tasks

| FR# | Title | Mapped to Task(s) | Status |
|-----|-------|-------------------|--------|
| FR-001 | golang-migrate/migrate library integration | T001, T004 | ✅ |
| FR-002 | migrations_ddl table creation | T002, T007 | ✅ |
| FR-003 | migrations_ddl table columns | T002, T007 | ✅ |
| FR-004 | migrations_dml table columns | T002, T007 | ✅ |
| FR-005 | Idempotent execution | T004, T005, T010 | ✅ |
| FR-005a | Per-service schema isolation | T010, T016 | ✅ |
| FR-005b | Independent migrations_ddl/dml per schema | T002, T010, T028 | ✅ |
| FR-005c | Schema creation if missing | T010 | ✅ |
| FR-006 | Folder structure: ddl/, dml/<env>/ | T011-T035, T036 | ✅ |
| FR-007 | File naming: <NNNNNN>_<slug>.{up,down}.sql | T005, T037 | ✅ |
| FR-008 | DML subfolder per environment | T029-T033 | ✅ |
| FR-009 | Auto-discovery without per-service config | T005, T037 | ✅ |
| FR-010 | Valid PostgreSQL SQL statements | T011-T015, T029-T033 | ✅ |
| FR-011 | DDL before DML execution | T010, T028 | ✅ |
| FR-012 | DML after DDL completes | T028 | ✅ |
| FR-013 | Only active environment DML | T028 | ✅ |
| FR-014 | Error: stop on DDL failure | T010, T016 | ✅ |
| FR-015 | Error: stop on DML failure | T028 | ✅ |
| FR-016 | migrate up command | T019, T022 | ✅ |
| FR-017 | migrate down command | T020, T022 | ✅ |
| FR-018 | migrate status command | T021, T022 | ✅ |
| FR-019 | --env flag | T019, T020 | ✅ |
| FR-020 | --service flag & all-services | T019, T020 | ✅ |
| FR-021 | Clear timestamped log output | T023, T027 | ✅ |
| FR-022 | Exit codes: 0 success, non-zero failure | T019-T021 | ✅ |
| FR-023 | Error message with context | T023 | ✅ |
| FR-024 | Failed migration NOT recorded | T010, T016, T028 | ✅ |
| FR-024a | Rollback failure: record unchanged, --force-rollback | T020, T026 | ✅ |
| FR-025 | Manual rollback support | T016, T020 | ✅ |
| FR-026 | Rollback in reverse order | T016 | ✅ |
| FR-026a | Rollback failure: --force-rollback flag | T020 | ✅ |
| FR-027 | Environment variable priority (APP_ENV → ENVIRONMENT → local) | T006, T008 | ✅ |
| FR-028 | Environment-specific DML | T029-T033 | ✅ |
| FR-029 | Production safety: APP_ENV=prd + --approve-production | T006, T019 | ✅ |
| FR-030 | Structured logging with timestamps | T023, T025 | ✅ |
| FR-031 | Audit log (migrations_ddl/dml tables) | T002, T007 | ✅ |
| FR-032 | Queryable migration history | T017 | ✅ |

---

## Coverage Matrix: Success Criteria → Tasks

| SC# | Title | Verified By Task(s) | Status |
|-----|-------|---------------------|--------|
| SC-001 | DDL migrations execute in order, schema matches expected state | T018 (Scenario 1.1), T040 | ✅ |
| SC-002 | DDL not re-executed, migrations_ddl prevents duplicates | T018 (Scenario 1.4), T040 | ✅ |
| SC-003 | DML executes only after all DDL complete | T034 (Scenario 2.1), T040 | ✅ |
| SC-004 | Environment-specific DML (local, dev, stg, prd) executes correctly | T034 (Scenario 2.2, 2.3), T040 | ✅ |
| SC-005 | Failure → atomic rollback + clear error logging | T018 (Scenario 1.5), T040 | ✅ |
| SC-006 | Rollback correctly reverts schema via .down.sql | T018 (Scenario 1.6, 1.7), T040 | ✅ |
| SC-007 | migrate status reports pending and applied versions accurately | T024 | ✅ |
| SC-008 | All migrations complete in < 30 seconds | T035 | ✅ |
| SC-009 | Standardized folder structure auto-discovered without errors | T039, T040 | ✅ |
| SC-010 | Seeded data present in local/dev, absent in prd | T034 (Scenario 2.6), T040 | ✅ |

---

## Task Dependencies & Sequential Ordering

```
Setup
├─ T001: Add golang-migrate/migrate dependency
├─ T002: Create migration tracking tables
├─ T003: Setup Migration Service Go module
│
└─> Foundational
    ├─ T004: Wrap golang-migrate library
    ├─ T005: Discovery algorithm
    ├─ T006: Production safety check
    ├─ T007: Migration record tracking
    ├─ T008: Environment variable reading
    ├─ T009: Cobra CLI scaffold
    │
    └─> US1-DDL (CRITICAL PATH)
        ├─ T010: DDL orchestration logic (core)
        ├─ T011: Bills DDL migrations
        ├─ T012: Files DDL migrations
        ├─ T013: Identity DDL migrations
        ├─ T014: Onboarding DDL migrations
        ├─ T015: Payments DDL migrations
        ├─ T016: Rollback logic for DDL
        ├─ T017: Status command
        └─ T018: US1 integration tests ← **MVP CHECKPOINT**
        
        ├─> [P] US3-CLI (parallel with US2 after MVP)
        │   ├─ T019: migrate up command
        │   ├─ T020: migrate down command
        │   ├─ T021: migrate status command
        │   ├─ T022: Makefile targets
        │   ├─ T023: Logging & structured output
        │   ├─ T024: US3 integration tests
        │   ├─ T025: OpenTelemetry tracing
        │   ├─ T026: Graceful shutdown
        │   └─ T027: CLI help documentation
        │
        ├─> [P] US2-DML (parallel with US3 after MVP)
        │   ├─ T028: DML orchestration logic
        │   ├─ T029: Bills DML migrations (local/dev/stg/prd)
        │   ├─ T030: Files DML migrations
        │   ├─ T031: Identity DML migrations
        │   ├─ T032: Onboarding DML migrations
        │   ├─ T033: Payments DML migrations
        │   ├─ T034: US2 integration tests
        │   └─ T035: Performance test (30s SLA)
        │
        └─> [P] US4-Folder (parallel with US2/US3)
            ├─ T036: Validate Bills folder structure
            ├─ T037: Folder structure validation tool
            ├─ T038: Makefile validate target
            └─ T039: US4 integration tests
        
        └─> Polish (after US2/US3/US4 complete)
            ├─ T040: E2E integration test
            ├─ T041: Documentation
            └─ T042: Code review & cleanup
```

---

## MVP Checkpoint: After T018

**Milestone**: US1 (DDL migrations) complete and tested

**What Works**:
- ✅ All DDL migrations apply to correct service schemas
- ✅ Schema state verified after migrations
- ✅ Idempotent execution proven
- ✅ Rollback functional
- ✅ Error handling with transaction rollback

**What NOT Yet Available**:
- ❌ DML environment-specific migrations (US2)
- ❌ CLI commands for user interaction (US3)
- ❌ Documentation and observability
- ❌ Performance optimization

**Testing to Proceed**: 
- All US1 integration tests pass (T018)
- Database schemas match expected state
- migrations_ddl tables populated correctly

**To Re-Plan**: If T018 has failures, address root causes before proceeding to US2/US3/US4

---

## Risk Mitigation & Rollback Strategy

### Risk 1: golang-migrate/migrate Integration Issues
- **Mitigation**: Early integration test (T004)
- **Rollback**: Revert to previous migration approach (if any); add compatibility layer

### Risk 2: Performance SLA Miss (30 seconds)
- **Mitigation**: Performance test (T035) identifies bottlenecks early
- **Rollback**: Optimize hot paths (batch inserts, parallel service processing)

### Risk 3: Production Safety Two-Factor Bypass
- **Mitigation**: Comprehensive test coverage (T019, T020)
- **Rollback**: Add mandatory code review for production approvals

### Risk 4: Data Loss from Incorrect Rollback
- **Mitigation**: Extensive rollback tests (T016, T018 Scenario 1.6-1.7)
- **Rollback**: Maintain database backups; require dry-run rollback in staging

---

## Success Metrics

| Metric | Target | Validation |
|--------|--------|-----------|
| All 42 tasks completed | 100% | Task checklist |
| All FR (FR-001 to FR-032) implemented | 100% | Code coverage matrix |
| All SC (SC-001 to SC-010) verified | 100% | Integration tests |
| Zero lint errors | 0 | CI pipeline golangci-lint |
| Zero security issues | 0 | gosec scan |
| Test pass rate | 100% | CI pipeline |
| Migration execution time < 30s | ✅ | T035 performance test |
| Code review approval | ✅ | T042 code review |

---

## Timeline Estimate

- **Phase 1 (Setup)**: 2 hours (sequential)
- **Phase 2 (Foundational)**: 4 hours (sequential)
- **Phase 3 (US1-DDL)**: 6 hours (sequential)
- **Checkpoint**: T018 validates MVP readiness
- **Phase 4 (US3-CLI)**: 5 hours [parallel with Phase 5]
- **Phase 5 (US2-DML)**: 5 hours [parallel with Phase 4]
- **Phase 6 (US4-Folder)**: 2 hours [parallel with Phase 4-5]
- **Phase 7 (Polish)**: 3 hours (after all phases)
- **Total Critical Path**: ~22 hours (Setup → Foundational → US1 → [US2||US3||US4] → Polish)

---

## Go Ahead Checklist

### Before Starting Phase 1
- ✅ All specification artifacts reviewed (spec.md, plan.md, research.md, data-model.md)
- ✅ All clarifications resolved and documented in session memory
- ✅ Team understands 6 phases and parallel execution opportunities
- ✅ Database connectivity verified (integration test environment ready)
- ✅ Go toolchain available (go 1.19+)

### Before Proceeding to Phase 4 (US3-CLI) & Phase 5 (US2-DML)
- ✅ T018 integration tests 100% passing
- ✅ All T010-T018 tasks code-reviewed and approved
- ✅ MVP checkpoint criteria met

### Before Phase 7 (Polish)
- ✅ All US2, US3, US4 tasks completed
- ✅ All integration tests passing
- ✅ Performance test (T035) in acceptable range

---

## Notes for Developers

1. **Test-First Approach**: Each task with [Test-First] marker requires writing tests BEFORE implementation. This ensures acceptance criteria are clear and testable.

2. **Parallelization**: Phases 4, 5, 6 can be distributed across team members after US1 completion. Coordinate via Git branches.

3. **File Paths**: All file paths in tasks are absolute within the monorepo root (`backend/`, `specs/`, etc.). Use these as templates—adjust based on actual project structure.

4. **Quality Gates**: All code must pass `golangci-lint` and `gosec` before PR submission. This is non-negotiable.

5. **Integration Tests**: Use a temporary database (Docker or local PostgreSQL) for integration tests. Do NOT use production database. See `testmain_test.go` for setup pattern.

6. **Rollback Failures**: The `--force-rollback` flag is a recovery mechanism. Use sparingly and only after investigating root cause. Always have a database backup available.

7. **Documentation**: T041 is critical for team onboarding. Make it comprehensive and keep it updated as implementation details emerge.

---

## Questions for Clarification During Implementation

If questions arise during tasks, follow this precedence:
1. Refer to spec.md, plan.md, research.md, data-model.md for clarification
2. Check project conventions in `.github/instructions/`
3. Ask tech lead or project owner (in order of availability)

---

**End of Tasks Document — Ready for Implementation**

Status: ✅ **APPROVED FOR DEVELOPMENT**
Date Generated: 2026-04-01
Feature: 003-backend-migration-system
