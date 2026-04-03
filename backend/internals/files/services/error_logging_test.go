package services

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"

	"github.com/ralvescosta/costa-financial-assistant/backend/internals/files/interfaces"
	filesv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/files/v1"
)

type bankRepoErrStub struct {
	createErr error
	listErr   error
	deleteErr error
}

func (m *bankRepoErrStub) Create(_ context.Context, account *filesv1.BankAccount) (*filesv1.BankAccount, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	return account, nil
}

func (m *bankRepoErrStub) ListByProject(_ context.Context, _ string) ([]*filesv1.BankAccount, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return []*filesv1.BankAccount{}, nil
}

func (m *bankRepoErrStub) FindByProjectAndID(_ context.Context, _, _ string) (*filesv1.BankAccount, error) {
	return nil, nil
}

func (m *bankRepoErrStub) Delete(_ context.Context, _, _ string) error {
	return m.deleteErr
}

var _ interfaces.BankAccountRepository = (*bankRepoErrStub)(nil)

func TestBankAccountServiceBoundaryLogsOnce(t *testing.T) {
	core, logs := observer.New(zap.ErrorLevel)
	logger := zap.New(core)
	svc := NewBankAccountService(&bankRepoErrStub{listErr: errors.New("db down")}, logger)

	_, _ = svc.ListBankAccounts(context.Background(), "project-1")

	if logs.Len() != 1 {
		t.Fatalf("expected exactly 1 boundary error log, got %d", logs.Len())
	}
	if logs.All()[0].Message != "bank_account.list: failed" {
		t.Fatalf("unexpected log message: %s", logs.All()[0].Message)
	}
}
