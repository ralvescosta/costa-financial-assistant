package repositories

import (
	"database/sql"
	"errors"

	apperrors "github.com/ralvescosta/costa-financial-assistant/backend/pkgs/errors"
)

// translateRepositoryError converts native database/SQL errors to AppError.
// This is used at the repository boundary to prevent raw dependency errors
// from leaking to service and transport layers.
//
// Usage:
//   if err != nil {
//       r.logger.Error("repository: context for logging", zap.Error(err))
//       return nil, translateRepositoryError(err)
//   }
func translateRepositoryError(err error) *apperrors.AppError {
	if err == nil {
		return nil
	}

	// Special cases: these are expected errors, not "errors", so we use
	// specific catalog entries directly
	if errors.Is(err, sql.ErrNoRows) {
		return apperrors.NewCatalogError(apperrors.ErrResourceNotFound)
	}

	// General case: use translation policy for dependency errors
	appErr := apperrors.TranslateError(err, "repository")
	if appErr == nil {
		appErr = apperrors.NewCatalogError(apperrors.ErrUnknown)
	}
	return appErr
}

// translateRepositoryErrorWithNative wraps the translated error with the native
// error preserved for logging. This should be logged at the boundary using:
//   logger.Error("repository: issue description", zap.Error(nativeErr))
func translateRepositoryErrorWithNative(err error) (*apperrors.AppError, error) {
	if err == nil {
		return nil, nil
	}
	return translateRepositoryError(err), err
}
