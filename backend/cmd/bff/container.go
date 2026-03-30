package bff

import (
	"context"
	"fmt"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humaecho"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	"go.uber.org/dig"
	"go.uber.org/zap"

	"github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/transport/http/controllers"
	bffmiddleware "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/transport/http/middleware"
	"github.com/ralvescosta/costa-financial-assistant/backend/pkgs/configs"
)

// run wires the dependency container and starts the BFF HTTP server.
func run(ctx context.Context) error {
	c := dig.New()

	// ─── Config ──────────────────────────────────────────────────────────────
	if err := c.Provide(func() (*configs.Config, error) {
		return configs.Load("BFF")
	}); err != nil {
		return fmt.Errorf("bff: provide config: %w", err)
	}

	// ─── Logger ──────────────────────────────────────────────────────────────
	if err := c.Provide(func(cfg *configs.Config) (*zap.Logger, error) {
		if cfg.Env == "production" {
			return zap.NewProduction()
		}
		return zap.NewDevelopment()
	}); err != nil {
		return fmt.Errorf("bff: provide logger: %w", err)
	}

	// ─── JWKS cache ──────────────────────────────────────────────────────────
	if err := c.Provide(bffmiddleware.NewJWKSCache); err != nil {
		return fmt.Errorf("bff: provide jwks cache: %w", err)
	}

	// ─── Controllers ─────────────────────────────────────────────────────────
	if err := c.Provide(controllers.NewDocumentsController); err != nil {
		return fmt.Errorf("bff: provide documents controller: %w", err)
	}

	// ─── Start ────────────────────────────────────────────────────────────────
	return c.Invoke(func(
		cfg *configs.Config,
		logger *zap.Logger,
		jwksCache *bffmiddleware.JWKSCache,
		docCtrl *controllers.DocumentsController,
	) error {
		e := echo.New()
		e.HideBanner = true
		e.HidePort = true

		// Middleware — OTel must be first for trace propagation
		e.Use(otelecho.Middleware(cfg.ServiceName))
		e.Use(middleware.RecoverWithConfig(middleware.RecoverConfig{
			LogErrorFunc: func(c echo.Context, err error, stack []byte) error {
				logger.Error("panic recovered", zap.Error(err))
				return nil
			},
		}))
		e.Use(middleware.RequestID())

		// Huma OpenAPI registration
		api := humaecho.New(e, huma.DefaultConfig("Costa Financial Assistant API", "1.0.0"))

		// Auth middleware group (applied per route via Huma middleware)
		authMiddleware := bffmiddleware.NewAuthMiddleware(jwksCache, logger)

		// Register controller routes
		docCtrl.Register(api, authMiddleware)

		addr := fmt.Sprintf("%s:%d", cfg.HTTP.Host, cfg.HTTP.Port)
		logger.Info("BFF HTTP server starting", zap.String("addr", addr))

		srv := &http.Server{
			Addr:    addr,
			Handler: e,
		}

		go func() {
			<-ctx.Done()
			_ = srv.Shutdown(context.Background())
		}()

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("bff: serve: %w", err)
		}
		return nil
	})
}
