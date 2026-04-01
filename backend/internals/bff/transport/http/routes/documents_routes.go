package routes

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"go.uber.org/zap"

	controllers "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/transport/http/controllers"
	bffmiddleware "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/transport/http/middleware"
)

// Compile-time assertion: *DocumentsController satisfies DocumentsCapability.
var _ DocumentsCapability = (*controllers.DocumentsController)(nil)

// DocumentsRoute owns all Huma operation registrations for the documents resource.
type DocumentsRoute struct {
	ctrl   DocumentsCapability
	logger *zap.Logger
}

// NewDocumentsRoute constructs a DocumentsRoute.
func NewDocumentsRoute(ctrl DocumentsCapability, logger *zap.Logger) *DocumentsRoute {
	return &DocumentsRoute{ctrl: ctrl, logger: logger}
}

// Register wires all document routes to the Huma API.
func (r *DocumentsRoute) Register(api huma.API, auth func(huma.Context, func(huma.Context))) {
	huma.Register(api, huma.Operation{
		OperationID: "upload-document",
		Method:      http.MethodPost,
		Path:        "/api/v1/documents/upload",
		Summary:     "Upload PDF and create pending analysis document record",
		Description: "Accepts a raw PDF body and registers the document metadata scoped to the caller's project.",
		Tags:        []string{"documents"},
		Middlewares: huma.Middlewares{auth, bffmiddleware.NewProjectGuard("update", r.logger)},
	}, r.ctrl.HandleUpload)

	huma.Register(api, huma.Operation{
		OperationID: "classify-document",
		Method:      http.MethodPost,
		Path:        "/api/v1/documents/{documentId}/classify",
		Summary:     "Set document type and attribution metadata (bill/statement)",
		Description: "Updates the kind of an uploaded document to bill or statement.",
		Tags:        []string{"documents"},
		Middlewares: huma.Middlewares{auth, bffmiddleware.NewProjectGuard("update", r.logger)},
	}, r.ctrl.HandleClassify)

	huma.Register(api, huma.Operation{
		OperationID: "list-documents",
		Method:      http.MethodGet,
		Path:        "/api/v1/documents",
		Summary:     "List project-scoped documents with status filters",
		Description: "Returns documents in reverse-chronological order scoped to the caller's project.",
		Tags:        []string{"documents"},
		Middlewares: huma.Middlewares{auth, bffmiddleware.NewProjectGuard("read_only", r.logger)},
	}, r.ctrl.HandleList)

	huma.Register(api, huma.Operation{
		OperationID: "get-document",
		Method:      http.MethodGet,
		Path:        "/api/v1/documents/{documentId}",
		Summary:     "Fetch document details and extraction fields",
		Description: "Returns full document metadata for a project-scoped document.",
		Tags:        []string{"documents"},
		Middlewares: huma.Middlewares{auth, bffmiddleware.NewProjectGuard("read_only", r.logger)},
	}, r.ctrl.HandleGet)
}
