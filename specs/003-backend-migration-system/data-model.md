# Data Model: Backend Migration System

**Feature**: 003-backend-migration-system  
**Created**: 2026-04-01  
**Status**: Design Phase  
**Architecture Alignment**: Per-service schema isolation (bills, files, identity, onboarding, payments)

## System Overview

The migration tracking model enforces per-service schema isolation with dedicated `migrations_ddl` and `migrations_dml` tables in each PostgreSQL schema. This architecture ensures:

- **Service Boundary Enforcement**: Each service's migrations and audit history are isolated within its schema
- **Multi-Environment Support**: DML migrations are environment-aware; same version can execute differently per environment
- **Idempotent Execution**: Tracking tables prevent re-execution of already-applied migrations
- **Audit Trail**: Complete history of who executed what, when, and how long it took
- **Safe Rollback**: Recording rollbacks enables recovery from failures via `--force-rollback`

---

## Migration Tracking Tables

### migrations_ddl Table

Tracks all applied Data Definition Language (DDL) schema migrations. Created once per service schema (bills.migrations_ddl, files.migrations_ddl, etc.).

```sql
CREATE TABLE IF NOT EXISTS migrations_ddl (
  version BIGINT PRIMARY KEY,
  name VARCHAR(255) NOT NULL UNIQUE,
  description TEXT,
  executed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  execution_time_ms BIGINT NOT NULL,
  success BOOLEAN NOT NULL DEFAULT TRUE,
  error_message TEXT,
  executed_by TEXT DEFAULT CURRENT_USER,
  checksum TEXT,
  sql_snippet TEXT
);

CREATE INDEX IF NOT EXISTS idx_migrations_ddl_executed_at 
  ON migrations_ddl(executed_at DESC);
CREATE INDEX IF NOT EXISTS idx_migrations_ddl_success 
  ON migrations_ddl(success, executed_at DESC);
```

#### Columns

| Column | Type | Purpose | Notes |
|--------|------|---------|-------|
| version | BIGINT | Unique migration version (primary key) | 6-digit padded number (1, 2, 3, ..., 999999). Per-service uniqueness. |
| name | VARCHAR(255) | Migration name from filename | e.g., "create_bills_table", "add_status_index". UNIQUE ensures no duplicates. |
| description | TEXT | Human-readable description | Extracted from filename slug or manually added for clarity |
| executed_at | TIMESTAMPTZ | Timestamp when migration applied | Auto-set via DEFAULT NOW(). Enables sorting and timeline queries. |
| execution_time_ms | BIGINT | Duration in milliseconds | Tracks performance; slow migrations (>500ms) flagged for optimization |
| success | BOOLEAN | Whether migration succeeded | TRUE for applied migrations. FALSE only if rolled back and not reapplied. |
| error_message | TEXT | Error context if execution failed | Populated during failure; helps debugging. Only for failed executions. |
| executed_by | TEXT | Operator or CI system that applied it | Defaults to database role (CURRENT_USER). For audit trail. |
| checksum | TEXT | SHA256 hash of SQL file | Optional; used to detect if migration file edited after application |
| sql_snippet | TEXT | First 500 characters of SQL | Reference for incident response; avoids reading files |

#### Sample Data (bills schema)

```sql
INSERT INTO bills.migrations_ddl 
  (version, name, executed_at, execution_time_ms, success, executed_by)
VALUES
  (1, 'create_bills_table', NOW(), 45, TRUE, 'postgres'),
  (2, 'add_status_index', NOW(), 12, TRUE, 'ci-system'),
  (3, 'create_bill_types_enum', NOW(), 8, TRUE, 'postgres');
```

---

### migrations_dml Table

Tracks all applied Data Manipulation Language (DML) data migrations per environment. Created once per service schema with composite primary key allowing same version in different environments.

```sql
CREATE TABLE IF NOT EXISTS migrations_dml (
  version BIGINT NOT NULL,
  environment VARCHAR(50) NOT NULL,
  name VARCHAR(255) NOT NULL,
  executed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  execution_time_ms BIGINT NOT NULL,
  success BOOLEAN NOT NULL DEFAULT TRUE,
  error_message TEXT,
  executed_by TEXT DEFAULT CURRENT_USER,
  checksum TEXT,
  sql_snippet TEXT,
  PRIMARY KEY (version, environment),
  UNIQUE(name, environment, version)
);

CREATE INDEX IF NOT EXISTS idx_migrations_dml_env_executed 
  ON migrations_dml(environment, executed_at DESC);
CREATE INDEX IF NOT EXISTS idx_migrations_dml_version_desc 
  ON migrations_dml(version DESC, environment);
```

