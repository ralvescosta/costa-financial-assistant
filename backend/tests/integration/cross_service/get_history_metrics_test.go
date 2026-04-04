//go:build integration

package cross_service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"google.golang.org/grpc"

	bffservices "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/services"
	bffmiddleware "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/transport/http/middleware"
	paymentsrepo "github.com/ralvescosta/costa-financial-assistant/backend/internals/payments/repositories"
	identityv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/identity/v1"
	paymentsv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/payments/v1"
)

type metricsPropagationClient struct {
	categoryRequest   *paymentsv1.GetHistoryCategoryBreakdownRequest
	complianceRequest *paymentsv1.GetHistoryComplianceRequest
}

func (c *metricsPropagationClient) GetCyclePreference(context.Context, *paymentsv1.GetCyclePreferenceRequest, ...grpc.CallOption) (*paymentsv1.GetCyclePreferenceResponse, error) {
	return &paymentsv1.GetCyclePreferenceResponse{}, nil
}

func (c *metricsPropagationClient) SetCyclePreference(context.Context, *paymentsv1.SetCyclePreferenceRequest, ...grpc.CallOption) (*paymentsv1.SetCyclePreferenceResponse, error) {
	return &paymentsv1.SetCyclePreferenceResponse{}, nil
}

func (c *metricsPropagationClient) GetHistoryTimeline(context.Context, *paymentsv1.GetHistoryTimelineRequest, ...grpc.CallOption) (*paymentsv1.GetHistoryTimelineResponse, error) {
	return &paymentsv1.GetHistoryTimelineResponse{}, nil
}

func (c *metricsPropagationClient) GetHistoryCategoryBreakdown(_ context.Context, in *paymentsv1.GetHistoryCategoryBreakdownRequest, _ ...grpc.CallOption) (*paymentsv1.GetHistoryCategoryBreakdownResponse, error) {
	c.categoryRequest = in
	return &paymentsv1.GetHistoryCategoryBreakdownResponse{}, nil
}

func (c *metricsPropagationClient) GetHistoryCompliance(_ context.Context, in *paymentsv1.GetHistoryComplianceRequest, _ ...grpc.CallOption) (*paymentsv1.GetHistoryComplianceResponse, error) {
	c.complianceRequest = in
	return &paymentsv1.GetHistoryComplianceResponse{}, nil
}

func (c *metricsPropagationClient) GetReconciliationSummary(context.Context, *paymentsv1.GetReconciliationSummaryRequest, ...grpc.CallOption) (*paymentsv1.GetReconciliationSummaryResponse, error) {
	return &paymentsv1.GetReconciliationSummaryResponse{}, nil
}

func (c *metricsPropagationClient) CreateManualLink(context.Context, *paymentsv1.CreateManualLinkRequest, ...grpc.CallOption) (*paymentsv1.CreateManualLinkResponse, error) {
	return &paymentsv1.CreateManualLinkResponse{}, nil
}

