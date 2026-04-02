package services

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

var validSchemaRegexp = regexp.MustCompile(`^[a-z_][a-z0-9_]*$`)

var defaultServiceOrder = []string{"bills", "files", "identity", "onboarding", "payments"}

// MigrateOptions are runtime controls for migration execution.
type MigrateOptions struct {
	Service           string
	Environment       string
	ApproveProduction bool
	ForceRollback     bool
	TargetVersion     *int
}

// ServiceMigrationStatus stores one service migration summary.
type ServiceMigrationStatus struct {
	AppliedDDL  int    `json:"appliedDDL"`
	PendingDDL  int    `json:"pendingDDL"`
	AppliedDML  int    `json:"appliedDML"`
	PendingDML  int    `json:"pendingDML"`
	LastVersion uint   `json:"lastVersion"`
	Dirty       bool   `json:"dirty"`
	Error       string `json:"error,omitempty"`
}

// MigrationStatus stores status for all services.
type MigrationStatus struct {
	ServiceStatuses map[string]ServiceMigrationStatus `json:"serviceStatuses"`
}

// MigrationService defines operations for migration orchestration.
type MigrationService interface {
	MigrateUp(ctx context.Context, opts MigrateOptions) error
	MigrateDown(ctx context.Context, opts MigrateOptions) error
	GetStatus(ctx context.Context) (*MigrationStatus, error)
}

type migrationService struct {
	db       *sql.DB
	logger   *zap.Logger
	basePath string
}

// NewMigrationService constructs the migration service implementation.
func NewMigrationService(db *sql.DB, logger *zap.Logger, basePath string) MigrationService {
	if basePath == "" {
		basePath = "internals"
	}
	return &migrationService{db: db, logger: logger, basePath: basePath}
}

// MigrateUp applies DDL migrations and then DML migrations for the requested service scope.
func (m *migrationService) MigrateUp(ctx context.Context, opts MigrateOptions) error {
	ctx, span := otel.Tracer("migrations").Start(ctx, "migrations.up")
	defer span.End()

	env := opts.Environment
	if env == "" {
		env = GetEnvironment()
	}
	if err := ValidateProductionAccess(ctx, env, opts.ApproveProduction); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "production safety validation failed")
		return err
	}

	services, err := resolveServices(opts.Service)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid service")
		return err
	}

	for _, serviceName := range services {
		if err := m.applyServiceDDL(ctx, serviceName); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "ddl execution failed")
			return err
		}
	}

	for _, serviceName := range services {
		if err := m.applyServiceDML(ctx, serviceName, env); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "dml execution failed")
			return err
		}
	}

	span.SetAttributes(attribute.String("environment", env), attribute.Int("services", len(services)))
	return nil
}

// MigrateDown rolls back DDL migrations for one service.
func (m *migrationService) MigrateDown(ctx context.Context, opts MigrateOptions) error {
	ctx, span := otel.Tracer("migrations").Start(ctx, "migrations.down")
	defer span.End()

	if opts.Service == "" {
		return fmt.Errorf("service is required for rollback")
	}
	if !isAllowedService(opts.Service) {
		return fmt.Errorf("unsupported service: %s", opts.Service)
	}

	set, err := DiscoverServiceMigrations(m.serviceMigrationsPath(opts.Service))
	if err != nil {
		return err
	}
	if len(set.DDL) == 0 {
		return nil
	}

	applied, err := m.getAppliedDDLVersions(ctx, opts.Service)
	if err != nil {
		return err
	}
	if len(applied) == 0 {
		return nil
	}

	downTargets := m.computeRollbackTargets(applied, opts.TargetVersion)
	for _, version := range downTargets {
		migration, found := findMigrationByVersion(set.DDL, version)
		if !found {
			return fmt.Errorf("missing down migration file for version %d", version)
		}
		if err := m.executeDownMigration(ctx, opts.Service, migration); err != nil {
			if !opts.ForceRollback {
				return err
			}
			m.logger.Warn("rollback failed, retrying with force flag", zap.String("service", opts.Service), zap.Int("version", version), zap.Error(err))
			if retryErr := m.executeDownMigration(ctx, opts.Service, migration); retryErr != nil {
				return retryErr
			}
		}
	}

	return nil
}

