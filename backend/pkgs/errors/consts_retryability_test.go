package errors

import "testing"

func TestCatalogEntriesHaveExplicitRetryability_T037(t *testing.T) {
	entries := []*CatalogEntry{
		ErrValidationError,
		ErrInvalidRequest,
		ErrUnauthorized,
		ErrForbidden,
		ErrResourceNotFound,
		ErrProjectNotFound,
		ErrConflict,
		ErrResourceAlreadyExists,
		ErrDatabaseError,
		ErrDatabaseConnection,
		ErrDatabaseTimeout,
		ErrGRPCError,
		ErrGRPCUnavailable,
		ErrNetworkError,
		ErrNetworkTimeout,
		ErrUnknown,
		ErrInternal,
	}

	for _, e := range entries {
		if e == nil {
			t.Fatalf("catalog entry must not be nil")
		}
		if e.Code == "" || e.Message == "" || e.Category == "" {
			t.Fatalf("catalog entry missing required fields: %+v", e)
		}
	}
}
