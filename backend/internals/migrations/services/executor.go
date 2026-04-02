package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// Executor is a thin wrapper around golang-migrate.
type Executor struct {
	migrator *migrate.Migrate
}

// NewExecutor creates a migration executor using an existing database connection and source URL.
func NewExecutor(db *sql.DB, sourceURL string) (*Executor, error) {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return nil, fmt.Errorf("create postgres migration driver: %w", err)
	}

	migrator, err := migrate.NewWithDatabaseInstance(sourceURL, "postgres", driver)
	if err != nil {
		return nil, fmt.Errorf("create migrator: %w", err)
	}

	return &Executor{migrator: migrator}, nil
}

// UpAll executes all pending migrations.
func (e *Executor) UpAll(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := e.migrator.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("apply migrations up: %w", err)
	}
	return nil
}

// Down rolls back one migration.
func (e *Executor) Down(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := e.migrator.Steps(-1); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("rollback migration: %w", err)
	}
	return nil
}

// DownN rolls back N migrations.
func (e *Executor) DownN(ctx context.Context, n int) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if n < 1 {
		return fmt.Errorf("rollback steps must be >= 1")
	}
	if err := e.migrator.Steps(-n); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("rollback %d migrations: %w", n, err)
	}
	return nil
}

// Version returns the current migration version and dirty status.
func (e *Executor) Version(_ context.Context) (uint, bool, error) {
	version, dirty, err := e.migrator.Version()
	if err != nil && !errors.Is(err, migrate.ErrNilVersion) {
		return 0, false, fmt.Errorf("read migration version: %w", err)
	}
	if errors.Is(err, migrate.ErrNilVersion) {
		return 0, false, nil
	}
	return version, dirty, nil
}