// GetStatus reports applied and pending migrations by service.
func (m *migrationService) GetStatus(ctx context.Context) (*MigrationStatus, error) {
	result := &MigrationStatus{ServiceStatuses: map[string]ServiceMigrationStatus{}}

	for _, serviceName := range defaultServiceOrder {
		set, err := DiscoverServiceMigrations(m.serviceMigrationsPath(serviceName))
		if err != nil {
			result.ServiceStatuses[serviceName] = ServiceMigrationStatus{Error: err.Error()}
			continue
		}

		ddlApplied, dmlApplied, lastVersion, dirty, statusErr := m.getAppliedCounts(ctx, serviceName)
		if statusErr != nil {
			result.ServiceStatuses[serviceName] = ServiceMigrationStatus{Error: statusErr.Error()}
			continue
		}

		pendingDML := 0
		for _, migrations := range set.DML {
			pendingDML += len(migrations)
		}
		pendingDML -= dmlApplied
		if pendingDML < 0 {
			pendingDML = 0
		}

		result.ServiceStatuses[serviceName] = ServiceMigrationStatus{
			AppliedDDL:  ddlApplied,
			PendingDDL:  max(len(set.DDL)-ddlApplied, 0),
			AppliedDML:  dmlApplied,
			PendingDML:  pendingDML,
			LastVersion: lastVersion,
			Dirty:       dirty,
		}
	}

	return result, nil
}

func (m *migrationService) applyServiceDDL(ctx context.Context, serviceName string) error {
	if err := m.ensureServiceSchema(ctx, serviceName); err != nil {
		return err
	}

	set, err := DiscoverServiceMigrations(m.serviceMigrationsPath(serviceName))
	if err != nil {
		return err
	}

	applied := 0
	for _, migration := range set.DDL {
		startedAt := time.Now()

		tx, beginErr := m.db.BeginTx(ctx, nil)
		if beginErr != nil {
			return fmt.Errorf("begin ddl transaction for %s: %w", serviceName, beginErr)
		}
		if err := setSearchPath(ctx, tx, serviceName); err != nil {
			_ = tx.Rollback()
			return err
		}

		exists, existsErr := IsDDLMigrationApplied(ctx, tx, migration.Version)
		if existsErr != nil {
			_ = tx.Rollback()
			return existsErr
		}
		if exists {
			_ = tx.Rollback()
			continue
		}

		content, readErr := os.ReadFile(migration.UpPath)
		if readErr != nil {
			_ = tx.Rollback()
			return fmt.Errorf("read ddl migration %s: %w", migration.UpPath, readErr)
		}
		if _, execErr := tx.ExecContext(ctx, string(content)); execErr != nil {
			_ = tx.Rollback()
			return fmt.Errorf("execute ddl migration %s: %w", migration.UpPath, execErr)
		}

		checksum := checksumFor(content)
		if err := RecordDDLMigration(ctx, tx, migration.Version, migration.Name, time.Since(startedAt), true, "", checksum); err != nil {
			_ = tx.Rollback()
			return err
		}

		if commitErr := tx.Commit(); commitErr != nil {
			return fmt.Errorf("commit ddl migration %s: %w", migration.UpPath, commitErr)
		}

		applied++
		m.logger.Info("applied ddl migration", zap.String("service", serviceName), zap.Int("version", migration.Version), zap.String("name", migration.Name), zap.Int64("duration_ms", time.Since(startedAt).Milliseconds()))
	}

	m.logger.Info("ddl phase completed", zap.String("service", serviceName), zap.Int("applied", applied))
	return nil
}

