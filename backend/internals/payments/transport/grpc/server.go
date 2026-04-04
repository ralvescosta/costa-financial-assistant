// Package grpc implements the gRPC server for the payments service.
package grpc

import (
	"context"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	paymentsinterfaces "github.com/ralvescosta/costa-financial-assistant/backend/internals/payments/interfaces"
	apperrors "github.com/ralvescosta/costa-financial-assistant/backend/pkgs/errors"
	commonv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/common/v1"
	paymentsv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/payments/v1"
)

// Server implements paymentsv1.PaymentsServiceServer backed by payments-domain services.
type Server struct {
	paymentsv1.UnimplementedPaymentsServiceServer
	cycleSvc          paymentsinterfaces.PaymentCycleService
	historySvc        paymentsinterfaces.HistoryService
	reconciliationSvc paymentsinterfaces.ReconciliationService
	logger            *zap.Logger
}

// NewServer constructs a payments gRPC server.
func NewServer(
	cycleSvc paymentsinterfaces.PaymentCycleService,
	historySvc paymentsinterfaces.HistoryService,
	reconciliationSvc paymentsinterfaces.ReconciliationService,
	logger *zap.Logger,
) *Server {
	return &Server{
		cycleSvc:          cycleSvc,
		historySvc:        historySvc,
		reconciliationSvc: reconciliationSvc,
		logger:            logger,
	}
}

// GetCyclePreference returns the preferred payment day for the project.
func (s *Server) GetCyclePreference(ctx context.Context, req *paymentsv1.GetCyclePreferenceRequest) (*paymentsv1.GetCyclePreferenceResponse, error) {
	pc := req.GetCtx()
	if pc == nil || pc.GetProjectId() == "" {
		return nil, status.Error(codes.InvalidArgument, "project_id is required")
	}
	if _, err := requireSessionUserID(req.GetSession()); err != nil {
		return nil, err
	}

	pref, err := s.cycleSvc.GetCyclePreference(ctx, pc.GetProjectId())
	if err != nil {
		if appErr := apperrors.AsAppError(err); appErr != nil {
			return nil, toGRPCStatusError(appErr)
		}
		s.logger.Error("grpc.GetCyclePreference failed",
			zap.String("project_id", pc.GetProjectId()),
			zap.Error(err))
		return nil, status.Error(codes.Internal, "internal service error")
	}

	return &paymentsv1.GetCyclePreferenceResponse{Preference: cyclePreferenceToProto(pref)}, nil
}

// SetCyclePreference creates or updates the preferred payment day for the project.
func (s *Server) SetCyclePreference(ctx context.Context, req *paymentsv1.SetCyclePreferenceRequest) (*paymentsv1.SetCyclePreferenceResponse, error) {
	pc := req.GetCtx()
	if pc == nil || pc.GetProjectId() == "" {
		return nil, status.Error(codes.InvalidArgument, "project_id is required")
	}
	sessionUserID, err := requireSessionUserID(req.GetSession())
	if err != nil {
		return nil, err
	}
	if req.GetPreferredDayOfMonth() < 1 || req.GetPreferredDayOfMonth() > 28 {
		return nil, status.Error(codes.InvalidArgument, "preferred_day_of_month must be between 1 and 28")
	}

	updatedBy := pc.GetUserId()
	if updatedBy == "" {
		updatedBy = sessionUserID
	}
	if audit := req.GetAudit(); audit != nil && audit.GetPerformedBy() != "" {
		updatedBy = audit.GetPerformedBy()
	}

	pref, err := s.cycleSvc.UpsertCyclePreference(ctx, pc.GetProjectId(), int(req.GetPreferredDayOfMonth()), updatedBy)
	if err != nil {
		if appErr := apperrors.AsAppError(err); appErr != nil {
			return nil, toGRPCStatusError(appErr)
		}
		s.logger.Error("grpc.SetCyclePreference failed",
			zap.String("project_id", pc.GetProjectId()),
			zap.Int32("preferred_day_of_month", req.GetPreferredDayOfMonth()),
			zap.Error(err))
		return nil, status.Error(codes.Internal, "internal service error")
	}

	return &paymentsv1.SetCyclePreferenceResponse{Preference: cyclePreferenceToProto(pref)}, nil
}

