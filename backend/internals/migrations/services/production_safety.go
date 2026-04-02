package services

import (
	"context"
	"fmt"
	"os"
)

var allowedEnvironments = map[string]struct{}{
	"local": {},
	"dev":   {},
	"stg":   {},
	"prd":   {},
}

// GetEnvironment resolves the active environment using APP_ENV, ENVIRONMENT, then local.
func GetEnvironment() string {
	if value := os.Getenv("APP_ENV"); value != "" {
		return value
	}
	if value := os.Getenv("ENVIRONMENT"); value != "" {
		return value
	}
	return "local"
}

// ValidateProductionAccess enforces explicit production safeguards for migrations.
func ValidateProductionAccess(_ context.Context, env string, approveProduction bool) error {
	if _, ok := allowedEnvironments[env]; !ok {
		return fmt.Errorf("invalid environment: %s. must be one of: local, dev, stg, prd", env)
	}

	if env != "prd" {
		return nil
	}

	if GetEnvironment() != "prd" || !approveProduction {
		return fmt.Errorf("production migration requires APP_ENV=prd and --approve-production flag")
	}

	return nil
}
