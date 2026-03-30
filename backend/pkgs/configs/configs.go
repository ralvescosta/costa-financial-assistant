package configs

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config holds runtime configuration resolved by Viper.
// All ${SECRET_KEY} sentinel values are resolved via pkgs/secrets at startup.
type Config struct {
	Env         string `mapstructure:"ENV"`
	ServiceName string `mapstructure:"SERVICE_NAME"`

	HTTP     HTTPConfig
	GRPC     GRPCConfig
	DB       DBConfig
	OTEL     OTELConfig
	Services ServicesConfig
}

// ServicesConfig holds upstream gRPC service addresses consumed by the BFF and other services.
type ServicesConfig struct {
	IdentityGRPCAddr string `mapstructure:"IDENTITY_GRPC_ADDR"`
	FilesGRPCAddr    string `mapstructure:"FILES_GRPC_ADDR"`
	BillsGRPCAddr    string `mapstructure:"BILLS_GRPC_ADDR"`
}

// HTTPConfig holds HTTP server binding settings.
type HTTPConfig struct {
	Port int    `mapstructure:"HTTP_PORT"`
	Host string `mapstructure:"HTTP_HOST"`
}

// GRPCConfig holds gRPC server binding settings.
type GRPCConfig struct {
	Port int    `mapstructure:"GRPC_PORT"`
	Host string `mapstructure:"GRPC_HOST"`
}

// DBConfig holds PostgreSQL connection settings.
type DBConfig struct {
	DSN             string `mapstructure:"DB_DSN"`
	MaxOpenConns    int    `mapstructure:"DB_MAX_OPEN_CONNS"`
	MaxIdleConns    int    `mapstructure:"DB_MAX_IDLE_CONNS"`
	ConnMaxLifetime int    `mapstructure:"DB_CONN_MAX_LIFETIME_SECONDS"`
}

// OTELConfig holds OpenTelemetry exporter settings.
type OTELConfig struct {
	Endpoint string `mapstructure:"OTEL_EXPORTER_OTLP_ENDPOINT"`
	Insecure bool   `mapstructure:"OTEL_INSECURE"`
}

// Load initialises Viper, reads environment variables, and returns a Config.
// prefix is the env-var namespace prefix (e.g. "BFF" gives BFF_HTTP_PORT).
func Load(prefix string) (*Config, error) {
	v := viper.New()
	v.SetEnvPrefix(prefix)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Defaults
	v.SetDefault("ENV", "development")
	v.SetDefault("HTTP_PORT", 8080)
	v.SetDefault("HTTP_HOST", "0.0.0.0")
	v.SetDefault("GRPC_PORT", 9090)
	v.SetDefault("GRPC_HOST", "0.0.0.0")
	v.SetDefault("DB_MAX_OPEN_CONNS", 10)
	v.SetDefault("DB_MAX_IDLE_CONNS", 5)
	v.SetDefault("DB_CONN_MAX_LIFETIME_SECONDS", 300)
	v.SetDefault("OTEL_INSECURE", true)
	v.SetDefault("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317")
	v.SetDefault("IDENTITY_GRPC_ADDR", "localhost:9091")
	v.SetDefault("FILES_GRPC_ADDR", "localhost:9092")
	v.SetDefault("BILLS_GRPC_ADDR", "localhost:9093")

	cfg := &Config{
		Env:         v.GetString("ENV"),
		ServiceName: v.GetString("SERVICE_NAME"),
		HTTP: HTTPConfig{
			Port: v.GetInt("HTTP_PORT"),
			Host: v.GetString("HTTP_HOST"),
		},
		GRPC: GRPCConfig{
			Port: v.GetInt("GRPC_PORT"),
			Host: v.GetString("GRPC_HOST"),
		},
		DB: DBConfig{
			DSN:             v.GetString("DB_DSN"),
			MaxOpenConns:    v.GetInt("DB_MAX_OPEN_CONNS"),
			MaxIdleConns:    v.GetInt("DB_MAX_IDLE_CONNS"),
			ConnMaxLifetime: v.GetInt("DB_CONN_MAX_LIFETIME_SECONDS"),
		},
		OTEL: OTELConfig{
			Endpoint: v.GetString("OTEL_EXPORTER_OTLP_ENDPOINT"),
			Insecure: v.GetBool("OTEL_INSECURE"),
		},
		Services: ServicesConfig{
			IdentityGRPCAddr: v.GetString("IDENTITY_GRPC_ADDR"),
			FilesGRPCAddr:    v.GetString("FILES_GRPC_ADDR"),
			BillsGRPCAddr:    v.GetString("BILLS_GRPC_ADDR"),
		},
	}

	if cfg.ServiceName == "" {
		return nil, fmt.Errorf("configs: SERVICE_NAME is required")
	}

	return cfg, nil
}
