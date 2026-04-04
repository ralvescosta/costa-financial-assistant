package payments

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/dig"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	_ "github.com/lib/pq"

	paymentsinterfaces "github.com/ralvescosta/costa-financial-assistant/backend/internals/payments/interfaces"
	paymentsrepo "github.com/ralvescosta/costa-financial-assistant/backend/internals/payments/repositories"
	paymentssvc "github.com/ralvescosta/costa-financial-assistant/backend/internals/payments/services"
	paymentsgrpc "github.com/ralvescosta/costa-financial-assistant/backend/internals/payments/transport/grpc"
	"github.com/ralvescosta/costa-financial-assistant/backend/pkgs/configs"
	pkglogger "github.com/ralvescosta/costa-financial-assistant/backend/pkgs/logger"
	pkgotel "github.com/ralvescosta/costa-financial-assistant/backend/pkgs/otel"
	paymentsv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/payments/v1"
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
	if err := c.Provide(pkglogger.New); err != nil {
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

	if err := c.Provide(paymentsrepo.NewHistoryRepository); err != nil {
		return fmt.Errorf("payments: provide history repository: %w", err)
	}

	// ─── Services ─────────────────────────────────────────────────────────────
	if err := c.Provide(paymentssvc.NewPaymentCycleService); err != nil {
		return fmt.Errorf("payments: provide payment cycle service: %w", err)
	}

	if err := c.Provide(paymentssvc.NewReconciliationService); err != nil {
		return fmt.Errorf("payments: provide reconciliation service: %w", err)
	}

	if err := c.Provide(paymentssvc.NewHistoryService); err != nil {
		return fmt.Errorf("payments: provide history service: %w", err)
	}

	// ─── gRPC handler ─────────────────────────────────────────────────────────
	if err := c.Provide(func(
		cycleSvc paymentsinterfaces.PaymentCycleService,
		historySvc paymentsinterfaces.HistoryService,
		reconciliationSvc paymentsinterfaces.ReconciliationService,
		logger *zap.Logger,
	) paymentsv1.PaymentsServiceServer {
		return paymentsgrpc.NewServer(cycleSvc, historySvc, reconciliationSvc, logger)
	}); err != nil {
		return fmt.Errorf("payments: provide grpc handler: %w", err)
	}

	// ─── Start ────────────────────────────────────────────────────────────────
	return c.Invoke(func(cfg *configs.Config, logger *zap.Logger, handler paymentsv1.PaymentsServiceServer) error {
		addr := fmt.Sprintf("%s:%d", cfg.GRPC.Host, cfg.GRPC.Port)
		lis, err := net.Listen("tcp", addr)
		if err != nil {
			return fmt.Errorf("payments: listen %s: %w", addr, err)
		}

		srv := grpc.NewServer(
			grpc.StatsHandler(otelgrpc.NewServerHandler()),
		)
		paymentsv1.RegisterPaymentsServiceServer(srv, handler)
		reflection.Register(srv)

		logger.Info("payments gRPC server starting", zap.String("addr", addr))

		go func() {
			<-ctx.Done()
			logger.Info("payments: shutting down gRPC server")
			srv.GracefulStop()
			_ = logger.Sync()
		}()

		return srv.Serve(lis)
	})
}
