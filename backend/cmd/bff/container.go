package bff

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humaecho"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	"go.uber.org/dig"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/transport/http/controllers"
	bffmiddleware "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/transport/http/middleware"
	bfftransportroutes "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/transport/http/routes"
	"github.com/ralvescosta/costa-financial-assistant/backend/pkgs/configs"
	pkglogger "github.com/ralvescosta/costa-financial-assistant/backend/pkgs/logger"
	pkgotel "github.com/ralvescosta/costa-financial-assistant/backend/pkgs/otel"
	billsv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/bills/v1"
	filesv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/files/v1"
	identityv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/identity/v1"
	onboardingv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/onboarding/v1"
	paymentsv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/payments/v1"

	bffinterfaces "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/interfaces"
	bffservices "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/services"
)

// run wires the dependency container and starts the BFF HTTP server.
func run(ctx context.Context) error {
	if err := pkgotel.RegisterServiceMetrics("bff"); err != nil {
		return fmt.Errorf("bff: register metrics: %w", err)
	}

	c := dig.New()

	// ─── Config ──────────────────────────────────────────────────────────────
	if err := c.Provide(func() (*configs.Config, error) {
		return configs.Load("BFF")
	}); err != nil {
		return fmt.Errorf("bff: provide config: %w", err)
	}

	// ─── Logger ──────────────────────────────────────────────────────────────
	if err := c.Provide(pkglogger.New); err != nil {
		return fmt.Errorf("bff: provide logger: %w", err)
	}

	// ─── JWKS cache ──────────────────────────────────────────────────────────
	if err := c.Provide(bffmiddleware.NewJWKSCache); err != nil {
		return fmt.Errorf("bff: provide jwks cache: %w", err)
	}

	// ─── Files gRPC client ───────────────────────────────────────────────────
	if err := c.Provide(func(cfg *configs.Config, logger *zap.Logger) (filesv1.FilesServiceClient, error) {
		conn, err := grpc.NewClient(
			cfg.Services.FilesGRPCAddr,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			return nil, fmt.Errorf("bff: dial files grpc: %w", err)
		}
		logger.Info("files gRPC client connected", zap.String("addr", cfg.Services.FilesGRPCAddr))
		return filesv1.NewFilesServiceClient(conn), nil
	}); err != nil {
		return fmt.Errorf("bff: provide files client: %w", err)
	}

	// ─── Onboarding gRPC client ──────────────────────────────────────────────
	if err := c.Provide(func(cfg *configs.Config, logger *zap.Logger) (onboardingv1.OnboardingServiceClient, error) {
		conn, err := grpc.NewClient(
			cfg.Services.OnboardingGRPCAddr,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			return nil, fmt.Errorf("bff: dial onboarding grpc: %w", err)
		}
		logger.Info("onboarding gRPC client connected", zap.String("addr", cfg.Services.OnboardingGRPCAddr))
		return onboardingv1.NewOnboardingServiceClient(conn), nil
	}); err != nil {
		return fmt.Errorf("bff: provide onboarding client: %w", err)
	}

	// ─── Identity gRPC client ────────────────────────────────────────────────
	if err := c.Provide(func(cfg *configs.Config, logger *zap.Logger) (identityv1.IdentityServiceClient, error) {
		conn, err := grpc.NewClient(
			cfg.Services.IdentityGRPCAddr,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			return nil, fmt.Errorf("bff: dial identity grpc: %w", err)
		}
		logger.Info("identity gRPC client connected", zap.String("addr", cfg.Services.IdentityGRPCAddr))
		return identityv1.NewIdentityServiceClient(conn), nil
	}); err != nil {
		return fmt.Errorf("bff: provide identity client: %w", err)
	}

	// ─── Bills gRPC client ───────────────────────────────────────────────────
	if err := c.Provide(func(cfg *configs.Config, logger *zap.Logger) (billsv1.BillsServiceClient, error) {
		conn, err := grpc.NewClient(
			cfg.Services.BillsGRPCAddr,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			return nil, fmt.Errorf("bff: dial bills grpc: %w", err)
		}
		logger.Info("bills gRPC client connected", zap.String("addr", cfg.Services.BillsGRPCAddr))
		return billsv1.NewBillsServiceClient(conn), nil
	}); err != nil {
		return fmt.Errorf("bff: provide bills client: %w", err)
	}

	// ─── Payments gRPC client ────────────────────────────────────────────────
	if err := c.Provide(func(cfg *configs.Config, logger *zap.Logger) (paymentsv1.PaymentsServiceClient, error) {
		conn, err := grpc.NewClient(
			cfg.Services.PaymentsGRPCAddr,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			return nil, fmt.Errorf("bff: dial payments grpc: %w", err)
		}
		logger.Info("payments gRPC client connected", zap.String("addr", cfg.Services.PaymentsGRPCAddr))
		return paymentsv1.NewPaymentsServiceClient(conn), nil
	}); err != nil {
		return fmt.Errorf("bff: provide payments client: %w", err)
	}

	// ─── BFF interface adapters (concrete gRPC → BFF narrow interfaces) ────────
	if err := c.Provide(func(c filesv1.FilesServiceClient) bffinterfaces.FilesClient {
		return c
	}); err != nil {
		return fmt.Errorf("bff: provide files bff interface: %w", err)
	}

	if err := c.Provide(func(c onboardingv1.OnboardingServiceClient) bffinterfaces.OnboardingClient {
		return c
	}); err != nil {
		return fmt.Errorf("bff: provide onboarding bff interface: %w", err)
	}

	if err := c.Provide(func(c identityv1.IdentityServiceClient) bffinterfaces.IdentityClient {
		return c
	}); err != nil {
		return fmt.Errorf("bff: provide identity bff interface: %w", err)
	}

	// ─── BFF Services ────────────────────────────────────────────────────────
	if err := c.Provide(bffservices.NewAuthService); err != nil {
		return fmt.Errorf("bff: provide auth service: %w", err)
	}

	if err := c.Provide(bffservices.NewDocumentsService); err != nil {
		return fmt.Errorf("bff: provide documents service: %w", err)
	}

	if err := c.Provide(bffservices.NewProjectsService); err != nil {
		return fmt.Errorf("bff: provide projects service: %w", err)
	}

	if err := c.Provide(bffservices.NewSettingsService); err != nil {
		return fmt.Errorf("bff: provide settings service: %w", err)
	}

	if err := c.Provide(bffservices.NewPaymentsService); err != nil {
		return fmt.Errorf("bff: provide payments service: %w", err)
	}

	if err := c.Provide(bffservices.NewReconciliationService); err != nil {
		return fmt.Errorf("bff: provide reconciliation service: %w", err)
	}

	if err := c.Provide(bffservices.NewHistoryService); err != nil {
		return fmt.Errorf("bff: provide history service: %w", err)
	}

	// ─── Validator ───────────────────────────────────────────────────────────
	if err := c.Provide(validator.New); err != nil {
		return fmt.Errorf("bff: provide validator: %w", err)
	}

	// ─── Controllers (provided as capability interfaces) ────────────────────
	if err := c.Provide(controllers.NewAuthController, dig.As(new(bfftransportroutes.AuthCapability))); err != nil {
		return fmt.Errorf("bff: provide auth controller: %w", err)
	}

	if err := c.Provide(controllers.NewDocumentsController, dig.As(new(bfftransportroutes.DocumentsCapability))); err != nil {
		return fmt.Errorf("bff: provide documents controller: %w", err)
	}

	if err := c.Provide(controllers.NewSettingsController, dig.As(new(bfftransportroutes.SettingsCapability))); err != nil {
		return fmt.Errorf("bff: provide settings controller: %w", err)
	}

	if err := c.Provide(controllers.NewProjectsController, dig.As(new(bfftransportroutes.ProjectsCapability))); err != nil {
		return fmt.Errorf("bff: provide projects controller: %w", err)
	}

	if err := c.Provide(controllers.NewPaymentsController, dig.As(new(bfftransportroutes.PaymentsCapability))); err != nil {
		return fmt.Errorf("bff: provide payments controller: %w", err)
	}

	if err := c.Provide(controllers.NewReconciliationController, dig.As(new(bfftransportroutes.ReconciliationCapability))); err != nil {
		return fmt.Errorf("bff: provide reconciliation controller: %w", err)
	}

	if err := c.Provide(controllers.NewHistoryController, dig.As(new(bfftransportroutes.HistoryCapability))); err != nil {
		return fmt.Errorf("bff: provide history controller: %w", err)
	}

	// ─── Route modules ────────────────────────────────────────────────────────
	if err := c.Provide(bfftransportroutes.NewAuthRoute); err != nil {
		return fmt.Errorf("bff: provide auth route: %w", err)
	}

	if err := c.Provide(bfftransportroutes.NewDocumentsRoute); err != nil {
		return fmt.Errorf("bff: provide documents route: %w", err)
	}

	if err := c.Provide(bfftransportroutes.NewSettingsRoute); err != nil {
		return fmt.Errorf("bff: provide settings route: %w", err)
	}

	if err := c.Provide(bfftransportroutes.NewProjectsRoute); err != nil {
		return fmt.Errorf("bff: provide projects route: %w", err)
	}

	if err := c.Provide(bfftransportroutes.NewPaymentsRoute); err != nil {
		return fmt.Errorf("bff: provide payments route: %w", err)
	}

	if err := c.Provide(bfftransportroutes.NewReconciliationRoute); err != nil {
		return fmt.Errorf("bff: provide reconciliation route: %w", err)
	}

	if err := c.Provide(bfftransportroutes.NewHistoryRoute); err != nil {
		return fmt.Errorf("bff: provide history route: %w", err)
	}

	// ─── Start ────────────────────────────────────────────────────────────────
	return c.Invoke(func(
		cfg *configs.Config,
		logger *zap.Logger,
		jwksCache *bffmiddleware.JWKSCache,
		authRoute *bfftransportroutes.AuthRoute,
		docRoute *bfftransportroutes.DocumentsRoute,
		settingsRoute *bfftransportroutes.SettingsRoute,
		projectsRoute *bfftransportroutes.ProjectsRoute,
		paymentsRoute *bfftransportroutes.PaymentsRoute,
		reconRoute *bfftransportroutes.ReconciliationRoute,
		historyRoute *bfftransportroutes.HistoryRoute,
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
		e.Use(middleware.Gzip())
		e.Use(middleware.RequestID())
		e.Use(middleware.CORS())
		e.Use(middleware.RequestLogger())

		// Huma OpenAPI registration
		api := humaecho.New(e, huma.DefaultConfig("Costa Financial Assistant API", "1.0.0"))

		// Auth middleware group (applied per route via Huma middleware)
		authMiddleware := bffmiddleware.NewAuthMiddleware(jwksCache, logger)

		// Register route modules
		authRoute.Register(api, authMiddleware)
		docRoute.Register(api, authMiddleware)
		settingsRoute.Register(api, authMiddleware)
		projectsRoute.Register(api, authMiddleware)
		paymentsRoute.Register(api, authMiddleware)
		reconRoute.Register(api, authMiddleware)
		historyRoute.Register(api, authMiddleware)

		addr := fmt.Sprintf("%s:%d", cfg.HTTP.Host, cfg.HTTP.Port)
		logger.Info("BFF HTTP server starting", zap.String("addr", addr))

		srv := &http.Server{
			Addr:    addr,
			Handler: e,
		}

		go func() {
			<-ctx.Done()
			logger.Info("bff: shutting down HTTP server")
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			_ = srv.Shutdown(shutdownCtx)
			_ = logger.Sync()
		}()

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("bff: serve: %w", err)
		}
		return nil
	})
}
