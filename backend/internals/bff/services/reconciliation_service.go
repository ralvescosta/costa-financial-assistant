package services

import (
	"context"

	"go.uber.org/zap"

	bffinterfaces "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/interfaces"
	bffcontracts "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/services/contracts"
	apperrors "github.com/ralvescosta/costa-financial-assistant/backend/pkgs/errors"
	commonv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/common/v1"
	paymentsv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/payments/v1"
)

// ReconciliationServiceImpl implements bffinterfaces.ReconciliationService.
type ReconciliationServiceImpl struct {
	logger         *zap.Logger
	paymentsClient paymentsv1.PaymentsServiceClient
}

// NewReconciliationService constructs a ReconciliationServiceImpl.
func NewReconciliationService(logger *zap.Logger, paymentsClient paymentsv1.PaymentsServiceClient) bffinterfaces.ReconciliationService {
	return &ReconciliationServiceImpl{logger: logger, paymentsClient: paymentsClient}
}

// GetSummary returns the reconciliation summary for the project and period.
func (s *ReconciliationServiceImpl) GetSummary(ctx context.Context, projectID, periodStart, periodEnd string) (*bffcontracts.ReconciliationSummaryResponse, error) {
	resp, err := s.paymentsClient.GetReconciliationSummary(ctx, &paymentsv1.GetReconciliationSummaryRequest{
		Ctx:         projectContextFromContext(ctx, projectID, ""),
		Session:     sessionFromContext(ctx),
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
	})
	if err != nil {
		s.logger.Error("reconciliation_svc: get summary failed",
			zap.String("project_id", projectID),
			zap.Error(err))
		if appErr := apperrors.AsAppError(err); appErr != nil {
			return nil, appErr
		}
		return nil, apperrors.TranslateError(err, "service")
	}

	summary := resp.GetSummary()
	if summary == nil {
		return &bffcontracts.ReconciliationSummaryResponse{ProjectID: projectID}, nil
	}

	result := &bffcontracts.ReconciliationSummaryResponse{
		ProjectID:   summary.GetProjectId(),
		PeriodStart: summary.GetPeriodStart(),
		PeriodEnd:   summary.GetPeriodEnd(),
		Entries:     make([]*bffcontracts.ReconciliationEntryResponse, 0, len(summary.GetEntries())),
	}
	for _, entry := range summary.GetEntries() {
		mappedEntry := &bffcontracts.ReconciliationEntryResponse{
			TransactionLineID:    entry.GetTransactionLineId(),
			TransactionDate:      entry.GetTransactionDate(),
			Description:          entry.GetDescription(),
			Amount:               entry.GetAmount(),
			Direction:            entry.GetDirection(),
			ReconciliationStatus: reconciliationStatusFromProto(entry.GetReconciliationStatus()),
		}
		if entry.LinkedBillId != nil {
			mappedEntry.LinkedBillID = entry.LinkedBillId
		}
		if entry.LinkedBillDueDate != nil {
			mappedEntry.LinkedBillDueDate = entry.LinkedBillDueDate
		}
		if entry.LinkedBillAmount != nil {
			mappedEntry.LinkedBillAmount = entry.LinkedBillAmount
		}
		if entry.LinkType != nil {
			linkType := reconciliationLinkTypeFromProto(*entry.LinkType)
			mappedEntry.LinkType = &linkType
		}
		result.Entries = append(result.Entries, mappedEntry)
	}
	return result, nil
}

// CreateManualLink manually links a statement transaction to a bill record.
func (s *ReconciliationServiceImpl) CreateManualLink(ctx context.Context, projectID, transactionLineID, billRecordID, linkedBy string) (*bffcontracts.ReconciliationLinkResponse, error) {
	resp, err := s.paymentsClient.CreateManualLink(ctx, &paymentsv1.CreateManualLinkRequest{
		Ctx:               projectContextFromContext(ctx, projectID, linkedBy),
		Session:           sessionFromContext(ctx),
		TransactionLineId: transactionLineID,
		BillRecordId:      billRecordID,
		Audit:             &commonv1.AuditMetadata{PerformedBy: linkedBy},
	})
	if err != nil {
		s.logger.Error("reconciliation_svc: create manual link failed",
			zap.String("project_id", projectID),
			zap.String("transaction_line_id", transactionLineID),
			zap.String("bill_record_id", billRecordID),
			zap.Error(err))
		if appErr := apperrors.AsAppError(err); appErr != nil {
			return nil, appErr
		}
		return nil, apperrors.TranslateError(err, "service")
	}

	link := resp.GetLink()
	if link == nil {
		return &bffcontracts.ReconciliationLinkResponse{}, nil
	}

	result := &bffcontracts.ReconciliationLinkResponse{
		ID:                link.GetId(),
		ProjectID:         link.GetProjectId(),
		TransactionLineID: link.GetTransactionLineId(),
		BillRecordID:      link.GetBillRecordId(),
		LinkType:          reconciliationLinkTypeFromProto(link.GetLinkType()),
		CreatedAt:         link.GetCreatedAt(),
	}
	if link.LinkedBy != nil {
		result.LinkedBy = link.LinkedBy
	}
	return result, nil
}

func reconciliationStatusFromProto(status paymentsv1.TransactionReconciliationStatus) string {
	switch status {
	case paymentsv1.TransactionReconciliationStatus_TRANSACTION_RECONCILIATION_STATUS_UNMATCHED:
		return "unmatched"
	case paymentsv1.TransactionReconciliationStatus_TRANSACTION_RECONCILIATION_STATUS_MATCHED_AUTO:
		return "matched_auto"
	case paymentsv1.TransactionReconciliationStatus_TRANSACTION_RECONCILIATION_STATUS_MATCHED_MANUAL:
		return "matched_manual"
	case paymentsv1.TransactionReconciliationStatus_TRANSACTION_RECONCILIATION_STATUS_AMBIGUOUS:
		return "ambiguous"
	default:
		return ""
	}
}

func reconciliationLinkTypeFromProto(linkType paymentsv1.ReconciliationLinkType) string {
	switch linkType {
	case paymentsv1.ReconciliationLinkType_RECONCILIATION_LINK_TYPE_AUTO:
		return "auto"
	case paymentsv1.ReconciliationLinkType_RECONCILIATION_LINK_TYPE_MANUAL:
		return "manual"
	default:
		return ""
	}
}