#### Columns

| Column | Type | Purpose | Notes |
|--------|------|---------|-------|
| version | BIGINT | Migration version (composite key part 1) | Per-environment versioning; version=1 can exist in both 'local' and 'dev' |
| environment | VARCHAR(50) | Target environment (composite key part 2) | Values: 'local', 'dev', 'stg', 'prd'. Composite key ensures isolated tracking per environment. |
| name | VARCHAR(255) | Migration name from filename | e.g., "seed_bill_types", "seed_test_users" |
| executed_at | TIMESTAMPTZ | Timestamp when applied | Auto-set via DEFAULT NOW() |
| execution_time_ms | BIGINT | Duration in milliseconds | For performance monitoring |
| success | BOOLEAN | Whether migration succeeded | TRUE for applied; FALSE for rolled back |
| error_message | TEXT | Error context if failed | Debugging aid |
| executed_by | TEXT | Operator or CI system | Audit trail |
| checksum | TEXT | SHA256 hash of SQL file | For integrity verification |
| sql_snippet | TEXT | First 500 chars of SQL | Reference for review |

#### Sample Data (bills schema)

```sql
INSERT INTO bills.migrations_dml 
  (version, environment, name, executed_at, execution_time_ms, success, executed_by)
VALUES
  (1, 'local', 'seed_bill_types', NOW(), 45, TRUE, 'postgres'),
  (2, 'local', 'seed_test_data', NOW(), 120, TRUE, 'developer'),
  (1, 'dev', 'seed_bill_types', NOW(), 50, TRUE, 'ci-system'),
  (1, 'stg', 'seed_staging_fixtures', NOW(), 85, TRUE, 'ci-system');
  -- Note: version=1 appears in multiple environments with isolated tracking
```

---

## Migration File Structure

### DDL Migration File

**Filename Convention**: `<NNNNNN>_<slug>.{up,down}.sql`

**Example**: `000001_create_documents_table.up.sql`

```sql
-- Create documents table with all required columns
CREATE TABLE IF NOT EXISTS documents (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  project_id UUID NOT NULL,
  filename VARCHAR(255) NOT NULL,
  file_hash VARCHAR(64) UNIQUE,
  classification VARCHAR(50) CHECK (classification IN ('bill', 'statement')),
  bill_type VARCHAR(100),
  analysis_status VARCHAR(50) DEFAULT 'pending' CHECK (analysis_status IN ('pending', 'analysed', 'failed')),
  uploaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
  UNIQUE(project_id, file_hash)
);

-- Create indexes for common queries
CREATE INDEX IF NOT EXISTS idx_documents_project_id ON documents(project_id);
CREATE INDEX IF NOT EXISTS idx_documents_analysis_status ON documents(analysis_status);
CREATE INDEX IF NOT EXISTS idx_documents_uploaded_at ON documents(uploaded_at DESC);
```

**Corresponding Rollback**: `000001_create_documents_table.down.sql`

```sql
-- Drop the documents table and its indexes (indexes dropped automatically)
DROP TABLE IF EXISTS documents;
```

---

### DML Migration File (Environment-Specific)

**Filename Convention**: Same as DDL; stored in `migrations/dml/<environment>/`

**Example**: `000001_seed_default_user.up.sql` (in `migrations/dml/local/`)

```sql
-- Seed default user for local development
INSERT INTO users (id, username, email, password_hash, created_at)
VALUES (
  'user-default-local',
  'demo',
  'demo@localhost',
  '$2a$10$encrypted_hash_here', -- bcrypt hash of 'demo123'
  CURRENT_TIMESTAMP
)
ON CONFLICT (username) DO NOTHING; -- Idempotent: skip if user exists

-- Seed default project
INSERT INTO projects (id, owner_id, name, type, created_at)
VALUES (
  'project-default-local',
  'user-default-local',
  'Personal Finance',
  'personal',
  CURRENT_TIMESTAMP
)
ON CONFLICT (owner_id, name) DO NOTHING;
```

