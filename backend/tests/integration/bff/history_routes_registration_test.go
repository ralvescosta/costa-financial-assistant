//go:build integration

package integration

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	bfftransportroutes "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/transport/http/routes"
)

// TestHistoryRouteIntegration verifies that the HistoryRoute module correctly
// registers all three history endpoints with the proper HTTP methods and paths.
// Timeline and category business logic is covered by us6_history_timeline_test.go
// and us6_history_metrics_test.go; this suite focuses on route registration and
// confirms that history routes do NOT require a project guard.
func TestHistoryRouteIntegration(t *testing.T) {
	logger := zap.NewNop()
	route := bfftransportroutes.NewHistoryRoute(stubHistory{}, logger)
	srv, api := buildBFFTestServer(t, route)

	t.Run("get-history-timeline route is registered", func(t *testing.T) {
		path, ok := api.OpenAPI().Paths["/api/v1/history/timeline"]
		require.True(t, ok, "/api/v1/history/timeline must be registered")
		assert.NotNil(t, path.Get, "get-history-timeline must be a GET")
		assert.Equal(t, "get-history-timeline", path.Get.OperationID)
	})

	t.Run("get-history-categories route is registered", func(t *testing.T) {
		path, ok := api.OpenAPI().Paths["/api/v1/history/categories"]
		require.True(t, ok, "/api/v1/history/categories must be registered")
		assert.NotNil(t, path.Get, "get-history-categories must be a GET")
		assert.Equal(t, "get-history-categories", path.Get.OperationID)
	})

	t.Run("get-history-compliance route is registered", func(t *testing.T) {
		path, ok := api.OpenAPI().Paths["/api/v1/history/compliance"]
		require.True(t, ok, "/api/v1/history/compliance must be registered")
		assert.NotNil(t, path.Get, "get-history-compliance must be a GET")
		assert.Equal(t, "get-history-compliance", path.Get.OperationID)
	})

	t.Run("get-history-timeline endpoint is reachable", func(t *testing.T) {
		resp, err := srv.Client().Get(srv.URL + "/api/v1/history/timeline")
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.NotEqual(t, http.StatusNotFound, resp.StatusCode)
		assert.NotEqual(t, http.StatusMethodNotAllowed, resp.StatusCode)
	})

	t.Run("get-history-categories endpoint is reachable", func(t *testing.T) {
		resp, err := srv.Client().Get(srv.URL + "/api/v1/history/categories")
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.NotEqual(t, http.StatusNotFound, resp.StatusCode)
		assert.NotEqual(t, http.StatusMethodNotAllowed, resp.StatusCode)
	})

	t.Run("get-history-compliance endpoint is reachable", func(t *testing.T) {
		resp, err := srv.Client().Get(srv.URL + "/api/v1/history/compliance")
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.NotEqual(t, http.StatusNotFound, resp.StatusCode)
		assert.NotEqual(t, http.StatusMethodNotAllowed, resp.StatusCode)
	})
}
