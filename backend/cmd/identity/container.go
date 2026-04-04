package identity

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"database/sql"
	"fmt"
	"net"
	"time"

	_ "github.com/lib/pq"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/dig"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	identityrepo "github.com/ralvescosta/costa-financial-assistant/backend/internals/identity/repositories"
	"github.com/ralvescosta/costa-financial-assistant/backend/internals/identity/services"
	identitygrpc "github.com/ralvescosta/costa-financial-assistant/backend/internals/identity/transport/grpc"
	"github.com/ralvescosta/costa-financial-assistant/backend/pkgs/configs"
	pkglogger "github.com/ralvescosta/costa-financial-assistant/backend/pkgs/logger"
	pkgotel "github.com/ralvescosta/costa-financial-assistant/backend/pkgs/otel"
	identityv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/identity/v1"
)

// run wires the dependency container and starts the identity gRPC server.
func run(ctx context.Context) error {
	if err := pkgotel.RegisterServiceMetrics("identity"); err != nil {
		return fmt.Errorf("identity: register metrics: %w", err)
	}

	c := dig.New()

	// ─── Config ──────────────────────────────────────────────────────────────
	if err := c.Provide(func() (*configs.Config, error) {
		return configs.Load("IDENTITY")
	}); err != nil {
		return fmt.Errorf("identity: provide config: %w", err)
	}

	// ─── Logger ──────────────────────────────────────────────────────────────
	if err := c.Provide(pkglogger.New); err != nil {
		return fmt.Errorf("identity: provide logger: %w", err)
	}

	// ─── Database ────────────────────────────────────────────────────────────
	if err := c.Provide(func(cfg *configs.Config) (*sql.DB, error) {
		db, err := sql.Open("postgres", cfg.DB.DSN)
		if err != nil {
			return nil, fmt.Errorf("identity: open database: %w", err)
		}
		db.SetMaxOpenConns(cfg.DB.MaxOpenConns)
		db.SetMaxIdleConns(cfg.DB.MaxIdleConns)
		db.SetConnMaxLifetime(time.Duration(cfg.DB.ConnMaxLifetime) * time.Second)
		return db, nil
	}); err != nil {
		return fmt.Errorf("identity: provide db: %w", err)
	}

	// ─── RSA signing key (Phase-1 generated in-process) ──────────────────────
	if err := c.Provide(func() (*rsa.PrivateKey, error) {
		return rsa.GenerateKey(rand.Reader, 2048)
	}); err != nil {
		return fmt.Errorf("identity: provide rsa key: %w", err)
	}

	// ─── Repositories ────────────────────────────────────────────────────────
	if err := c.Provide(identityrepo.NewBootstrapAuthRepository); err != nil {
		return fmt.Errorf("identity: provide bootstrap auth repository: %w", err)
	}

	// ─── Services ─────────────────────────────────────────────────────────────
	if err := c.Provide(services.NewTokenServiceWithRepository); err != nil {
		return fmt.Errorf("identity: provide token service: %w", err)
	}

	// ─── gRPC handler ─────────────────────────────────────────────────────────
	if err := c.Provide(func(svc services.TokenServiceIface, logger *zap.Logger) identityv1.IdentityServiceServer {
		return identitygrpc.NewServer(svc, logger)
	}); err != nil {
		return fmt.Errorf("identity: provide grpc server: %w", err)
	}

	// ─── Start ────────────────────────────────────────────────────────────────
	return c.Invoke(func(cfg *configs.Config, logger *zap.Logger, db *sql.DB, handler identityv1.IdentityServiceServer) error {
		addr := fmt.Sprintf("%s:%d", cfg.GRPC.Host, cfg.GRPC.Port)
		lis, err := net.Listen("tcp", addr)
		if err != nil {
			return fmt.Errorf("identity: listen %s: %w", addr, err)
		}

		srv := grpc.NewServer(
			grpc.StatsHandler(otelgrpc.NewServerHandler()),
		)
		identityv1.RegisterIdentityServiceServer(srv, handler)
		reflection.Register(srv)

		logger.Info("identity gRPC server starting", zap.String("addr", addr))

		go func() {
			<-ctx.Done()
			logger.Info("identity: shutting down gRPC server")
			srv.GracefulStop()
			if err := db.Close(); err != nil {
				logger.Error("identity: close db failed", zap.Error(err))
			}
			_ = logger.Sync()
		}()

		return srv.Serve(lis)
	})
}
