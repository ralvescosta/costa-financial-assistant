package bills

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

	billsinterfaces "github.com/ralvescosta/costa-financial-assistant/backend/internals/bills/interfaces"
	billsrepo "github.com/ralvescosta/costa-financial-assistant/backend/internals/bills/repositories"
	billssvc "github.com/ralvescosta/costa-financial-assistant/backend/internals/bills/services"
	billsgrpc "github.com/ralvescosta/costa-financial-assistant/backend/internals/bills/transport/grpc"
	"github.com/ralvescosta/costa-financial-assistant/backend/pkgs/configs"
	pkgotel "github.com/ralvescosta/costa-financial-assistant/backend/pkgs/otel"
	billsv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/bills/v1"
)

// run wires the dependency container and starts the bills gRPC service.
func run(ctx context.Context) error {
	if err := pkgotel.RegisterServiceMetrics("bills"); err != nil {
		return fmt.Errorf("bills: register metrics: %w", err)
	}

	c := dig.New()

	// ─── Config ──────────────────────────────────────────────────────────────
	if err := c.Provide(func() (*configs.Config, error) {
		return configs.Load("BILLS")
	}); err != nil {
		return fmt.Errorf("bills: provide config: %w", err)
	}

	// ─── Logger ──────────────────────────────────────────────────────────────
	if err := c.Provide(func(cfg *configs.Config) (*zap.Logger, error) {
		if cfg.Env == "production" {
			return zap.NewProduction()
		}
		return zap.NewDevelopment()
	}); err != nil {
		return fmt.Errorf("bills: provide logger: %w", err)
	}

	// ─── Database ─────────────────────────────────────────────────────────────
	if err := c.Provide(func(cfg *configs.Config, logger *zap.Logger) (*sql.DB, error) {
		db, err := sql.Open("postgres", cfg.DB.DSN)
		if err != nil {
			return nil, fmt.Errorf("bills: open db: %w", err)
		}
		db.SetMaxOpenConns(cfg.DB.MaxOpenConns)
		db.SetMaxIdleConns(cfg.DB.MaxIdleConns)
		db.SetConnMaxLifetime(time.Duration(cfg.DB.ConnMaxLifetime) * time.Second)
		logger.Info("bills DB configured")
		return db, nil
	}); err != nil {
		return fmt.Errorf("bills: provide db: %w", err)
	}

	// ─── Repository ──────────────────────────────────────────────────────────
	if err := c.Provide(billsrepo.NewBillPaymentRepository); err != nil {
		return fmt.Errorf("bills: provide bill payment repository: %w", err)
	}

	// ─── Service ─────────────────────────────────────────────────────────────
	if err := c.Provide(billssvc.NewBillPaymentService); err != nil {
		return fmt.Errorf("bills: provide bill payment service: %w", err)
	}

	// ─── gRPC handler ─────────────────────────────────────────────────────────
	if err := c.Provide(func(svc billsinterfaces.BillPaymentService, logger *zap.Logger) billsv1.BillsServiceServer {
		return billsgrpc.NewServer(svc, logger)
	}); err != nil {
		return fmt.Errorf("bills: provide grpc handler: %w", err)
	}

	// ─── Start ────────────────────────────────────────────────────────────────
	return c.Invoke(func(cfg *configs.Config, logger *zap.Logger, handler billsv1.BillsServiceServer) error {
		addr := fmt.Sprintf("%s:%d", cfg.GRPC.Host, cfg.GRPC.Port)
		lis, err := net.Listen("tcp", addr)
		if err != nil {
			return fmt.Errorf("bills: listen %s: %w", addr, err)
		}

		srv := grpc.NewServer(
			grpc.StatsHandler(otelgrpc.NewServerHandler()),
		)
		billsv1.RegisterBillsServiceServer(srv, handler)
		reflection.Register(srv)

		logger.Info("bills gRPC server starting", zap.String("addr", addr))

		go func() {
			<-ctx.Done()
			logger.Info("bills: shutting down gRPC server")
			srv.GracefulStop()
			_ = logger.Sync()
		}()

		return srv.Serve(lis)
	})
}
