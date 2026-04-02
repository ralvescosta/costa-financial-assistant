//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humaecho"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	bffmiddleware "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/transport/http/middleware"
	identityv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/identity/v1"
)

// TestUS7_RoleEnforcement verifies the role permission matrix via a lightweight
// in-memory Echo + Huma server with pre-injected JWT claims.
//
// Permission matrix under test:
//
//	read_only  → GET /test/read   ✓   POST /test/write  ✗
//	update     → GET /test/read   ✓   POST /test/write  ✗  (write requires "write")
//	write      → GET /test/read   ✓   POST /test/write  ✓
func TestUS7_RoleEnforcement(t *testing.T) {
	logger := zaptest.NewLogger(t)
	e, api := buildTestServer(t)

	// Register a read endpoint guarded by "read_only" (minimum privilege)
	huma.Register(api, huma.Operation{
		OperationID: "test-read",
		Method:      http.MethodGet,
		Path:        "/test/read",
		Middlewares: huma.Middlewares{bffmiddleware.NewProjectGuard("read_only", logger)},
	}, func(ctx context.Context, _ *struct{}) (*struct{ Body string }, error) {
		return &struct{ Body string }{Body: "ok"}, nil
	})

	// Register a write endpoint guarded by "write" (highest privilege required)
	huma.Register(api, huma.Operation{
		OperationID: "test-write",
		Method:      http.MethodPost,
		Path:        "/test/write",
		Middlewares: huma.Middlewares{bffmiddleware.NewProjectGuard("write", logger)},
	}, func(ctx context.Context, _ *struct{}) (*struct{ Body string }, error) {
		return &struct{ Body string }{Body: "ok"}, nil
	})

	srv := httptest.NewServer(e)
	t.Cleanup(srv.Close)

	const projectID = "00000000-0000-0000-0000-000000000010"

	tests := []struct {
		name         string
		role         string
		method       string
		path         string
		expectStatus int
	}{
		// read_only role
		{"read_only can read", "read_only", http.MethodGet, "/test/read", http.StatusOK},
		{"read_only cannot write", "read_only", http.MethodPost, "/test/write", http.StatusForbidden},

		// update role
		{"update can read", "update", http.MethodGet, "/test/read", http.StatusOK},
		{"update cannot write", "update", http.MethodPost, "/test/write", http.StatusForbidden},

		// write role
		{"write can read", "write", http.MethodGet, "/test/read", http.StatusOK},
		{"write can write", "write", http.MethodPost, "/test/write", http.StatusOK},

		// no claims — missing authentication
		{"no claims denied read", "", http.MethodGet, "/test/read", http.StatusForbidden},
		{"no claims denied write", "", http.MethodPost, "/test/write", http.StatusForbidden},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(tc.method, srv.URL+tc.path, nil)
			require.NoError(t, err)

			if tc.role != "" {
				req.Header.Set(testClaimsHeader, buildTestClaimsHeader(projectID, tc.role))
			}

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tc.expectStatus, resp.StatusCode,
				"role=%q endpoint=%s %s", tc.role, tc.method, tc.path)
		})
	}
}

// ─── test server helpers ──────────────────────────────────────────────────────

const testClaimsHeader = "X-Test-Claims"

// buildTestServer creates a minimal Echo + Huma server with a test-claims middleware
// that reads JWT claims from the X-Test-Claims header (base64-encoded JSON) and
// injects them into the context — bypassing real JWT validation for role tests.
func buildTestServer(t *testing.T) (*echo.Echo, huma.API) {
	t.Helper()
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	api := humaecho.New(e, huma.DefaultConfig("test-api", "1.0.0"))

	// Inject a synthetic claims-injection middleware that reads from X-Test-Claims.
	// This replaces the real auth middleware for role-guard unit tests.
	claimsInjector := func(ctx huma.Context, next func(huma.Context)) {
		raw := ctx.Header(testClaimsHeader)
		if raw == "" {
			next(ctx)
			return
		}
		var claims identityv1.JwtClaims
		if err := json.Unmarshal([]byte(raw), &claims); err != nil {
			next(ctx)
			return
		}
		newCtx := context.WithValue(ctx.Context(), bffmiddleware.ProjectContextKey, &claims)
		next(huma.WithContext(ctx, newCtx))
	}
	_ = claimsInjector // used inline in huma.Register Middlewares

	// Override the HumaAPI to embed the test claims injector per-route.
	// We do this by registering routes with the injector prepended.
	// Build the Echo server with the injector as a global Echo middleware.
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			raw := c.Request().Header.Get(testClaimsHeader)
			if raw != "" {
				var claims identityv1.JwtClaims
				if err := json.Unmarshal([]byte(raw), &claims); err == nil {
					newCtx := context.WithValue(c.Request().Context(), bffmiddleware.ProjectContextKey, &claims)
					c.SetRequest(c.Request().WithContext(newCtx))
				}
			}
			return next(c)
		}
	})

	return e, api
}

// buildTestClaimsHeader encodes minimal JwtClaims as JSON for the test claims header.
func buildTestClaimsHeader(projectID, role string) string {
	b, _ := json.Marshal(&identityv1.JwtClaims{
		ProjectId: projectID,
		Role:      role,
		Subject:   "00000000-0000-0000-0000-000000000001",
	})
	return string(b)
}

// drainClose reads and discards the body, then closes it.
func drainClose(body io.ReadCloser) {
	_, _ = io.Copy(io.Discard, body)
	_ = body.Close()
}
