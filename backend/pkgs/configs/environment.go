package configs

import (
	"fmt"
	"os"

	"go.uber.org/zap"
)

var validAppEnvironments = map[string]struct{}{
	"local": {},
	"dev":   {},
	"stg":   {},
	"prd":   {},
}

// GetAppEnvironment resolves the application environment and validates supported values.
func GetAppEnvironment(logger *zap.Logger) (string, error) {
	env := "local"
	source := "default"

	if value := os.Getenv("APP_ENV"); value != "" {
		env = value
		source = "APP_ENV"
	} else if value := os.Getenv("ENVIRONMENT"); value != "" {
		env = value
		source = "ENVIRONMENT"
	}

	if _, ok := validAppEnvironments[env]; !ok {
		return "", fmt.Errorf("invalid environment: %s. must be one of: local, dev, stg, prd", env)
	}

	if logger != nil {
		logger.Info("environment_determined", zap.String("env", env), zap.String("source", source))
	}

	return env, nil
}
