package services

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"

	"github.com/ralvescosta/costa-financial-assistant/backend/internals/payments/interfaces"
)

type paymentCycleRepoErrStub struct{}

func (paymentCycleRepoErrStub) GetByProjectID(context.Context, string) (*interfaces.CyclePreference, error) {
	return nil, errors.New("db unavailable")
}

func (paymentCycleRepoErrStub) Upsert(context.Context, string, int, string) (*interfaces.CyclePreference, error) {
	return nil, errors.New("db unavailable")
}

type historyRepoErrStub struct{}

func (historyRepoErrStub) GetTimeline(context.Context, string, int) ([]interfaces.MonthlyTimelineEntry, error) {
	return nil, errors.New("db down")
}

func (historyRepoErrStub) GetCategoryBreakdown(context.Context, string, int) ([]interfaces.CategoryBreakdownEntry, error) {
	return nil, errors.New("db down")
}

func (historyRepoErrStub) GetComplianceMetrics(context.Context, string, int) ([]interfaces.MonthlyComplianceEntry, error) {
	return nil, errors.New("db down")
}

type reconRepoErrStub struct{}

func (reconRepoErrStub) GetUnmatchedTransactionLines(context.Context, string, string) ([]interfaces.ReconciliationSummaryEntry, error) {
	return nil, errors.New("db down")
}
func (reconRepoErrStub) GetBillsForPeriod(context.Context, string, string, string) ([]interfaces.ReconciliationSummaryEntry, error) {
	return nil, errors.New("db down")
}
func (reconRepoErrStub) CreateLink(context.Context, interfaces.ReconciliationLink) (*interfaces.ReconciliationLink, error) {
	return nil, errors.New("db down")
}
func (reconRepoErrStub) UpdateTransactionStatus(context.Context, string, string, interfaces.TransactionReconciliationStatus) error {
	return errors.New("db down")
}
func (reconRepoErrStub) GetSummary(context.Context, string, string, string) (*interfaces.ReconciliationSummary, error) {
	return nil, errors.New("db down")
}

func TestPaymentCycleServiceBoundaryLogsOnce(t *testing.T) {
	core, logs := observer.New(zap.ErrorLevel)
	logger := zap.New(core)
	svc := NewPaymentCycleService(paymentCycleRepoErrStub{}, nil, logger)

	_, _ = svc.GetCyclePreference(context.Background(), "project-1")

	if logs.Len() != 1 {
		t.Fatalf("expected exactly 1 boundary error log, got %d", logs.Len())
	}
	if logs.All()[0].Message != "cycle_service: get preference failed" {
		t.Fatalf("unexpected log message: %s", logs.All()[0].Message)
	}
}

func TestHistoryServiceBoundaryLogsOnce(t *testing.T) {
	core, logs := observer.New(zap.ErrorLevel)
	logger := zap.New(core)
	svc := NewHistoryService(historyRepoErrStub{}, logger)

	_, _ = svc.GetTimeline(context.Background(), "project-1", 6)

	if logs.Len() != 1 {
		t.Fatalf("expected exactly 1 boundary error log, got %d", logs.Len())
	}
	if logs.All()[0].Message != "history_service: get timeline failed" {
		t.Fatalf("unexpected log message: %s", logs.All()[0].Message)
	}
}

func TestReconciliationServiceBoundaryLogsOnce(t *testing.T) {
	core, logs := observer.New(zap.ErrorLevel)
	logger := zap.New(core)
	svc := NewReconciliationService(reconRepoErrStub{}, logger)

	_, _ = svc.GetSummary(context.Background(), "project-1", "", "")

	if logs.Len() != 1 {
		t.Fatalf("expected exactly 1 boundary error log, got %d", logs.Len())
	}
	if logs.All()[0].Message != "reconciliation_service: get summary failed" {
		t.Fatalf("unexpected log message: %s", logs.All()[0].Message)
	}
}
