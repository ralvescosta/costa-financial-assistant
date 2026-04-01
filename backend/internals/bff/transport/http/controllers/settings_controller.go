package controllers

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"go.uber.org/zap"

	bffmiddleware "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/transport/http/middleware"
	commonv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/common/v1"
	filesv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/files/v1"
)

// ─── Input / Output types ─────────────────────────────────────────────────────

// CreateBankAccountInput carries the label for a new bank account.
type CreateBankAccountInput struct {
	Body struct {
		Label string `json:"label" minLength:"1" maxLength:"100" doc:"Display label for the bank account"`
	}
}

// DeleteBankAccountInput carries the bank account ID path parameter.
type DeleteBankAccountInput struct {
	BankAccountID string `path:"bankAccountId" doc:"Bank account UUID"`
}

// BankAccountResponse is the JSON shape returned for a single bank account.
type BankAccountResponse struct {
	ID        string `json:"id"`
	ProjectID string `json:"projectId"`
	Label     string `json:"label"`
	CreatedBy string `json:"createdBy,omitempty"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

// ListBankAccountsResponse is the JSON body for the list endpoints.
type ListBankAccountsResponse struct {
	Items []BankAccountResponse `json:"items"`
}

// ─── Controller ───────────────────────────────────────────────────────────────

// SettingsController handles BFF settings HTTP endpoints.
type SettingsController struct {
	BaseController
	filesClient filesv1.FilesServiceClient
}

// NewSettingsController constructs a SettingsController.
func NewSettingsController(logger *zap.Logger, filesClient filesv1.FilesServiceClient) *SettingsController {
	return &SettingsController{BaseController: BaseController{logger: logger}, filesClient: filesClient}
}

// ─── Handlers ─────────────────────────────────────────────────────────────────

// HandleList returns all bank account labels for the caller's project.
func (c *SettingsController) HandleList(ctx context.Context, _ *struct{}) (*struct{ Body ListBankAccountsResponse }, error) {
	claims := bffmiddleware.ClaimsFromContext(ctx)
	if claims == nil {
		return nil, huma.Error403Forbidden("missing project context")
	}

	resp, err := c.filesClient.ListBankAccounts(ctx, &filesv1.ListBankAccountsRequest{
		Ctx: &commonv1.ProjectContext{
			ProjectId: claims.GetProjectId(),
		},
	})
	if err != nil {
		return nil, c.grpcToHumaError(err, "list bank accounts failed")
	}

	items := make([]BankAccountResponse, 0, len(resp.BankAccounts))
	for _, a := range resp.BankAccounts {
		items = append(items, protoBankAccountToResponse(a))
	}
	return &struct{ Body ListBankAccountsResponse }{Body: ListBankAccountsResponse{Items: items}}, nil
}

// HandleCreate registers a new bank account label for the caller's project.
func (c *SettingsController) HandleCreate(ctx context.Context, input *CreateBankAccountInput) (*struct{ Body BankAccountResponse }, error) {
	claims := bffmiddleware.ClaimsFromContext(ctx)
	if claims == nil {
		return nil, huma.Error403Forbidden("missing project context")
	}

	if input.Body.Label == "" {
		return nil, huma.Error400BadRequest("label is required")
	}

	resp, err := c.filesClient.CreateBankAccount(ctx, &filesv1.CreateBankAccountRequest{
		Ctx: &commonv1.ProjectContext{
			ProjectId: claims.GetProjectId(),
		},
		Label: input.Body.Label,
		Audit: &commonv1.AuditMetadata{
			PerformedBy: claims.GetSubject(),
		},
	})
	if err != nil {
		return nil, c.grpcToHumaError(err, "create bank account failed")
	}

	c.logger.Info("settings: bank account created",
		zap.String("bank_account_id", resp.BankAccount.Id),
		zap.String("project_id", claims.GetProjectId()))
	return &struct{ Body BankAccountResponse }{Body: protoBankAccountToResponse(resp.BankAccount)}, nil
}

// HandleDelete removes a bank account label from the caller's project.
func (c *SettingsController) HandleDelete(ctx context.Context, input *DeleteBankAccountInput) (*struct{}, error) {
	claims := bffmiddleware.ClaimsFromContext(ctx)
	if claims == nil {
		return nil, huma.Error403Forbidden("missing project context")
	}

	_, err := c.filesClient.DeleteBankAccount(ctx, &filesv1.DeleteBankAccountRequest{
		Ctx: &commonv1.ProjectContext{
			ProjectId: claims.GetProjectId(),
		},
		BankAccountId: input.BankAccountID,
	})
	if err != nil {
		return nil, c.grpcToHumaError(err, "delete bank account failed")
	}

	c.logger.Info("settings: bank account deleted",
		zap.String("bank_account_id", input.BankAccountID),
		zap.String("project_id", claims.GetProjectId()))
	return &struct{}{}, nil
}

// ── helpers ──────────────────────────────────────────────────────────────────

func protoBankAccountToResponse(a *filesv1.BankAccount) BankAccountResponse {
	return BankAccountResponse{
		ID:        a.Id,
		ProjectID: a.ProjectId,
		Label:     a.Label,
		CreatedBy: a.CreatedBy,
		CreatedAt: a.CreatedAt,
		UpdatedAt: a.UpdatedAt,
	}
}

