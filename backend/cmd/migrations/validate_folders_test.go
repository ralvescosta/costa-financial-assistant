package migrations

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateFolderStructure(t *testing.T) {
	t.Parallel()

	t.Run("GivenValidStructure_WhenValidate_ThenSucceeds", func(t *testing.T) {
		// Given
		baseDir := t.TempDir()
		services := []string{"bills", "files", "identity", "onboarding", "payments"}
		envs := []string{"local", "dev", "stg", "prd"}
		for _, serviceName := range services {
			ddlPath := filepath.Join(baseDir, serviceName, "migrations", "ddl")
			if err := os.MkdirAll(ddlPath, 0o755); err != nil {
				t.Fatalf("create ddl path: %v", err)
			}
			if err := os.WriteFile(filepath.Join(ddlPath, "000001_init.up.sql"), []byte("SELECT 1;"), 0o644); err != nil {
				t.Fatalf("write up file: %v", err)
			}
			if err := os.WriteFile(filepath.Join(ddlPath, "000001_init.down.sql"), []byte("SELECT 1;"), 0o644); err != nil {
				t.Fatalf("write down file: %v", err)
			}
			for _, env := range envs {
				dmlPath := filepath.Join(baseDir, serviceName, "migrations", "dml", env)
				if err := os.MkdirAll(dmlPath, 0o755); err != nil {
					t.Fatalf("create dml path: %v", err)
				}
			}
		}

		// When
		err := validateFolderStructure(baseDir, false)

		// Then
		if err != nil {
			t.Fatalf("expected success, got error: %v", err)
		}
	})
}
