package logger

import (
	"fmt"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/ralvescosta/costa-financial-assistant/backend/pkgs/configs"
)

const (
	defaultLogLevel = "info"
	fieldService    = "service"
	fieldEnv        = "env"
)

// New builds a shared zap logger configured from the runtime config.
func New(cfg *configs.Config) (*zap.Logger, error) {
	if cfg == nil {
		return nil, fmt.Errorf("logger: config is required")
	}

	zapConfig, err := newZapConfig(cfg)
	if err != nil {
		return nil, err
	}

	builtLogger, err := zapConfig.Build()
	if err != nil {
		return nil, fmt.Errorf("logger: build zap logger: %w", err)
	}

	fields := make([]zap.Field, 0, 2)
	if cfg.ServiceName != "" {
		fields = append(fields, zap.String(fieldService, cfg.ServiceName))
	}

	normalizedEnv := normalizeEnvironment(cfg.Env)
	if normalizedEnv != "" {
		fields = append(fields, zap.String(fieldEnv, normalizedEnv))
	}

	return builtLogger.With(fields...), nil
}

func newZapConfig(cfg *configs.Config) (zap.Config, error) {
	var zapConfig zap.Config
	if isProductionEnvironment(cfg.Env) {
		zapConfig = zap.NewProductionConfig()
	} else {
		zapConfig = zap.NewDevelopmentConfig()
	}

	level, err := resolveLevel(cfg.LogLevel)
	if err != nil {
		return zap.Config{}, fmt.Errorf("logger: resolve log level %q: %w", cfg.LogLevel, err)
	}

	zapConfig.Level = zap.NewAtomicLevelAt(level)
	zapConfig.OutputPaths = []string{"stdout"}
	zapConfig.ErrorOutputPaths = []string{"stderr"}

	return zapConfig, nil
}

func resolveLevel(input string) (zapcore.Level, error) {
	normalizedLevel := strings.ToLower(strings.TrimSpace(input))
	if normalizedLevel == "" {
		normalizedLevel = defaultLogLevel
	}

	switch normalizedLevel {
	case "trace", "debug":
		return zap.DebugLevel, nil
	case "info":
		return zap.InfoLevel, nil
	case "warn", "warning":
		return zap.WarnLevel, nil
	case "error":
		return zap.ErrorLevel, nil
	case "dpanic":
		return zap.DPanicLevel, nil
	case "panic":
		return zap.PanicLevel, nil
	case "fatal":
		return zap.FatalLevel, nil
	default:
		return zap.InfoLevel, fmt.Errorf("unsupported log level %q", input)
	}
}

func isProductionEnvironment(env string) bool {
	switch normalizeEnvironment(env) {
	case "stg", "prd":
		return true
	default:
		return false
	}
}

func normalizeEnvironment(env string) string {
	normalizedEnv := strings.ToLower(strings.TrimSpace(env))

	switch normalizedEnv {
	case "", "local":
		return "local"
	case "dev", "development":
		return "dev"
	case "stg", "stage", "staging":
		return "stg"
	case "prd", "prod", "production":
		return "prd"
	default:
		return normalizedEnv
	}
}