**Corresponding Rollback**: `000001_seed_default_user.down.sql`

```sql
-- Delete seeded data
DELETE FROM project_members WHERE project_id = 'project-default-local';
DELETE FROM projects WHERE id = 'project-default-local';
DELETE FROM users WHERE id = 'user-default-local';
```

---

## Migration State Transitions

### DDL Migration Lifecycle

```
Pending (not yet applied)
   ↓
Applying (in-progress)
   ↓↙ (error)
Failed (rolled back, dirty = TRUE)
   ↓
Applied (succeeded, inserted in migrations_ddl, success = TRUE)
   ↓
Rollback Pending (user requested)
   ↓
Rolling Back (in-progress)
   ↓
Rolled Back (down.sql executed, dirty = TRUE until next up)
```

### DML Migration Lifecycle

Same as DDL, but tracked per environment:
- `(version=1, environment='local')` can be rolled back independently
- `(version=1, environment='dev')` is separate
- Both can be rolled back without affecting the other

---

## Query Examples

### Check Applied Migrations

```sql
-- All DDL migrations in order
SELECT version, name, executed_at, execution_time_ms, success
FROM migrations_ddl
ORDER BY version ASC;

-- DML migrations for specific environment
SELECT version, name, executed_at, execution_time_ms, success
FROM migrations_dml
WHERE environment = 'local'
ORDER BY version ASC;

-- All pending migrations (not yet applied)
SELECT * FROM migrations_ddl
WHERE version NOT IN (SELECT MAX(version) FROM migrations_ddl)
ORDER BY version DESC
LIMIT 1;
```

### Check Migration Status

```sql
-- Total DDL migrations applied
SELECT COUNT(*) as total_ddl FROM migrations_ddl WHERE success = TRUE;

-- Total DML migrations for dev
SELECT COUNT(*) as total_dml_dev FROM migrations_dml WHERE environment = 'dev' AND success = TRUE;

-- Migrations in dirty state (need manual intervention)
SELECT * FROM migrations_ddl WHERE dirty = TRUE
UNION ALL
SELECT version, environment,dirty, name, executed_at, execution_time_ms, success FROM migrations_dml WHERE dirty = TRUE;
```

### Performance Metrics

```sql
-- Slowest migrations (by execution time)
SELECT name, execution_time_ms FROM migrations_ddl
ORDER BY execution_time_ms DESC
LIMIT 10;

-- Migration timeline
SELECT executed_at, COUNT(*) as migrations_count, SUM(execution_time_ms) as total_ms
FROM migrations_ddl
GROUP BY DATE(executed_at)
ORDER BY executed_at DESC;
```

---

## Execution Model

### Per-Service Schema Isolation

Each backend service owns its own PostgreSQL schema:

```
PostgreSQL Instance
├── bills (schema)
│   ├── migrations_ddl, migrations_dml
│   ├── bills, bill_records, bill_types, idempotency_keys
│   └── (all bill-related tables)
├── files (schema)
│   ├── migrations_ddl, migrations_dml
│   ├── documents, document_metadata, file_events
│   └── (all file-related tables)
├── identity (schema)
│   ├── migrations_ddl, migrations_dml
│   ├── users, roles, permissions, jwks_cache
│   └── (all identity-related tables)
├── onboarding (schema)
│   ├── migrations_ddl, migrations_dml
│   ├── registrations, projects, team_members
│   └── (all onboarding-related tables)
├── payments (schema)
│   ├── migrations_ddl, migrations_dml
│   ├── transactions, reconciliations, payment_methods
│   └── (all payment-related tables)
└── public (schema) — RESERVED
    └── (no application tables; cluster-wide objects only)
```

**Benefits**:
- Service boundaries enforced at DB layer
- RBAC: Role can be granted access to specific schema only
- Isolation: One service's migrations don't affect another's
- Catastrophic recovery: Can drop entire schema if needed
- Cross-schema queries forbidden at application layer (prevents accidental coupling)

### Multi-Service Execution Order

When `make migrate/up` is called without specifying a service, migrations execute in **lexicographic (alphabetical) order** for determinism:

```
1. bills
2. files
3. identity
4. onboarding
5. payments
```

**Why lexicographic?**
- Deterministic without explicit dependency configuration
- Easy to predict on any run
- No need for a separate dependency graph config
- Services can be added without ordering reconfiguration