func (m *migrationService) applyServiceDML(ctx context.Context, serviceName string, env string) error {
	set, err := DiscoverServiceMigrations(m.serviceMigrationsPath(serviceName))
	if err != nil {
		return err
	}

	migrations := set.DML[env]
	for _, migration := range migrations {
		startedAt := time.Now()

		tx, beginErr := m.db.BeginTx(ctx, nil)
		if beginErr != nil {
			return fmt.Errorf("begin dml transaction for %s: %w", serviceName, beginErr)
		}
		if err := setSearchPath(ctx, tx, serviceName); err != nil {
			_ = tx.Rollback()
			return err
		}

		exists, existsErr := IsDMLMigrationApplied(ctx, tx, migration.Version, env)
		if existsErr != nil {
			_ = tx.Rollback()
			return existsErr
		}
		if exists {
			_ = tx.Rollback()
			continue
		}

		content, readErr := os.ReadFile(migration.UpPath)
		if readErr != nil {
			_ = tx.Rollback()
			return fmt.Errorf("read dml migration %s: %w", migration.UpPath, readErr)
		}
		if _, execErr := tx.ExecContext(ctx, string(content)); execErr != nil {
			_ = tx.Rollback()
			return fmt.Errorf("execute dml migration %s: %w", migration.UpPath, execErr)
		}

		checksum := checksumFor(content)
		if err := RecordDMLMigration(ctx, tx, migration.Version, migration.Name, env, time.Since(startedAt), true, "", checksum); err != nil {
			_ = tx.Rollback()
			return err
		}

		if commitErr := tx.Commit(); commitErr != nil {
			return fmt.Errorf("commit dml migration %s: %w", migration.UpPath, commitErr)
		}

		m.logger.Info("applied dml migration", zap.String("service", serviceName), zap.String("environment", env), zap.Int("version", migration.Version), zap.String("name", migration.Name), zap.Int64("duration_ms", time.Since(startedAt).Milliseconds()))
	}

	return nil
}

