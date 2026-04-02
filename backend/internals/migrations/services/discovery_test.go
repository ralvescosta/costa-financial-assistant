package services

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanMigrationFolder(t *testing.T) {
	t.Parallel()

	// Given a valid migration folder with one up/down pair
	tempDir := t.TempDir()
	folder := filepath.Join(tempDir, "ddl")
	if err := os.MkdirAll(folder, 0o755); err != nil {
		t.Fatalf("create temp folder: %v", err)
	}
	if err := os.WriteFile(filepath.Join(folder, "000001_create_table.up.sql"), []byte("SELECT 1;"), 0o644); err != nil {
		t.Fatalf("write up migration: %v", err)
	}
	if err := os.WriteFile(filepath.Join(folder, "000001_create_table.down.sql"), []byte("SELECT 1;"), 0o644); err != nil {
		t.Fatalf("write down migration: %v", err)
	}

	// When scanning the folder
	migrations, err := scanMigrationFolder(folder)

	// Then the pair is discovered and sorted
	if err != nil {
		t.Fatalf("unexpected scan error: %v", err)
	}
	if len(migrations) != 1 {
		t.Fatalf("expected one migration, got %d", len(migrations))
	}
	if migrations[0].Version != 1 {
		t.Fatalf("expected version 1, got %d", migrations[0].Version)
	}
}

func TestScanMigrationFolder_MissingPair(t *testing.T) {
	t.Parallel()

	// Given a folder with only one side of the migration pair
	tempDir := t.TempDir()
	folder := filepath.Join(tempDir, "ddl")
	if err := os.MkdirAll(folder, 0o755); err != nil {
		t.Fatalf("create temp folder: %v", err)
	}
	if err := os.WriteFile(filepath.Join(folder, "000001_create_table.up.sql"), []byte("SELECT 1;"), 0o644); err != nil {
		t.Fatalf("write up migration: %v", err)
	}

	// When scanning the folder
	_, err := scanMigrationFolder(folder)

	// Then an error is returned
	if err == nil {
		t.Fatal("expected missing pair error, got nil")
	}
}
