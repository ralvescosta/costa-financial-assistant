package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	bffinterfaces "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/interfaces"
	"github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/services"
	filesv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/files/v1"
)

// ─── helpers ──────────────────────────────────────────────────────────────────

func newSettingsService(t *testing.T, client bffinterfaces.FilesClient) bffinterfaces.SettingsService {
	t.Helper()
	return services.NewSettingsService(zaptest.NewLogger(t), client)
}

// ─── ListBankAccounts ─────────────────────────────────────────────────────────

func TestSettingsService_ListBankAccounts_ReturnsAccounts(t *testing.T) {
	// Arrange
	client := &mockFilesClient{}
	svc := newSettingsService(t, client)
	ctx := context.Background()

	client.On("ListBankAccounts", ctx, mock.AnythingOfType("*filesv1.ListBankAccountsRequest")).Return(
		&filesv1.ListBankAccountsResponse{
			BankAccounts: []*filesv1.BankAccount{
				{Id: "ba-1", Label: "Nubank"},
				{Id: "ba-2", Label: "Itaú"},
			},
		}, nil)

	// Act
	result, err := svc.ListBankAccounts(ctx, "proj-1")

	// Assert
	require.NoError(t, err)
	assert.Len(t, result.Items, 2)
	assert.Equal(t, "ba-1", result.Items[0].ID)
}

func TestSettingsService_ListBankAccounts_ClientError(t *testing.T) {
	// Arrange
	client := &mockFilesClient{}
	svc := newSettingsService(t, client)
	ctx := context.Background()

	client.On("ListBankAccounts", ctx, mock.Anything).Return(nil, errors.New("downstream error"))

	// Act
	result, err := svc.ListBankAccounts(ctx, "proj-1")

	// Assert
	assert.Nil(t, result)
	assert.Error(t, err)
}

// ─── CreateBankAccount ────────────────────────────────────────────────────────

func TestSettingsService_CreateBankAccount_Success(t *testing.T) {
	// Arrange
	client := &mockFilesClient{}
	svc := newSettingsService(t, client)
	ctx := context.Background()

	client.On("CreateBankAccount", ctx, mock.AnythingOfType("*filesv1.CreateBankAccountRequest")).Return(
		&filesv1.CreateBankAccountResponse{
			BankAccount: &filesv1.BankAccount{Id: "ba-new", Label: "Bradesco"},
		}, nil)

	// Act
	result, err := svc.CreateBankAccount(ctx, "proj-1", "user-1", "Bradesco")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "ba-new", result.ID)
	assert.Equal(t, "Bradesco", result.Label)
}

// ─── DeleteBankAccount ────────────────────────────────────────────────────────

func TestSettingsService_DeleteBankAccount_Success(t *testing.T) {
	// Arrange
	client := &mockFilesClient{}
	svc := newSettingsService(t, client)
	ctx := context.Background()

	client.On("DeleteBankAccount", ctx, mock.AnythingOfType("*filesv1.DeleteBankAccountRequest")).Return(
		&filesv1.DeleteBankAccountResponse{}, nil)

	// Act
	err := svc.DeleteBankAccount(ctx, "proj-1", "ba-1")

	// Assert
	require.NoError(t, err)
	client.AssertExpectations(t)
}
