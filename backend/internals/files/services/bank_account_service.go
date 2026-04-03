package services

import (
	"context"
	"errors"

	"go.uber.org/zap"

	"github.com/ralvescosta/costa-financial-assistant/backend/internals/files/interfaces"
	"github.com/ralvescosta/costa-financial-assistant/backend/internals/files/repositories"
	apperrors "github.com/ralvescosta/costa-financial-assistant/backend/pkgs/errors"
	filesv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/files/v1"
)

// BankAccountServiceIface is the narrow interface consumed by the gRPC server.
type BankAccountServiceIface interface {
	CreateBankAccount(ctx context.Context, projectID, label, createdBy string) (*filesv1.BankAccount, error)
	ListBankAccounts(ctx context.Context, projectID string) ([]*filesv1.BankAccount, error)
	DeleteBankAccount(ctx context.Context, projectID, bankAccountID string) error
}

// BankAccountService implements BankAccountServiceIface.
type BankAccountService struct {
	repo   interfaces.BankAccountRepository
	logger *zap.Logger
}

// NewBankAccountService constructs a BankAccountService.
func NewBankAccountService(repo interfaces.BankAccountRepository, logger *zap.Logger) BankAccountServiceIface {
	return &BankAccountService{repo: repo, logger: logger}
}

// CreateBankAccount creates a project-scoped bank account label.
func (s *BankAccountService) CreateBankAccount(ctx context.Context, projectID, label, createdBy string) (*filesv1.BankAccount, error) {
	if label == "" {
		return nil, apperrors.NewCatalogError(apperrors.ErrValidationError)
	}

	account := &filesv1.BankAccount{
		ProjectId: projectID,
		Label:     label,
		CreatedBy: createdBy,
	}

	result, err := s.repo.Create(ctx, account)
	if err != nil {
		if errors.Is(err, repositories.ErrDuplicateBankAccount) {
			return nil, apperrors.NewCatalogError(apperrors.ErrResourceAlreadyExists).WithError(err)
		}
		if appErr := apperrors.AsAppError(err); appErr != nil {
			return nil, appErr
		}
		s.logger.Error("bank_account.create: failed",
			zap.String("project_id", projectID),
			zap.Error(err))
		return nil, apperrors.TranslateError(err, "service")
	}
	return result, nil
}

// ListBankAccounts returns all project-scoped bank account labels.
func (s *BankAccountService) ListBankAccounts(ctx context.Context, projectID string) ([]*filesv1.BankAccount, error) {
	accounts, err := s.repo.ListByProject(ctx, projectID)
	if err != nil {
		if appErr := apperrors.AsAppError(err); appErr != nil {
			return nil, appErr
		}
		s.logger.Error("bank_account.list: failed",
			zap.String("project_id", projectID),
			zap.Error(err))
		return nil, apperrors.TranslateError(err, "service")
	}
	return accounts, nil
}

// DeleteBankAccount removes a bank account label. Returns ErrBankAccountInUse if referenced.
func (s *BankAccountService) DeleteBankAccount(ctx context.Context, projectID, bankAccountID string) error {
	if bankAccountID == "" {
		return apperrors.NewCatalogError(apperrors.ErrValidationError)
	}

	if err := s.repo.Delete(ctx, projectID, bankAccountID); err != nil {
		if errors.Is(err, repositories.ErrBankAccountNotFound) {
			return apperrors.NewCatalogError(apperrors.ErrResourceNotFound).WithError(err)
		}
		if errors.Is(err, repositories.ErrBankAccountInUse) {
			return apperrors.NewCatalogError(apperrors.ErrConflict).WithError(err)
		}
		if appErr := apperrors.AsAppError(err); appErr != nil {
			return appErr
		}
		s.logger.Error("bank_account.delete: failed",
			zap.String("project_id", projectID),
			zap.String("bank_account_id", bankAccountID),
			zap.Error(err))
		return apperrors.TranslateError(err, "service")
	}
	return nil
}
