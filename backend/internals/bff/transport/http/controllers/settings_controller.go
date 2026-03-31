package controllers

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	bffmiddleware "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/transport/http/middleware"
	commonv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/common/v1"
	filesv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/files/v1"
)

// ─── Input / Output types ─────────────────────────────────────────────────────

// createBankAccountInput carries the label for a new bank account.
type createBankAccountInput struct {
	Body struct {
		Label string `json:"label" minLength:"1" maxLength:"100" doc:"Display label for the bank account"`
	}
}

// bankAccountResponse is the JSON shape returned for a single bank account.
type bankAccountResponse struct {
	ID        string `json:"id"`
	ProjectID string `json:"projectId"`
	Label     string `json:"label"`
	CreatedBy string `json:"createdBy,omitempty"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

// listBankAccountsResponse is the JSON body for the list endpoints.
type listBankAccountsResponse struct {
	Items []bankAccountResponse `json:"items"`
}

// ─── Controller ───────────────────────────────────────────────────────────────

// SettingsController registers and handles all settings HTTP routes.
type SettingsController struct {
	logger      *zap.Logger
	filesClient filesv1.FilesServiceClient
}

// NewSettingsController constructs a SettingsController.
func NewSettingsController(logger *zap.Logger, filesClient filesv1.FilesServiceClient) *SettingsController {
	return &SettingsController{logger: logger, filesClient: filesClient}
}

// Register wires all settings routes to the Huma API with auth + role middleware.
func (c *SettingsController) Register(api huma.API, auth func(huma.Context, func(huma.Context))) {
	huma.Register(api, huma.Operation{
		OperationID: "list-bank-accounts",
		Method:      http.MethodGet,
		Path:        "/api/v1/bank-accounts",
		Summary:     "List project-scoped bank account labels",
		Description: "Returns all bank account labels registered for the caller's project.",
		Tags:        []string{"settings"},
		Middlewares: huma.Middlewares{auth, bffmiddleware.NewProjectGuard("read_only", c.logger)},
	}, c.handleList)

	huma.Register(api, huma.Operation{
		OperationID: "create-bank-account",
		Method:      http.MethodPost,
		Path:        "/api/v1/bank-accounts",
		Summary:     "Create a new project-scoped bank account label",
		Description: "Registers a new bank account label for attaching to statement records.",
		Tags:        []string{"settings"},
		Middlewares: huma.Middlewares{auth, bffmiddleware.NewProjectGuard("update", c.logger)},
	}, c.handleCreate)

	huma.Register(api, huma.Operation{
		OperationID: "delete-bank-account",
		Method:      http.MethodDelete,
		Path:        "/api/v1/bank-accounts/{bankAccountId}",
		Summary:     "Delete a project-scoped bank account label",
		Description: "Removes the bank account label. Fails if referenced by statement records.",
		Tags:        []string{"settings"},
		Middlewares: huma.Middlewares{auth, bffmiddleware.NewProjectGuard("update", c.logger)},
	}, c.handleDelete)
}

// ─── Handlers ─────────────────────────────────────────────────────────────────

func (c *SettingsController) handleList(ctx context.Context, _ *struct{}) (*struct{ Body listBankAccountsResponse }, error) {
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

	items := make([]bankAccountResponse, 0, len(resp.BankAccounts))
	for _, a := range resp.BankAccounts {
		items = append(items, protoBankAccountToResponse(a))
	}
	return &struct{ Body listBankAccountsResponse }{Body: listBankAccountsResponse{Items: items}}, nil
}

func (c *SettingsController) handleCreate(ctx context.Context, input *createBankAccountInput) (*struct{ Body bankAccountResponse }, error) {
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
	return &struct{ Body bankAccountResponse }{Body: protoBankAccountToResponse(resp.BankAccount)}, nil
}

func (c *SettingsController) handleDelete(ctx context.Context, input *struct {
	BankAccountID string `path:"bankAccountId" doc:"Bank account UUID"`
}) (*struct{}, error) {
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

func protoBankAccountToResponse(a *filesv1.BankAccount) bankAccountResponse {
	return bankAccountResponse{
		ID:        a.Id,
		ProjectID: a.ProjectId,
		Label:     a.Label,
		CreatedBy: a.CreatedBy,
		CreatedAt: a.CreatedAt,
		UpdatedAt: a.UpdatedAt,
	}
}

// grpcToHumaError maps gRPC status codes to Huma HTTP errors.
func (c *SettingsController) grpcToHumaError(err error, fallback string) error {
	st, ok := status.FromError(err)
	if !ok {
		c.logger.Error(fallback, zap.Error(err))
		return huma.Error500InternalServerError(fallback)
	}
	switch st.Code() {
	case codes.NotFound:
		return huma.Error404NotFound(st.Message())
	case codes.AlreadyExists:
		return huma.Error409Conflict(st.Message())
	case codes.InvalidArgument:
		return huma.Error400BadRequest(st.Message())
	case codes.FailedPrecondition:
		return huma.Error409Conflict(st.Message())
	case codes.PermissionDenied:
		return huma.Error403Forbidden(st.Message())
	case codes.Unauthenticated:
		return huma.Error401Unauthorized(st.Message())
	default:
		c.logger.Error(fallback, zap.Error(err))
		return huma.Error500InternalServerError(fallback)
	}
}
