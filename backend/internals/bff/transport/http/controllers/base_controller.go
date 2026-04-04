package controllers

import (
	"github.com/danielgtaylor/huma/v2"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	apperrors "github.com/ralvescosta/costa-financial-assistant/backend/pkgs/errors"
)

// BaseController provides shared behavior embedded by all BFF controllers.
// It holds a logger, a validator instance, and shared error-mapping helpers.
type BaseController struct {
	logger   *zap.Logger
	validate *validator.Validate
}

// validateInput runs struct-tag validation on v and returns a Huma 400 error on failure.
func (b *BaseController) validateInput(v interface{}) error {
	if err := b.validate.Struct(v); err != nil {
		return huma.Error400BadRequest(err.Error())
	}
	return nil
}

// grpcToHumaError maps a gRPC status error to the appropriate Huma HTTP error.
// It falls back to HTTP 500 for any code not explicitly handled.
func (b *BaseController) grpcToHumaError(err error, fallback string) error {
	if appErr := apperrors.AsAppError(err); appErr != nil {
		switch appErr.Category {
		case apperrors.CategoryValidation:
			return huma.Error400BadRequest(appErr.Message)
		case apperrors.CategoryAuth:
			return huma.Error401Unauthorized(appErr.Message)
		case apperrors.CategoryNotFound:
			return huma.Error404NotFound(appErr.Message)
		case apperrors.CategoryConflict:
			return huma.Error409Conflict(appErr.Message)
		default:
			b.logger.Error(fallback, zap.Error(err))
			return huma.Error500InternalServerError(fallback)
		}
	}

	st, ok := status.FromError(err)
	if !ok {
		b.logger.Error(fallback, zap.Error(err))
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
		b.logger.Error(fallback, zap.Error(err))
		return huma.Error500InternalServerError(fallback)
	}
}