**Order within services:**
1. All DDL migrations (sorted by version ascending)
2. Then all DML migrations for active environment (sorted by version ascending)

### Environment-Aware DML Filtering

### Environment-Aware DML Filtering

DML migrations are scoped by environment. The system determines which DML migrations to execute based on the environment variable:

```
Environment: APP_ENV (or ENVIRONMENT, or default 'local')
│
├─ APP_ENV=local → Execute: dml/local/*.sql
│  └─ Examples: seed_default_user, seed_test_projects
│
├─ APP_ENV=dev → Execute: dml/dev/*.sql
│  └─ Examples: seed_dev_test_data, seed_integration_fixtures
│
├─ APP_ENV=stg → Execute: dml/stg/*.sql (requires explicit approval)
│  └─ Examples: seed_staging_fixtures, seed_reference_data
│
└─ APP_ENV=prd → Execute: dml/prd/*.sql (requires APP_ENV=prd + --approve-production)
   └─ Examples: seed_production_defaults, seed_system_roles
```

**Safeguard**: Even if a `dml/prd/` file exists, it will NOT execute unless BOTH conditions are met:
- Environment variable is explicitly set to `prd`
- CLI flag `--approve-production` is provided

This two-factor safety prevents accidental production data changes.

---

## Transaction Semantics

### DDL Migration Transaction

```sql
BEGIN TRANSACTION(isolation_level = SERIALIZABLE);
  -- Execute user's .up.sql
  CREATE TABLE bills (...);
  CREATE INDEX idx_bills_project_id ON bills(project_id);
  
  -- Atomic audit record insertion
  INSERT INTO <schema>.migrations_ddl 
    (version, name, executed_at, execution_time_ms, success, ...)
  VALUES (1, 'create_bills_table', NOW(), 45, TRUE, ...);
  
COMMIT;
-- If ANY statement fails: ROLLBACK entire transaction
-- Result: migrations_ddl has NO record for this version (idempotent)
```

**Guarantees**:
- DDL + audit record are atomic (all-or-nothing)
- No partial schema changes
- Re-running same migration produces identical state

### DML Migration Transaction

```sql
BEGIN TRANSACTION(isolation_level = SERIALIZABLE);
  -- Execute user's .up.sql (may have multiple INSERTs, UPDATEs, etc.)
  IF APP_ENV = 'local' THEN
    INSERT INTO users (id, email, name) VALUES (...);
    INSERT INTO projects (id, owner_id, name) VALUES (...);
  END IF;
  
  -- Atomic audit record insertion
  INSERT INTO <schema>.migrations_dml 
    (version, environment, name, executed_at, execution_time_ms, success, ...)
  VALUES (1, 'local', 'seed_default_users', NOW(), 85, TRUE, ...);
  
COMMIT;
-- If ANY statement fails: ROLLBACK entire transaction
// Result: migrations_dml has NO record for this (version, environment)
```

---

## Query Patterns for Operators

### Query 1: Show All Applied Migrations for a Service

```sql
-- Show migration timeline for bills service
SELECT version, name, executed_at, execution_time_ms, executed_by  
FROM bills.migrations_ddl  
ORDER BY version ASC;

-- Show DML migrations applied in local environment
SELECT version, name, executed_at, execution_time_ms, executed_by  
FROM bills.migrations_dml  
WHERE environment = 'local' AND success = TRUE
ORDER BY version ASC;
```

### Query 2: Find Pending Migrations

```sql
-- Assuming discovery found versions 1-5 in folder but DB has only 1-3
-- Pending = discovered but not in tracking table
SELECT version FROM discovered_migrations
WHERE version NOT IN (SELECT version FROM bills.migrations_ddl);
-- Result: 4, 5 (these should execute next)
```

### Query 3: Slow Migrations (Performance Analysis)

```sql
SELECT name, execution_time_ms 
FROM bills.migrations_ddl
WHERE execution_time_ms > 500
ORDER BY execution_time_ms DESC;
```

### Query 4: Migration Status Summary

