//go:build integration

package helpers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// SuiteResources stores shared resources required by integration suites.
type SuiteResources struct {
	DB        *sql.DB
	DSN       string
	terminate func(context.Context) error
}

// SetupPostgresSuite starts an ephemeral PostgreSQL container and returns
// resources required by integration packages.
func SetupPostgresSuite(ctx context.Context) (*SuiteResources, error) {
	container, err := postgres.Run(
		ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("financial_test"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(2*time.Minute),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("start postgres container: %w", err)
	}

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, fmt.Errorf("resolve postgres dsn: %w", err)
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, fmt.Errorf("open postgres db: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()
	if err := pingWithRetry(pingCtx, db); err != nil {
		_ = db.Close()
		_ = container.Terminate(ctx)
		return nil, fmt.Errorf("ping postgres db: %w", err)
	}

	return &SuiteResources{
		DB:  db,
		DSN: dsn,
		terminate: func(c context.Context) error {
			return container.Terminate(c)
		},
	}, nil
}

// Close tears down all suite resources.
func (r *SuiteResources) Close(ctx context.Context) {
	if r == nil {
		return
	}
	if r.DB != nil {
		_ = r.DB.Close()
	}
	if r.terminate != nil {
		_ = r.terminate(ctx)
	}
}

// RunMigrations applies migrations from sourcePath into the provided DSN.
func RunMigrations(dsn, sourcePath string) error {
	sourcePath, err := normalizeSourcePath(sourcePath)
	if err != nil {
		return err
	}

	rawSource := strings.TrimPrefix(sourcePath, "file://")
	rawSource, err = resolveMigrationDir(rawSource)
	if err != nil {
		return err
	}

	if empty, err := isMigrationDirEmpty(rawSource); err != nil {
		return err
	} else if empty {
		return nil
	}

	sourcePath = "file://" + rawSource

	svcName := "schema"
	parts := strings.Split(strings.TrimPrefix(sourcePath, "file://"), "/")
	for i, part := range parts {
		if part == "internals" && i+1 < len(parts) {
			svcName = parts[i+1]
			break
		}
	}

	tableParam := "x-migrations-table=" + svcName + "_schema_migrations"
	migrDSN := dsn
	if strings.Contains(dsn, "?") {
		migrDSN = dsn + "&" + tableParam
	} else {
		migrDSN = dsn + "?" + tableParam
	}

	mig, err := migrate.New(sourcePath, migrDSN)
	if err != nil {
		return fmt.Errorf("migrate.New: %w", err)
	}
	defer mig.Close()

	if err := ensurePreMigrationCompatibility(dsn, sourcePath); err != nil {
		return err
	}

	if err := mig.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate.Up: %w", err)
	}

	if err := applyIntegrationCompatibilitySchema(dsn, sourcePath); err != nil {
		return err
	}

	return nil
}

func ensurePreMigrationCompatibility(dsn, sourcePath string) error {
	if !strings.Contains(sourcePath, "/internals/payments/") {
		return nil
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("pre-migration compat: open db: %w", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	stmt := `CREATE TABLE IF NOT EXISTS users (
		id UUID PRIMARY KEY,
		project_id UUID,
		email TEXT,
		role TEXT,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	)`

	if _, err := db.ExecContext(ctx, stmt); err != nil {
		return fmt.Errorf("pre-migration compat: users table: %w", err)
	}

	return nil
}

func applyIntegrationCompatibilitySchema(dsn, sourcePath string) error {
	if strings.Contains(sourcePath, "/internals/onboarding/") {
		if err := applyOnboardingCompatibilitySchema(dsn); err != nil {
			return err
		}
	}

	if !strings.Contains(sourcePath, "/internals/bills/") {
		return nil
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("compat schema: open db: %w", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	statements := []string{
		`CREATE TABLE IF NOT EXISTS bill_types (
			id UUID PRIMARY KEY,
			project_id UUID NOT NULL,
			name TEXT NOT NULL,
			created_by UUID,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`ALTER TABLE bill_types ADD COLUMN IF NOT EXISTS created_by UUID`,
		`ALTER TABLE bill_types ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`,
		`CREATE INDEX IF NOT EXISTS idx_bill_types_project_id ON bill_types (project_id)`,
		`ALTER TABLE bill_records ADD COLUMN IF NOT EXISTS bill_type_id UUID`,
		`CREATE INDEX IF NOT EXISTS idx_bill_records_bill_type_id ON bill_records (bill_type_id)`,
		`CREATE TABLE IF NOT EXISTS idempotency_keys (
			project_id UUID NOT NULL,
			operation TEXT NOT NULL,
			idempotency_key TEXT NOT NULL,
			response_hash TEXT NOT NULL,
			expires_at TIMESTAMPTZ NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			PRIMARY KEY (project_id, operation, idempotency_key)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_idempotency_keys_expires_at ON idempotency_keys (expires_at)`,
	}

	for _, stmt := range statements {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("compat schema: exec: %w", err)
		}
	}

	return nil
}

func applyOnboardingCompatibilitySchema(dsn string) error {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("onboarding compat: open db: %w", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	statements := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY,
			project_id UUID,
			email TEXT,
			role TEXT,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`DO $$ BEGIN
			CREATE TYPE project_type AS ENUM ('personal', 'conjugal', 'shared');
		EXCEPTION WHEN duplicate_object THEN NULL;
		END $$`,
		`DO $$ BEGIN
			CREATE TYPE project_member_role AS ENUM ('read_only', 'update', 'write');
		EXCEPTION WHEN duplicate_object THEN NULL;
		END $$`,
		`ALTER TABLE projects ADD COLUMN IF NOT EXISTS owner_id UUID`,
		`ALTER TABLE projects ADD COLUMN IF NOT EXISTS type project_type NOT NULL DEFAULT 'personal'`,
		`ALTER TABLE projects ALTER COLUMN id SET DEFAULT gen_random_uuid()`,
		`ALTER TABLE projects ALTER COLUMN owner_user_id DROP NOT NULL`,
		`UPDATE projects SET owner_id = owner_user_id WHERE owner_id IS NULL`,
		`CREATE TABLE IF NOT EXISTS project_members (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
			user_id UUID NOT NULL,
			role project_member_role NOT NULL DEFAULT 'read_only',
			invited_by UUID,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			CONSTRAINT uq_project_members_project_user UNIQUE (project_id, user_id)
		)`,
		`INSERT INTO users (id, project_id, email, role)
		 VALUES ('00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000010', 'integration@example.com', 'write')
		 ON CONFLICT (id) DO NOTHING`,
		`INSERT INTO projects (id, owner_user_id, owner_id, name, type)
		 VALUES ('00000000-0000-0000-0000-000000000010', '00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000001', 'Integration Project', 'personal')
		 ON CONFLICT (id) DO NOTHING`,
		`INSERT INTO project_members (project_id, user_id, role, invited_by)
		 VALUES ('00000000-0000-0000-0000-000000000010', '00000000-0000-0000-0000-000000000001', 'write', '00000000-0000-0000-0000-000000000001')
		 ON CONFLICT (project_id, user_id) DO NOTHING`,
	}

	for _, stmt := range statements {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("onboarding compat: exec: %w", err)
		}
	}

	return nil
}

func isMigrationDirEmpty(path string) (bool, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return false, fmt.Errorf("read migration dir %q: %w", path, err)
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.HasSuffix(e.Name(), ".sql") {
			return false, nil
		}
	}

	return true, nil
}

func resolveMigrationDir(path string) (string, error) {
	empty, err := isMigrationDirEmpty(path)
	if err != nil {
		return "", err
	}
	if !empty {
		return path, nil
	}

	ddlPath := filepath.Join(path, "ddl")
	if info, err := os.Stat(ddlPath); err == nil && info.IsDir() {
		return ddlPath, nil
	}

	return path, nil
}

func normalizeSourcePath(sourcePath string) (string, error) {
	const prefix = "file://"
	if !strings.HasPrefix(sourcePath, prefix) {
		return sourcePath, nil
	}

	raw := strings.TrimPrefix(sourcePath, prefix)
	if filepath.IsAbs(raw) {
		return sourcePath, nil
	}

	if idx := strings.Index(raw, "internals/"); idx >= 0 {
		raw = raw[idx:]
	}

	root, err := backendRootFromRuntime()
	if err != nil {
		return "", err
	}

	abs := filepath.Join(root, raw)
	return prefix + abs, nil
}

func backendRootFromRuntime() (string, error) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("resolve helper file path")
	}

	// helpers/lifecycle.go -> integration -> tests -> backend
	backendRoot := filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", "..", ".."))
	return backendRoot, nil
}

func pingWithRetry(ctx context.Context, db *sql.DB) error {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		if err := db.PingContext(ctx); err == nil {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}
