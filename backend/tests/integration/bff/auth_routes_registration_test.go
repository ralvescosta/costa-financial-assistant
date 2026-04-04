//go:build integration

package integration

import (
	"bytes"
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	bfftransportroutes "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/transport/http/routes"
	views "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/transport/http/views"
)

type stubAuth struct{}

func (stubAuth) HandleLogin(_ context.Context, _ *views.LoginInput) (*views.LoginOutput, error) {
	return &views.LoginOutput{}, nil
}

func (stubAuth) HandleRefresh(_ context.Context, _ *views.RefreshInput) (*views.RefreshOutput, error) {
	return &views.RefreshOutput{}, nil
}

func TestAuthRouteIntegration(t *testing.T) {
	logger := zap.NewNop()
	route := bfftransportroutes.NewAuthRoute(stubAuth{}, logger)
	srv, api := buildBFFTestServer(t, route)

	t.Run("login route is registered", func(t *testing.T) {
		path, ok := api.OpenAPI().Paths["/api/auth/login"]
		require.True(t, ok, "/api/auth/login must be registered")
		assert.NotNil(t, path.Post)
		assert.Equal(t, "login-user", path.Post.OperationID)
	})

	t.Run("refresh route is registered", func(t *testing.T) {
		path, ok := api.OpenAPI().Paths["/api/auth/refresh"]
		require.True(t, ok, "/api/auth/refresh must be registered")
		assert.NotNil(t, path.Post)
		assert.Equal(t, "refresh-session", path.Post.OperationID)
	})

	t.Run("login endpoint is reachable", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPost, srv.URL+"/api/auth/login", bytes.NewBufferString(`{"username":"ralvescosta","password":"mudar@1234"}`))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err := srv.Client().Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.NotEqual(t, http.StatusNotFound, resp.StatusCode)
		assert.NotEqual(t, http.StatusMethodNotAllowed, resp.StatusCode)
	})
}
