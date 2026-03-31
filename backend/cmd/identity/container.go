package identity

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"net"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/dig"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/ralvescosta/costa-financial-assistant/backend/internals/identity/services"
	identitygrpc "github.com/ralvescosta/costa-financial-assistant/backend/internals/identity/transport/grpc"
	"github.com/ralvescosta/costa-financial-assistant/backend/pkgs/configs"
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
	if err := c.Provide(func(cfg *configs.Config) (*zap.Logger, error) {
		if cfg.Env == "production" {
			return zap.NewProduction()
		}
		return zap.NewDevelopment()
	}); err != nil {
		return fmt.Errorf("identity: provide logger: %w", err)
	}

	// ─── RSA signing key (Phase-1 generated in-process) ──────────────────────
	if err := c.Provide(func() (*rsa.PrivateKey, error) {
		return rsa.GenerateKey(rand.Reader, 2048)
	}); err != nil {
		return fmt.Errorf("identity: provide rsa key: %w", err)
	}

	// ─── Services ─────────────────────────────────────────────────────────────
	if err := c.Provide(services.NewTokenService); err != nil {
		return fmt.Errorf("identity: provide token service: %w", err)
	}

	// ─── gRPC handler ─────────────────────────────────────────────────────────
	if err := c.Provide(func(svc services.TokenServiceIface, logger *zap.Logger) identityv1.IdentityServiceServer {
		return identitygrpc.NewServer(svc, logger)
	}); err != nil {
		return fmt.Errorf("identity: provide grpc server: %w", err)
	}

	// ─── Start ────────────────────────────────────────────────────────────────
	return c.Invoke(func(cfg *configs.Config, logger *zap.Logger, handler identityv1.IdentityServiceServer) error {
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
			srv.GracefulStop()
		}()

		return srv.Serve(lis)
	})
}
