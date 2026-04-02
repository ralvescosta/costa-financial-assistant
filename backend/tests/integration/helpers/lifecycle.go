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
	if empty, err := isMigrationDirEmpty(rawSource); err != nil {
		return err
	} else if empty {
		return nil
	}

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

	if err := mig.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate.Up: %w", err)
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
