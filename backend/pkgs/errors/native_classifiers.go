package errors

import (
	"database/sql"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ClassifyDatabaseError maps database errors to appropriate catalog entries.
// Implements deterministic classification for pq (PostgreSQL driver) errors.
func ClassifyDatabaseError(err error) *CatalogEntry {
	if err == nil {
		return nil
	}

	// Check for sql.ErrNoRows (not-found, not retryable)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrResourceNotFound
	}

	// Check for connection errors (retryable)
	if errors.Is(err, sql.ErrConnDone) {
		return ErrDatabaseConnection
	}

	// For pq-specific errors, check the error message patterns
	errStr := err.Error()

	// Connection errors (retryable)
	if containsSubstring(errStr, "connection refused") ||
		containsSubstring(errStr, "connection reset") ||
		containsSubstring(errStr, "broken pipe") ||
		containsSubstring(errStr, "network unreachable") {
		return ErrDatabaseConnection
	}

	// Timeout errors (retryable)
	if containsSubstring(errStr, "timeout") ||
		containsSubstring(errStr, "timed out") {
		return ErrDatabaseTimeout
	}

	// Validation/constraint errors (not retryable)
	if containsSubstring(errStr, "constraint") ||
		containsSubstring(errStr, "violates") ||
		containsSubstring(errStr, "unique violation") {
		return ErrValidationError
	}

	// Default fallback for database errors
	return ErrDatabaseError
}

// ClassifyGRPCError maps gRPC error codes to appropriate catalog entries.
// Implements deterministic classification for gRPC status codes.
func ClassifyGRPCError(err error) *CatalogEntry {
	if err == nil {
		return nil
	}

	// Extract status from error
	st, ok := status.FromError(err)
	if !ok {
		// Not a gRPC error, fallback to generic gRPC error
		return ErrGRPCError
	}

	code := st.Code()

	switch code {
	case codes.Unavailable:
		// Service temporarily unavailable (retryable)
		return ErrGRPCUnavailable
	case codes.DeadlineExceeded:
		// Request timeout (retryable)
		return ErrGRPCUnavailable
	case codes.ResourceExhausted:
		// Rate limited or resource exhausted (retryable)
		return ErrGRPCUnavailable
	case codes.Internal:
		// Internal server error (often retryable, but could be permanent)
		return ErrGRPCError
	case codes.Unknown:
		// Unknown error (treat as non-retryable internal)
		return ErrInternal
	case codes.InvalidArgument:
		// Invalid request (not retryable)
		return ErrValidationError
	case codes.NotFound:
		// Resource not found (not retryable)
		return ErrResourceNotFound
	case codes.AlreadyExists:
		// Resource already exists (not retryable)
		return ErrResourceAlreadyExists
	case codes.PermissionDenied:
		// Permission denied (not retryable)
		return ErrForbidden
	case codes.Unauthenticated:
		// Not authenticated (not retryable, caller should re-auth)
		return ErrUnauthorized
	case codes.FailedPrecondition:
		// Failed precondition (often not retryable)
		return ErrConflict
	case codes.Aborted:
		// Transaction aborted (potentially retryable)
		return ErrGRPCError
	case codes.OK:
		// Success (should not reach here)
		return nil
	default:
		// Unknown code
		return ErrGRPCError
	}
}

// ClassifyNetworkError maps network-related errors to appropriate catalog entries.
// This is a generic classifier for net package and similar errors.
func ClassifyNetworkError(err error) *CatalogEntry {
	if err == nil {
		return nil
	}

	errStr := err.Error()

	// Timeout errors (retryable)
	if containsSubstring(errStr, "timeout") ||
		containsSubstring(errStr, "timed out") ||
		containsSubstring(errStr, "i/o timeout") {
		return ErrNetworkTimeout
	}

	// Connection refused (retryable initially, but may also indicate service down)
	if containsSubstring(errStr, "connection refused") ||
		containsSubstring(errStr, "connection reset") {
		return ErrNetworkError
	}

	// DNS resolution errors (retryable)
	if containsSubstring(errStr, "no such host") ||
		containsSubstring(errStr, "name resolution") {
		return ErrNetworkError
	}

	// Generic network error
	return ErrNetworkError
}

// containsSubstring is a simple substring check.
func containsSubstring(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}
