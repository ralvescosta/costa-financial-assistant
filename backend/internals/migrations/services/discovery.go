package services

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
)

var migrationFileRegexp = regexp.MustCompile(`^(\d{6})_(.+)\.(up|down)\.sql$`)

// Migration represents one versioned migration pair.
type Migration struct {
	Version  int
	Name     string
	UpPath   string
	DownPath string
}

// MigrationSet groups discovered DDL and DML migrations for one service.
type MigrationSet struct {
	Service string
	DDL     []Migration
	DML     map[string][]Migration
}

// DiscoverMigrations discovers migration sets for all supported services.
func DiscoverMigrations(basePath string) (map[string]*MigrationSet, error) {
	services := make(map[string]*MigrationSet, len(defaultServiceOrder))
	for _, serviceName := range defaultServiceOrder {
		set, err := DiscoverServiceMigrations(filepath.Join(basePath, serviceName, "migrations"))
		if err != nil {
			return nil, fmt.Errorf("discover migrations for %s: %w", serviceName, err)
		}
		set.Service = serviceName
		services[serviceName] = set
	}

	return services, nil
}

// DiscoverServiceMigrations discovers DDL and DML migrations for one service path.
func DiscoverServiceMigrations(servicePath string) (*MigrationSet, error) {
	ddl, err := ScanDDL(servicePath)
	if err != nil {
		return nil, err
	}

	envs := []string{"local", "dev", "stg", "prd"}
	dml := make(map[string][]Migration, len(envs))
	for _, env := range envs {
		scanned, scanErr := ScanDML(servicePath, env)
		if scanErr != nil {
			return nil, scanErr
		}
		dml[env] = scanned
	}

	return &MigrationSet{DDL: ddl, DML: dml}, nil
}

// ScanDDL scans the standard DDL migration folder.
func ScanDDL(servicePath string) ([]Migration, error) {
	return scanMigrationFolder(filepath.Join(servicePath, "ddl"))
}

// ScanDML scans one environment-specific DML migration folder.
func ScanDML(servicePath string, env string) ([]Migration, error) {
	return scanMigrationFolder(filepath.Join(servicePath, "dml", env))
}

func scanMigrationFolder(folderPath string) ([]Migration, error) {
	if _, err := os.Stat(folderPath); err != nil {
		if os.IsNotExist(err) {
			return []Migration{}, nil
		}
		return nil, fmt.Errorf("stat migration folder %s: %w", folderPath, err)
	}

	entries, err := os.ReadDir(folderPath)
	if err != nil {
		return nil, fmt.Errorf("read migration folder %s: %w", folderPath, err)
	}

	paired := map[int]*Migration{}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		matches := migrationFileRegexp.FindStringSubmatch(entry.Name())
		if len(matches) != 4 {
			continue
		}

		version, convErr := strconv.Atoi(matches[1])
		if convErr != nil {
			return nil, fmt.Errorf("invalid migration version in %s: %w", entry.Name(), convErr)
		}

		name := matches[2]
		direction := matches[3]
		current, ok := paired[version]
		if !ok {
			current = &Migration{Version: version, Name: name}
			paired[version] = current
		}
		if current.Name != name {
			return nil, fmt.Errorf("migration version %06d has mismatched names (%s and %s)", version, current.Name, name)
		}

		fullPath := filepath.Join(folderPath, entry.Name())
		if direction == "up" {
			current.UpPath = fullPath
		} else {
			current.DownPath = fullPath
		}
	}

	migrations := make([]Migration, 0, len(paired))
	for version, pair := range paired {
		if pair.UpPath == "" || pair.DownPath == "" {
			return nil, fmt.Errorf("migration version %06d is missing up/down pair", version)
		}
		migrations = append(migrations, *pair)
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}
