package mappers_test

import (
	"testing"

	bffcontracts "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/services/contracts"
	controllermappers "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/transport/http/controllers/mappers"
	views "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/transport/http/views"
)

func TestPaymentsMapperNilSafety(t *testing.T) {
	t.Parallel()

	t.Run("GivenNilDashboardInputWhenMappedThenDefaultsAreApplied", func(t *testing.T) {
		// Given
		var input *views.GetPaymentDashboardInput

		// Arrange
		// no additional setup

		// Act
		cycleStart, cycleEnd, pageSize, pageToken := controllermappers.ToPaymentDashboardRequest(input)

		// Then
		if cycleStart != "" || cycleEnd != "" || pageSize != 20 || pageToken != "" {
			t.Fatalf("expected dashboard mapper defaults for nil input")
		}
	})

	t.Run("GivenNilPaymentDashboardResponseWhenMappedThenReturnsEmptyEntries", func(t *testing.T) {
		// Given
		var response *bffcontracts.PaymentDashboardResponse

		// Arrange
		// no additional setup

		// Act
		mapped := controllermappers.ToPaymentDashboardResponse(response)

		// Then
		if mapped.Entries == nil || len(mapped.Entries) != 0 {
			t.Fatalf("expected non-nil empty dashboard entries")
		}
	})
}

func TestProjectsReconciliationSettingsMapperNilSafety(t *testing.T) {
	t.Parallel()

	t.Run("GivenNilProjectsInputWhenMappedThenProjectRequestsAreEmpty", func(t *testing.T) {
		// Given
		var invite *views.InviteMemberInput
		var update *views.UpdateMemberRoleInput

		// Arrange
		// no additional setup

		// Act
		email, role := controllermappers.ToInviteMemberRequest(invite)
		memberID, newRole := controllermappers.ToUpdateMemberRoleRequest(update)

		// Then
		if email != "" || role != "" || memberID != "" || newRole != "" {
			t.Fatalf("expected empty projects mapping for nil inputs")
		}
	})

	t.Run("GivenNilReconciliationInputWhenMappedThenLinkRequestIsEmpty", func(t *testing.T) {
		// Given
		var input *views.CreateReconciliationLinkInput

		// Arrange
		// no additional setup

		// Act
		transactionLineID, billRecordID := controllermappers.ToCreateReconciliationLinkRequest(input)

		// Then
		if transactionLineID != "" || billRecordID != "" {
			t.Fatalf("expected empty reconciliation mapping for nil input")
		}
	})

	t.Run("GivenNilSettingsResponsesWhenMappedThenCollectionsAndObjectsAreSafe", func(t *testing.T) {
		// Given
		var listResponse *bffcontracts.ListBankAccountsResponse
		var itemResponse *bffcontracts.BankAccountResponse

		// Arrange
		// no additional setup

		// Act
		listMapped := controllermappers.ToListBankAccountsResponse(listResponse)
		itemMapped := controllermappers.ToBankAccountResponse(itemResponse)

		// Then
		if listMapped.Items == nil || len(listMapped.Items) != 0 {
			t.Fatalf("expected non-nil empty settings items")
		}
		if itemMapped.ID != "" {
			t.Fatalf("expected zero-value bank account for nil response")
		}
	})
}
