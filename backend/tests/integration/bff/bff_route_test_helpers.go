//go:build integration

package integration

import (
	"net/http/httptest"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humaecho"
	"github.com/labstack/echo/v4"

	bfftransportroutes "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/transport/http/routes"
)

// buildBFFTestServer constructs an in-process Echo/Huma server pre-seeded with
// the provided route modules. The noopAuth middleware passes every request
// through without JWT validation so route-level tests can focus on handler
// behaviour rather than token infrastructure.
//
// Returns an *httptest.Server (already started) and the backing huma.API so
// tests can inspect the registered operation set.
//
// Example:
//
//	srv, _ := buildBFFTestServer(t, myRoute)
//	defer srv.Close()
//	resp, err := srv.Client().Get(srv.URL + "/api/v1/documents")
func buildBFFTestServer(t *testing.T, routeModules ...bfftransportroutes.Route) (*httptest.Server, huma.API) {
	t.Helper()

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	api := humaecho.New(e, huma.DefaultConfig("BFF Test", "0.0.0"))

	noopAuth := func(ctx huma.Context, next func(huma.Context)) {
		next(ctx)
	}

	for _, r := range routeModules {
		r.Register(api, noopAuth)
	}

	srv := httptest.NewServer(e)
	t.Cleanup(srv.Close)

	return srv, api
}
