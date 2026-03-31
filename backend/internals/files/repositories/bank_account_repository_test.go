package repositories_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"

	"github.com/ralvescosta/costa-financial-assistant/backend/internals/files/mocks"
	"github.com/ralvescosta/costa-financial-assistant/backend/internals/files/repositories"
	"github.com/ralvescosta/costa-financial-assistant/backend/internals/files/services"
	filesv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/files/v1"
)

const (
	testProjectID    = "00000000-0000-0000-0000-000000000010"
	testAccountID    = "00000000-0000-0000-0000-000000000100"
	testCreatedBy    = "00000000-0000-0000-0000-000000000001"
	testAccountLabel = "Savings Account"
)

// ── Create ────────────────────────────────────────────────────────────────────

func TestBankAccountService_Create_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockBankAccountRepository(ctrl)
	svc := services.NewBankAccountService(repo, zaptest.NewLogger(t))

	expected := &filesv1.BankAccount{
		Id:        testAccountID,
		ProjectId: testProjectID,
		Label:     testAccountLabel,
		CreatedBy: testCreatedBy,
		CreatedAt: "2024-01-15T10:00:00Z",
		UpdatedAt: "2024-01-15T10:00:00Z",
	}

	repo.EXPECT().
		Create(gomock.Any(), gomock.Any()).
		Return(expected, nil)

	result, err := svc.CreateBankAccount(context.Background(), testProjectID, testAccountLabel, testCreatedBy)
	require.NoError(t, err)
	assert.Equal(t, testAccountID, result.Id)
	assert.Equal(t, testAccountLabel, result.Label)
}

func TestBankAccountService_Create_EmptyLabel(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockBankAccountRepository(ctrl)
	svc := services.NewBankAccountService(repo, zaptest.NewLogger(t))

	// Repository should NOT be called when label validation fails.
	_, err := svc.CreateBankAccount(context.Background(), testProjectID, "", testCreatedBy)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "label is required")
}

func TestBankAccountService_Create_DuplicateLabel(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockBankAccountRepository(ctrl)
	svc := services.NewBankAccountService(repo, zaptest.NewLogger(t))

	repo.EXPECT().
		Create(gomock.Any(), gomock.Any()).
		Return(nil, repositories.ErrDuplicateBankAccount)

	_, err := svc.CreateBankAccount(context.Background(), testProjectID, testAccountLabel, testCreatedBy)
	require.Error(t, err)
	assert.True(t, errors.Is(err, repositories.ErrDuplicateBankAccount))
}

// ── List ─────────────────────────────────────────────────────────────────────

func TestBankAccountService_List_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockBankAccountRepository(ctrl)
	svc := services.NewBankAccountService(repo, zaptest.NewLogger(t))

	expected := []*filesv1.BankAccount{
		{Id: testAccountID, ProjectId: testProjectID, Label: "Checking"},
		{Id: "00000000-0000-0000-0000-000000000101", ProjectId: testProjectID, Label: "Savings"},
	}

	repo.EXPECT().
		ListByProject(gomock.Any(), testProjectID).
		Return(expected, nil)

	result, err := svc.ListBankAccounts(context.Background(), testProjectID)
	require.NoError(t, err)
	require.Len(t, result, 2)
	assert.Equal(t, "Checking", result[0].Label)
	assert.Equal(t, "Savings", result[1].Label)
}

func TestBankAccountService_List_ReturnsEmptySlice(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockBankAccountRepository(ctrl)
	svc := services.NewBankAccountService(repo, zaptest.NewLogger(t))

	repo.EXPECT().
		ListByProject(gomock.Any(), testProjectID).
		Return(nil, nil)

	result, err := svc.ListBankAccounts(context.Background(), testProjectID)
	require.NoError(t, err)
	assert.Nil(t, result)
}

// ── Delete ────────────────────────────────────────────────────────────────────

func TestBankAccountService_Delete_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockBankAccountRepository(ctrl)
	svc := services.NewBankAccountService(repo, zaptest.NewLogger(t))

	repo.EXPECT().
		Delete(gomock.Any(), testProjectID, testAccountID).
		Return(nil)

	err := svc.DeleteBankAccount(context.Background(), testProjectID, testAccountID)
	require.NoError(t, err)
}

func TestBankAccountService_Delete_EmptyID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockBankAccountRepository(ctrl)
	svc := services.NewBankAccountService(repo, zaptest.NewLogger(t))

	// Repository should NOT be called when ID validation fails.
	err := svc.DeleteBankAccount(context.Background(), testProjectID, "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "bank_account_id is required")
}

func TestBankAccountService_Delete_AttributionGuard(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockBankAccountRepository(ctrl)
	svc := services.NewBankAccountService(repo, zaptest.NewLogger(t))

	repo.EXPECT().
		Delete(gomock.Any(), testProjectID, testAccountID).
		Return(repositories.ErrBankAccountInUse)

	err := svc.DeleteBankAccount(context.Background(), testProjectID, testAccountID)
	require.Error(t, err)
	assert.True(t, errors.Is(err, repositories.ErrBankAccountInUse))
}

func TestBankAccountService_Delete_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockBankAccountRepository(ctrl)
	svc := services.NewBankAccountService(repo, zaptest.NewLogger(t))

	repo.EXPECT().
		Delete(gomock.Any(), testProjectID, testAccountID).
		Return(repositories.ErrBankAccountNotFound)

	err := svc.DeleteBankAccount(context.Background(), testProjectID, testAccountID)
	require.Error(t, err)
	assert.True(t, errors.Is(err, repositories.ErrBankAccountNotFound))
}
