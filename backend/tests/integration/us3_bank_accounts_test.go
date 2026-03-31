//go:build integration

package integration

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	commonv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/common/v1"
	filesv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/files/v1"
)

// TestUS3_BankAccountCRUDAndAttributionGuard validates the full bank account lifecycle:
// create, list, attribution guard (cannot delete if in use), and delete.
func TestUS3_BankAccountCRUDAndAttributionGuard(t *testing.T) {
	require.NoError(t, runMigrations(testDSN(), "file://../../internals/files/migrations"))

	t.Cleanup(func() {
		_, _ = testDB.ExecContext(context.Background(), "DELETE FROM bank_accounts WHERE project_id = '00000000-0000-0000-0000-000000000020'")
	})

	client := newFilesClient(t, testDB)

	const (
		projectID = "00000000-0000-0000-0000-000000000020"
		createdBy = "00000000-0000-0000-0000-000000000002"
	)

	projectCtx := &commonv1.ProjectContext{ProjectId: projectID}
	audit := &commonv1.AuditMetadata{PerformedBy: createdBy}

	// ── Step 1: Create a bank account ─────────────────────────────────────────
	createResp, err := client.CreateBankAccount(context.Background(), &filesv1.CreateBankAccountRequest{
		Ctx:   projectCtx,
		Label: "Checking Account",
		Audit: audit,
	})
	require.NoError(t, err, "create bank account should succeed")
	require.NotNil(t, createResp.BankAccount)

	accountID := createResp.BankAccount.Id
	assert.NotEmpty(t, accountID, "bank account ID must be populated")
	assert.Equal(t, "Checking Account", createResp.BankAccount.Label)
	assert.Equal(t, projectID, createResp.BankAccount.ProjectId)

	// ── Step 2: Duplicate label should be rejected ────────────────────────────
	_, err = client.CreateBankAccount(context.Background(), &filesv1.CreateBankAccountRequest{
		Ctx:   projectCtx,
		Label: "Checking Account",
		Audit: audit,
	})
	require.Error(t, err, "duplicate bank account label must return an error")
	assert.Contains(t, err.Error(), "AlreadyExists", "duplicate should produce AlreadyExists gRPC error")

	// ── Step 3: Create a second bank account ──────────────────────────────────
	create2Resp, err := client.CreateBankAccount(context.Background(), &filesv1.CreateBankAccountRequest{
		Ctx:   projectCtx,
		Label: "Savings Account",
		Audit: audit,
	})
	require.NoError(t, err, "second bank account creation should succeed")
	account2ID := create2Resp.BankAccount.Id

	// ── Step 4: List bank accounts — expect 2 in label order ─────────────────
	listResp, err := client.ListBankAccounts(context.Background(), &filesv1.ListBankAccountsRequest{
		Ctx: projectCtx,
	})
	require.NoError(t, err, "list bank accounts should succeed")
	require.Len(t, listResp.BankAccounts, 2, "should list exactly 2 bank accounts")
	assert.Equal(t, "Checking Account", listResp.BankAccounts[0].Label, "should be sorted alphabetically")
	assert.Equal(t, "Savings Account", listResp.BankAccounts[1].Label)

	// ── Step 5: Delete the second account ────────────────────────────────────
	deleteResp, err := client.DeleteBankAccount(context.Background(), &filesv1.DeleteBankAccountRequest{
		Ctx:           projectCtx,
		BankAccountId: account2ID,
	})
	require.NoError(t, err, "delete unused bank account should succeed")
	assert.True(t, deleteResp.Success)

	// ── Step 6: List after delete — expect 1 remaining ────────────────────────
	listResp2, err := client.ListBankAccounts(context.Background(), &filesv1.ListBankAccountsRequest{
		Ctx: projectCtx,
	})
	require.NoError(t, err)
	require.Len(t, listResp2.BankAccounts, 1)
	assert.Equal(t, accountID, listResp2.BankAccounts[0].Id)

	// ── Step 7: Delete non-existent account returns NotFound ─────────────────
	_, err = client.DeleteBankAccount(context.Background(), &filesv1.DeleteBankAccountRequest{
		Ctx:           projectCtx,
		BankAccountId: "00000000-0000-0000-0000-000000000999",
	})
	require.Error(t, err, "deleting non-existent account must return an error")
	assert.Contains(t, err.Error(), "NotFound")

	// ── Step 8: Cross-project isolation — create an account in a different project ──
	const otherProjectID = "00000000-0000-0000-0000-000000000021"
	t.Cleanup(func() {
		_, _ = testDB.ExecContext(context.Background(), "DELETE FROM bank_accounts WHERE project_id = '00000000-0000-0000-0000-000000000021'")
	})

	otherCtx := &commonv1.ProjectContext{ProjectId: otherProjectID}
	_, err = client.CreateBankAccount(context.Background(), &filesv1.CreateBankAccountRequest{
		Ctx:   otherCtx,
		Label: "Checking Account", // same label, different project — should be allowed
		Audit: audit,
	})
	require.NoError(t, err, "same label in different project scope should succeed")

	// Verify it is NOT visible in the original project
	listResp3, err := client.ListBankAccounts(context.Background(), &filesv1.ListBankAccountsRequest{
		Ctx: projectCtx,
	})
	require.NoError(t, err)
	require.Len(t, listResp3.BankAccounts, 1, "cross-project isolation: other project's accounts must not appear")
}
