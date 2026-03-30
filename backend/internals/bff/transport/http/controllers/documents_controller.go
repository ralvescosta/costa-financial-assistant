package controllers

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"go.uber.org/zap"

	bffmiddleware "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/transport/http/middleware"
)

// DocumentsController handles HTTP routes for document upload and classification.
type DocumentsController struct {
	logger *zap.Logger
}

// NewDocumentsController constructs a DocumentsController.
func NewDocumentsController(logger *zap.Logger) *DocumentsController {
	return &DocumentsController{logger: logger}
}

// Register wires all document routes to the Huma API with the provided auth middleware.
func (c *DocumentsController) Register(api huma.API, auth func(huma.Context, func(huma.Context))) {
	huma.Register(api, huma.Operation{
		OperationID: "upload-document",
		Method:      http.MethodPost,
		Path:        "/api/v1/documents/upload",
		Summary:     "Upload PDF and create pending analysis document record",
		Tags:        []string{"documents"},
		Middlewares: huma.Middlewares{auth, bffmiddleware.NewProjectGuard("update", c.logger)},
	}, func(ctx context.Context, _ *struct{}) (*struct{ Body map[string]any }, error) {
		// TODO(T034): implement — delegate to files gRPC service
		c.logger.Info("upload-document called (stub)")
		return &struct{ Body map[string]any }{Body: map[string]any{"status": "not_implemented"}}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "classify-document",
		Method:      http.MethodPost,
		Path:        "/api/v1/documents/{documentId}/classify",
		Summary:     "Set document type and attribution metadata (bill/statement)",
		Tags:        []string{"documents"},
		Middlewares: huma.Middlewares{auth, bffmiddleware.NewProjectGuard("update", c.logger)},
	}, func(ctx context.Context, input *struct {
		DocumentID string `path:"documentId"`
	}) (*struct{ Body map[string]any }, error) {
		c.logger.Info("classify-document called (stub)", zap.String("document_id", input.DocumentID))
		return &struct{ Body map[string]any }{Body: map[string]any{"status": "not_implemented"}}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "list-documents",
		Method:      http.MethodGet,
		Path:        "/api/v1/documents",
		Summary:     "List project-scoped documents with status filters",
		Tags:        []string{"documents"},
		Middlewares: huma.Middlewares{auth, bffmiddleware.NewProjectGuard("read_only", c.logger)},
	}, func(ctx context.Context, _ *struct{}) (*struct{ Body map[string]any }, error) {
		c.logger.Info("list-documents called (stub)")
		return &struct{ Body map[string]any }{Body: map[string]any{"items": []any{}}}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "get-document",
		Method:      http.MethodGet,
		Path:        "/api/v1/documents/{documentId}",
		Summary:     "Fetch document details and extraction fields",
		Tags:        []string{"documents"},
		Middlewares: huma.Middlewares{auth, bffmiddleware.NewProjectGuard("read_only", c.logger)},
	}, func(ctx context.Context, input *struct {
		DocumentID string `path:"documentId"`
	}) (*struct{ Body map[string]any }, error) {
		c.logger.Info("get-document called (stub)", zap.String("document_id", input.DocumentID))
		return &struct{ Body map[string]any }{Body: map[string]any{"status": "not_implemented"}}, nil
	})
}
