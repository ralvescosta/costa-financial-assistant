# Feature Specification: Backend Migration System Overhaul

**Feature Branch**: `003-backend-migration-system`
**Created**: 2026-04-01
**Status**: Draft
**Input**: User description: "We must put the migration to work. The migration project is doing nothing right now, and we need to work on the migration as our priority. First of all, each module in the backend must organize the migrations folder. Files/migrations organized as DDL and DML (local, dev, stg, prd environments). The migration module will configure the services to execute up and down operations. First execute DDL, then DML based on the environment. The migration module must use golang-migrate/migrate to manage migrations. We must create two tables to manage the version: one called migrations_ddl and one called migrations_dml."

## Clarifications

### Session 2026-04-01

- Q: How should production migrations be protected from accidental execution? → A: Require explicit CLI flag (`--approve-production` or `--force`) AND verify `APP_ENV=prd` before executing production migrations. Implement two-factor safety pattern.
- Q: In what order should multi-service migrations execute when no specific service is selected? → A: Lexicographic/alphabetical order (bills, files, identity, onboarding, payments) for determinism and simplicity.
- Q: Should each service have its own isolated schema or share a single schema? → A: Each service owns its own PostgreSQL schema (bills, files, identity, onboarding, payments) for explicit service boundaries, security isolation, and RBAC.
- Q: How should rollback failures (e.g., constraint violations) be handled and recovered? → A: On rollback failure, leave migration record unchanged, log extensively, exit non-zero. Require manual intervention + explicit `--force-rollback` flag for retry.
- Q: What environment variable should control the active environment and what is the fallback behavior? → A: Check `APP_ENV` first, then `ENVIRONMENT`, then default to `local` for safety. Priority order prevents production mutations on unset vars.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Migrate DDL Schema Changes Across Environments (Priority: P0)

The backend migration system must execute all Data Definition Language (DDL) scripts — CREATE TABLE, ALTER TABLE, CREATE INDEX, etc. — across all environments in the correct order. Each service's DDL migrations are discovered from a standardized folder structure (`<service>/migrations/ddl/`), versioned sequentially, and executed before any DML operations. The system tracks DDL migration version history in a dedicated `migrations_ddl` table to prevent re-execution and enable rollback.

**Why this priority**: DDL schema changes are the foundation upon which all data operations depend. Without working DDL migrations, the database schema cannot be established, evolved, or rolled back safely. This is the critical path blocker.

**Independent Test**: Can be fully tested by running `make migrate/up/files` (or another service) in a clean database environment, verifying that all DDL migrations execute in order, the database schema matches the expected state, and the `migrations_ddl` table records all applied migrations with timestamps.

**Acceptance Scenarios**:

