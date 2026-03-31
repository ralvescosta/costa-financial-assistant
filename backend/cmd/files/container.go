package files

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

	filesrepo "github.com/ralvescosta/costa-financial-assistant/backend/internals/files/repositories"
	filessvc "github.com/ralvescosta/costa-financial-assistant/backend/internals/files/services"
	filesgrpc "github.com/ralvescosta/costa-financial-assistant/backend/internals/files/transport/grpc"
	"github.com/ralvescosta/costa-financial-assistant/backend/pkgs/configs"
	pkgotel "github.com/ralvescosta/costa-financial-assistant/backend/pkgs/otel"
	filesv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/files/v1"
)

// run wires the dependency container and starts the files gRPC service.
func run(ctx context.Context) error {
	if err := pkgotel.RegisterServiceMetrics("files"); err != nil {
		return fmt.Errorf("files: register metrics: %w", err)
	}

	c := dig.New()

	// ─── Config ──────────────────────────────────────────────────────────────
	if err := c.Provide(func() (*configs.Config, error) {
		return configs.Load("FILES")
	}); err != nil {
		return fmt.Errorf("files: provide config: %w", err)
	}

	// ─── Logger ──────────────────────────────────────────────────────────────
	if err := c.Provide(func(cfg *configs.Config) (*zap.Logger, error) {
		if cfg.Env == "production" {
			return zap.NewProduction()
		}
		return zap.NewDevelopment()
	}); err != nil {
		return fmt.Errorf("files: provide logger: %w", err)
	}

	// ─── Database ─────────────────────────────────────────────────────────────
	if err := c.Provide(func(cfg *configs.Config, logger *zap.Logger) (*sql.DB, error) {
		db, err := sql.Open("postgres", cfg.DB.DSN)
		if err != nil {
			return nil, fmt.Errorf("files: open db: %w", err)
		}
		db.SetMaxOpenConns(cfg.DB.MaxOpenConns)
		db.SetMaxIdleConns(cfg.DB.MaxIdleConns)
		db.SetConnMaxLifetime(time.Duration(cfg.DB.ConnMaxLifetime) * time.Second)
		logger.Info("files DB configured")
		return db, nil
	}); err != nil {
		return fmt.Errorf("files: provide db: %w", err)
	}

	// ─── Unit of Work ─────────────────────────────────────────────────────────
	if err := c.Provide(filesrepo.NewUnitOfWork); err != nil {
		return fmt.Errorf("files: provide unit of work: %w", err)
	}

	// ─── Repositories ─────────────────────────────────────────────────────────
	if err := c.Provide(filesrepo.NewDocumentRepository); err != nil {
		return fmt.Errorf("files: provide document repository: %w", err)
	}

	if err := c.Provide(filesrepo.NewAnalysisJobRepository); err != nil {
		return fmt.Errorf("files: provide analysis job repository: %w", err)
	}

	if err := c.Provide(filesrepo.NewBillRecordRepository); err != nil {
		return fmt.Errorf("files: provide bill record repository: %w", err)
	}

	if err := c.Provide(filesrepo.NewStatementRecordRepository); err != nil {
		return fmt.Errorf("files: provide statement record repository: %w", err)
	}

	if err := c.Provide(filesrepo.NewBankAccountRepository); err != nil {
		return fmt.Errorf("files: provide bank account repository: %w", err)
	}

	// ─── PDF Extractor (stub for development) ────────────────────────────────
	if err := c.Provide(filessvc.NewStubPDFExtractor); err != nil {
		return fmt.Errorf("files: provide pdf extractor: %w", err)
	}

	// ─── Services ─────────────────────────────────────────────────────────────
	if err := c.Provide(filessvc.NewDocumentService); err != nil {
		return fmt.Errorf("files: provide document service: %w", err)
	}

	if err := c.Provide(filessvc.NewExtractionService); err != nil {
		return fmt.Errorf("files: provide extraction service: %w", err)
	}

	if err := c.Provide(filessvc.NewBankAccountService); err != nil {
		return fmt.Errorf("files: provide bank account service: %w", err)
	}

	// ─── gRPC handler ─────────────────────────────────────────────────────────
	if err := c.Provide(func(
		svc filessvc.DocumentServiceIface,
		extSvc filessvc.ExtractionServiceIface,
		bankSvc filessvc.BankAccountServiceIface,
		logger *zap.Logger,
	) filesv1.FilesServiceServer {
		return filesgrpc.NewServer(svc, extSvc, bankSvc, logger)
	}); err != nil {
		return fmt.Errorf("files: provide grpc handler: %w", err)
	}

	// ─── Start ────────────────────────────────────────────────────────────────
	return c.Invoke(func(cfg *configs.Config, logger *zap.Logger, handler filesv1.FilesServiceServer) error {
		addr := fmt.Sprintf("%s:%d", cfg.GRPC.Host, cfg.GRPC.Port)
		lis, err := net.Listen("tcp", addr)
		if err != nil {
			return fmt.Errorf("files: listen %s: %w", addr, err)
		}

		srv := grpc.NewServer(
			grpc.StatsHandler(otelgrpc.NewServerHandler()),
		)
		filesv1.RegisterFilesServiceServer(srv, handler)
		reflection.Register(srv)

		logger.Info("files gRPC server starting", zap.String("addr", addr))

		go func() {
			<-ctx.Done()
			srv.GracefulStop()
		}()

		return srv.Serve(lis)
	})
}
