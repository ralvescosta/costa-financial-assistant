//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	billsrepo "github.com/ralvescosta/costa-financial-assistant/backend/internals/bills/repositories"
	billssvc "github.com/ralvescosta/costa-financial-assistant/backend/internals/bills/services"
	billsv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/bills/v1"
	commonv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/common/v1"
)

// TestUS4_MarkPaidIdempotency validates the mark-paid workflow:
//   - First call marks the bill as paid and returns the updated record
//   - Subsequent identical calls return the same bill without side effects
//   - Mark-paid on a different bill is independent
func TestUS4_MarkPaidIdempotency(t *testing.T) {
	require.NoError(t, runMigrations(testDSN(), "file://../../../internals/files/migrations"))
	require.NoError(t, runMigrations(testDSN(), "file://../../../internals/bills/migrations"))

	const (
		projectID = "00000000-0000-0000-0000-000000000010"
		userID    = "00000000-0000-0000-0000-000000000001"
		billID1   = "00000000-0000-0000-0000-000000000301"
		billID2   = "00000000-0000-0000-0000-000000000302"
		docID1    = "00000000-0000-0000-0000-000000000401"
		docID2    = "00000000-0000-0000-0000-000000000402"
	)

	ctx := context.Background()
	t.Cleanup(func() {
		_, _ = testDB.ExecContext(ctx, "DELETE FROM idempotency_keys")
		_, _ = testDB.ExecContext(ctx, "DELETE FROM bill_records WHERE project_id = $1", projectID)
		_, _ = testDB.ExecContext(ctx, "DELETE FROM documents WHERE id IN ($1, $2)", docID1, docID2)
	})

	// Seed a document and two bill records
	_, err := testDB.ExecContext(ctx, `
		INSERT INTO documents (id, project_id, file_name, file_hash, storage_provider, storage_key, kind, analysis_status, uploaded_by)
		VALUES
			($1, $3, 'idempotent_bill1.pdf', 'idem001hash', 'local', 'local/idem001', 'bill', 'analysed', $4),
			($2, $3, 'idempotent_bill2.pdf', 'idem002hash', 'local', 'local/idem002', 'bill', 'analysed', $4)
	`, docID1, docID2, projectID, userID)
	require.NoError(t, err, "seed documents")

	dueDate := time.Now().AddDate(0, 0, 7).Format("2006-01-02")
	_, err = testDB.ExecContext(ctx, `
		INSERT INTO bill_records (id, project_id, document_id, due_date, amount_due, payment_status)
		VALUES
			($1, $3, $5, $6, 300.00, 'unpaid'),
			($2, $3, $4, $6, 400.00, 'unpaid')
	`, billID1, billID2, projectID, docID2, docID1, dueDate)
	require.NoError(t, err, "seed bill_records")

	logger := zaptest.NewLogger(t)
	repo := billsrepo.NewBillPaymentRepository(testDB, logger)
	svc := billssvc.NewBillPaymentService(repo, logger)
	client := newBillsClient(t, testDB)

	projectCtx := &commonv1.ProjectContext{ProjectId: projectID, UserId: userID}
	audit := &commonv1.AuditMetadata{PerformedBy: userID}

	_ = svc
	_ = repo

	// ── Scenario 1: First mark-paid call succeeds ─────────────────────────────
	t.Run("GivenUnpaidBill WhenMarkPaid ThenStatusPaid", func(t *testing.T) {
		resp, err := client.MarkBillPaid(ctx, &billsv1.MarkBillPaidRequest{
			Ctx:    projectCtx,
			BillId: billID1,
			Audit:  audit,
		})
		require.NoError(t, err, "first mark-paid call must succeed")
		require.NotNil(t, resp.GetBill())
		assert.Equal(t, billID1, resp.GetBill().GetId())
		assert.Equal(t, billsv1.PaymentStatus_PAYMENT_STATUS_PAID, resp.GetBill().GetPaymentStatus(),
			"bill must be in paid status after mark-paid")
		assert.NotEmpty(t, resp.GetBill().GetPaidAt(), "paid_at must be set")
	})

	// ── Scenario 2: Duplicate mark-paid call is idempotent ────────────────────
	t.Run("GivenAlreadyPaidBill WhenMarkPaidAgain ThenSameResult", func(t *testing.T) {
		// Call mark-paid a second time for the same bill
		resp, err := client.MarkBillPaid(ctx, &billsv1.MarkBillPaidRequest{
			Ctx:    projectCtx,
			BillId: billID1,
			Audit:  audit,
		})
		require.NoError(t, err, "idempotent mark-paid must not return an error")
		require.NotNil(t, resp.GetBill())
		assert.Equal(t, billID1, resp.GetBill().GetId())
		assert.Equal(t, billsv1.PaymentStatus_PAYMENT_STATUS_PAID, resp.GetBill().GetPaymentStatus(),
			"bill must still report paid on repeated call")
	})

	// ── Scenario 3: Mark a different bill — not affected by previous call ─────
	t.Run("GivenSecondBillUnpaid WhenMarkFirstPaid ThenSecondBillUnchanged", func(t *testing.T) {
		// Verify bill2 is still unpaid
		listResp, err := client.ListBills(ctx, &billsv1.ListBillsRequest{
			Ctx:          projectCtx,
			StatusFilter: billsv1.PaymentStatus_PAYMENT_STATUS_UNPAID,
			Pagination:   &commonv1.Pagination{PageSize: 10},
		})
		require.NoError(t, err)

		var bill2Found bool
		for _, b := range listResp.GetBills() {
			if b.GetId() == billID2 {
				bill2Found = true
				assert.Equal(t, billsv1.PaymentStatus_PAYMENT_STATUS_UNPAID, b.GetPaymentStatus(),
					"bill2 must still be unpaid")
			}
		}
		assert.True(t, bill2Found, "bill2 must appear in unpaid list")
	})
}