// GetHistoryTimeline returns the monthly expenditure totals for the project.
func (s *Server) GetHistoryTimeline(ctx context.Context, req *paymentsv1.GetHistoryTimelineRequest) (*paymentsv1.GetHistoryTimelineResponse, error) {
	pc := req.GetCtx()
	if pc == nil || pc.GetProjectId() == "" {
		return nil, status.Error(codes.InvalidArgument, "project_id is required")
	}
	if _, err := requireSessionUserID(req.GetSession()); err != nil {
		return nil, err
	}
	if req.GetMonths() < 0 {
		return nil, status.Error(codes.InvalidArgument, "months must be greater than or equal to 0")
	}

	entries, err := s.historySvc.GetTimeline(ctx, pc.GetProjectId(), int(req.GetMonths()))
	if err != nil {
		if appErr := apperrors.AsAppError(err); appErr != nil {
			return nil, toGRPCStatusError(appErr)
		}
		s.logger.Error("grpc.GetHistoryTimeline failed",
			zap.String("project_id", pc.GetProjectId()),
			zap.Int32("months", req.GetMonths()),
			zap.Error(err))
		return nil, status.Error(codes.Internal, "internal service error")
	}

	return &paymentsv1.GetHistoryTimelineResponse{
		ProjectId: pc.GetProjectId(),
		Months:    req.GetMonths(),
		Entries:   timelineEntriesToProto(entries),
	}, nil
}

// GetHistoryCategoryBreakdown returns the monthly category totals for the project.
func (s *Server) GetHistoryCategoryBreakdown(ctx context.Context, req *paymentsv1.GetHistoryCategoryBreakdownRequest) (*paymentsv1.GetHistoryCategoryBreakdownResponse, error) {
	pc := req.GetCtx()
	if pc == nil || pc.GetProjectId() == "" {
		return nil, status.Error(codes.InvalidArgument, "project_id is required")
	}
	if _, err := requireSessionUserID(req.GetSession()); err != nil {
		return nil, err
	}
	if req.GetMonths() < 0 {
		return nil, status.Error(codes.InvalidArgument, "months must be greater than or equal to 0")
	}

	entries, err := s.historySvc.GetCategoryBreakdown(ctx, pc.GetProjectId(), int(req.GetMonths()))
	if err != nil {
		if appErr := apperrors.AsAppError(err); appErr != nil {
			return nil, toGRPCStatusError(appErr)
		}
		s.logger.Error("grpc.GetHistoryCategoryBreakdown failed",
			zap.String("project_id", pc.GetProjectId()),
			zap.Int32("months", req.GetMonths()),
			zap.Error(err))
		return nil, status.Error(codes.Internal, "internal service error")
	}

	return &paymentsv1.GetHistoryCategoryBreakdownResponse{
		ProjectId: pc.GetProjectId(),
		Months:    req.GetMonths(),
		Entries:   categoryEntriesToProto(entries),
	}, nil
}

// GetHistoryCompliance returns the monthly compliance metrics for the project.
func (s *Server) GetHistoryCompliance(ctx context.Context, req *paymentsv1.GetHistoryComplianceRequest) (*paymentsv1.GetHistoryComplianceResponse, error) {
	pc := req.GetCtx()
	if pc == nil || pc.GetProjectId() == "" {
		return nil, status.Error(codes.InvalidArgument, "project_id is required")
	}
	if _, err := requireSessionUserID(req.GetSession()); err != nil {
		return nil, err
	}
	if req.GetMonths() < 0 {
		return nil, status.Error(codes.InvalidArgument, "months must be greater than or equal to 0")
	}

	entries, err := s.historySvc.GetComplianceMetrics(ctx, pc.GetProjectId(), int(req.GetMonths()))
	if err != nil {
		if appErr := apperrors.AsAppError(err); appErr != nil {
			return nil, toGRPCStatusError(appErr)
		}
		s.logger.Error("grpc.GetHistoryCompliance failed",
			zap.String("project_id", pc.GetProjectId()),
			zap.Int32("months", req.GetMonths()),
			zap.Error(err))
		return nil, status.Error(codes.Internal, "internal service error")
	}

	return &paymentsv1.GetHistoryComplianceResponse{
		ProjectId: pc.GetProjectId(),
		Months:    req.GetMonths(),
		Entries:   complianceEntriesToProto(entries),
	}, nil
}

