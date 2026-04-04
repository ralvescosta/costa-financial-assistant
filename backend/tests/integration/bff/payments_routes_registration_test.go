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

// TestPaymentsRouteIntegration verifies that the PaymentsRoute module correctly
// registers all four payment and cycle endpoints.
func TestPaymentsRouteIntegration(t *testing.T) {
	logger := zap.NewNop()
	route := bfftransportroutes.NewPaymentsRoute(stubPayments{}, logger)
	srv, api := buildBFFTestServer(t, route)

	t.Run("get-payment-dashboard route is registered", func(t *testing.T) {
		path, ok := api.OpenAPI().Paths["/api/v1/bills/payment-dashboard"]
		require.True(t, ok, "/api/v1/bills/payment-dashboard must be registered")
		assert.NotNil(t, path.Get, "get-payment-dashboard must be a GET")
		assert.Equal(t, "get-payment-dashboard", path.Get.OperationID)
	})

	t.Run("mark-bill-paid route is registered", func(t *testing.T) {
		path, ok := api.OpenAPI().Paths["/api/v1/bills/{billId}/mark-paid"]
		require.True(t, ok, "/api/v1/bills/{billId}/mark-paid must be registered")
		assert.NotNil(t, path.Post, "mark-bill-paid must be a POST")
		assert.Equal(t, "mark-bill-paid", path.Post.OperationID)
	})

	t.Run("get-preferred-payment-day route is registered", func(t *testing.T) {
		path, ok := api.OpenAPI().Paths["/api/v1/payment-cycle/preferred-day"]
		require.True(t, ok, "/api/v1/payment-cycle/preferred-day must be registered")
		assert.NotNil(t, path.Get, "get-preferred-payment-day must be a GET")
		assert.Equal(t, "get-preferred-payment-day", path.Get.OperationID)
	})

	t.Run("set-preferred-payment-day route is registered", func(t *testing.T) {
		path, ok := api.OpenAPI().Paths["/api/v1/payment-cycle/preferred-day"]
		require.True(t, ok, "/api/v1/payment-cycle/preferred-day must be registered")
		assert.NotNil(t, path.Put, "set-preferred-payment-day must be a PUT")
		assert.Equal(t, "set-preferred-payment-day", path.Put.OperationID)
	})

	t.Run("get-payment-dashboard endpoint is reachable", func(t *testing.T) {
		resp, err := srv.Client().Get(srv.URL + "/api/v1/bills/payment-dashboard")
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.NotEqual(t, http.StatusNotFound, resp.StatusCode)
		assert.NotEqual(t, http.StatusMethodNotAllowed, resp.StatusCode)
	})

	t.Run("set-preferred-payment-day endpoint is reachable", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPut, srv.URL+"/api/v1/payment-cycle/preferred-day", bytes.NewBufferString("{}"))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		resp, err := srv.Client().Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.NotEqual(t, http.StatusNotFound, resp.StatusCode)
		assert.NotEqual(t, http.StatusMethodNotAllowed, resp.StatusCode)
	})

	t.Run("get-payment-dashboard endpoint enforces auth semantics", func(t *testing.T) {
		resp, err := srv.Client().Get(srv.URL + "/api/v1/bills/payment-dashboard")
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("get-payment-dashboard accepts seeded owner context with omitted pagination", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, srv.URL+"/api/v1/bills/payment-dashboard", nil)
		require.NoError(t, err)
		req.Header.Set(testAuthHeader, "ralvescosta")

		resp, err := srv.Client().Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.NotEqual(t, http.StatusForbidden, resp.StatusCode)
		assert.NotEqual(t, http.StatusUnauthorized, resp.StatusCode)
	})
}