```sql
SELECT 
  'bills' as service,
  (SELECT MAX(version) FROM bills.migrations_ddl) as latest_ddl_version,
  (SELECT MAX(version) FROM bills.migrations_dml WHERE environment = 'local') as latest_dml_version
UNION ALL
SELECT 
  'files' as service,
  (SELECT MAX(version) FROM files.migrations_ddl) as latest_ddl_version,
  (SELECT MAX(version) FROM files.migrations_dml WHERE environment = 'local') as latest_dml_version;
-- ... repeat for all services
```

---

## Validation Rules & Constraints

### During Discovery (File System Scan)

1. **Filename Format**
   - ✅ Pattern: `<NNNNNN>_<slug>.(up|down).sql`
   - ✅ Version is 6-digit zero-padded integer
   - ❌ Fail: `/etc/migrations/ddl/create_table.sql` (missing version)
   - ❌ Fail: `/etc/migrations/ddl/1_create.up.sql` (not zero-padded)

2. **File Pairing**
   - ✅ Every `.up.sql` has a paired `.down.sql`
   - ⚠️ Warning: `.up.sql` without `.down.sql` (can still execute; rollback will fail)

3. **Folder Structure**
   - ✅ `migrations/ddl/` exists or can be created
   - ✅ `migrations/dml/` has subdirectories: `local/`, `dev/`, `stg/`, `prd/`
   - ❌ Fail: Invalid environment folder like `dml/prod/` (should be `prd`)

4. **Version Uniqueness**
   - ✅ No duplicate version numbers within same service's DDL folder
   - ✅ No duplicate version numbers within same service's DML folder per environment
   - ⚠️ Warning: Large gaps (e.g., jump from 1 to 100) may indicate copy-paste error

### During Execution (Before Running Migration)

1. **Idempotency Check**
   ```sql
   SELECT 1 FROM <schema>.migrations_ddl WHERE version = ? AND success = TRUE;
   -- If found: Skip migration (already applied)
   -- If not found: Execute migration
   ```

2. **Production Safety Check** (for prd environment DML)
   ```
   IF APP_ENV = 'prd':
      IF --approve-production flag NOT provided:
         ERROR: Production migrations require explicit approval
         FAIL and exit with code 1
   ELSE IF APP_ENV != 'prd' AND migrations exist in dml/prd/:
      WARN: Production migrations available but skipped (not in prd environment)
   ```

3. **Transaction Integrity Check**
   - Verify database is accessible and schema exists
   - Verify migrations_ddl and migrations_dml tables exist (or create them)
   - Verify no uncommitted transactions blocking migrations

### During Rollback (Down Execution)

### During Rollback (Down Execution)

1. **Find Rollback Target**
   ```sql
   -- Find latest applied migration (default target)
   SELECT version, name FROM <schema>.migrations_ddl 
   WHERE success = TRUE 
   ORDER BY version DESC LIMIT 1;
   
   -- OR find specific version if provided
   SELECT version FROM <schema>.migrations_ddl 
   WHERE version = ? AND success = TRUE;
   ```

2. **Reverse Order Execution**
   - Rollback in reverse version order (most recent first)
   - Stop immediately on first failure
   - Don't attempt cascading rollbacks

3. **Failed Rollback Handling**
   - ❌ Don't delete tracking record on failure
   - ✅ Leave record as-is (preserves audit trail)
   - ✅ Log error with full context
   - ✅ Require `--force-rollback` flag to retry
   - ✅ Operator must investigate data constraints and clean up manually

---

## Idempotency Guarantees

### Guarantee 1: Re-running Migrations is Safe

```
Scenario: Run `migrate up` twice
Run 1: DDL 1,2,3 execute and record inserted into tracking table
Run 2: DDL 1,2,3 skipped (found in tracking table), no duplicates
Result: Same final state, no errors ✓
```

### Guarantee 2: Failed Migrations Don't Corrupt State

```
Scenario: Migration 4 fails midway (syntax error)
Result: 
  - DDL changes rolled back automatically (transaction semantics)
  - migrations_ddl table has NO record for version 4
  - Operator can fix SQL and re-run; no cleanup needed ✓
```

### Guarantee 3: Environment Isolation

```
Scenario: APP_ENV=local during local development
Result:
  - Only dml/local/* migrations executed
  - dml/dev/*, dml/stg/*, dml/prd/* are ignored
  - No accidental data leaks across environments ✓

Scenario: APP_ENV=prd without --approve-production
Result:
  - Migration fails immediately with clear warning
  - dml/prd/* migrations NOT executed
  - No accidental production data mutations ✓
```

