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

// TestSettingsRouteIntegration verifies that the SettingsRoute module correctly
// registers all three bank account endpoints with the proper HTTP methods and paths.
// Full CRUD business logic is covered by us3_bank_accounts_test.go; this suite
// focuses on route registration and HTTP reachability.
func TestSettingsRouteIntegration(t *testing.T) {
	logger := zap.NewNop()
	route := bfftransportroutes.NewSettingsRoute(stubSettings{}, logger)
	srv, api := buildBFFTestServer(t, route)

	t.Run("list-bank-accounts route is registered", func(t *testing.T) {
		path, ok := api.OpenAPI().Paths["/api/v1/bank-accounts"]
		require.True(t, ok, "/api/v1/bank-accounts must be registered")
		assert.NotNil(t, path.Get, "list-bank-accounts must be a GET")
		assert.Equal(t, "list-bank-accounts", path.Get.OperationID)
	})

	t.Run("create-bank-account route is registered", func(t *testing.T) {
		path, ok := api.OpenAPI().Paths["/api/v1/bank-accounts"]
		require.True(t, ok, "/api/v1/bank-accounts must be registered")
		assert.NotNil(t, path.Post, "create-bank-account must be a POST")
		assert.Equal(t, "create-bank-account", path.Post.OperationID)
	})

	t.Run("delete-bank-account route is registered", func(t *testing.T) {
		path, ok := api.OpenAPI().Paths["/api/v1/bank-accounts/{bankAccountId}"]
		require.True(t, ok, "/api/v1/bank-accounts/{bankAccountId} must be registered")
		assert.NotNil(t, path.Delete, "delete-bank-account must be a DELETE")
		assert.Equal(t, "delete-bank-account", path.Delete.OperationID)
	})

	t.Run("list-bank-accounts endpoint is reachable", func(t *testing.T) {
		resp, err := srv.Client().Get(srv.URL + "/api/v1/bank-accounts")
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.NotEqual(t, http.StatusNotFound, resp.StatusCode)
		assert.NotEqual(t, http.StatusMethodNotAllowed, resp.StatusCode)
	})

	t.Run("create-bank-account endpoint is reachable", func(t *testing.T) {
		resp, err := srv.Client().Post(srv.URL+"/api/v1/bank-accounts", "application/json", bytes.NewBufferString(`{}`))
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.NotEqual(t, http.StatusNotFound, resp.StatusCode)
		assert.NotEqual(t, http.StatusMethodNotAllowed, resp.StatusCode)
	})
}
