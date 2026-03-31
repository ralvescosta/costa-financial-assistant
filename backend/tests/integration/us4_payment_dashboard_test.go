//go:build integration

package integration

import (
	"context"
	"database/sql"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	billsrepo "github.com/ralvescosta/costa-financial-assistant/backend/internals/bills/repositories"
	billssvc "github.com/ralvescosta/costa-financial-assistant/backend/internals/bills/services"
	billsgrpc "github.com/ralvescosta/costa-financial-assistant/backend/internals/bills/transport/grpc"
	billsv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/bills/v1"
	commonv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/common/v1"
)

// TestUS4_PaymentDashboard validates the payment dashboard endpoint:
//   - Returns outstanding (unpaid) bills for a cycle date range
//   - Marks overdue bills correctly (due_date < today)
//   - Does NOT return already-paid bills
func TestUS4_PaymentDashboard(t *testing.T) {
	require.NoError(t, runMigrations(testDSN(), "file://../../internals/files/migrations"))
	require.NoError(t, runMigrations(testDSN(), "file://../../internals/bills/migrations"))

	const (
		projectID  = "00000000-0000-0000-0000-000000000010"
		documentID = "00000000-0000-0000-0000-000000000099"
		userID     = "00000000-0000-0000-0000-000000000001"
	)

	ctx := context.Background()
	t.Cleanup(func() {
		_, _ = testDB.ExecContext(ctx, "DELETE FROM bill_records WHERE project_id = $1", projectID)
		_, _ = testDB.ExecContext(ctx, "DELETE FROM documents WHERE project_id = $1", projectID)
	})

	client := newBillsClient(t, testDB)

	// ── Seed: insert a document and two bill_records ──────────────────────────
	_, err := testDB.ExecContext(ctx, `
		INSERT INTO documents (id, project_id, file_name, file_hash, storage_provider, storage_key, kind, analysis_status, uploaded_by)
		VALUES
			('00000000-0000-0000-0000-000000000099', $1, 'bill1.pdf', 'aabbcc', 'local', 'local/aabbcc', 'bill', 'analysed', $2),
			('00000000-0000-0000-0000-000000000098', $1, 'bill2.pdf', 'ddeeff', 'local', 'local/ddeeff', 'bill', 'analysed', $2),
			('00000000-0000-0000-0000-000000000097', $1, 'bill3.pdf', 'ffee00', 'local', 'local/ffee00', 'bill', 'analysed', $2)
	`, projectID, userID)
	require.NoError(t, err, "seed documents")

	overdueDate := time.Now().AddDate(0, 0, -5).Format("2006-01-02")   // 5 days ago — overdue
	upcomingDate := time.Now().AddDate(0, 0, 10).Format("2006-01-02")  // 10 days from now — upcoming
	paidDate := time.Now().AddDate(0, 0, -1).Format("2006-01-02")      // yesterday — paid

	_, err = testDB.ExecContext(ctx, `
		INSERT INTO bill_records (id, project_id, document_id, due_date, amount_due, payment_status)
		VALUES
			('00000000-0000-0000-0000-000000000201', $1, '00000000-0000-0000-0000-000000000099', $2, 150.00, 'unpaid'),
			('00000000-0000-0000-0000-000000000202', $1, '00000000-0000-0000-0000-000000000098', $3, 200.00, 'unpaid'),
			('00000000-0000-0000-0000-000000000203', $1, '00000000-0000-0000-0000-000000000097', $4, 100.00, 'paid')
	`, projectID, overdueDate, upcomingDate, paidDate)
	require.NoError(t, err, "seed bill_records")

	// ── Scenario: dashboard returns only outstanding bills ────────────────────
	t.Run("GivenOutstandingBills WhenGetDashboard ThenOverdueFlagIsCorrect", func(t *testing.T) {
		cycleStart := time.Now().AddDate(0, 0, -30).Format("2006-01-02")
		cycleEnd := time.Now().AddDate(0, 0, 30).Format("2006-01-02")

		resp, err := client.GetPaymentDashboard(ctx, &billsv1.GetPaymentDashboardRequest{
			Ctx: &commonv1.ProjectContext{
				ProjectId: projectID,
				UserId:    userID,
			},
			CycleStart: cycleStart,
			CycleEnd:   cycleEnd,
			Pagination: &commonv1.Pagination{PageSize: 10},
		})
		require.NoError(t, err, "get dashboard should succeed")
		require.NotNil(t, resp)

		// Only 2 unpaid bills should be returned (paid bill excluded)
		assert.Len(t, resp.GetEntries(), 2, "expected 2 unpaid bills in dashboard")

		// Identify bills by ID in response
		var overdueBill, upcomingBill *billsv1.PaymentDashboardEntry
		for _, e := range resp.GetEntries() {
			if e.GetBill().GetId() == "00000000-0000-0000-0000-000000000201" {
				overdueBill = e
			}
			if e.GetBill().GetId() == "00000000-0000-0000-0000-000000000202" {
				upcomingBill = e
			}
		}

		require.NotNil(t, overdueBill, "overdue bill must be present in dashboard")
		assert.True(t, overdueBill.GetIsOverdue(), "bill with past due_date must be flagged overdue")
		assert.True(t, overdueBill.GetDaysUntilDue() < 0, "overdue bill days_until_due must be negative")

		require.NotNil(t, upcomingBill, "upcoming bill must be present in dashboard")
		assert.False(t, upcomingBill.GetIsOverdue(), "upcoming bill must not be flagged overdue")
		assert.True(t, upcomingBill.GetDaysUntilDue() > 0, "upcoming bill days_until_due must be positive")
	})

	// ── Scenario: empty result when no bills in cycle ─────────────────────────
	t.Run("GivenBillsExist WhenCycleExcludesAll ThenEmptyDashboard", func(t *testing.T) {
		// Cycle 3 years in the future — no bills there
		cycleStart := time.Now().AddDate(3, 0, 0).Format("2006-01-02")
		cycleEnd := time.Now().AddDate(3, 1, 0).Format("2006-01-02")

		resp, err := client.GetPaymentDashboard(ctx, &billsv1.GetPaymentDashboardRequest{
			Ctx:        &commonv1.ProjectContext{ProjectId: projectID, UserId: userID},
			CycleStart: cycleStart,
			CycleEnd:   cycleEnd,
			Pagination: &commonv1.Pagination{PageSize: 10},
		})
		require.NoError(t, err)
		assert.Empty(t, resp.GetEntries(), "no bills in future cycle")
	})
}

// ─── helpers ──────────────────────────────────────────────────────────────────

// newBillsClient starts an in-process gRPC bills server backed by testDB and
// returns a client connected to it.
func newBillsClient(t *testing.T, db *sql.DB) billsv1.BillsServiceClient {
	t.Helper()
	logger := zaptest.NewLogger(t)

	repo := billsrepo.NewBillPaymentRepository(db, logger)
	svc := billssvc.NewBillPaymentService(repo, logger)
	srv := billsgrpc.NewServer(svc, logger)

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	grpcSrv := grpc.NewServer()
	billsv1.RegisterBillsServiceServer(grpcSrv, srv)

	t.Cleanup(func() { grpcSrv.Stop() })
	go func() { _ = grpcSrv.Serve(lis) }()

	conn, err := grpc.NewClient(
		lis.Addr().String(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	return billsv1.NewBillsServiceClient(conn)
}
