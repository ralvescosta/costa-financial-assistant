// Package grpc implements the gRPC server for the bills service.
package grpc

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	billsinterfaces "github.com/ralvescosta/costa-financial-assistant/backend/internals/bills/interfaces"
	apperrors "github.com/ralvescosta/costa-financial-assistant/backend/pkgs/errors"
	billsv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/bills/v1"
	commonv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/common/v1"
)

// Server implements billsv1.BillsServiceServer backed by the BillPaymentService interface.
type Server struct {
	billsv1.UnimplementedBillsServiceServer
	svc    billsinterfaces.BillPaymentService
	logger *zap.Logger
}

// NewServer constructs a bills gRPC server.
func NewServer(svc billsinterfaces.BillPaymentService, logger *zap.Logger) *Server {
	return &Server{svc: svc, logger: logger}
}

// GetPaymentDashboard returns outstanding and overdue bills for the project's active cycle.
func (s *Server) GetPaymentDashboard(ctx context.Context, req *billsv1.GetPaymentDashboardRequest) (*billsv1.GetPaymentDashboardResponse, error) {
	pc := req.GetCtx()
	if pc == nil || pc.GetProjectId() == "" {
		return nil, status.Error(codes.InvalidArgument, "project_id is required")
	}

	var pageSize int32 = 20
	var pageToken string
	if p := req.GetPagination(); p != nil {
		if p.GetPageSize() > 0 {
			pageSize = p.GetPageSize()
		}
		pageToken = p.GetPageToken()
	}

	entries, nextToken, err := s.svc.GetPaymentDashboard(
		ctx,
		pc.GetProjectId(),
		req.GetCycleStart(),
		req.GetCycleEnd(),
		pageSize,
		pageToken,
	)
	if err != nil {
		if appErr := apperrors.AsAppError(err); appErr != nil {
			return nil, toGRPCStatusError(appErr)
		}
		s.logger.Error("grpc.GetPaymentDashboard failed",
			zap.String("project_id", pc.GetProjectId()),
			zap.Error(err))
		return nil, status.Error(codes.Internal, "internal service error")
	}

	var pagination *commonv1.PaginationResult
	if nextToken != "" {
		pagination = &commonv1.PaginationResult{NextPageToken: nextToken}
	}

	return &billsv1.GetPaymentDashboardResponse{
		Entries:    entries,
		Pagination: pagination,
	}, nil
}

// MarkBillPaid idempotently marks a bill as paid.
func (s *Server) MarkBillPaid(ctx context.Context, req *billsv1.MarkBillPaidRequest) (*billsv1.MarkBillPaidResponse, error) {
	pc := req.GetCtx()
	if pc == nil || pc.GetProjectId() == "" {
		return nil, status.Error(codes.InvalidArgument, "project_id is required")
	}
	if req.GetBillId() == "" {
		return nil, status.Error(codes.InvalidArgument, "bill_id is required")
	}

	markedBy := ""
	if a := req.GetAudit(); a != nil {
		markedBy = a.GetPerformedBy()
	}

	bill, err := s.svc.MarkBillPaid(ctx, pc.GetProjectId(), req.GetBillId(), markedBy)
	if err != nil {
		if appErr := apperrors.AsAppError(err); appErr != nil {
			return nil, toGRPCStatusError(appErr)
		}
		s.logger.Error("grpc.MarkBillPaid failed",
			zap.String("bill_id", req.GetBillId()),
			zap.Error(err))
		return nil, status.Error(codes.Internal, "internal service error")
	}

	return &billsv1.MarkBillPaidResponse{Bill: bill}, nil
}

// GetBill returns a single bill record by ID.
func (s *Server) GetBill(ctx context.Context, req *billsv1.GetBillRequest) (*billsv1.GetBillResponse, error) {
	pc := req.GetCtx()
	if pc == nil || pc.GetProjectId() == "" {
		return nil, status.Error(codes.InvalidArgument, "project_id is required")
	}
	if req.GetBillId() == "" {
		return nil, status.Error(codes.InvalidArgument, "bill_id is required")
	}

	bill, err := s.svc.GetBill(ctx, pc.GetProjectId(), req.GetBillId())
	if err != nil {
		if appErr := apperrors.AsAppError(err); appErr != nil {
			return nil, toGRPCStatusError(appErr)
		}
		s.logger.Error("grpc.GetBill failed",
			zap.String("bill_id", req.GetBillId()),
			zap.Error(err))
		return nil, status.Error(codes.Internal, "internal service error")
	}
	if bill == nil {
		return nil, status.Error(codes.NotFound, "bill not found")
	}

	return &billsv1.GetBillResponse{Bill: bill}, nil
}

// ListBills returns project-scoped bill records with optional status filter.
func (s *Server) ListBills(ctx context.Context, req *billsv1.ListBillsRequest) (*billsv1.ListBillsResponse, error) {
	pc := req.GetCtx()
	if pc == nil || pc.GetProjectId() == "" {
		return nil, status.Error(codes.InvalidArgument, "project_id is required")
	}

	var pageSize int32 = 20
	var pageToken string
	if p := req.GetPagination(); p != nil {
		if p.GetPageSize() > 0 {
			pageSize = p.GetPageSize()
		}
		pageToken = p.GetPageToken()
	}

	bills, nextToken, err := s.svc.ListBills(ctx, pc.GetProjectId(), req.GetStatusFilter(), pageSize, pageToken)
	if err != nil {
		if appErr := apperrors.AsAppError(err); appErr != nil {
			return nil, toGRPCStatusError(appErr)
		}
		s.logger.Error("grpc.ListBills failed",
			zap.String("project_id", pc.GetProjectId()),
			zap.Error(err))
		return nil, status.Error(codes.Internal, "internal service error")
	}

	var pagination *commonv1.PaginationResult
	if nextToken != "" {
		pagination = &commonv1.PaginationResult{NextPageToken: nextToken}
	}

	return &billsv1.ListBillsResponse{
		Bills:      bills,
		Pagination: pagination,
	}, nil
}

func toGRPCStatusError(appErr *apperrors.AppError) error {
	if appErr == nil {
		return status.Error(codes.Internal, "internal service error")
	}

	message := appErr.Message
	if message == "" {
		message = "internal service error"
	}

	switch appErr.Category {
	case apperrors.CategoryValidation:
		return status.Error(codes.InvalidArgument, message)
	case apperrors.CategoryAuth:
		return status.Error(codes.PermissionDenied, message)
	case apperrors.CategoryNotFound:
		return status.Error(codes.NotFound, message)
	case apperrors.CategoryConflict:
		return status.Error(codes.AlreadyExists, message)
	case apperrors.CategoryDependencyDB, apperrors.CategoryDependencyGRPC, apperrors.CategoryDependencyNet:
		if appErr.Retryable {
			return status.Error(codes.Unavailable, message)
		}
		return status.Error(codes.FailedPrecondition, message)
	default:
		return status.Error(codes.Internal, message)
	}
}
