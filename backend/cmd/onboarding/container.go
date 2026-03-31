package onboarding

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"time"

	"go.uber.org/dig"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	_ "github.com/lib/pq"

	onboardingrepo "github.com/ralvescosta/costa-financial-assistant/backend/internals/onboarding/repositories"
	onboardingsvc "github.com/ralvescosta/costa-financial-assistant/backend/internals/onboarding/services"
	onboardinggrpc "github.com/ralvescosta/costa-financial-assistant/backend/internals/onboarding/transport/grpc"
	"github.com/ralvescosta/costa-financial-assistant/backend/pkgs/configs"
	pkgotel "github.com/ralvescosta/costa-financial-assistant/backend/pkgs/otel"
)

// run wires the dependency container and starts the onboarding gRPC service.
func run(ctx context.Context) error {
	if err := pkgotel.RegisterServiceMetrics("onboarding"); err != nil {
		return fmt.Errorf("onboarding: register metrics: %w", err)
	}

	c := dig.New()

	// ─── Config ──────────────────────────────────────────────────────────────
	if err := c.Provide(func() (*configs.Config, error) {
		return configs.Load("ONBOARDING")
	}); err != nil {
		return fmt.Errorf("onboarding: provide config: %w", err)
	}

	// ─── Logger ──────────────────────────────────────────────────────────────
	if err := c.Provide(func(cfg *configs.Config) (*zap.Logger, error) {
		if cfg.Env == "production" {
			return zap.NewProduction()
		}
		return zap.NewDevelopment()
	}); err != nil {
		return fmt.Errorf("onboarding: provide logger: %w", err)
	}

	// ─── Database ─────────────────────────────────────────────────────────────
	if err := c.Provide(func(cfg *configs.Config, logger *zap.Logger) (*sql.DB, error) {
		db, err := sql.Open("postgres", cfg.DB.DSN)
		if err != nil {
			return nil, fmt.Errorf("onboarding: open db: %w", err)
		}
		db.SetMaxOpenConns(cfg.DB.MaxOpenConns)
		db.SetMaxIdleConns(cfg.DB.MaxIdleConns)
		db.SetConnMaxLifetime(time.Duration(cfg.DB.ConnMaxLifetime) * time.Second)
		logger.Info("onboarding DB configured")
		return db, nil
	}); err != nil {
		return fmt.Errorf("onboarding: provide db: %w", err)
	}

	// ─── Repository ──────────────────────────────────────────────────────────
	if err := c.Provide(onboardingrepo.NewProjectMembersRepository); err != nil {
		return fmt.Errorf("onboarding: provide project members repository: %w", err)
	}

	// ─── Service ─────────────────────────────────────────────────────────────
	if err := c.Provide(onboardingsvc.NewProjectMembersService); err != nil {
		return fmt.Errorf("onboarding: provide project members service: %w", err)
	}

	// ─── gRPC server ─────────────────────────────────────────────────────────
	if err := c.Provide(func(svc onboardingsvc.ProjectMembersServiceIface, logger *zap.Logger) *grpc.Server {
		srv, _ := onboardinggrpc.NewServer(svc, logger)
		reflection.Register(srv)
		return srv
	}); err != nil {
		return fmt.Errorf("onboarding: provide grpc server: %w", err)
	}

	// ─── Start ────────────────────────────────────────────────────────────────
	return c.Invoke(func(cfg *configs.Config, logger *zap.Logger, srv *grpc.Server) error {
		addr := fmt.Sprintf("%s:%d", cfg.GRPC.Host, cfg.GRPC.Port)
		lis, err := net.Listen("tcp", addr)
		if err != nil {
			return fmt.Errorf("onboarding: listen %s: %w", addr, err)
		}

		logger.Info("onboarding gRPC server starting", zap.String("addr", addr))

		go func() {
			<-ctx.Done()
			logger.Info("onboarding: shutting down gRPC server")
			srv.GracefulStop()
			_ = logger.Sync()
		}()

		return srv.Serve(lis)
	})
}
