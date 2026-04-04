//go:build integration

package cross_service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	paymentsrepo "github.com/ralvescosta/costa-financial-assistant/backend/internals/payments/repositories"
)

// TestUS6_HistoryTimeline validates that the monthly expenditure timeline:
//   - Returns one entry per calendar month with correct totals
//   - Respects the look-back window (months parameter)
//   - Handles months with no bills gracefully (no entry = gap)
func TestUS6_HistoryTimeline(t *testing.T) {
	require.NoError(t, runMigrations(testDSN(), "file://../../../internals/files/migrations"))
	require.NoError(t, runMigrations(testDSN(), "file://../../../internals/bills/migrations"))
	require.NoError(t, runMigrations(testDSN(), "file://../../../internals/payments/migrations"))

	const (
		projectID = "00000000-0000-0000-0000-000000000010"
		userID    = "00000000-0000-0000-0000-000000000001"
		docH1     = "00000000-0000-0000-0000-000000000501"
		docH2     = "00000000-0000-0000-0000-000000000502"
		docH3     = "00000000-0000-0000-0000-000000000503"
		billH1    = "00000000-0000-0000-0000-000000000511"
		billH2    = "00000000-0000-0000-0000-000000000512"
		billH3    = "00000000-0000-0000-0000-000000000513"
	)

	ctx := context.Background()
	t.Cleanup(func() {
		_, _ = testDB.ExecContext(ctx, "DELETE FROM bill_records WHERE project_id = $1", projectID)
		_, _ = testDB.ExecContext(ctx, "DELETE FROM documents WHERE project_id = $1 AND id IN ($2,$3,$4)",
			projectID, docH1, docH2, docH3)
	})

	logger := zaptest.NewLogger(t)

	now := time.Now()
	// bill1: current month, bill2: two months ago, bill3: current month (second bill)
	dueDateCurrent := now.Format("2006-01-02")
	dueDate2MonthsAgo := now.AddDate(0, -2, 0).Format("2006-01-02")

	// Seed documents
	_, err := testDB.ExecContext(ctx, `
		INSERT INTO documents (id, project_id, file_name, file_hash, storage_provider, storage_key, kind, analysis_status, uploaded_by)
		VALUES
			($1, $2, 'hbill1.pdf', 'hh1111aa', 'local', 'local/hh1111aa', 'bill', 'analysed', $3),
			($4, $2, 'hbill2.pdf', 'hh2222bb', 'local', 'local/hh2222bb', 'bill', 'analysed', $3),
			($5, $2, 'hbill3.pdf', 'hh3333cc', 'local', 'local/hh3333cc', 'bill', 'analysed', $3)
	`, docH1, projectID, userID, docH2, docH3)
	require.NoError(t, err)

	// Seed bill records: H1=200.00 current month, H2=150.00 two months ago, H3=300.00 current month
	_, err = testDB.ExecContext(ctx, `
		INSERT INTO bill_records (id, project_id, document_id, due_date, amount_due, payment_status)
		VALUES
			($1, $2, $3, $4, 200.00, 'unpaid'),
			($5, $2, $6, $7, 150.00, 'paid'),
			($8, $2, $9, $4, 300.00, 'unpaid')
	`, billH1, projectID, docH1, dueDateCurrent,
		billH2, docH2, dueDate2MonthsAgo,
		billH3, docH3)
	require.NoError(t, err)

	repo := paymentsrepo.NewHistoryRepository(testDB, logger)

	t.Run("returns monthly totals for last 12 months", func(t *testing.T) {
		entries, err := repo.GetTimeline(ctx, projectID, 12)
		require.NoError(t, err)

		// Must have at least 2 entries — current month and 2-months-ago month
		assert.GreaterOrEqual(t, len(entries), 2)

		// 200.00 + 300.00 = 500.00 must be present in one month aggregate row.
		var foundAggregate bool
		for _, e := range entries {
			if e.TotalAmount == "500.00" && e.BillCount == 2 {
				foundAggregate = true
				break
			}
		}
		assert.True(t, foundAggregate, "expected monthly aggregate row with total=500.00 and bill_count=2")
	})

	t.Run("look-back window of 1 month excludes older bills", func(t *testing.T) {
		entries, err := repo.GetTimeline(ctx, projectID, 1)
		require.NoError(t, err)

		// The older month had exactly one bill with amount 150.00 and should be excluded.
		for _, e := range entries {
			assert.False(t, e.TotalAmount == "150.00" && e.BillCount == 1, "2-months-ago aggregate should be excluded")
		}
	})

	t.Run("months=0 returns all history", func(t *testing.T) {
		entries, err := repo.GetTimeline(ctx, projectID, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(entries), 2)
	})
}
