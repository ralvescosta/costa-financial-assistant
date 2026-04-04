package migrations

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"go.uber.org/dig"
	"go.uber.org/zap"

	migrationsvc "github.com/ralvescosta/costa-financial-assistant/backend/internals/migrations/services"
	"github.com/ralvescosta/costa-financial-assistant/backend/pkgs/configs"
	pkglogger "github.com/ralvescosta/costa-financial-assistant/backend/pkgs/logger"
)

type appContext struct {
	service migrationsvc.MigrationService
	logger  *zap.Logger
}

// run wires the dependency container and keeps the command runtime alive.
func run(ctx context.Context) (*appContext, error) {
	container := dig.New()

	if err := container.Provide(func() (*configs.Config, error) {
		return configs.Load("MIGRATIONS")
	}); err != nil {
		return nil, fmt.Errorf("migrations: provide config: %w", err)
	}

	if err := container.Provide(pkglogger.New); err != nil {
		return nil, fmt.Errorf("migrations: provide logger: %w", err)
	}

	if err := container.Provide(func(cfg *configs.Config) (*sql.DB, error) {
		db, err := sql.Open("postgres", cfg.DB.DSN)
		if err != nil {
			return nil, fmt.Errorf("migrations: open database: %w", err)
		}
		db.SetMaxOpenConns(cfg.DB.MaxOpenConns)
		db.SetMaxIdleConns(cfg.DB.MaxIdleConns)
		db.SetConnMaxLifetime(time.Duration(cfg.DB.ConnMaxLifetime) * time.Second)
		return db, nil
	}); err != nil {
		return nil, fmt.Errorf("migrations: provide db: %w", err)
	}

	if err := container.Provide(func(db *sql.DB, logger *zap.Logger) migrationsvc.MigrationService {
		return migrationsvc.NewMigrationService(db, logger, "internals")
	}); err != nil {
		return nil, fmt.Errorf("migrations: provide migration service: %w", err)
	}

	result := &appContext{}
	if err := container.Invoke(func(db *sql.DB, service migrationsvc.MigrationService, logger *zap.Logger) {
		result.service = service
		result.logger = logger

		go func() {
			<-ctx.Done()
			logger.Info("migrations: shutting down resources")
			if closeErr := db.Close(); closeErr != nil {
				logger.Error("migrations: close db failed", zap.Error(closeErr))
			}
			_ = logger.Sync()
		}()
	}); err != nil {
		return nil, fmt.Errorf("migrations: invoke dependencies: %w", err)
	}

	return result, nil
}

func runUp(ctx context.Context, opts migrationsvc.MigrateOptions) error {
	app, err := run(ctx)
	if err != nil {
		return err
	}
	return app.service.MigrateUp(ctx, opts)
}

func runDown(ctx context.Context, opts migrationsvc.MigrateOptions) error {
	app, err := run(ctx)
	if err != nil {
		return err
	}
	return app.service.MigrateDown(ctx, opts)
}

func runStatus(ctx context.Context, serviceFilter string, format string) error {
	app, err := run(ctx)
	if err != nil {
		return err
	}
	status, err := app.service.GetStatus(ctx)
	if err != nil {
		return err
	}
	return printMigrationStatus(format, serviceFilter, status)
}
