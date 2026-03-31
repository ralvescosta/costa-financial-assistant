package interfaces

import (
	"context"

	filesv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/files/v1"
)

// BankAccountService defines the contract for bank account label management.
// It is implemented by services.BankAccountService and consumed by the BFF gRPC client handlers.
type BankAccountService interface {
	CreateBankAccount(ctx context.Context, projectID, label, createdBy string) (*filesv1.BankAccount, error)
	ListBankAccounts(ctx context.Context, projectID string) ([]*filesv1.BankAccount, error)
	DeleteBankAccount(ctx context.Context, projectID, bankAccountID string) error
}

// BankAccountRepository defines the project-scoped persistence contract for bank accounts.
// It is implemented by repositories.PostgresBankAccountRepository.
type BankAccountRepository interface {
	Create(ctx context.Context, account *filesv1.BankAccount) (*filesv1.BankAccount, error)
	ListByProject(ctx context.Context, projectID string) ([]*filesv1.BankAccount, error)
	FindByProjectAndID(ctx context.Context, projectID, id string) (*filesv1.BankAccount, error)
	Delete(ctx context.Context, projectID, id string) error
}
