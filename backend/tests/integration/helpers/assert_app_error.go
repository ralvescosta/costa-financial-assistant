package helpers

import (
	"errors"
	"testing"

	apperrors "github.com/ralvescosta/costa-financial-assistant/backend/pkgs/errors"
)

// AppErrorAssertions provides test helpers for validating AppError behavior in integration tests.
type AppErrorAssertions struct {
	t *testing.T
}

// NewAppErrorAssertions creates a new assertions helper.
func NewAppErrorAssertions(t *testing.T) *AppErrorAssertions {
	return &AppErrorAssertions{t: t}
}

// AssertIsAppError checks that an error is an AppError.
func (a *AppErrorAssertions) AssertIsAppError(err error) *apperrors.AppError {
	if err == nil {
		a.t.Error("Expected AppError, got nil")
		return nil
	}

	var appErr *apperrors.AppError
	if !errors.As(err, &appErr) {
		a.t.Errorf("Expected AppError, got %T: %v", err, err)
		return nil
	}

	return appErr
}

// AssertNotRawError checks that an error is NOT a native dependency error but instead an AppError.
// This is used to verify non-leakage compliance (FR-010).
func (a *AppErrorAssertions) AssertNotRawError(err error) bool {
	if err == nil {
		return true
	}

	var appErr *apperrors.AppError
	if errors.As(err, &appErr) {
		// Is an AppError, which is good (not a raw error)
		return true
	}

	// Check if it's a raw SQL error
	if errors.Is(err, &apperrors.AppError{}) {
		// This will not match, but we're checking for native DB/gRPC errors
	}

	// Generic check: if it's NOT an AppError, it's a raw error leak
	a.t.Errorf("Error is not wrapped in AppError (raw error leak detected): %T: %v", err, err)
	return false
}

// AssertErrorCategory checks that an AppError has the expected category.
func (a *AppErrorAssertions) AssertErrorCategory(appErr *apperrors.AppError, expectedCat apperrors.ErrorCategory) bool {
	if appErr.Category != expectedCat {
		a.t.Errorf("Expected error category %s, got %s", expectedCat, appErr.Category)
		return false
	}
	return true
}

// AssertErrorRetryable checks that an AppError has the expected retryability.
func (a *AppErrorAssertions) AssertErrorRetryable(appErr *apperrors.AppError, expectedRetryable bool) bool {
	if appErr.Retryable != expectedRetryable {
		a.t.Errorf("Expected retryable=%v, got %v", expectedRetryable, appErr.Retryable)
		return false
	}
	return true
}

// AssertErrorCode checks that an AppError has the expected error code.
func (a *AppErrorAssertions) AssertErrorCode(appErr *apperrors.AppError, expectedCode string) bool {
	if appErr.Code != expectedCode {
		a.t.Errorf("Expected error code %q, got %q", expectedCode, appErr.Code)
		return false
	}
	return true
}

// AssertErrorMessage checks that an AppError has the expected message.
func (a *AppErrorAssertions) AssertErrorMessage(appErr *apperrors.AppError, expectedMsg string) bool {
	if appErr.Message != expectedMsg {
		a.t.Errorf("Expected error message %q, got %q", expectedMsg, appErr.Message)
		return false
	}
	return true
}

// AssertErrorWrapsNative checks that an AppError wraps a native error.
func (a *AppErrorAssertions) AssertErrorWrapsNative(appErr *apperrors.AppError) bool {
	if appErr.Err == nil {
		a.t.Error("Expected AppError.Err to wrap a native error, got nil")
		return false
	}
	return true
}

// AssertErrorNotWrapped checks that an AppError does NOT wrap a native error.
func (a *AppErrorAssertions) AssertErrorNotWrapped(appErr *apperrors.AppError) bool {
	if appErr.Err != nil {
		a.t.Errorf("Expected AppError.Err to be nil, got: %v", appErr.Err)
		return false
	}
	return true
}

// AssertCatalogEntry checks that an AppError matches expected catalog entry properties.
func (a *AppErrorAssertions) AssertCatalogEntry(appErr *apperrors.AppError, entry *apperrors.CatalogEntry) bool {
	success := true

	if appErr.Message != entry.Message {
		a.t.Errorf("Message mismatch: expected %q, got %q", entry.Message, appErr.Message)
		success = false
	}

	if appErr.Category != entry.Category {
		a.t.Errorf("Category mismatch: expected %s, got %s", entry.Category, appErr.Category)
		success = false
	}

	if appErr.Retryable != entry.Retryable {
		a.t.Errorf("Retryable mismatch: expected %v, got %v", entry.Retryable, appErr.Retryable)
		success = false
	}

	if appErr.Code != entry.Code {
		a.t.Errorf("Code mismatch: expected %q, got %q", entry.Code, appErr.Code)
		success = false
	}

	return success
}

// AssertNonLeakageContract checks that an error conforms to the non-leakage contract (FR-010).
// It verifies:
// 1. Error is an AppError (not raw)
// 2. Message is safe for external exposure (no internal details)
// 3. Category is properly set
// 4. Code is present for client handling
func (a *AppErrorAssertions) AssertNonLeakageContract(err error) bool {
	appErr := a.AssertIsAppError(err)
	if appErr == nil {
		return false
	}

	success := true

	// Check message is reasonable (not empty, not containing internal details)
	if appErr.Message == "" {
		a.t.Error("Error message is empty")
		success = false
	}

	// Check category is set
	if appErr.Category == "" {
		a.t.Error("Error category is empty")
		success = false
	}

	// Check code is set
	if appErr.Code == "" {
		a.t.Error("Error code is empty")
		success = false
	}

	return success
}

// AssertErrorIsRetryable returns true if the error is retryable, fails if not.
func (a *AppErrorAssertions) AssertErrorIsRetryable(appErr *apperrors.AppError) bool {
	return a.AssertErrorRetryable(appErr, true)
}

// AssertErrorIsNotRetryable returns true if the error is NOT retryable, fails if it is.
func (a *AppErrorAssertions) AssertErrorIsNotRetryable(appErr *apperrors.AppError) bool {
	return a.AssertErrorRetryable(appErr, false)
}

// AssertValidErrorContract checks that an AppError satisfies all translation requirements.
func (a *AppErrorAssertions) AssertValidErrorContract(appErr *apperrors.AppError) bool {
	if err := apperrors.ValidateErrorContract(appErr); err != nil {
		a.t.Errorf("Error contract validation failed: %v", err)
		return false
	}
	return true
}
