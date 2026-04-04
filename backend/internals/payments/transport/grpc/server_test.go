package grpc

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	paymentsinterfaces "github.com/ralvescosta/costa-financial-assistant/backend/internals/payments/interfaces"
	apperrors "github.com/ralvescosta/costa-financial-assistant/backend/pkgs/errors"
	commonv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/common/v1"
	paymentsv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/payments/v1"
)

type cycleServiceStub struct {
	pref *paymentsinterfaces.CyclePreference
	err  error
}

func (s cycleServiceStub) GetCyclePreference(context.Context, string) (*paymentsinterfaces.CyclePreference, error) {
	return s.pref, s.err
}

func (s cycleServiceStub) UpsertCyclePreference(context.Context, string, int, string) (*paymentsinterfaces.CyclePreference, error) {
	return s.pref, s.err
}

type historyServiceStub struct {
	timeline   []paymentsinterfaces.MonthlyTimelineEntry
	categories []paymentsinterfaces.CategoryBreakdownEntry
	compliance []paymentsinterfaces.MonthlyComplianceEntry
	err        error
}

func (s historyServiceStub) GetTimeline(context.Context, string, int) ([]paymentsinterfaces.MonthlyTimelineEntry, error) {
	return s.timeline, s.err
}

func (s historyServiceStub) GetCategoryBreakdown(context.Context, string, int) ([]paymentsinterfaces.CategoryBreakdownEntry, error) {
	return s.categories, s.err
}

func (s historyServiceStub) GetComplianceMetrics(context.Context, string, int) ([]paymentsinterfaces.MonthlyComplianceEntry, error) {
	return s.compliance, s.err
}

type reconciliationServiceStub struct {
	summary *paymentsinterfaces.ReconciliationSummary
	link    *paymentsinterfaces.ReconciliationLink
	err     error
}

func (s reconciliationServiceStub) AutoReconcile(context.Context, string, string) (*paymentsinterfaces.ReconciliationSummary, error) {
	return s.summary, s.err
}

func (s reconciliationServiceStub) GetSummary(context.Context, string, string, string) (*paymentsinterfaces.ReconciliationSummary, error) {
	return s.summary, s.err
}

func (s reconciliationServiceStub) CreateManualLink(context.Context, string, string, string, string) (*paymentsinterfaces.ReconciliationLink, error) {
	return s.link, s.err
}

func TestServer_GetCyclePreference_RequiresProjectID(t *testing.T) {
	srv := NewServer(cycleServiceStub{}, historyServiceStub{}, reconciliationServiceStub{}, zaptest.NewLogger(t))

	resp, err := srv.GetCyclePreference(context.Background(), &paymentsv1.GetCyclePreferenceRequest{})

	assert.Nil(t, resp)
	require.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestServer_GetHistoryTimeline_ReturnsEntries(t *testing.T) {
	srv := NewServer(
		cycleServiceStub{},
		historyServiceStub{timeline: []paymentsinterfaces.MonthlyTimelineEntry{{
			Month:       "2026-03-01",
			TotalAmount: "120.50",
			BillCount:   2,
		}}},
		reconciliationServiceStub{},
		zaptest.NewLogger(t),
	)

	resp, err := srv.GetHistoryTimeline(context.Background(), &paymentsv1.GetHistoryTimelineRequest{
		Ctx:     &commonv1.ProjectContext{ProjectId: "proj-1"},
		Session: &commonv1.Session{Id: "user-1"},
		Months:  6,
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "proj-1", resp.GetProjectId())
	assert.Equal(t, int32(6), resp.GetMonths())
	assert.Len(t, resp.GetEntries(), 1)
	assert.Equal(t, "120.50", resp.GetEntries()[0].GetTotalAmount())
}

func TestServer_CreateManualLink_MapsConflictToAlreadyExists(t *testing.T) {
	srv := NewServer(
		cycleServiceStub{},
		historyServiceStub{},
		reconciliationServiceStub{err: apperrors.NewCatalogError(apperrors.ErrConflict)},
		zaptest.NewLogger(t),
	)

	resp, err := srv.CreateManualLink(context.Background(), &paymentsv1.CreateManualLinkRequest{
		Ctx:               &commonv1.ProjectContext{ProjectId: "proj-1", UserId: "user-1"},
		Session:           &commonv1.Session{Id: "user-1"},
		TransactionLineId: "txn-1",
		BillRecordId:      "bill-1",
	})

	assert.Nil(t, resp)
	require.Error(t, err)
	assert.Equal(t, codes.AlreadyExists, status.Code(err))
}

func TestServer_SetCyclePreference_ReturnsPersistedPreference(t *testing.T) {
	srv := NewServer(
		cycleServiceStub{pref: &paymentsinterfaces.CyclePreference{
			ID:                  "pref-1",
			ProjectID:           "proj-1",
			PreferredDayOfMonth: 18,
			UpdatedBy:           "user-1",
			UpdatedAt:           time.Date(2026, 4, 4, 12, 0, 0, 0, time.UTC),
		}},
		historyServiceStub{},
		reconciliationServiceStub{},
		zaptest.NewLogger(t),
	)

	resp, err := srv.SetCyclePreference(context.Background(), &paymentsv1.SetCyclePreferenceRequest{
		Ctx:                 &commonv1.ProjectContext{ProjectId: "proj-1", UserId: "user-1"},
		Session:             &commonv1.Session{Id: "user-1"},
		PreferredDayOfMonth: 18,
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "proj-1", resp.GetPreference().GetProjectId())
	assert.Equal(t, int32(18), resp.GetPreference().GetPreferredDayOfMonth())
	assert.Equal(t, "2026-04-04T12:00:00Z", resp.GetPreference().GetUpdatedAt())
}
