package services

import (
	"context"

	"go.uber.org/zap"

	bffinterfaces "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/interfaces"
	bffcontracts "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/services/contracts"
	commonv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/common/v1"
	filesv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/files/v1"
)

// SettingsServiceImpl implements bffinterfaces.SettingsService using the Files gRPC client.
type SettingsServiceImpl struct {
	logger      *zap.Logger
	filesClient bffinterfaces.FilesClient
}

// NewSettingsService constructs a SettingsServiceImpl.
func NewSettingsService(logger *zap.Logger, filesClient bffinterfaces.FilesClient) bffinterfaces.SettingsService {
	return &SettingsServiceImpl{logger: logger, filesClient: filesClient}
}

// ListBankAccounts returns all bank accounts for the project.
func (s *SettingsServiceImpl) ListBankAccounts(ctx context.Context, projectID string) (*bffcontracts.ListBankAccountsResponse, error) {
	resp, err := s.filesClient.ListBankAccounts(ctx, &filesv1.ListBankAccountsRequest{
		Ctx: &commonv1.ProjectContext{ProjectId: projectID},
	})
	if err != nil {
		return nil, err
	}

	items := make([]*bffcontracts.BankAccountResponse, 0, len(resp.BankAccounts))
	for _, a := range resp.BankAccounts {
		v := bankAccountToView(a)
		items = append(items, &v)
	}
	return &bffcontracts.ListBankAccountsResponse{Items: items}, nil
}

// CreateBankAccount registers a new bank account label for the project.
func (s *SettingsServiceImpl) CreateBankAccount(ctx context.Context, projectID, createdBy, label string) (*bffcontracts.BankAccountResponse, error) {
	resp, err := s.filesClient.CreateBankAccount(ctx, &filesv1.CreateBankAccountRequest{
		Ctx:   &commonv1.ProjectContext{ProjectId: projectID},
		Label: label,
		Audit: &commonv1.AuditMetadata{PerformedBy: createdBy},
	})
	if err != nil {
		return nil, err
	}
	s.logger.Info("settings_svc: bank account created",
		zap.String("bank_account_id", resp.BankAccount.Id),
		zap.String("project_id", projectID))
	result := bankAccountToView(resp.BankAccount)
	return &result, nil
}

// DeleteBankAccount removes a bank account from the project.
func (s *SettingsServiceImpl) DeleteBankAccount(ctx context.Context, projectID, bankAccountID string) error {
	_, err := s.filesClient.DeleteBankAccount(ctx, &filesv1.DeleteBankAccountRequest{
		Ctx:           &commonv1.ProjectContext{ProjectId: projectID},
		BankAccountId: bankAccountID,
	})
	if err != nil {
		return err
	}
	s.logger.Info("settings_svc: bank account deleted",
		zap.String("bank_account_id", bankAccountID),
		zap.String("project_id", projectID))
	return nil
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func bankAccountToView(a *filesv1.BankAccount) bffcontracts.BankAccountResponse {
	return bffcontracts.BankAccountResponse{
		ID:        a.Id,
		ProjectID: a.ProjectId,
		Label:     a.Label,
		CreatedBy: a.CreatedBy,
		CreatedAt: a.CreatedAt,
		UpdatedAt: a.UpdatedAt,
	}
}
