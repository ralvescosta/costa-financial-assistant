package errors

import "fmt"

// ErrorCategory defines the classification of an error for translation and handling.
type ErrorCategory string

const (
	// Validation errors: invalid input, malformed request, constraint violation.
	CategoryValidation ErrorCategory = "validation"
	// Auth errors: authentication failed, authorization denied, token issues.
	CategoryAuth ErrorCategory = "auth"
	// Conflict errors: resource already exists, version conflict, state conflict.
	CategoryConflict ErrorCategory = "conflict"
	// NotFound errors: requested resource does not exist.
	CategoryNotFound ErrorCategory = "not_found"
	// Dependency errors: database, gRPC, network, external services.
	CategoryDependencyDB    ErrorCategory = "dependency_db"
	CategoryDependencyGRPC  ErrorCategory = "dependency_grpc"
	CategoryDependencyNet   ErrorCategory = "dependency_network"
	// Unknown errors: unclassified or unexpected failures.
	CategoryUnknown ErrorCategory = "unknown"
)

// AppError is the canonical cross-layer error type for all backend services.
// It sanitizes dependency-native errors before propagating across layer boundaries,
// ensuring no sensitive information leaks and enabling deterministic retry semantics.
type AppError struct {
	// Message is a sanitized, stable error message safe for upper-layer exposure.
	Message string
	// Category classifies the error for translation rules and retry logic.
	Category ErrorCategory
	// Retryable indicates whether this error represents a transient failure
	// that should be retried by the caller (future policy support).
	Retryable bool
	// Err is the wrapped native dependency error (e.g., sql.Error, grpc error).
	// It is used for internal diagnosis and structured logging at boundaries
	// but is NOT propagated across layer boundaries.
	Err error
	// Code is a stable machine-readable error code for programmatic handling.
	Code string
}

// New creates a non-retryable generic AppError.
func New(message string) *AppError {
	return &AppError{
		Message:  message,
		Category: CategoryUnknown,
		Code:     "generic_error",
	}
}

// NewWithCategory creates a non-retryable AppError with an explicit category.
func NewWithCategory(message string, category ErrorCategory) *AppError {
	return &AppError{
		Message:  message,
		Category: category,
		Code:     string(category),
	}
}

// NewRetryable creates a retryable AppError (typically for transient dependency failures).
func NewRetryable(message string) *AppError {
	return &AppError{
		Message:   message,
		Retryable: true,
		Category:  CategoryDependencyDB,
		Code:      "retryable_error",
	}
}

// NewRetryableWithCategory creates a retryable AppError with specific category.
func NewRetryableWithCategory(message string, category ErrorCategory) *AppError {
	return &AppError{
		Message:   message,
		Category:  category,
		Retryable: true,
		Code:      string(category),
	}
}

// NewCatalogError creates an AppError from a predefined catalog entry.
func NewCatalogError(entry *CatalogEntry) *AppError {
	return &AppError{
		Message:   entry.Message,
		Category:  entry.Category,
		Retryable: entry.Retryable,
		Code:      entry.Code,
	}
}

// WithError wraps the native dependency error in the AppError for internal diagnosis.
func (e *AppError) WithError(err error) *AppError {
	e.Err = err
	return e
}

// Error implements the error interface.
func (e *AppError) Error() string {
	return e.Message
}

// String returns the sanitized error message (suitable for external exposure).
func (e *AppError) String() string {
	return e.Message
}

// Unwrap returns the wrapped native error for errors.Is/As compatibility.
func (e *AppError) Unwrap() error {
	return e.Err
}

// Details returns a debug string including category and code (for logging/debugging only).
func (e *AppError) Details() string {
	return fmt.Sprintf("%s (category=%s, code=%s, retryable=%v)", e.Message, e.Category, e.Code, e.Retryable)
}