func (m *migrationService) executeDownMigration(ctx context.Context, serviceName string, migration Migration) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin rollback transaction: %w", err)
	}
	if err := setSearchPath(ctx, tx, serviceName); err != nil {
		_ = tx.Rollback()
		return err
	}

	content, err := os.ReadFile(migration.DownPath)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("read down migration %s: %w", migration.DownPath, err)
	}
	if _, err := tx.ExecContext(ctx, string(content)); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("execute down migration %s: %w", migration.DownPath, err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM migrations_ddl WHERE version = $1`, migration.Version); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("delete rollback record version=%d: %w", migration.Version, err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit rollback transaction: %w", err)
	}

	m.logger.Info("rolled back migration", zap.String("service", serviceName), zap.Int("version", migration.Version), zap.String("name", migration.Name))
	return nil
}

func (m *migrationService) ensureServiceSchema(ctx context.Context, serviceName string) error {
	if !validSchemaRegexp.MatchString(serviceName) {
		return fmt.Errorf("invalid service schema name: %s", serviceName)
	}

	if _, err := m.db.ExecContext(ctx, fmt.Sprintf(`CREATE SCHEMA IF NOT EXISTS "%s"`, serviceName)); err != nil {
		return fmt.Errorf("create schema %s: %w", serviceName, err)
	}

	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin schema setup transaction: %w", err)
	}
	if err := setSearchPath(ctx, tx, serviceName); err != nil {
		_ = tx.Rollback()
		return err
	}

	const ddlTracking = `
		CREATE TABLE IF NOT EXISTS migrations_ddl (
			version BIGINT PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			executed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			execution_time_ms BIGINT NOT NULL,
			success BOOLEAN NOT NULL DEFAULT TRUE,
			error_message TEXT,
			executed_by TEXT,
			checksum TEXT
		);
		CREATE INDEX IF NOT EXISTS idx_migrations_ddl_executed_at ON migrations_ddl (executed_at DESC);
	`
	const dmlTracking = `
		CREATE TABLE IF NOT EXISTS migrations_dml (
			version BIGINT NOT NULL,
			name TEXT NOT NULL,
			environment TEXT NOT NULL,
			executed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			execution_time_ms BIGINT NOT NULL,
			success BOOLEAN NOT NULL DEFAULT TRUE,
			error_message TEXT,
			executed_by TEXT,
			checksum TEXT,
			PRIMARY KEY (version, environment)
		);
		CREATE INDEX IF NOT EXISTS idx_migrations_dml_env_executed ON migrations_dml (environment, executed_at DESC);
	`

	if _, err := tx.ExecContext(ctx, ddlTracking); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("create migrations_ddl in %s: %w", serviceName, err)
	}
	if _, err := tx.ExecContext(ctx, dmlTracking); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("create migrations_dml in %s: %w", serviceName, err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tracking tables for %s: %w", serviceName, err)
	}
	return nil
}

func (m *migrationService) getAppliedDDLVersions(ctx context.Context, serviceName string) ([]int, error) {
	query := fmt.Sprintf(`SELECT version FROM "%s".migrations_ddl WHERE success = TRUE ORDER BY version DESC`, serviceName)
	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query applied ddl versions for %s: %w", serviceName, err)
	}
	defer rows.Close()

	versions := []int{}
	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return nil, fmt.Errorf("scan applied ddl version for %s: %w", serviceName, err)
		}
		versions = append(versions, version)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate applied ddl versions for %s: %w", serviceName, err)
	}
	return versions, nil
}

func (m *migrationService) getAppliedCounts(ctx context.Context, serviceName string) (int, int, uint, bool, error) {
	ddlApplied := 0
	dmlApplied := 0
	lastVersion := uint(0)

	ddlQuery := fmt.Sprintf(`SELECT COUNT(*), COALESCE(MAX(version), 0) FROM "%s".migrations_ddl WHERE success = TRUE`, serviceName)
	if err := m.db.QueryRowContext(ctx, ddlQuery).Scan(&ddlApplied, &lastVersion); err != nil {
		if !strings.Contains(err.Error(), "does not exist") {
			return 0, 0, 0, false, fmt.Errorf("status ddl count for %s: %w", serviceName, err)
		}
	}

	dmlQuery := fmt.Sprintf(`SELECT COUNT(*) FROM "%s".migrations_dml WHERE success = TRUE`, serviceName)
	if err := m.db.QueryRowContext(ctx, dmlQuery).Scan(&dmlApplied); err != nil {
		if !strings.Contains(err.Error(), "does not exist") {
			return 0, 0, 0, false, fmt.Errorf("status dml count for %s: %w", serviceName, err)
		}
	}

	return ddlApplied, dmlApplied, lastVersion, false, nil
}

func (m *migrationService) computeRollbackTargets(applied []int, targetVersion *int) []int {
	if targetVersion == nil {
		if len(applied) == 0 {
			return nil
		}
		return []int{applied[0]}
	}

	targets := []int{}
	for _, version := range applied {
		if version >= *targetVersion {
			targets = append(targets, version)
		}
	}
	sort.Slice(targets, func(i, j int) bool { return targets[i] > targets[j] })
	return targets
}

func (m *migrationService) serviceMigrationsPath(serviceName string) string {
	return filepath.Join(m.basePath, serviceName, "migrations")
}

func resolveServices(requested string) ([]string, error) {
	if requested == "" {
		return defaultServiceOrder, nil
	}
	if !isAllowedService(requested) {
		return nil, fmt.Errorf("unsupported service: %s", requested)
	}
	return []string{requested}, nil
}

func isAllowedService(serviceName string) bool {
	for _, item := range defaultServiceOrder {
		if item == serviceName {
			return true
		}
	}
	return false
}

func setSearchPath(ctx context.Context, tx *sql.Tx, schemaName string) error {
	if !validSchemaRegexp.MatchString(schemaName) {
		return fmt.Errorf("invalid schema name: %s", schemaName)
	}
	if _, err := tx.ExecContext(ctx, fmt.Sprintf(`SET LOCAL search_path TO "%s", public`, schemaName)); err != nil {
		return fmt.Errorf("set search path for schema %s: %w", schemaName, err)
	}
	return nil
}

func checksumFor(content []byte) string {
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:])
}

func findMigrationByVersion(migrations []Migration, version int) (Migration, bool) {
	for _, migration := range migrations {
		if migration.Version == version {
			return migration, true
		}
	}
	return Migration{}, false
}

func max(left int, right int) int {
	if left > right {
		return left
	}
	return right
}
