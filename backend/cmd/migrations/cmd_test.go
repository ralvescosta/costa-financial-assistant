package migrations

import (
	"testing"

	migrationsvc "github.com/ralvescosta/costa-financial-assistant/backend/internals/migrations/services"
)

func TestNewCommand_SubcommandsPresent(t *testing.T) {
	t.Parallel()

	// Given
	cmd := NewCommand()

	// When
	hasUp := false
	hasDown := false
	hasStatus := false
	hasValidate := false
	for _, sub := range cmd.Commands() {
		switch sub.Name() {
		case "up":
			hasUp = true
		case "down":
			hasDown = true
		case "status":
			hasStatus = true
		case "validate":
			hasValidate = true
		}
	}

	// Then
	if !hasUp || !hasDown || !hasStatus || !hasValidate {
		t.Fatalf("missing expected subcommands: up=%t down=%t status=%t validate=%t", hasUp, hasDown, hasStatus, hasValidate)
	}
}

func TestPrintMigrationStatus_Table(t *testing.T) {
	t.Parallel()

	// Given
	status := &migrationsvc.MigrationStatus{
		ServiceStatuses: map[string]migrationsvc.ServiceMigrationStatus{
			"bills": {
				AppliedDDL: 1,
			},
		},
	}

	// When
	err := printMigrationStatus("table", "", status)

	// Then
	if err != nil {
		t.Fatalf("print table failed: %v", err)
	}
}
