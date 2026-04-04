package controllers

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"

	bffinterfaces "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/interfaces"
	controllermappers "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/transport/http/controllers/mappers"
	bffmiddleware "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/transport/http/middleware"
	views "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/transport/http/views"
)

// SettingsController handles BFF settings HTTP endpoints.
// It is a pure HTTP adapter: it extracts claims, delegates to SettingsService, and returns view types.
type SettingsController struct {
	BaseController
	svc bffinterfaces.SettingsService
}

// NewSettingsController constructs a SettingsController.
func NewSettingsController(logger *zap.Logger, validate *validator.Validate, svc bffinterfaces.SettingsService) *SettingsController {
	return &SettingsController{BaseController: BaseController{logger: logger, validate: validate}, svc: svc}
}

// HandleList returns all bank account labels for the caller's project.
func (c *SettingsController) HandleList(ctx context.Context, _ *struct{}) (*struct {
	Body views.ListBankAccountsResponse
}, error) {
	claims := bffmiddleware.ClaimsFromContext(ctx)
	if claims == nil {
		return nil, huma.Error403Forbidden("missing project context")
	}

	resp, err := c.svc.ListBankAccounts(ctx, claims.GetProjectId())
	if err != nil {
		return nil, c.grpcToHumaError(err, "list bank accounts failed")
	}

	return &struct {
		Body views.ListBankAccountsResponse
	}{Body: controllermappers.ToListBankAccountsResponse(resp)}, nil
}

// HandleCreate registers a new bank account label for the caller's project.
func (c *SettingsController) HandleCreate(ctx context.Context, input *views.CreateBankAccountInput) (*struct{ Body views.BankAccountResponse }, error) {
	claims := bffmiddleware.ClaimsFromContext(ctx)
	if claims == nil {
		return nil, huma.Error403Forbidden("missing project context")
	}

	label := controllermappers.ToCreateBankAccountRequest(input)
	if label == "" {
		return nil, huma.Error400BadRequest("label is required")
	}

	resp, err := c.svc.CreateBankAccount(ctx, claims.GetProjectId(), claims.GetSubject(), label)
	if err != nil {
		return nil, c.grpcToHumaError(err, "create bank account failed")
	}

	body := controllermappers.ToBankAccountResponse(resp)
	c.logger.Info("settings: bank account created",
		zap.String("bank_account_id", body.ID),
		zap.String("project_id", claims.GetProjectId()))
	return &struct{ Body views.BankAccountResponse }{Body: body}, nil
}

// HandleDelete removes a bank account label from the caller's project.
func (c *SettingsController) HandleDelete(ctx context.Context, input *views.DeleteBankAccountInput) (*struct{}, error) {
	claims := bffmiddleware.ClaimsFromContext(ctx)
	if claims == nil {
		return nil, huma.Error403Forbidden("missing project context")
	}

	bankAccountID := controllermappers.ToDeleteBankAccountRequest(input)
	if err := c.svc.DeleteBankAccount(ctx, claims.GetProjectId(), bankAccountID); err != nil {
		return nil, c.grpcToHumaError(err, "delete bank account failed")
	}

	c.logger.Info("settings: bank account deleted",
		zap.String("bank_account_id", bankAccountID),
		zap.String("project_id", claims.GetProjectId()))
	return &struct{}{}, nil
}
