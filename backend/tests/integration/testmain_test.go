//go:build integration

package integration

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

var testDB *sql.DB

// TestMain provisions an ephemeral test database, runs all migrations, executes
// the test suite, then tears down the database.
func TestMain(m *testing.M) {
	dsn := testDSN()

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "integration: open db: %v\n", err)
		os.Exit(1)
	}

	if err := db.PingContext(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "integration: ping db: %v\n", err)
		os.Exit(1)
	}
	testDB = db

	// Run onboarding migrations (creates users/projects/project_members + seed data)
	if err := runMigrations(dsn, "file://../../internals/onboarding/migrations"); err != nil {
		fmt.Fprintf(os.Stderr, "integration: migrate onboarding: %v\n", err)
		os.Exit(1)
	}

	code := m.Run()

	_ = db.Close()
	os.Exit(code)
}

func testDSN() string {
	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5433/financial_test?sslmode=disable"
	}
	return dsn
}

func runMigrations(dsn, sourcePath string) error {
	// Derive a service-specific migrations table from the source path to prevent
	// version conflicts when multiple services share the same test database.
	// Path format: "file://../../internals/<service>/migrations"
	svcName := "schema"
	parts := strings.Split(strings.TrimPrefix(sourcePath, "file://"), "/")
	for i, p := range parts {
		if p == "internals" && i+1 < len(parts) {
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

	if err := mig.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migrate.Up: %w", err)
	}
	return nil
}
