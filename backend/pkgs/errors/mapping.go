package errors

import (
	"fmt"

	"google.golang.org/grpc/status"
)

// TranslationPolicy defines the deterministic rules for mapping native errors
// to standardized AppError catalog entries at layer boundaries.
type TranslationPolicy struct {
	databaseRules *DatabaseErrorTranslationRules
	grpcRules     *GRPCErrorTranslationRules
	networkRules  *NetworkErrorTranslationRules
}

// DatabaseErrorTranslationRules encapsulates database error classification rules.
type DatabaseErrorTranslationRules struct {
	// Rule set for common database scenarios
}

// GRPCErrorTranslationRules encapsulates gRPC error classification rules.
type GRPCErrorTranslationRules struct {
	// Rule set for gRPC status codes
}

// NetworkErrorTranslationRules encapsulates network error classification rules.
type NetworkErrorTranslationRules struct {
	// Rule set for network-related errors
}

// NewTranslationPolicy creates a new TranslationPolicy with default rules.
func NewTranslationPolicy() *TranslationPolicy {
	return &TranslationPolicy{
		databaseRules: &DatabaseErrorTranslationRules{},
		grpcRules:     &GRPCErrorTranslationRules{},
		networkRules:  &NetworkErrorTranslationRules{},
	}
}

// TranslateRepositoryError translates a native error from the repository layer
// to a standardized AppError. This is the PRIMARY translation boundary.
func (p *TranslationPolicy) TranslateRepositoryError(nativeErr error) *AppError {
	if nativeErr == nil {
		return nil
	}

	// Try database classifiers first (most repository errors are DB errors)
	if dbCatalogEntry := ClassifyDatabaseError(nativeErr); dbCatalogEntry != nil {
		appErr := NewCatalogError(dbCatalogEntry)
		appErr.WithError(nativeErr)
		return appErr
	}

	// Fallback to unknown
	appErr := NewCatalogError(ErrUnknown)
	appErr.WithError(nativeErr)
	return appErr
}

// TranslateServiceError translates a native error from the service layer
// to a standardized AppError. Services may receive database or gRPC errors.
func (p *TranslationPolicy) TranslateServiceError(nativeErr error, source string) *AppError {
	if nativeErr == nil {
		return nil
	}

	// Attempt gRPC classification if the error mentions gRPC
	if isGRPCError(nativeErr) {
		if grpcCatalogEntry := ClassifyGRPCError(nativeErr); grpcCatalogEntry != nil {
			appErr := NewCatalogError(grpcCatalogEntry)
			appErr.WithError(nativeErr)
			return appErr
		}
	}

	// Attempt database classification
	if dbCatalogEntry := ClassifyDatabaseError(nativeErr); dbCatalogEntry != nil {
		appErr := NewCatalogError(dbCatalogEntry)
		appErr.WithError(nativeErr)
		return appErr
	}

	// Attempt network classification
	if netCatalogEntry := ClassifyNetworkError(nativeErr); netCatalogEntry != nil {
		appErr := NewCatalogError(netCatalogEntry)
		appErr.WithError(nativeErr)
		return appErr
	}

	// Fallback to unknown
	appErr := NewCatalogError(ErrUnknown)
	appErr.WithError(nativeErr)
	return appErr
}

// TranslateTransportError translates a native error from the transport layer
// (gRPC handlers, HTTP controllers) to a standardized AppError.
// At this boundary, we should rarely see native errors (they should be translated
// by service layer), but we handle them gracefully.
func (p *TranslationPolicy) TranslateTransportError(nativeErr error) *AppError {
	if nativeErr == nil {
		return nil
	}

	// Assume if it's already an AppError, return as-is
	if appErr, isAppErr := nativeErr.(*AppError); isAppErr {
		return appErr
	}

	// Otherwise, use service layer translation logic
	return p.TranslateServiceError(nativeErr, "transport")
}

