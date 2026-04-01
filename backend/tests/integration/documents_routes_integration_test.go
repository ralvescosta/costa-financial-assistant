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

// TestDocumentsRouteIntegration verifies that the DocumentsRoute module correctly
// registers all four document endpoints with the proper HTTP methods and paths.
func TestDocumentsRouteIntegration(t *testing.T) {
	logger := zap.NewNop()
	route := bfftransportroutes.NewDocumentsRoute(stubDocuments{}, logger)
	srv, api := buildBFFTestServer(t, route)

	t.Run("upload-document route is registered", func(t *testing.T) {
		path, ok := api.OpenAPI().Paths["/api/v1/documents/upload"]
		require.True(t, ok, "/api/v1/documents/upload must be registered")
		assert.NotNil(t, path.Post, "upload-document must be a POST")
		assert.Equal(t, "upload-document", path.Post.OperationID)
	})

	t.Run("classify-document route is registered", func(t *testing.T) {
		path, ok := api.OpenAPI().Paths["/api/v1/documents/{documentId}/classify"]
		require.True(t, ok, "/api/v1/documents/{documentId}/classify must be registered")
		assert.NotNil(t, path.Post, "classify-document must be a POST")
		assert.Equal(t, "classify-document", path.Post.OperationID)
	})

	t.Run("list-documents route is registered", func(t *testing.T) {
		path, ok := api.OpenAPI().Paths["/api/v1/documents"]
		require.True(t, ok, "/api/v1/documents must be registered")
		assert.NotNil(t, path.Get, "list-documents must be a GET")
		assert.Equal(t, "list-documents", path.Get.OperationID)
	})

	t.Run("get-document route is registered", func(t *testing.T) {
		path, ok := api.OpenAPI().Paths["/api/v1/documents/{documentId}"]
		require.True(t, ok, "/api/v1/documents/{documentId} must be registered")
		assert.NotNil(t, path.Get, "get-document must be a GET")
		assert.Equal(t, "get-document", path.Get.OperationID)
	})

	t.Run("list-documents endpoint is reachable", func(t *testing.T) {
		resp, err := srv.Client().Get(srv.URL + "/api/v1/documents")
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.NotEqual(t, http.StatusNotFound, resp.StatusCode)
		assert.NotEqual(t, http.StatusMethodNotAllowed, resp.StatusCode)
	})
}
