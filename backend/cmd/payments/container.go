package payments

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"go.uber.org/dig"
	"go.uber.org/zap"

	_ "github.com/lib/pq"

	paymentsrepo "github.com/ralvescosta/costa-financial-assistant/backend/internals/payments/repositories"
	paymentssvc "github.com/ralvescosta/costa-financial-assistant/backend/internals/payments/services"
	"github.com/ralvescosta/costa-financial-assistant/backend/pkgs/configs"
	pkgotel "github.com/ralvescosta/costa-financial-assistant/backend/pkgs/otel"
)

// run wires the dependency container and starts the payments service.
func run(ctx context.Context) error {
	if err := pkgotel.RegisterServiceMetrics("payments"); err != nil {
		return fmt.Errorf("payments: register metrics: %w", err)
	}

	c := dig.New()

	// ─── Config ──────────────────────────────────────────────────────────────
	if err := c.Provide(func() (*configs.Config, error) {
		return configs.Load("PAYMENTS")
	}); err != nil {
		return fmt.Errorf("payments: provide config: %w", err)
	}

	// ─── Logger ──────────────────────────────────────────────────────────────
	if err := c.Provide(func(cfg *configs.Config) (*zap.Logger, error) {
		if cfg.Env == "production" {
			return zap.NewProduction()
		}
		return zap.NewDevelopment()
	}); err != nil {
		return fmt.Errorf("payments: provide logger: %w", err)
	}

	// ─── Database ─────────────────────────────────────────────────────────────
	if err := c.Provide(func(cfg *configs.Config, logger *zap.Logger) (*sql.DB, error) {
		db, err := sql.Open("postgres", cfg.DB.DSN)
		if err != nil {
			return nil, fmt.Errorf("payments: open db: %w", err)
		}
		db.SetMaxOpenConns(cfg.DB.MaxOpenConns)
		db.SetMaxIdleConns(cfg.DB.MaxIdleConns)
		db.SetConnMaxLifetime(time.Duration(cfg.DB.ConnMaxLifetime) * time.Second)
		logger.Info("payments DB configured")
		return db, nil
	}); err != nil {
		return fmt.Errorf("payments: provide db: %w", err)
	}

	// ─── Repositories ─────────────────────────────────────────────────────────
	if err := c.Provide(paymentsrepo.NewPaymentCycleRepository); err != nil {
		return fmt.Errorf("payments: provide payment cycle repository: %w", err)
	}

	if err := c.Provide(paymentsrepo.NewReconciliationRepository); err != nil {
		return fmt.Errorf("payments: provide reconciliation repository: %w", err)
	}

	// ─── Services ─────────────────────────────────────────────────────────────
	if err := c.Provide(paymentssvc.NewPaymentCycleService); err != nil {
		return fmt.Errorf("payments: provide payment cycle service: %w", err)
	}

	if err := c.Provide(paymentssvc.NewReconciliationService); err != nil {
		return fmt.Errorf("payments: provide reconciliation service: %w", err)
	}

	// ─── Start ────────────────────────────────────────────────────────────────
	return c.Invoke(func(cfg *configs.Config, logger *zap.Logger) error {
		logger.Info("payments service started",
			zap.String("env", cfg.Env),
			zap.String("service", cfg.ServiceName))
		<-ctx.Done()
		return nil
	})
}