// TranslateAsyncConsumerError translates errors from RabbitMQ consumers or event handlers.
func (p *TranslationPolicy) TranslateAsyncConsumerError(nativeErr error) *AppError {
	if nativeErr == nil {
		return nil
	}

	// Try database first (common for DB-backed consumers)
	if dbCatalogEntry := ClassifyDatabaseError(nativeErr); dbCatalogEntry != nil {
		appErr := NewCatalogError(dbCatalogEntry)
		appErr.WithError(nativeErr)
		return appErr
	}

	// Try gRPC (if consumer calls services)
	if isGRPCError(nativeErr) {
		if grpcCatalogEntry := ClassifyGRPCError(nativeErr); grpcCatalogEntry != nil {
			appErr := NewCatalogError(grpcCatalogEntry)
			appErr.WithError(nativeErr)
			return appErr
		}
	}

	// Fallback to unknown
	appErr := NewCatalogError(ErrUnknown)
	appErr.WithError(nativeErr)
	return appErr
}

// isGRPCError checks if an error is a gRPC error without relying on deep inspection.
func isGRPCError(err error) bool {
	_, ok := status.FromError(err)
	return ok
}

// TranslationBoundaryLogger defines structured logging at translation boundaries.
// This should be called ONCE before translating and propagating an error.
type TranslationBoundaryLogger interface {
	// LogTranslation logs a native error translation event with context.
	LogTranslation(ctx interface{}, layer string, nativeErr error, appErr *AppError)
}

// CategoryMapping provides safe category transitions for error handling policies.
type CategoryMapping struct {
	from ErrorCategory
	to   ErrorCategory
	note string
}

// GetCategoryMappings returns recognized category transitions (useful for documentation).
func GetCategoryMappings() []CategoryMapping {
	return []CategoryMapping{
		{
			from: CategoryDependencyDB,
			to:   CategoryDependencyDB,
			note: "Database errors remain database category",
		},
		{
			from: CategoryDependencyGRPC,
			to:   CategoryDependencyGRPC,
			note: "gRPC errors remain gRPC category",
		},
		{
			from: CategoryValidation,
			to:   CategoryValidation,
			note: "Validation errors remain validation category",
		},
		{
			from: CategoryUnknown,
			to:   CategoryUnknown,
			note: "Unknown errors remain unknown category (mandatory fallback)",
		},
	}
}

// ValidateErrorContract checks that an AppError conforms to translation requirements.
// Returns nil if valid, otherwise returns a validation error describing the violation.
func ValidateErrorContract(appErr *AppError) error {
	if appErr == nil {
		return fmt.Errorf("AppError must not be nil")
	}

	if appErr.Message == "" {
		return fmt.Errorf("AppError.Message must not be empty")
	}

	if appErr.Code == "" {
		return fmt.Errorf("AppError.Code must not be empty")
	}

	if appErr.Category == "" {
		return fmt.Errorf("AppError.Category must not be empty")
	}

	return nil
}

// DefaultTranslationPolicy is the singleton used by all services for consistency.
var DefaultTranslationPolicy = NewTranslationPolicy()

// TranslateError is a convenience function to translate using the default policy.
// Layer should be one of: "repository", "service", "transport", "async_consumer".
func TranslateError(nativeErr error, layer string) *AppError {
	if nativeErr == nil {
		return nil
	}

	switch layer {
	case "repository":
		return DefaultTranslationPolicy.TranslateRepositoryError(nativeErr)
	case "service":
		return DefaultTranslationPolicy.TranslateServiceError(nativeErr, layer)
	case "transport":
		return DefaultTranslationPolicy.TranslateTransportError(nativeErr)
	case "async_consumer":
		return DefaultTranslationPolicy.TranslateAsyncConsumerError(nativeErr)
	default:
		// Unknown layer, use service layer rules
		return DefaultTranslationPolicy.TranslateServiceError(nativeErr, layer)
	}
}