// GetReconciliationSummary returns the reconciliation summary for the project and period.
func (s *Server) GetReconciliationSummary(ctx context.Context, req *paymentsv1.GetReconciliationSummaryRequest) (*paymentsv1.GetReconciliationSummaryResponse, error) {
	pc := req.GetCtx()
	if pc == nil || pc.GetProjectId() == "" {
		return nil, status.Error(codes.InvalidArgument, "project_id is required")
	}
	if _, err := requireSessionUserID(req.GetSession()); err != nil {
		return nil, err
	}

	summary, err := s.reconciliationSvc.GetSummary(ctx, pc.GetProjectId(), req.GetPeriodStart(), req.GetPeriodEnd())
	if err != nil {
		if appErr := apperrors.AsAppError(err); appErr != nil {
			return nil, toGRPCStatusError(appErr)
		}
		s.logger.Error("grpc.GetReconciliationSummary failed",
			zap.String("project_id", pc.GetProjectId()),
			zap.Error(err))
		return nil, status.Error(codes.Internal, "internal service error")
	}

	return &paymentsv1.GetReconciliationSummaryResponse{Summary: reconciliationSummaryToProto(summary)}, nil
}

// CreateManualLink creates a user-confirmed reconciliation link.
func (s *Server) CreateManualLink(ctx context.Context, req *paymentsv1.CreateManualLinkRequest) (*paymentsv1.CreateManualLinkResponse, error) {
	pc := req.GetCtx()
	if pc == nil || pc.GetProjectId() == "" {
		return nil, status.Error(codes.InvalidArgument, "project_id is required")
	}
	sessionUserID, err := requireSessionUserID(req.GetSession())
	if err != nil {
		return nil, err
	}
	if req.GetTransactionLineId() == "" {
		return nil, status.Error(codes.InvalidArgument, "transaction_line_id is required")
	}
	if req.GetBillRecordId() == "" {
		return nil, status.Error(codes.InvalidArgument, "bill_record_id is required")
	}

	linkedBy := pc.GetUserId()
	if linkedBy == "" {
		linkedBy = sessionUserID
	}
	if audit := req.GetAudit(); audit != nil && audit.GetPerformedBy() != "" {
		linkedBy = audit.GetPerformedBy()
	}

	link, err := s.reconciliationSvc.CreateManualLink(ctx, pc.GetProjectId(), req.GetTransactionLineId(), req.GetBillRecordId(), linkedBy)
	if err != nil {
		if appErr := apperrors.AsAppError(err); appErr != nil {
			return nil, toGRPCStatusError(appErr)
		}
		s.logger.Error("grpc.CreateManualLink failed",
			zap.String("project_id", pc.GetProjectId()),
			zap.String("transaction_line_id", req.GetTransactionLineId()),
			zap.String("bill_record_id", req.GetBillRecordId()),
			zap.Error(err))
		return nil, status.Error(codes.Internal, "internal service error")
	}

	return &paymentsv1.CreateManualLinkResponse{Link: reconciliationLinkToProto(link)}, nil
}

func requireSessionUserID(session *commonv1.Session) (string, error) {
	if session == nil || session.GetId() == "" {
		return "", status.Error(codes.Unauthenticated, "session is required")
	}
	return session.GetId(), nil
}

func cyclePreferenceToProto(pref *paymentsinterfaces.CyclePreference) *paymentsv1.CyclePreference {
	if pref == nil {
		return nil
	}
	return &paymentsv1.CyclePreference{
		Id:                  pref.ID,
		ProjectId:           pref.ProjectID,
		PreferredDayOfMonth: int32(pref.PreferredDayOfMonth),
		UpdatedBy:           pref.UpdatedBy,
		UpdatedAt:           formatTime(pref.UpdatedAt),
	}
}

func timelineEntriesToProto(entries []paymentsinterfaces.MonthlyTimelineEntry) []*paymentsv1.MonthlyTimelineEntry {
	result := make([]*paymentsv1.MonthlyTimelineEntry, 0, len(entries))
	for _, entry := range entries {
		result = append(result, &paymentsv1.MonthlyTimelineEntry{
			Month:       entry.Month,
			TotalAmount: entry.TotalAmount,
			BillCount:   int32(entry.BillCount),
		})
	}
	return result
}

func categoryEntriesToProto(entries []paymentsinterfaces.CategoryBreakdownEntry) []*paymentsv1.CategoryBreakdownEntry {
	result := make([]*paymentsv1.CategoryBreakdownEntry, 0, len(entries))
	for _, entry := range entries {
		result = append(result, &paymentsv1.CategoryBreakdownEntry{
			Month:        entry.Month,
			BillTypeName: entry.BillTypeName,
			TotalAmount:  entry.TotalAmount,
			BillCount:    int32(entry.BillCount),
		})
	}
	return result
}