### Guarantee 4: Rollback Tracks All Reversals

```
Scenario: Rollback migrations 5,4,3 in order
Result:
  - Each successful rollback deletes its tracking record
  - migrations_ddl has versions 3,4,5 deleted in sequence
  - Timestamps logged for each rollback for audit trail
  - Failed rollback: record stays; manual recovery required ✓
```

---

## Schema Change Examples

### Example 1: Add New Table to bills Service

**File**: `backend/internals/bills/migrations/ddl/000004_create_invoices_table.up.sql`

```sql
CREATE TABLE IF NOT EXISTS invoices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    bill_id UUID NOT NULL REFERENCES bills(id) ON DELETE CASCADE,
    amount DECIMAL(10, 2) NOT NULL,
    due_date DATE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_invoices_project_id ON invoices(project_id);
CREATE INDEX IF NOT EXISTS idx_invoices_due_date ON invoices(due_date);
```

**Rollback**: `000004_create_invoices_table.down.sql`

```sql
DROP TABLE IF EXISTS invoices CASCADE;
```

**Execution Flow**:
1. Discovers file: `000004_create_invoices_table.up.sql`
2. Checks: SELECT version FROM bills.migrations_ddl WHERE version=4 — Not found
3. Executes: BEGIN TRANSACTION → CREATE TABLE → INSERT tracking record → COMMIT
4. Result: bills.migrations_ddl now has (version=4, name='create_invoices_table', success=TRUE)

---

### Example 2: Seed Default Roles for Local Dev

**File**: `backend/internals/identity/migrations/dml/local/000001_seed_roles.up.sql`

```sql
INSERT INTO roles (name, description) VALUES
('owner', 'Full access to project'),
('editor', 'Can edit bills and documents'),
('viewer', 'Read-only access')
ON CONFLICT (name) DO NOTHING;
```

**Rollback**: `000001_seed_roles.down.sql`

```sql
DELETE FROM roles WHERE name IN ('owner', 'editor', 'viewer');
```

**Execution Flow** (when APP_ENV=local):
1. Discovers: `dml/local/000001_seed_roles.up.sql`
2. Checks: SELECT version FROM identity.migrations_dml WHERE (version=1, environment='local') — Not found
3. Checks: APP_ENV=local (matches dml/local/) ✓
4. Executes: BEGIN TRANSACTION → INSERT → INSERT tracking record → COMMIT
5. Result: identity.migrations_dml now has (version=1, environment='local', name='seed_roles', success=TRUE)

**If APP_ENV=dev**:
- File is ignored (only `dml/dev/` scanned)
- Different seed file executed instead

---

## Performance Characteristics

| Operation | Target | Method |
|-----------|--------|--------|
| Migration Discovery | O(n) | Directory scan; n = number of files |
| Idempotency Check | O(1) | Primary key lookup in tracking table |
| DDL Execution | O(schema_complexity) | Depends on SQL statements |
| DML Execution | O(data_volume) | Depends on INSERT/UPDATE volume |
| Status Query | O(1) | Single row lookup per service |

**Typical Performance**:
- Discovery: 10-50ms for all services
- Single DDL migration: 10-500ms (mostly depends on index creation)
- Single DML migration: 20-300ms (depends on data volume)
- Total for all 6 services: < 30 seconds (SC-008 target)

---

## Data Integrity & Compliance

### ACID Properties

| Property | Guarantee |
|----------|-----------|
| **Atomicity** | Each migration + tracking record is one transaction; all-or-nothing semantics |
| **Consistency** | Migrations committed to tracking table only after success; idempotency prevents duplicate execution |
| **Isolation** | Per-schema isolation; concurrent services don't interfere; transaction-level isolation within PostgreSQL |
| **Durability** | Tracking records persisted immediately; recovery possible even if process crashes |

### Audit Trail

```sql
-- WHO executed what, WHEN, and HOW LONG
SELECT executed_by, name, executed_at, execution_time_ms, success 
FROM bills.migrations_ddl
UNION ALL
SELECT executed_by, name, executed_at, execution_time_ms, success 
FROM bills.migrations_dml
ORDER BY executed_at DESC;

-- Result: Complete timeline of database evolution with operator context
```

This enables compliance reporting, SLA verification, and incident forensics.
