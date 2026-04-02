package controllers

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"

	bffinterfaces "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/interfaces"
	bffmiddleware "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/transport/http/middleware"
	views "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/transport/http/views"
)

// DocumentsController handles BFF document HTTP endpoints.
// It is a pure HTTP adapter: it extracts claims, delegates to DocumentsService, and returns view types.
type DocumentsController struct {
	BaseController
	svc bffinterfaces.DocumentsService
}

// NewDocumentsController constructs a DocumentsController.
func NewDocumentsController(logger *zap.Logger, validate *validator.Validate, svc bffinterfaces.DocumentsService) *DocumentsController {
	return &DocumentsController{BaseController: BaseController{logger: logger, validate: validate}, svc: svc}
}

// HandleUpload processes a raw PDF upload and registers the document record.
func (c *DocumentsController) HandleUpload(ctx context.Context, input *views.UploadDocumentInput) (*struct{ Body views.DocumentResponse }, error) {
	claims := bffmiddleware.ClaimsFromContext(ctx)
	if claims == nil {
		return nil, huma.Error403Forbidden("missing project context")
	}

	if len(input.RawBody) == 0 {
		return nil, huma.Error400BadRequest("request body must be the PDF file bytes")
	}

	doc, err := c.svc.UploadDocument(ctx, claims.GetProjectId(), claims.GetSubject(), input.FileName, input.RawBody)
	if err != nil {
		return nil, c.grpcToHumaError(err, "upload failed")
	}

	c.logger.Info("upload: document registered",
		zap.String("document_id", doc.ID),
		zap.String("project_id", claims.GetProjectId()))
	return &struct{ Body views.DocumentResponse }{Body: *doc}, nil
}

// HandleClassify updates a document's kind (bill or statement).
func (c *DocumentsController) HandleClassify(ctx context.Context, input *views.ClassifyDocumentInput) (*struct{ Body views.DocumentResponse }, error) {
	claims := bffmiddleware.ClaimsFromContext(ctx)
	if claims == nil {
		return nil, huma.Error403Forbidden("missing project context")
	}

	doc, err := c.svc.ClassifyDocument(ctx, claims.GetProjectId(), input.DocumentID, input.Body.Kind)
	if err != nil {
		return nil, c.grpcToHumaError(err, "classify failed")
	}

	c.logger.Info("classify: document classified",
		zap.String("document_id", input.DocumentID),
		zap.String("kind", input.Body.Kind))
	return &struct{ Body views.DocumentResponse }{Body: *doc}, nil
}

// HandleList returns project-scoped documents with pagination.
func (c *DocumentsController) HandleList(ctx context.Context, input *views.ListDocumentsInput) (*struct{ Body views.ListDocumentsResponse }, error) {
	claims := bffmiddleware.ClaimsFromContext(ctx)
	if claims == nil {
		return nil, huma.Error403Forbidden("missing project context")
	}

	pageSize := input.PageSize
	if pageSize == 0 {
		pageSize = 25
	}

	resp, err := c.svc.ListDocuments(ctx, claims.GetProjectId(), pageSize, input.PageToken)
	if err != nil {
		return nil, c.grpcToHumaError(err, "list documents failed")
	}

	return &struct{ Body views.ListDocumentsResponse }{Body: *resp}, nil
}

// HandleGet returns full document metadata including extraction fields.
func (c *DocumentsController) HandleGet(ctx context.Context, input *views.GetDocumentInput) (*struct{ Body views.DocumentDetailResponse }, error) {
	claims := bffmiddleware.ClaimsFromContext(ctx)
	if claims == nil {
		return nil, huma.Error403Forbidden("missing project context")
	}

	detail, err := c.svc.GetDocument(ctx, claims.GetProjectId(), input.DocumentID)
	if err != nil {
		return nil, c.grpcToHumaError(err, "get document failed")
	}

	return &struct{ Body views.DocumentDetailResponse }{Body: *detail}, nil
}