func complianceEntriesToProto(entries []paymentsinterfaces.MonthlyComplianceEntry) []*paymentsv1.MonthlyComplianceEntry {
	result := make([]*paymentsv1.MonthlyComplianceEntry, 0, len(entries))
	for _, entry := range entries {
		result = append(result, &paymentsv1.MonthlyComplianceEntry{
			Month:          entry.Month,
			TotalBills:     int32(entry.TotalBills),
			PaidOnTime:     int32(entry.PaidOnTime),
			Overdue:        int32(entry.Overdue),
			ComplianceRate: entry.ComplianceRate,
		})
	}
	return result
}

func reconciliationSummaryToProto(summary *paymentsinterfaces.ReconciliationSummary) *paymentsv1.ReconciliationSummary {
	if summary == nil {
		return nil
	}

	result := &paymentsv1.ReconciliationSummary{
		ProjectId:   summary.ProjectID,
		PeriodStart: summary.PeriodStart,
		PeriodEnd:   summary.PeriodEnd,
		Entries:     make([]*paymentsv1.ReconciliationSummaryEntry, 0, len(summary.Entries)),
	}
	for _, entry := range summary.Entries {
		result.Entries = append(result.Entries, reconciliationSummaryEntryToProto(entry))
	}
	return result
}

func reconciliationSummaryEntryToProto(entry paymentsinterfaces.ReconciliationSummaryEntry) *paymentsv1.ReconciliationSummaryEntry {
	result := &paymentsv1.ReconciliationSummaryEntry{
		TransactionLineId:    entry.TransactionLineID,
		TransactionDate:      entry.TransactionDate,
		Description:          entry.Description,
		Amount:               entry.Amount,
		Direction:            entry.Direction,
		ReconciliationStatus: toProtoReconciliationStatus(entry.ReconciliationStatus),
	}
	if entry.LinkedBillID != nil {
		result.LinkedBillId = entry.LinkedBillID
	}
	if entry.LinkedBillDueDate != nil {
		result.LinkedBillDueDate = entry.LinkedBillDueDate
	}
	if entry.LinkedBillAmount != nil {
		result.LinkedBillAmount = entry.LinkedBillAmount
	}
	if entry.LinkType != nil {
		linkType := toProtoLinkType(*entry.LinkType)
		result.LinkType = &linkType
	}
	return result
}

func reconciliationLinkToProto(link *paymentsinterfaces.ReconciliationLink) *paymentsv1.ReconciliationLink {
	if link == nil {
		return nil
	}
	result := &paymentsv1.ReconciliationLink{
		Id:                link.ID,
		ProjectId:         link.ProjectID,
		TransactionLineId: link.TransactionLineID,
		BillRecordId:      link.BillRecordID,
		LinkType:          toProtoLinkType(link.LinkType),
		CreatedAt:         formatTime(link.CreatedAt),
	}
	if link.LinkedBy != nil {
		result.LinkedBy = link.LinkedBy
	}
	return result
}

func toProtoReconciliationStatus(statusValue paymentsinterfaces.TransactionReconciliationStatus) paymentsv1.TransactionReconciliationStatus {
	switch statusValue {
	case paymentsinterfaces.TransactionUnmatched:
		return paymentsv1.TransactionReconciliationStatus_TRANSACTION_RECONCILIATION_STATUS_UNMATCHED
	case paymentsinterfaces.TransactionMatchedAuto:
		return paymentsv1.TransactionReconciliationStatus_TRANSACTION_RECONCILIATION_STATUS_MATCHED_AUTO
	case paymentsinterfaces.TransactionMatchedManual:
		return paymentsv1.TransactionReconciliationStatus_TRANSACTION_RECONCILIATION_STATUS_MATCHED_MANUAL
	case paymentsinterfaces.TransactionAmbiguous:
		return paymentsv1.TransactionReconciliationStatus_TRANSACTION_RECONCILIATION_STATUS_AMBIGUOUS
	default:
		return paymentsv1.TransactionReconciliationStatus_TRANSACTION_RECONCILIATION_STATUS_UNSPECIFIED
	}
}

func toProtoLinkType(linkTypeValue paymentsinterfaces.ReconciliationLinkType) paymentsv1.ReconciliationLinkType {
	switch linkTypeValue {
	case paymentsinterfaces.ReconciliationLinkTypeAuto:
		return paymentsv1.ReconciliationLinkType_RECONCILIATION_LINK_TYPE_AUTO
	case paymentsinterfaces.ReconciliationLinkTypeManual:
		return paymentsv1.ReconciliationLinkType_RECONCILIATION_LINK_TYPE_MANUAL
	default:
		return paymentsv1.ReconciliationLinkType_RECONCILIATION_LINK_TYPE_UNSPECIFIED
	}
}

func formatTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339)
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
