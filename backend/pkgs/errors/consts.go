package errors

// CatalogEntry represents a predefined error in the centralized error catalog.
// All backend services must use these entries or extend them through
// safe translation boundaries without leaking raw dependency errors.
type CatalogEntry struct {
	// Name is the canonical identifier for this error entry (e.g., "resource_not_found").
	Name string
	// Code is a machine-readable stable code for programmatic client handling.
	Code string
	// Message is the sanitized error message safe for user exposure.
	Message string
	// Category classifies the error for routing and retry policy.
	Category ErrorCategory
	// Retryable indicates whether this error represents a transient failure.
	Retryable bool
}

// Catalog of predefined backend errors. These MUST be used as translation targets
// at all layer boundaries (repository→service, service→transport).

// Validation errors
var (
	ErrValidationError = &CatalogEntry{
		Name:       "validation_error",
		Code:       "validation_failed",
		Message:    "request validation failed",
		Category:   CategoryValidation,
		Retryable:  false,
	}
	ErrInvalidRequest = &CatalogEntry{
		Name:       "invalid_request",
		Code:       "invalid_request",
		Message:    "invalid request format",
		Category:   CategoryValidation,
		Retryable:  false,
	}
)

// Authorization/Authentication errors
var (
	ErrUnauthorized = &CatalogEntry{
		Name:       "unauthorized",
		Code:       "unauthorized",
		Message:    "authentication required",
		Category:   CategoryAuth,
		Retryable:  false,
	}
	ErrForbidden = &CatalogEntry{
		Name:       "forbidden",
		Code:       "forbidden",
		Message:    "access denied",
		Category:   CategoryAuth,
		Retryable:  false,
	}
)

// Not found errors
var (
	ErrResourceNotFound = &CatalogEntry{
		Name:       "resource_not_found",
		Code:       "not_found",
		Message:    "resource not found",
		Category:   CategoryNotFound,
		Retryable:  false,
	}
	ErrProjectNotFound = &CatalogEntry{
		Name:       "project_not_found",
		Code:       "project_not_found",
		Message:    "project not found",
		Category:   CategoryNotFound,
		Retryable:  false,
	}
)

// Conflict errors
var (
	ErrConflict = &CatalogEntry{
		Name:       "conflict",
		Code:       "conflict",
		Message:    "resource conflict",
		Category:   CategoryConflict,
		Retryable:  false,
	}
	ErrResourceAlreadyExists = &CatalogEntry{
		Name:       "resource_already_exists",
		Code:       "resource_already_exists",
		Message:    "resource already exists",
		Category:   CategoryConflict,
		Retryable:  false,
	}
)

// Database dependency errors (retryable by default)
var (
	ErrDatabaseError = &CatalogEntry{
		Name:       "database_error",
		Code:       "database_error",
		Message:    "database operation failed",
		Category:   CategoryDependencyDB,
		Retryable:  true,
	}
	ErrDatabaseConnection = &CatalogEntry{
		Name:       "database_connection",
		Code:       "database_connection_failed",
		Message:    "database connection failed",
		Category:   CategoryDependencyDB,
		Retryable:  true,
	}
	ErrDatabaseTimeout = &CatalogEntry{
		Name:       "database_timeout",
		Code:       "database_timeout",
		Message:    "database operation timeout",
		Category:   CategoryDependencyDB,
		Retryable:  true,
	}
)

// gRPC dependency errors (transient failures are retryable)
var (
	ErrGRPCError = &CatalogEntry{
		Name:       "grpc_error",
		Code:       "grpc_error",
		Message:    "gRPC service error",
		Category:   CategoryDependencyGRPC,
		Retryable:  true,
	}
	ErrGRPCUnavailable = &CatalogEntry{
		Name:       "grpc_unavailable",
		Code:       "grpc_unavailable",
		Message:    "service temporarily unavailable",
		Category:   CategoryDependencyGRPC,
		Retryable:  true,
	}
)

// Network dependency errors (retryable)
var (
	ErrNetworkError = &CatalogEntry{
		Name:       "network_error",
		Code:       "network_error",
		Message:    "network request failed",
		Category:   CategoryDependencyNet,
		Retryable:  true,
	}
	ErrNetworkTimeout = &CatalogEntry{
		Name:       "network_timeout",
		Code:       "network_timeout",
		Message:    "network request timeout",
		Category:   CategoryDependencyNet,
		Retryable:  true,
	}
)

// Unknown/Internal errors (mandatory fallback for unclassified failures)
var (
	ErrUnknown = &CatalogEntry{
		Name:       "unknown",
		Code:       "internal_error",
		Message:    "an unexpected error occurred",
		Category:   CategoryUnknown,
		Retryable:  false,
	}
	ErrInternal = &CatalogEntry{
		Name:       "internal",
		Code:       "internal_error",
		Message:    "internal service error",
		Category:   CategoryUnknown,
		Retryable:  false,
	}
)

// Deprecated: ErrGenericError remains for backwards compatibility.
// New code MUST use catalog entries instead.
var ErrGenericError = NewWithCategory("an error occurred", CategoryUnknown)

// Deprecated: ErrUnformattedRequest remains for backwards compatibility.
var ErrUnformattedRequest = NewWithCategory("unformatted request body", CategoryValidation)
