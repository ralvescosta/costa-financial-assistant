package migrations

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	migrationsvc "github.com/ralvescosta/costa-financial-assistant/backend/internals/migrations/services"
	"github.com/spf13/cobra"
)

func newValidateCommand() *cobra.Command {
	var autoCreate bool

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate migration folder structure",
		Long:  "Checks migration folder and file pair conventions across backend services.",
		RunE: func(_ *cobra.Command, _ []string) error {
			return validateFolderStructure("internals", autoCreate)
		},
	}

	cmd.Flags().BoolVar(&autoCreate, "auto-create", false, "Create missing migration folders for known services")

	return cmd
}

func validateFolderStructure(basePath string, autoCreate bool) error {
	services := []string{"bills", "files", "identity", "onboarding", "payments"}
	envs := []string{"local", "dev", "stg", "prd"}
	issues := []string{}

	for _, serviceName := range services {
		serviceMigrationPath := filepath.Join(basePath, serviceName, "migrations")
		ddlPath := filepath.Join(serviceMigrationPath, "ddl")
		if err := ensureDirectory(ddlPath, autoCreate); err != nil {
			issues = append(issues, err.Error())
			continue
		}

		for _, env := range envs {
			dmlPath := filepath.Join(serviceMigrationPath, "dml", env)
			if err := ensureDirectory(dmlPath, autoCreate); err != nil {
				issues = append(issues, err.Error())
			}
		}

		if _, err := migrationsvc.DiscoverServiceMigrations(serviceMigrationPath); err != nil {
			issues = append(issues, fmt.Sprintf("%s: %v", serviceName, err))
		}
	}

	if len(issues) > 0 {
		return fmt.Errorf("folder validation failed:\n%s", strings.Join(issues, "\n"))
	}

	return nil
}

func ensureDirectory(path string, autoCreate bool) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	}

	if !autoCreate {
		return fmt.Errorf("missing directory: %s", path)
	}

	if err := os.MkdirAll(path, 0o755); err != nil {
		return fmt.Errorf("create directory %s: %w", path, err)
	}

	return nil
}
