package errors

import (
	nativeerrors "errors"
	"testing"
)

func TestNew(t *testing.T) {
	err := New("test error")
	if err.Message != "test error" {
		t.Fatalf("unexpected message: %s", err.Message)
	}
	if err.Category != CategoryUnknown {
		t.Fatalf("unexpected category: %s", err.Category)
	}
	if err.Retryable {
		t.Fatalf("expected non-retryable")
	}
}

func TestNewWithCategory(t *testing.T) {
	err := NewWithCategory("validation", CategoryValidation)
	if err.Code != string(CategoryValidation) {
		t.Fatalf("unexpected code: %s", err.Code)
	}
	if err.Category != CategoryValidation {
		t.Fatalf("unexpected category: %s", err.Category)
	}
}

func TestNewRetryable(t *testing.T) {
	err := NewRetryable("db timeout")
	if !err.Retryable {
		t.Fatalf("expected retryable=true")
	}
	if err.Category != CategoryDependencyDB {
		t.Fatalf("unexpected category: %s", err.Category)
	}
}

func TestWithErrorAndUnwrap(t *testing.T) {
	native := nativeerrors.New("native")
	appErr := New("wrapped").WithError(native)

	if !nativeerrors.Is(appErr, native) {
		t.Fatalf("errors.Is should match wrapped native error")
	}
}

func TestNewCatalogError(t *testing.T) {
	appErr := NewCatalogError(ErrDatabaseTimeout)
	if appErr.Code != ErrDatabaseTimeout.Code {
		t.Fatalf("unexpected code: %s", appErr.Code)
	}
	if appErr.Category != ErrDatabaseTimeout.Category {
		t.Fatalf("unexpected category: %s", appErr.Category)
	}
	if appErr.Retryable != ErrDatabaseTimeout.Retryable {
		t.Fatalf("unexpected retryable")
	}
}

func TestIsAppErrorAndAsAppError(t *testing.T) {
	appErr := New("app error")
	if !IsAppError(appErr) {
		t.Fatalf("expected IsAppError=true")
	}
	if AsAppError(appErr) == nil {
		t.Fatalf("expected AsAppError to return value")
	}
	if IsAppError(nativeerrors.New("native")) {
		t.Fatalf("native error must not be AppError")
	}
}
