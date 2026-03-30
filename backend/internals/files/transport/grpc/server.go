package grpc

import (
	"context"
	"errors"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/ralvescosta/costa-financial-assistant/backend/internals/files/repositories"
	"github.com/ralvescosta/costa-financial-assistant/backend/internals/files/services"
	commonv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/common/v1"
	filesv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/files/v1"
)

// Server implements filesv1.FilesServiceServer.
type Server struct {
	filesv1.UnimplementedFilesServiceServer
	svc    services.DocumentServiceIface
	logger *zap.Logger
}

// NewServer constructs a files gRPC server.
func NewServer(svc services.DocumentServiceIface, logger *zap.Logger) *Server {
	return &Server{svc: svc, logger: logger}
}

// UploadDocument registers a PDF upload and persists metadata.
func (s *Server) UploadDocument(ctx context.Context, req *filesv1.UploadDocumentRequest) (*filesv1.UploadDocumentResponse, error) {
	if req.GetCtx() == nil || req.GetCtx().GetProjectId() == "" {
		return nil, status.Error(codes.InvalidArgument, "project_id is required")
	}
	if req.GetFileName() == "" || req.GetFileHash() == "" {
		return nil, status.Error(codes.InvalidArgument, "file_name and file_hash are required")
	}

	input := &services.UploadDocumentInput{
		ProjectID:       req.GetCtx().GetProjectId(),
		UploadedBy:      req.GetAudit().GetPerformedBy(),
		FileName:        req.GetFileName(),
		FileHash:        req.GetFileHash(),
		StorageProvider: req.GetStorageProvider(),
		StorageKey:      req.GetStorageKey(),
	}

	doc, err := s.svc.UploadDocument(ctx, input)
	if err != nil {
		if errors.Is(err, repositories.ErrDuplicateDocument) {
			return nil, status.Error(codes.AlreadyExists, "document already uploaded in this project")
		}
		s.logger.Error("grpc.UploadDocument failed",
			zap.String("project_id", input.ProjectID),
			zap.Error(err))
		return nil, status.Error(codes.Internal, "upload failed")
	}
	return &filesv1.UploadDocumentResponse{Document: doc}, nil
}

// ClassifyDocument updates the document kind.
func (s *Server) ClassifyDocument(ctx context.Context, req *filesv1.ClassifyDocumentRequest) (*filesv1.ClassifyDocumentResponse, error) {
	if req.GetCtx() == nil || req.GetCtx().GetProjectId() == "" {
		return nil, status.Error(codes.InvalidArgument, "project_id is required")
	}
	if req.GetDocumentId() == "" {
		return nil, status.Error(codes.InvalidArgument, "document_id is required")
	}

	doc, err := s.svc.ClassifyDocument(ctx, req.GetCtx().GetProjectId(), req.GetDocumentId(), req.GetKind())
	if err != nil {
		if errors.Is(err, repositories.ErrDocumentNotFound) {
			return nil, status.Error(codes.NotFound, "document not found")
		}
		s.logger.Error("grpc.ClassifyDocument failed",
			zap.String("document_id", req.GetDocumentId()),
			zap.Error(err))
		return nil, status.Error(codes.Internal, "classify failed")
	}
	return &filesv1.ClassifyDocumentResponse{Document: doc}, nil
}

// GetDocument returns a single project-scoped document.
func (s *Server) GetDocument(ctx context.Context, req *filesv1.GetDocumentRequest) (*filesv1.GetDocumentResponse, error) {
	if req.GetCtx() == nil || req.GetCtx().GetProjectId() == "" {
		return nil, status.Error(codes.InvalidArgument, "project_id is required")
	}
	if req.GetDocumentId() == "" {
		return nil, status.Error(codes.InvalidArgument, "document_id is required")
	}

	doc, err := s.svc.GetDocument(ctx, req.GetCtx().GetProjectId(), req.GetDocumentId())
	if err != nil {
		if errors.Is(err, repositories.ErrDocumentNotFound) {
			return nil, status.Error(codes.NotFound, "document not found")
		}
		s.logger.Error("grpc.GetDocument failed",
			zap.String("document_id", req.GetDocumentId()),
			zap.Error(err))
		return nil, status.Error(codes.Internal, "get document failed")
	}
	return &filesv1.GetDocumentResponse{Document: doc}, nil
}

// ListDocuments returns project-scoped documents with keyset pagination.
func (s *Server) ListDocuments(ctx context.Context, req *filesv1.ListDocumentsRequest) (*filesv1.ListDocumentsResponse, error) {
	if req.GetCtx() == nil || req.GetCtx().GetProjectId() == "" {
		return nil, status.Error(codes.InvalidArgument, "project_id is required")
	}

	pageSize := int32(25)
	if req.GetPagination() != nil && req.GetPagination().GetPageSize() > 0 {
		pageSize = req.GetPagination().GetPageSize()
	}

	docs, err := s.svc.ListDocuments(ctx, req.GetCtx().GetProjectId(), pageSize, req.GetPagination().GetPageToken())
	if err != nil {
		s.logger.Error("grpc.ListDocuments failed",
			zap.String("project_id", req.GetCtx().GetProjectId()),
			zap.Error(err))
		return nil, status.Error(codes.Internal, "list documents failed")
	}

	resp := &filesv1.ListDocumentsResponse{Documents: docs}
	// Set next page token when the page is full — more records may exist.
	if int32(len(docs)) == pageSize && len(docs) > 0 {
		resp.Pagination = &commonv1.PaginationResult{
			NextPageToken: docs[len(docs)-1].UploadedAt,
		}
	}
	return resp, nil
}
