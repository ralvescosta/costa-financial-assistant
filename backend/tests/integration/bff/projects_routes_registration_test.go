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

// TestProjectsRouteIntegration verifies that the ProjectsRoute module correctly
// registers all four project endpoints with the proper HTTP methods and paths.
// Role enforcement and project isolation are covered by us7_role_enforcement_test.go
// and us7_project_isolation_test.go; this suite focuses on route registration.
func TestProjectsRouteIntegration(t *testing.T) {
	logger := zap.NewNop()
	route := bfftransportroutes.NewProjectsRoute(stubProjects{}, logger)
	srv, api := buildBFFTestServer(t, route)

	t.Run("get-current-project route is registered", func(t *testing.T) {
		path, ok := api.OpenAPI().Paths["/api/v1/projects/current"]
		require.True(t, ok, "/api/v1/projects/current must be registered")
		assert.NotNil(t, path.Get, "get-current-project must be a GET")
		assert.Equal(t, "get-current-project", path.Get.OperationID)
	})

	t.Run("list-project-members route is registered", func(t *testing.T) {
		path, ok := api.OpenAPI().Paths["/api/v1/projects/members"]
		require.True(t, ok, "/api/v1/projects/members must be registered")
		assert.NotNil(t, path.Get, "list-project-members must be a GET")
		assert.Equal(t, "list-project-members", path.Get.OperationID)
	})

	t.Run("invite-project-member route is registered", func(t *testing.T) {
		path, ok := api.OpenAPI().Paths["/api/v1/projects/members/invite"]
		require.True(t, ok, "/api/v1/projects/members/invite must be registered")
		assert.NotNil(t, path.Post, "invite-project-member must be a POST")
		assert.Equal(t, "invite-project-member", path.Post.OperationID)
	})

	t.Run("update-project-member-role route is registered", func(t *testing.T) {
		path, ok := api.OpenAPI().Paths["/api/v1/projects/members/{memberId}/role"]
		require.True(t, ok, "/api/v1/projects/members/{memberId}/role must be registered")
		assert.NotNil(t, path.Patch, "update-project-member-role must be a PATCH")
		assert.Equal(t, "update-project-member-role", path.Patch.OperationID)
	})

	t.Run("get-current-project endpoint is reachable", func(t *testing.T) {
		resp, err := srv.Client().Get(srv.URL + "/api/v1/projects/current")
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.NotEqual(t, http.StatusNotFound, resp.StatusCode)
		assert.NotEqual(t, http.StatusMethodNotAllowed, resp.StatusCode)
	})

	t.Run("invite-project-member endpoint is reachable", func(t *testing.T) {
		resp, err := srv.Client().Post(srv.URL+"/api/v1/projects/members/invite", "application/json", bytes.NewBufferString(`{}`))
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.NotEqual(t, http.StatusNotFound, resp.StatusCode)
		assert.NotEqual(t, http.StatusMethodNotAllowed, resp.StatusCode)
	})

	t.Run("get-current-project endpoint enforces auth semantics", func(t *testing.T) {
		resp, err := srv.Client().Get(srv.URL + "/api/v1/projects/current")
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})
}
