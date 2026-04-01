//go:build integration

package integration

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	bfftransportroutes "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/transport/http/routes"
)

// TestReconciliationRouteIntegration verifies that the ReconciliationRoute module
// correctly registers both reconciliation endpoints.
func TestReconciliationRouteIntegration(t *testing.T) {
	logger := zap.NewNop()
	route := bfftransportroutes.NewReconciliationRoute(stubReconciliation{}, logger)
	srv, api := buildBFFTestServer(t, route)

	t.Run("get-reconciliation-summary route is registered", func(t *testing.T) {
		path, ok := api.OpenAPI().Paths["/api/v1/reconciliation/summary"]
		require.True(t, ok, "/api/v1/reconciliation/summary must be registered")
		assert.NotNil(t, path.Get, "get-reconciliation-summary must be a GET")
		assert.Equal(t, "get-reconciliation-summary", path.Get.OperationID)
	})

	t.Run("create-reconciliation-link route is registered", func(t *testing.T) {
		path, ok := api.OpenAPI().Paths["/api/v1/reconciliation/links"]
		require.True(t, ok, "/api/v1/reconciliation/links must be registered")
		assert.NotNil(t, path.Post, "create-reconciliation-link must be a POST")
		assert.Equal(t, "create-reconciliation-link", path.Post.OperationID)
	})

	t.Run("get-reconciliation-summary endpoint is reachable", func(t *testing.T) {
		resp, err := srv.Client().Get(srv.URL + "/api/v1/reconciliation/summary")
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.NotEqual(t, http.StatusNotFound, resp.StatusCode)
		assert.NotEqual(t, http.StatusMethodNotAllowed, resp.StatusCode)
	})

	t.Run("create-reconciliation-link endpoint is reachable", func(t *testing.T) {
		resp, err := srv.Client().Post(srv.URL+"/api/v1/reconciliation/links", "application/json", bytes.NewBufferString("{}"))
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.NotEqual(t, http.StatusNotFound, resp.StatusCode)
		assert.NotEqual(t, http.StatusMethodNotAllowed, resp.StatusCode)
	})
}
