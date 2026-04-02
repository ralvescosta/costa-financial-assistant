//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/ralvescosta/costa-financial-assistant/backend/internals/payments/interfaces"
	paymentsrepo "github.com/ralvescosta/costa-financial-assistant/backend/internals/payments/repositories"
	paymentssvc "github.com/ralvescosta/costa-financial-assistant/backend/internals/payments/services"
)

// TestUS5_AutoReconciliation validates that the reconciliation engine:
//   - Automatically links exactly-matching debit transactions to bills
//   - Flags ambiguous matches (multiple bills with same amount) without auto-linking
//   - Leaves unmatched transactions with status 'unmatched'
func TestUS5_AutoReconciliation(t *testing.T) {
	require.NoError(t, runMigrations(testDSN(), "file://../../../internals/files/migrations"))
	require.NoError(t, runMigrations(testDSN(), "file://../../../internals/bills/migrations"))
	require.NoError(t, runMigrations(testDSN(), "file://../../../internals/payments/migrations"))

	const (
		projectID = "00000000-0000-0000-0000-000000000010"
		userID    = "00000000-0000-0000-0000-000000000001"
		docBill1  = "00000000-0000-0000-0000-000000000301"
		docBill2  = "00000000-0000-0000-0000-000000000302"
		docBill3  = "00000000-0000-0000-0000-000000000304"
		docStmt   = "00000000-0000-0000-0000-000000000303"
		billID1   = "00000000-0000-0000-0000-000000000311"
		billID2   = "00000000-0000-0000-0000-000000000312"
		stmtID    = "00000000-0000-0000-0000-000000000321"
		txLine1   = "00000000-0000-0000-0000-000000000331"
		txLine2   = "00000000-0000-0000-0000-000000000332"
		txAmbig   = "00000000-0000-0000-0000-000000000333"
	)

	ctx := context.Background()
	t.Cleanup(func() {
		_, _ = testDB.ExecContext(ctx, "DELETE FROM reconciliation_links WHERE project_id = $1", projectID)
		_, _ = testDB.ExecContext(ctx, "DELETE FROM transaction_lines WHERE project_id = $1", projectID)
		_, _ = testDB.ExecContext(ctx, "DELETE FROM statement_records WHERE project_id = $1", projectID)
		_, _ = testDB.ExecContext(ctx, "DELETE FROM bill_records WHERE project_id = $1", projectID)
		_, _ = testDB.ExecContext(ctx, "DELETE FROM documents WHERE project_id = $1", projectID)
	})

	logger := zaptest.NewLogger(t)
	dueDate := time.Now().AddDate(0, 0, 5).Format("2006-01-02")

	// Seed documents
	_, err := testDB.ExecContext(ctx, `
		INSERT INTO documents (id, project_id, file_name, file_hash, storage_provider, storage_key, kind, analysis_status, uploaded_by)
		VALUES
			($1, $2, 'bill1.pdf', 'aabbcc01', 'local', 'local/aabbcc01', 'bill', 'analysed', $3),
			($4, $2, 'bill2.pdf', 'aabbcc02', 'local', 'local/aabbcc02', 'bill', 'analysed', $3),
			($5, $2, 'stmt.pdf',  'ddeeff03', 'local', 'local/ddeeff03', 'statement', 'analysed', $3),
			($6, $2, 'bill3.pdf', 'aabbcc04', 'local', 'local/aabbcc04', 'bill', 'analysed', $3)
	`, docBill1, projectID, userID, docBill2, docStmt, docBill3)
	require.NoError(t, err)

	// Seed bills: bill1 = 150.00, bill2 = 99.99
	_, err = testDB.ExecContext(ctx, `
		INSERT INTO bill_records (id, project_id, document_id, due_date, amount_due, payment_status)
		VALUES
			($1, $2, $3, $4, 150.00, 'unpaid'),
			($5, $2, $6, $4,  99.99, 'unpaid')
	`, billID1, projectID, docBill1, dueDate, billID2, docBill2)
	require.NoError(t, err)

	// Seed statement record
	_, err = testDB.ExecContext(ctx, `
		INSERT INTO statement_records (id, project_id, document_id, period_start, period_end)
		VALUES ($1, $2, $3, $4, $5)
	`, stmtID, projectID, docStmt,
		time.Now().AddDate(0, -1, 0).Format("2006-01-02"),
		time.Now().Format("2006-01-02"))
	require.NoError(t, err)

	// Seed transaction lines
	_, err = testDB.ExecContext(ctx, `
		INSERT INTO transaction_lines (id, project_id, statement_id, transaction_date, description, amount, direction)
		VALUES
			($1, $2, $3, $4, 'Electric company',  150.00, 'debit'),
			($5, $2, $3, $4, 'Unknown withdrawal', 500.00, 'debit'),
			($6, $2, $3, $4, 'Ambiguous payment',   99.99, 'debit')
	`, txLine1, projectID, stmtID, time.Now().AddDate(0, 0, -3).Format("2006-01-02"),
		txLine2, txAmbig)
	require.NoError(t, err)

	// Add a second bill with the same amount as txAmbig to create ambiguity
	_, err = testDB.ExecContext(ctx, `
		INSERT INTO bill_records (id, project_id, document_id, due_date, amount_due, payment_status)
		VALUES ('00000000-0000-0000-0000-000000000313', $1, $2, $3, 99.99, 'unpaid')
	`, projectID, docBill3, dueDate)
	require.NoError(t, err)

	// Run auto-reconcile
	repo := paymentsrepo.NewReconciliationRepository(testDB, logger)
	svc := paymentssvc.NewReconciliationService(repo, logger)

	summary, err := svc.AutoReconcile(ctx, projectID, stmtID)
	require.NoError(t, err)
	require.NotNil(t, summary)

	t.Run("GivenExactMatch WhenAutoReconcile ThenTransactionStatusIsMatchedAuto", func(t *testing.T) {
		var status string
		err := testDB.QueryRowContext(ctx,
			"SELECT reconciliation_status FROM transaction_lines WHERE id = $1", txLine1,
		).Scan(&status)
		require.NoError(t, err)
		assert.Equal(t, string(interfaces.TransactionMatchedAuto), status)
	})

	t.Run("GivenNoMatch WhenAutoReconcile ThenTransactionStatusRemainsUnmatched", func(t *testing.T) {
		var status string
		err := testDB.QueryRowContext(ctx,
			"SELECT reconciliation_status FROM transaction_lines WHERE id = $1", txLine2,
		).Scan(&status)
		require.NoError(t, err)
		assert.Equal(t, string(interfaces.TransactionUnmatched), status)
	})

	t.Run("GivenAmbiguousMatch WhenAutoReconcile ThenTransactionStatusIsAmbiguous", func(t *testing.T) {
		var status string
		err := testDB.QueryRowContext(ctx,
			"SELECT reconciliation_status FROM transaction_lines WHERE id = $1", txAmbig,
		).Scan(&status)
		require.NoError(t, err)
		assert.Equal(t, string(interfaces.TransactionAmbiguous), status)
	})

	t.Run("GivenExactMatch WhenAutoReconcile ThenReconciliationLinkExists", func(t *testing.T) {
		var count int
		err := testDB.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM reconciliation_links WHERE transaction_line_id = $1 AND link_type = 'auto'",
			txLine1,
		).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	})
}
