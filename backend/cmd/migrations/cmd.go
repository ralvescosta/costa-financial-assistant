package migrations

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	migrationsvc "github.com/ralvescosta/costa-financial-assistant/backend/internals/migrations/services"
	"github.com/spf13/cobra"
)

// NewCommand returns the cobra command for managing database migrations.
func NewCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "migrations",
		Short: "Manages database migrations",
		Long:  "Executes DDL and DML migrations with environment-aware safeguards and status reporting.",
	}

	rootCmd.AddCommand(newUpCommand(), newDownCommand(), newStatusCommand(), newValidateCommand())

	return rootCmd
}

func newUpCommand() *cobra.Command {
	var service string
	var env string
	var approveProduction bool

	cmd := &cobra.Command{
		Use:   "up",
		Short: "Apply pending migrations",
		Long:  "Applies DDL migrations first and then DML migrations for the selected service scope.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			opts := migrationsvc.MigrateOptions{
				Service:           service,
				Environment:       env,
				ApproveProduction: approveProduction,
			}
			return runUp(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVar(&service, "service", "", "Target service (bills, files, identity, onboarding, payments); empty applies all")
	cmd.Flags().StringVar(&env, "env", "", "Target environment (local, dev, stg, prd). Defaults from APP_ENV/ENVIRONMENT")
	cmd.Flags().BoolVar(&approveProduction, "approve-production", false, "Required explicit approval when env=prd")

	return cmd
}

func newDownCommand() *cobra.Command {
	var service string
	var env string
	var version int
	var forceRollback bool

	cmd := &cobra.Command{
		Use:   "down",
		Short: "Rollback applied migrations",
		Long:  "Rolls back the latest migration or all migrations down to a target version.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			var targetVersion *int
			if version > 0 {
				targetVersion = &version
			}

			opts := migrationsvc.MigrateOptions{
				Service:       service,
				Environment:   env,
				ForceRollback: forceRollback,
				TargetVersion: targetVersion,
			}
			return runDown(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVar(&service, "service", "", "Target service for rollback (required)")
	cmd.Flags().StringVar(&env, "env", "", "Target environment (optional, used by guards)")
	cmd.Flags().IntVar(&version, "version", 0, "Rollback down to this version (inclusive)")
	cmd.Flags().BoolVar(&forceRollback, "force-rollback", false, "Retry rollback once when the first attempt fails")
	_ = cmd.MarkFlagRequired("service")

	return cmd
}

func newStatusCommand() *cobra.Command {
	var service string
	var format string

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show migration status",
		Long:  "Displays applied and pending migration counts by service.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runStatus(cmd.Context(), service, format)
		},
	}

	cmd.Flags().StringVar(&service, "service", "", "Optional service filter")
	cmd.Flags().StringVar(&format, "format", "table", "Output format: table or json")

	return cmd
}

func printMigrationStatus(format string, serviceFilter string, status *migrationsvc.MigrationStatus) error {
	if strings.EqualFold(format, "json") {
		encoded, err := json.MarshalIndent(status, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal status json: %w", err)
		}
		fmt.Println(string(encoded))
		return nil
	}

	if !strings.EqualFold(format, "table") {
		return fmt.Errorf("invalid format: %s (supported: table, json)", format)
	}

	fmt.Println("SERVICE | APPLIED DDL | PENDING DDL | APPLIED DML | PENDING DML | LAST VERSION | DIRTY")
	keys := make([]string, 0, len(status.ServiceStatuses))
	for serviceName := range status.ServiceStatuses {
		if serviceFilter != "" && serviceFilter != serviceName {
			continue
		}
		keys = append(keys, serviceName)
	}
	sort.Strings(keys)

	for _, serviceName := range keys {
		current := status.ServiceStatuses[serviceName]
		fmt.Printf("%s | %d | %d | %d | %d | %d | %t\n", serviceName, current.AppliedDDL, current.PendingDDL, current.AppliedDML, current.PendingDML, current.LastVersion, current.Dirty)
	}

	return nil
}
