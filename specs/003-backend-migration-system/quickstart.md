# Quickstart: Backend Migration System

**Feature**: 003-backend-migration-system

## For Developers

### Prerequisites

- Go 1.21+
- PostgreSQL running locally
- `golang-migrate/migrate` library (will be added to go.mod)
- Backend directory structure ready

### Quick Start

```bash
# Navigate to backend directory
cd backend

# Install migration tool
go get -u github.com/golang-migrate/migrate/v4
go get -u github.com/golang-migrate/migrate/v4/database/postgres
go get -u github.com/golang-migrate/migrate/v4/source/file

# Create folder structure for all services
mkdir -p internals/files/migrations/{ddl,dml/{local,dev,stg,prd}}
mkdir -p internals/bills/migrations/{ddl,dml/{local,dev,stg,prd}}
mkdir -p internals/identity/migrations/{ddl,dml/{local,dev,stg,prd}}
mkdir -p internals/onboarding/migrations/{ddl,dml/{local,dev,stg,prd}}
mkdir -p internals/payments/migrations/{ddl,dml/{local,dev,stg,prd}}

# Run migrations for local development
make migrate/up env=local

# Check migration status
make migrate/status

# Rollback last migration (if needed)
make migrate/down
```

### Testing the Feature

```bash
# Run migration tests
make test/migrations

# Test DDL execution
make migrate/up env=local service=files

# Test environment-specific DML
make migrate/up env=dev service=identity

# Check that default user was seeded in dev
make db/query query="SELECT * FROM users LIMIT 1"

# Verify migrations were recorded
make db/query query="SELECT * FROM migrations_ddl"
make db/query query="SELECT * FROM migrations_dml"
```

### Database Inspection

```bash
# Check migration tables exist
psql -U postgres -d financial_assistant -c "\\dt migrations_*"

# View applied migrations
psql -U postgres -d financial_assistant -c "SELECT * FROM migrations_ddl ORDER BY executed_at"
psql -U postgres -d financial_assistant -c "SELECT * FROM migrations_dml ORDER BY executed_at WHERE environment = 'local'"

# Check schema for a service
psql -U postgres -d financial_assistant -c "\\dt files_*"
```

## For DevOps / Infrastructure

### Environment Variables

```bash
export APP_ENV=local           # or dev, stg, prd
export DATABASE_URL=postgres://user:pass@localhost/financial_assistant
export MIGRATE_DATABASE=postgres
```

### CI/CD Integration

```bash
# In GitHub Actions or similar CI system:

# Run migrations on staging environment
make migrate/up env=stg

# Verify migration status before deployment
make migrate/status

# Rollback in case of failure
make migrate/down
```

### Production Safety

```bash
# Migrations for prd environment should be reviewed manually
# Migration module prevents accidental prd execution unless explicitly set

export APP_ENV=prd
make migrate/status  # Shows pending prd migrations
make migrate/validate  # (Optional) Validate migration syntax before execution

# Apply prd migrations only after approval
make migrate/up env=prd
```

## For QA / Testers

### Migration Validation Checklist

- [ ] DDL migrations create all expected tables
- [ ] DML migrations seed data correctly for the environment
- [ ] Each environment has only its own DML data (no cross-env pollution)
- [ ] Migrations are idempotent (running twice gives same result)
- [ ] Rollback reverts schema and data correctly
- [ ] Error handling prevents partial application
- [ ] Status command accurately reports pending migrations
- [ ] Migration tables track all applied migrations

### Test Scenarios

```bash
# Scenario 1: Fresh database setup
rm -f database.db  # (or drop postgres DB)
make migrate/up env=local
# Verify: All tables exist, default user is present in users table

# Scenario 2: Idempotency
make migrate/up env=local
make migrate/up env=local  # Should skip already-applied migrations
# Verify: Same number of records, no duplicates

# Scenario 3: Rollback
make migrate/up env=local
make migrate/down
# Verify: Schema is reverted

# Scenario 4: Environment isolation
make migrate/up env=local
# Verify: Only local DML data is present
make migrate/up env=dev  # (separate DB or with DB reset)
# Verify: Only dev DML data is present
```

## For Product Managers

### Benefits

- **Reproducible**: Same migrations work consistently across environments
- **Reversible**: Rollback capability enables safe deployment recovery
- **Auditable**: All migrations tracked and timestamped in database
- **Safe**: Environment flags prevent accidental data loss
- **Scalable**: Supports adding services and environments without code changes

## Troubleshooting

| Issue | Solution |
|-------|----------|
| Migration fails with "table already exists" | Check migration tracking tables; may be a duplicate filename or environment mismatch |
| DML migrations not executing | Verify APP_ENV is set correctly; check that dml/<environment>/ subfolder exists |
| Rollback fails | Check .down.sql file exists and is valid; may need manual intervention |
| Status shows wrong count | Query migrations_ddl/migrations_dml tables directly; check for stale records |

## Next Steps

- Add migration files for each service DDL schema
- Add DML seed data for local/dev environments
- Integrate migrations into CI/CD pipeline
- Run `/speckit.plan` to generate detailed design and task breakdown
- Execute tasks from tasks.md
- Run integration tests before marking feature complete