// TestUS6_HistoryMetrics validates that:
//   - Category breakdown groups bills by bill_type and month
//   - Compliance metrics correctly compute on-time vs overdue rates
func TestUS6_HistoryMetrics(t *testing.T) {
	require.NoError(t, runMigrations(testDSN(), "file://../../../internals/files/migrations"))
	require.NoError(t, runMigrations(testDSN(), "file://../../../internals/bills/migrations"))
	require.NoError(t, runMigrations(testDSN(), "file://../../../internals/payments/migrations"))

	const (
		projectID    = "00000000-0000-0000-0000-000000000010"
		userID       = "00000000-0000-0000-0000-000000000001"
		docM1        = "00000000-0000-0000-0000-000000000601"
		docM2        = "00000000-0000-0000-0000-000000000602"
		docM3        = "00000000-0000-0000-0000-000000000603"
		billM1       = "00000000-0000-0000-0000-000000000611"
		billM2       = "00000000-0000-0000-0000-000000000612"
		billM3       = "00000000-0000-0000-0000-000000000613"
		typeEnergy   = "00000000-0000-0000-0000-000000000621"
		typeInternet = "00000000-0000-0000-0000-000000000622"
	)

	ctx := context.Background()
	t.Cleanup(func() {
		_, _ = testDB.ExecContext(ctx, "DELETE FROM bill_records WHERE project_id = $1 AND id IN ($2,$3,$4)",
			projectID, billM1, billM2, billM3)
		_, _ = testDB.ExecContext(ctx, "DELETE FROM documents WHERE project_id = $1 AND id IN ($2,$3,$4)",
			projectID, docM1, docM2, docM3)
		_, _ = testDB.ExecContext(ctx, "DELETE FROM bill_types WHERE id IN ($1,$2)", typeEnergy, typeInternet)
	})

	logger := zaptest.NewLogger(t)

	now := time.Now()
	dueDatePast := now.AddDate(0, 0, -5).Format("2006-01-02")   // 5 days ago (overdue if unpaid)
	dueDateFuture := now.AddDate(0, 0, 10).Format("2006-01-02") // 10 days from now (not overdue)
	paidOnTime := now.AddDate(0, 0, -6).Format("2006-01-02")    // paid before due date

	// Seed bill types
	_, err := testDB.ExecContext(ctx, `
		INSERT INTO bill_types (id, project_id, name, created_by)
		VALUES ($1, $2, 'Energy', $3), ($4, $2, 'Internet', $3)
	`, typeEnergy, projectID, userID, typeInternet)
	require.NoError(t, err)

	// Seed documents
	_, err = testDB.ExecContext(ctx, `
		INSERT INTO documents (id, project_id, file_name, file_hash, storage_provider, storage_key, kind, analysis_status, uploaded_by)
		VALUES
			($1, $2, 'menergy.pdf',   'mm1111ee', 'local', 'local/mm1111ee', 'bill', 'analysed', $3),
			($4, $2, 'minternet.pdf', 'mm2222ii', 'local', 'local/mm2222ii', 'bill', 'analysed', $3),
			($5, $2, 'menergy2.pdf',  'mm3333ee', 'local', 'local/mm3333ee', 'bill', 'analysed', $3)
	`, docM1, projectID, userID, docM2, docM3)
	require.NoError(t, err)

	// Seed bills:
	//   M1: Energy, past due date, paid before due → on-time
	//   M2: Internet, past due date, unpaid         → overdue
	//   M3: Energy, future due date, unpaid         → not overdue (upcoming)
	_, err = testDB.ExecContext(ctx, `
		INSERT INTO bill_records (id, project_id, document_id, due_date, amount_due, payment_status, paid_at, bill_type_id)
		VALUES
			($1,  $2, $3,  $4,  120.00, 'paid',   $5::timestamptz, $6),
			($7,  $2, $8,  $4,   90.00, 'unpaid', NULL,            $9),
			($10, $2, $11, $12,  80.00, 'unpaid', NULL,            $6)
	`, billM1, projectID, docM1, dueDatePast, paidOnTime+"T00:00:00Z", typeEnergy,
		billM2, docM2, typeInternet,
		billM3, docM3, dueDateFuture)
	require.NoError(t, err)

	repo := paymentsrepo.NewHistoryRepository(testDB, logger)

	t.Run("category breakdown groups bills by type for current month", func(t *testing.T) {
		cats, err := repo.GetCategoryBreakdown(ctx, projectID, 12)
		require.NoError(t, err)

		// Should have at least Energy and Internet entries
		var energyTotal, internetTotal float64
		for _, c := range cats {
			switch c.BillTypeName {
			case "Energy":
				energyTotal += parseAmount(c.TotalAmount)
			case "Internet":
				internetTotal += parseAmount(c.TotalAmount)
			}
		}
		// Energy bills: 120 + 80 = 200 across both months
		assert.Greater(t, energyTotal, 0.0, "energy total should be positive")
		assert.Greater(t, internetTotal, 0.0, "internet total should be positive")
	})

	t.Run("compliance metrics count on-time vs overdue correctly", func(t *testing.T) {
		compliance, err := repo.GetComplianceMetrics(ctx, projectID, 12)
		require.NoError(t, err)

		require.NotEmpty(t, compliance)

		var foundOverdue bool
		for _, c := range compliance {
			if c.Overdue >= 1 {
				foundOverdue = true
				assert.GreaterOrEqual(t, c.TotalBills, 1)
			}
		}
		assert.True(t, foundOverdue, "expected at least one compliance row with overdue bills")
	})
}

func TestUS6_HistoryMetrics_DefaultPaginationPropagation(t *testing.T) {
	// Given
	client := &metricsPropagationClient{}
	svc := bffservices.NewHistoryService(zaptest.NewLogger(t), client)
	ctx := context.WithValue(context.Background(), bffmiddleware.ProjectContextKey, &identityv1.JwtClaims{
		Subject:   "00000000-0000-0000-0000-000000000001",
		ProjectId: "00000000-0000-0000-0000-000000000010",
		Role:      "write",
		Email:     "ralvescosta@local.dev",
		Username:  "ralvescosta",
	})

	// When
	_, err := svc.GetCategoryBreakdown(ctx, "00000000-0000-0000-0000-000000000010", 6)
	require.NoError(t, err)
	_, err = svc.GetComplianceMetrics(ctx, "00000000-0000-0000-0000-000000000010", 6)

	// Then
	require.NoError(t, err)
	require.NotNil(t, client.categoryRequest)
	require.NotNil(t, client.complianceRequest)
	assert.EqualValues(t, 20, client.categoryRequest.GetPagination().GetPageSize())
	assert.EqualValues(t, 20, client.complianceRequest.GetPagination().GetPageSize())
	assert.Equal(t, "ralvescosta@local.dev", client.categoryRequest.GetSession().GetEmail())
	assert.Equal(t, "ralvescosta", client.complianceRequest.GetSession().GetUsername())
}

// parseAmount converts a decimal string like "120.00" to float64 for test assertions.
func parseAmount(s string) float64 {
	var f float64
	_, _ = fmt.Sscanf(s, "%f", &f)
	return f
}
