//go:build integration

package integration

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	integrationhelpers "github.com/ralvescosta/costa-financial-assistant/backend/tests/integration/helpers"
)

var testDB *sql.DB
var suiteDSN string

func TestMain(m *testing.M) {
	ctx := context.Background()
	resources, err := integrationhelpers.SetupPostgresSuite(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "integration/payments: setup postgres: %v\n", err)
		os.Exit(1)
	}
	defer resources.Close(ctx)

	testDB = resources.DB
	suiteDSN = resources.DSN

	if err := integrationhelpers.RunMigrations(suiteDSN, "file://../../../internals/onboarding/migrations"); err != nil {
		fmt.Fprintf(os.Stderr, "integration/payments: migrate onboarding: %v\n", err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func testDSN() string {
	return suiteDSN
}

func runMigrations(dsn, sourcePath string) error {
	return integrationhelpers.RunMigrations(dsn, sourcePath)
}