1. **Given** a clean database with no schema, **When** the DDL migration for a service is executed, **Then** all `.up.sql` files in `<service>/migrations/ddl/` are discovered, sorted by filename prefix (sequential number), and executed in ascending order.
2. **Given** the first DDL migration is executed, **When** it completes successfully, **Then** the `migrations_ddl` table is created (if it doesn't exist) and a record is inserted documenting the migration version, name, and execution timestamp.
3. **Given** multiple DDL migrations exist for a service (e.g., `000001_create_documents.up.sql`, `000002_add_index_status.up.sql`), **When** migrations are executed, **Then** each migration is run in numeric order and recorded in `migrations_ddl` with its respective version.
4. **Given** a DDL migration has already been executed and is recorded in `migrations_ddl`, **When** the migration command is run again, **Then** the already-executed migration is skipped and not re-executed.
5. **Given** a DDL migration execution fails (e.g., SQL syntax error, constraint violation), **When** the error is detected, **Then** the transaction is rolled back, the error is logged with context, and the `migrations_ddl` record is NOT inserted for the failed migration.
6. **Given** a service's DDL migration is marked as needing rollback, **When** the rollback command is executed, **Then** the corresponding `.down.sql` script is run, the database schema is reverted, and the `migrations_ddl` record for that migration is deleted.
7. **Given** the `.down.sql` files exist alongside `.up.sql` files, **When** a rollback is executed, **Then** the `.down.sql` file for the migration is executed in reverse order (most recent migration first).
8. **Given** multiple services exist (files, bills, payments, etc.), **When** migrations are executed, **Then** each service's DDL migrations are run independently, and the system correctly tracks versions per service (or uses a unified `migrations_ddl` table with a service identifier).

---

### User Story 2 - Migrate DML Data Changes Based on Environment (Priority: P1)

After DDL migrations are complete, the system executes Data Manipulation Language (DML) scripts — INSERT, UPDATE, DELETE — specific to the current environment (local, dev, stg, prd). Environment-specific folders (`<service>/migrations/dml/<environment>/`) contain seeded data, sample records, and environment-specific configurations. The `migrations_dml` table tracks DML migration versions independently from DDL.

**Why this priority**: DML migrations enable environment-aware data seeding and transformations. This allows each environment to have appropriate baseline data (e.g., default users in local, production fixtures in stg) without manual setup. It is P1 because DDL (P0) must complete first.

**Independent Test**: Can be fully tested by running `make migrate/up/files env=local` (or another environment), verifying that all global DDL executes first, then DML migrations from `migrations/dml/local/` execute afterward, and the `migrations_dml` table records only the applied environment-specific migrations.

**Acceptance Scenarios**:

1. **Given** DDL migrations have completed successfully, **When** DML migrations are executed, **Then** all `.up.sql` files in `<service>/migrations/dml/<environment>/` are discovered and executed after DDL migrations complete.
2. **Given** the current environment is `local`, **When** DML migrations are executed, **Then** migrations from `<service>/migrations/dml/local/` are run (e.g., seeding a default user, creating bootstrap projects).
3. **Given** the current environment is `dev`, **When** migrations are executed, **Then** migrations from `<service>/migrations/dml/dev/` are run (e.g., seeding test data, creating sample records).
4. **Given** the environment is `prd` (production), **When** migrations are run, **Then** only migrations from `<service>/migrations/dml/prd/` are executed, and sensitive seed data (e.g., default credentials) is NOT included from other environments.
5. **Given** a DML migration has executed successfully, **When** the migration is recorded, **Then** the `migrations_dml` table inserts a record with the version, name, environment, and execution timestamp.
6. **Given** the same DML migration filename exists in multiple environments, **When** migrations are executed in different environments, **Then** each environment's version of the migration is tracked independently in `migrations_dml` (e.g., `local` and `dev` versions are separate records).
7. **Given** a DML migration for `local` has already executed, **When** migrations are re-run in the `local` environment, **Then** the already-executed migration is skipped.
8. **Given** a DML migration fails (e.g., constraint violation during an INSERT), **When** the error occurs, **Then** the transaction is rolled back, the error is logged, and no `migrations_dml` record is inserted.

---

### User Story 3 - Migration Module CLI Commands (Priority: P1)

The migration module provides CLI commands to trigger migrations up/down, rollback to a specific version, and query migration status. The BFF or other services can invoke these commands via the CLI (Cobra) or internal API to manage database schemas consistently across the stack.

**Why this priority**: CLI commands provide the operational interface for developers and CI/CD systems to control migrations safely. This must work early to unblock all database setup and testing workflows.

**Independent Test**: Can be fully tested by running `make migrate/up`, `make migrate/down`, `make migrate/status`, and verifying the correct command-line output and side effects (schema changes, migration records updated).

**Acceptance Scenarios**:

1. **Given** the migration CLI is available, **When** the `migrate up` command is executed with a service name (e.g., `migrate up --service files`), **Then** all pending DDL migrations for that service are applied, followed by all pending DML migrations for the configured environment, in order.
2. **Given** the migrate `up` command is executed, **When** all migrations complete successfully, **Then** the CLI exits with code 0 and logs a summary message (e.g., "Applied 3 DDL migrations and 2 DML migrations for files service").
3. **Given** the migrate `down` command is executed with a service name and optional version, **When** the rollback completes, **Then** the specified migration (or the most recent migration) is rolled back by executing its `.down.sql` script.
4. **Given** a migration during rollback fails, **When** the error is detected, **Then** the CLI logs the error, exits with a non-zero exit code, and does NOT apply any further rollbacks.
5. **Given** the `migrate status` command is executed, **When** it completes, **Then** the CLI outputs a table or formatted list showing the migration status for all services, including the number of pending migrations and the last applied migration version.
6. **Given** the migrations are executed for all services, **When** a user runs `make migrate/up` without specifying a service, **Then** migrations for all discovered services are applied in order (or the help text specifies that a service name is required).
7. **Given** the migrations are configured for a specific environment (e.g., `local`), **When** DML environment-specific migrations are applied, **Then** only DML migrations for that environment are applied (not all environments).

---

### User Story 4 - Folder Structure Standardization and Discovery (Priority: P0)

All backend services must organize their migration files into a consistent folder structure: `<service>/migrations/ddl/` for schema changes and `<service>/migrations/dml/<environment>/` for data changes. The migration system automatically discovers and executes migrations from this structure without manual configuration per service.

**Why this priority**: A standardized folder structure is foundational for the migration system's automated discovery logic. Without it, the system cannot reliably locate and execute migrations. This is the structural prerequisite for P0 DDL execution.

**Independent Test**: Can be fully tested by verifying that the `backend/internals/<service>/migrations/` folder structure follows the standard convention for all six services (files, bills, identity, onboarding, payments), and that the migration system's discovery logic correctly identifies all migrations without errors.

**Acceptance Scenarios**:

1. **Given** a service directory (e.g., `backend/internals/files/`), **When** the migration system initializes, **Then** it looks for a `migrations/ddl/` subdirectory and a `migrations/dml/<environment>/` subdirectory structure.
2. **Given** migration files exist in the standardized folders, **When** the discovery process runs, **Then** all `.up.sql` and `.down.sql` files are identified, sorted by numeric prefix (000001_, 000002_, etc.), and prepared for execution.
3. **Given** a service does not yet have a migrations folder, **When** the system is initialized, **Then** the folder structure is created automatically (or the system logs a warning and creates it on first run).
4. **Given** two services (files, bills) both have migrations, **When** migrations are executed, **Then** each service's migrations are discovered and tracked independently in the database.
5. **Given** the migration files are named with a sequential numeric prefix (e.g., `000001_create_documents.up.sql`), **When** they are discovered, **Then** they are sorted and executed in ascending numeric order.
6. **Given** migration files exist in both `ddl/` and `dml/` folders for a service, **When** migrations run, **Then** all `ddl/` migrations are executed first, followed by `dml/` migrations for the active environment.
7. **Given** the `dml/` folder has subfolders for different environments (local, dev, stg, prd), **When** migrations are executed in a specific environment, **Then** only the migrations in that environment's subfolder are executed (e.g., `dml/local/` if the environment is `local`).

---

## Functional Requirements *(mandatory)*

**Migration System Architecture & Storage**

- **FR-001**: The migration system MUST use `golang-migrate/migrate` library as the underlying migration execution engine, configured and wrapped by the migration service.
- **FR-002**: Two database tables MUST be created to track migration versions: `migrations_ddl` (for Data Definition Language schemas) and `migrations_dml` (for Data Manipulation Language data changes).
- **FR-003**: The `migrations_ddl` table MUST contain at minimum: `version` (integer, unique), `name` (string), `executed_at` (timestamp), `execution_time_ms` (integer), and `success` (boolean).
- **FR-004**: The `migrations_dml` table MUST contain at minimum: `version` (integer, unique), `name` (string), `environment` (string), `executed_at` (timestamp), `execution_time_ms` (integer), and `success` (boolean).
- **FR-005**: The system MUST support idempotent execution — running migrations multiple times against the same database must result in the same final state (no duplicate inserts or redundant schema changes).

**Schema Isolation & Multi-Service Architecture**

- **FR-005a**: Each service (bills, files, identity, onboarding, payments) MUST have its own isolated PostgreSQL schema (e.g., schema `bills`, schema `files`). All tables and objects for a service reside in that schema with no table prefix convention required.
- **FR-005b**: Each service's schema MUST contain its own `migrations_ddl` and `migrations_dml` tables for independent version tracking. Cross-service queries require explicit schema-qualified names and are tracked for audit purposes.
- **FR-005c**: The migration system MUST create the service's schema if it does not exist before applying migrations to that schema.

**Folder Structure & File Conventions**

- **FR-006**: All migrations for each service MUST be organized under `backend/internals/<service>/migrations/` with subdirectories: `ddl/` for schema changes and `dml/<environment>/` for environment-specific data changes.
- **FR-007**: Migration files MUST follow the naming convention: `<NNNNNN>_<descriptive_slug>.up.sql` and `<NNNNNN>_<descriptive_slug>.down.sql`, where `<NNNNNN>` is a 6-digit zero-padded sequential number.
- **FR-008**: DML migrations MUST be organized within `dml/` as subdirectories for each environment: `dml/local/`, `dml/dev/`, `dml/stg/`, `dml/prd/`.
- **FR-009**: The system MUST auto-discover migration files from the standardized folder structure without requiring explicit per-service configuration.
- **FR-010**: Each migration file MUST contain valid SQL statements appropriate for the target database (PostgreSQL for this project).

**Execution Order and Strategy**

- **FR-011**: When migrations are executed for a service, ALL DDL migrations MUST be executed first, in numeric order, before any DML migrations are attempted.
- **FR-012**: After all DDL migrations complete successfully, DML migrations from the active environment MUST be executed in numeric order (e.g., all `dml/local/` migrations if the environment is `local`).
- **FR-013**: Only DML migrations for the active environment MUST be executed; migrations from other environments MUST be skipped.
- **FR-014**: If a DDL migration fails at any point, the transaction MUST be rolled back and the system MUST stop execution; no DML migrations MUST be attempted.
- **FR-015**: If a DML migration fails, the transaction MUST be rolled back and the system MUST stop execution; subsequent DML migrations MUST NOT be attempted unless explicitly retried.

**CLI Commands & Interface**

- **FR-016**: A `migrate up` command MUST be available (via `make migrate/up/<service>` or direct CLI invocation) to apply all pending DDL and DML migrations for a specified service.
- **FR-017**: A `migrate down` command MUST be available to rollback the most recent migration (or a specific version) by executing the corresponding `.down.sql` script.
- **FR-018**: A `migrate status` command MUST be available to display the current migration state, including: total pending migrations, last applied DDL/DML version per service, and next migration to be applied.
- **FR-019**: The CLI MUST accept an `--environment` or `--env` flag to specify the active environment (local, dev, stg, prd), controlling which DML migrations are executed.
- **FR-020**: The CLI MUST accept a `--service` flag to specify which service's migrations to run; if omitted, migrations for all services MUST be executed in lexicographic (alphabetical) order: bills, files, identity, onboarding, payments.
- **FR-021**: The CLI MUST provide clear, timestamped log output for each migration step (discovery, execution, completion) and include execution time metrics.
- **FR-022**: The CLI MUST exit with code 0 on success and a non-zero exit code on failure, enabling programmatic detection of migration failures in CI/CD pipelines.

**Error Handling & Rollback**

- **FR-023**: If a migration fails during execution, the system MUST display a clear error message with the migration filename, SQL error details, and the line number where the error occurred.
- **FR-024**: Failed migrations MUST NOT be recorded in the migration tracking tables, allowing re-execution and recovery without duplicate entries.
- **FR-024a**: If a rollback fails, the migration record MUST remain in the migration tracking table as-is (not deleted or modified). The failure is logged with full context for manual investigation. A `--force-rollback` flag enables retry after manual intervention.
- **FR-025**: The system MUST support manual rollback of one or more migrations without requiring complete database reset.
- **FR-026**: Rollback MUST respect the reverse of the apply order: most recent migrations are rolled back first. If a rollback fails (e.g., due to foreign key constraint violations), the system MUST stop immediately, log the detailed error with context (migration name, SQL error, line number), and exit with a non-zero exit code. The migration record MUST remain in the tracking table untouched to preserve audit history.
- **FR-026a**: On rollback failure, the system MUST require an explicit `--force-rollback` flag to retry the failed rollback. This flag MUST be accompanied by a warning message indicating that manual data cleanup may be required. Automatic retry without operator intervention is NOT permitted.

**Environment Differentiation**

- **FR-027**: The migration system MUST read the current environment using this priority order: (1) `APP_ENV` environment variable, (2) `ENVIRONMENT` environment variable, (3) default to `local`. The system MUST log the selected environment and its source (env var name or default) at startup. If the environment is not one of [local, dev, stg, prd], the system MUST exit with a clear error message and exit code 1.
- **FR-028**: Different environments (local, dev, stg, prd) MUST have distinct DML data seeds and configurations without code duplication.
- **FR-029**: The system MUST NOT allow production (prd) migrations to execute without BOTH: (1) an explicit approval flag (`--approve-production` or equivalent) AND (2) the environment variable `APP_ENV=prd`. If either condition is unmet, migrations must fail with a clear warning message and exit code 1. This implements a two-factor safety pattern to prevent accidental production schema changes.

**Observability & Auditing**

- **FR-030**: All migration execution events MUST be logged with timestamps, migration names, execution duration, and status (success or failure).
- **FR-031**: The migration tracking tables (`migrations_ddl` and `migrations_dml`) MUST serve as an audit log for all applied and rolled-back migrations.
- **FR-032**: The system MUST provide a queryable interface to retrieve migration history for a given time range, service, or environment.

## Key Entities

- **Migration**: A single `.sql` file representing one atomic schema change (DDL) or data operation (DML), versioned with a 6-digit prefix.
- **MigrationDDL**: A record in the `migrations_ddl` table tracking a completed DDL migration's version, name, execution timestamp, and status.
- **MigrationDML**: A record in the `migrations_dml` table tracking a completed DML migration's version, name, environment, execution timestamp, and status.
- **Service**: A backend domain module (files, bills, identity, onboarding, payments) with its own isolated PostgreSQL schema and dedicated migrations folder structure.
- **Schema**: Each service owns its own PostgreSQL schema (e.g., `bills` schema, `files` schema) with no table prefix convention needed. Each schema has its own `migrations_ddl` and `migrations_dml` tracking tables.
- **Environment**: A deployment context (local, dev, stg, prd) that determines which DML migrations are applied.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001** [CI]: All DDL migrations for a service execute in the correct numeric order without errors, and the database schema matches the expected final state. *Verified by*: US1 integration test (T001) running migrations on a fresh database and verifying schema completeness.
- **SC-002** [CI]: DDL migrations are never re-executed on repeated `migrate up` invocations; the `migrations_ddl` table prevents duplicate executions. *Verified by*: US1 integration test (T003) running migrate twice and confirming identical state and no duplicate records.
- **SC-003** [CI]: DML migrations execute only after all DDL migrations complete successfully. *Verified by*: US2 integration test (T010) inspecting migration log output and confirming all DDL versions precede DML versions.
- **SC-004** [CI]: Environment-specific DML migrations (local, dev, stg, prd) execute correctly when the environment is specified, with other environments' migrations skipped. *Verified by*: US2 integration test (T012) running migrate in `local` and `dev` environments and confirming only the respective migrations execute.
- **SC-005** [CI]: Migration failures result in atomic rollback (no partial changes persisted) and clear error logging. *Verified by*: US1 integration test (T004) executing a malformed migration and confirming the transaction rolled back and error was logged.
- **SC-006** [CI]: A rollback command correctly reverts a migration to its previous state using the `.down.sql` script. *Verified by*: US1 integration test (T005) applying then rolling back a migration and confirming the schema is restored.
- **SC-007** [CI]: The `migrate status` command accurately reports pending migrations and the last applied version for all services. *Verified by*: US3 integration test (T020) checking status output against the migration tracking tables and confirming accuracy.
- **SC-008** [OPS]: Migration execution for all six services (files, bills, identity, onboarding, payments, migrations) completes in under 30 seconds from a clean database state. *Measurement method*: Total elapsed time from `make migrate/up` start to completion (all DDL and DML for all services). Validated via integration test or CI performance metric.
- **SC-009** [CI]: The standardized folder structure (`migrations/ddl/`, `migrations/dml/<environment>/`) is correctly auto-discovered for all services without manual configuration. *Verified by*: US4 integration test (T030) verifying migration discovery identifies all files without errors.
- **SC-010** [CI]: Default user and bootstrap data are successfully seeded in the `local` and `dev` environments via DML migrations, and are NOT present in `prd` environment migrations. *Verified by*: US2 integration test (T015) querying for seeded users after migrate and confirming presence/absence per environment.

## Assumptions

- **A1**: The `golang-migrate/migrate` library is already imported or will be added to the project's `go.mod` dependencies.
- **A2**: All six services (files, bills, identity, onboarding, payments, migrations) will have their migrations organized in the standardized folder structure.
- **A3**: The backend uses PostgreSQL as the target database; migration SQL is PostgreSQL-compatible.
- **A4**: The current environment (local, dev, stg, prd) is determinable via an environment variable (e.g., `APP_ENV`) or configuration file.
- **A5**: Each service has a dedicated migrations folder; shared/cross-service migrations are not required (each service owns its schema and data).
- **A6**: The system has database connection credentials and connectivity before migration execution; credential injection is handled separately (e.g., via environment variables or secrets manager).
- **A7**: Migration filenames uniquely identify migrations globally across all services (the 6-digit prefix is sufficient to prevent collisions).

## Notes

- The migration system is foundational infrastructure that unblocks all database development and testing workflows. It must be fully operational before any complex data operations are introduced.
- The separation of DDL and DML migrations allows flexible schema management while enabling environment-specific data seeding.
- The `golang-migrate/migrate` library provides battle-tested, production-grade migration management; leveraging it reduces the risk of custom migration logic bugs.
- Default user seeding in local/dev environments enables rapid onboarding and testing without manual database setup steps.
- The design supports horizontal scaling if needed in the future (migrations are idempotent and can be safely re-run).
