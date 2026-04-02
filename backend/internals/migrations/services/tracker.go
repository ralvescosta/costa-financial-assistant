package services

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// RecordDDLMigration inserts one successful or failed DDL migration execution record.
func RecordDDLMigration(ctx context.Context, tx *sql.Tx, version int, name string, duration time.Duration, success bool, errorMsg string, checksum string) error {
	const query = `
		INSERT INTO migrations_ddl (version, name, executed_at, execution_time_ms, success, error_message, executed_by, checksum)
		VALUES ($1, $2, NOW(), $3, $4, $5, CURRENT_USER, $6)
	`
	if _, err := tx.ExecContext(ctx, query, version, name, duration.Milliseconds(), success, nullString(errorMsg), nullString(checksum)); err != nil {
		return fmt.Errorf("insert ddl migration record: %w", err)
	}
	return nil
}

// RecordDMLMigration inserts one successful or failed DML migration execution record.
func RecordDMLMigration(ctx context.Context, tx *sql.Tx, version int, name string, env string, duration time.Duration, success bool, errorMsg string, checksum string) error {
	const query = `
		INSERT INTO migrations_dml (version, name, environment, executed_at, execution_time_ms, success, error_message, executed_by, checksum)
		VALUES ($1, $2, $3, NOW(), $4, $5, $6, CURRENT_USER, $7)
	`
	if _, err := tx.ExecContext(ctx, query, version, name, env, duration.Milliseconds(), success, nullString(errorMsg), nullString(checksum)); err != nil {
		return fmt.Errorf("insert dml migration record: %w", err)
	}
	return nil
}

// IsDDLMigrationApplied checks whether one DDL migration version is already recorded.
func IsDDLMigrationApplied(ctx context.Context, tx *sql.Tx, version int) (bool, error) {
	const query = `SELECT EXISTS(SELECT 1 FROM migrations_ddl WHERE version = $1 AND success = TRUE)`
	var exists bool
	if err := tx.QueryRowContext(ctx, query, version).Scan(&exists); err != nil {
		return false, fmt.Errorf("query ddl migration record: %w", err)
	}
	return exists, nil
}

// IsDMLMigrationApplied checks whether one DML migration version/environment tuple is already recorded.
func IsDMLMigrationApplied(ctx context.Context, tx *sql.Tx, version int, env string) (bool, error) {
	const query = `SELECT EXISTS(SELECT 1 FROM migrations_dml WHERE version = $1 AND environment = $2 AND success = TRUE)`
	var exists bool
	if err := tx.QueryRowContext(ctx, query, version, env).Scan(&exists); err != nil {
		return false, fmt.Errorf("query dml migration record: %w", err)
	}
	return exists, nil
}

func nullString(value string) interface{} {
	if value == "" {
		return nil
	}
	return value
}
