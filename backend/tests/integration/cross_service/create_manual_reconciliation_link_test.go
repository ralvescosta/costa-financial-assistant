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

// TestUS5_ManualReconciliation validates that:
//   - A user can manually link a transaction to a specific bill
//   - The transaction is marked as matched_manual
//   - Duplicate link attempts return conflict error
func TestUS5_ManualReconciliation(t *testing.T) {
	require.NoError(t, runMigrations(testDSN(), "file://../../../internals/files/migrations"))
	require.NoError(t, runMigrations(testDSN(), "file://../../../internals/bills/migrations"))
	require.NoError(t, runMigrations(testDSN(), "file://../../../internals/payments/migrations"))

	const (
		projectID = "00000000-0000-0000-0000-000000000010"
		userID    = "00000000-0000-0000-0000-000000000001"
		docBill   = "00000000-0000-0000-0000-000000000401"
		docStmt   = "00000000-0000-0000-0000-000000000402"
		billID    = "00000000-0000-0000-0000-000000000411"
		stmtID    = "00000000-0000-0000-0000-000000000421"
		txLine    = "00000000-0000-0000-0000-000000000431"
	)

	ctx := context.Background()
	t.Cleanup(func() {
		_, _ = testDB.ExecContext(ctx, "DELETE FROM reconciliation_links WHERE project_id = $1", projectID)
		_, _ = testDB.ExecContext(ctx, "DELETE FROM transaction_lines WHERE project_id = $1", projectID)
		_, _ = testDB.ExecContext(ctx, "DELETE FROM statement_records WHERE project_id = $1", projectID)
		_, _ = testDB.ExecContext(ctx, "DELETE FROM bill_records WHERE id = $1", billID)
		_, _ = testDB.ExecContext(ctx, "DELETE FROM documents WHERE id IN ($1, $2)", docBill, docStmt)
	})

	logger := zaptest.NewLogger(t)
	dueDate := time.Now().AddDate(0, 0, 7).Format("2006-01-02")

	// Seed documents.
	_, err := testDB.ExecContext(ctx, `
		INSERT INTO documents (id, project_id, file_name, file_hash, storage_provider, storage_key, kind, analysis_status, uploaded_by)
		VALUES
			($1, $2, 'bill.pdf',  'zz0001', 'local', 'local/zz0001', 'bill',      'analysed', $3),
			($4, $2, 'stmt.pdf',  'zz0002', 'local', 'local/zz0002', 'statement', 'analysed', $3)
	`, docBill, projectID, userID, docStmt)
	require.NoError(t, err)

	_, err = testDB.ExecContext(ctx, `
		INSERT INTO bill_records (id, project_id, document_id, due_date, amount_due, payment_status)
		VALUES ($1, $2, $3, $4, 250.00, 'unpaid')
	`, billID, projectID, docBill, dueDate)
	require.NoError(t, err)

	_, err = testDB.ExecContext(ctx, `
		INSERT INTO statement_records (id, project_id, document_id, period_start, period_end)
		VALUES ($1, $2, $3, $4, $5)
	`, stmtID, projectID, docStmt,
		time.Now().AddDate(0, -1, 0).Format("2006-01-02"),
		time.Now().Format("2006-01-02"))
	require.NoError(t, err)

	_, err = testDB.ExecContext(ctx, `
		INSERT INTO transaction_lines (id, project_id, statement_id, transaction_date, description, amount, direction)
		VALUES ($1, $2, $3, $4, 'Water company', 250.00, 'debit')
	`, txLine, projectID, stmtID, time.Now().AddDate(0, 0, -2).Format("2006-01-02"))
	require.NoError(t, err)

	repo := paymentsrepo.NewReconciliationRepository(testDB, logger)
	svc := paymentssvc.NewReconciliationService(repo, logger)

	t.Run("GivenUnlinkedTransaction WhenCreateManualLink ThenLinkIsCreated", func(t *testing.T) {
		link, createErr := svc.CreateManualLink(ctx, projectID, txLine, billID, userID)
		require.NoError(t, createErr)
		require.NotNil(t, link)
		assert.Equal(t, txLine, link.TransactionLineID)
		assert.Equal(t, billID, link.BillRecordID)
		assert.Equal(t, interfaces.ReconciliationLinkTypeManual, link.LinkType)
		assert.NotNil(t, link.LinkedBy)
		assert.Equal(t, userID, *link.LinkedBy)
	})

	t.Run("GivenLinkedTransaction WhenCreateManualLink ThenTransactionIsMatchedManual", func(t *testing.T) {
		var status string
		queryErr := testDB.QueryRowContext(ctx,
			"SELECT reconciliation_status FROM transaction_lines WHERE id = $1", txLine,
		).Scan(&status)
		require.NoError(t, queryErr)
		assert.Equal(t, string(interfaces.TransactionMatchedManual), status)
	})

	t.Run("GivenExistingLink WhenCreateDuplicateLink ThenConflictErrorIsReturned", func(t *testing.T) {
		_, createErr := svc.CreateManualLink(ctx, projectID, txLine, billID, userID)
		require.Error(t, createErr, "duplicate link should return an error")
	})
}
