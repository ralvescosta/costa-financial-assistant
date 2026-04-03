package errors

import (
	"errors"
	"fmt"
)

// Translator is responsible for converting native dependency errors
// to standardized AppError instances using catalog entries and translation rules.
type Translator interface {
	// Translate converts a native dependency error to an AppError using
	// the appropriate translation rule based on source layer and error type.
	Translate(sourceLayer string, nativeErr error) *AppError
	// TranslateWithFallback translates an error, falling back to the ErrUnknown
	// catalog entry if no specific translation rule matches.
	TranslateWithFallback(sourceLayer string, nativeErr error) *AppError
}

// SimpleTranslator provides basic translation functionality.
// It can be extended or replaced with more sophisticated translation logic.
type SimpleTranslator struct {
	rules map[string]*TranslationRule
}

// TranslationRule defines how a specific dependency error maps to a catalog entry.
type TranslationRule struct {
	// SourceLayer specifies where the error originated (e.g., "repository", "service", "transport").
	SourceLayer string
	// ErrorMatcher is a function that returns true if this rule applies to the native error.
	ErrorMatcher func(error) bool
	// TargetCatalog is the catalog entry to use for translation.
	TargetCatalog *CatalogEntry
	// LogRequired indicates whether a boundary log must be emitted before translation.
	LogRequired bool
}

// NewSimpleTranslator creates a new SimpleTranslator instance.
func NewSimpleTranslator() *SimpleTranslator {
	return &SimpleTranslator{
		rules: make(map[string]*TranslationRule),
	}
}

// RegisterRule adds a translation rule to the translator.
func (t *SimpleTranslator) RegisterRule(rule *TranslationRule) {
	key := fmt.Sprintf("%s:%s", rule.SourceLayer, rule.TargetCatalog.Code)
	t.rules[key] = rule
}

// Translate converts a native error to an AppError using registered rules.
// If no rule matches, it returns nil (caller should check and use fallback).
func (t *SimpleTranslator) Translate(sourceLayer string, nativeErr error) *AppError {
	if nativeErr == nil {
		return nil
	}

	// Attempt to find a matching rule
	for _, rule := range t.rules {
		if rule.SourceLayer == sourceLayer && rule.ErrorMatcher(nativeErr) {
			appErr := NewCatalogError(rule.TargetCatalog)
			appErr.WithError(nativeErr)
			return appErr
		}
	}

	return nil
}

// TranslateWithFallback translates an error to AppError, falling back to ErrUnknown
// if no specific rule matches. This ensures a deterministic response for all errors.
func (t *SimpleTranslator) TranslateWithFallback(sourceLayer string, nativeErr error) *AppError {
	if nativeErr == nil {
		return nil
	}

	translated := t.Translate(sourceLayer, nativeErr)
	if translated != nil {
		return translated
	}

	// Fallback to unknown error
	appErr := NewCatalogError(ErrUnknown)
	appErr.WithError(nativeErr)
	return appErr
}

// TranslateMultiError handles translation of wrapped errors (errors.Is/errors.As semantics).
// It attempts to unwrap and identify the most specific error category.
func TranslateMultiError(sourceLayer string, nativeErr error, translator Translator) *AppError {
	if nativeErr == nil {
		return nil
	}

	// Try the translator first
	if translated := translator.TranslateWithFallback(sourceLayer, nativeErr); translated != nil {
		return translated
	}

	// If translator didn't match, return unknown with wrapped error
	return NewCatalogError(ErrUnknown).WithError(nativeErr)
}

// IsAppError returns true if the error is an AppError or can be unwrapped to one.
func IsAppError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr)
}

// AsAppError unwraps and returns the AppError if present, nil otherwise.
func AsAppError(err error) *AppError {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	return nil
}

// DefaultTranslator is a singleton instance for convenient access.
// Services can register their translation rules during initialization.
var DefaultTranslator = NewSimpleTranslator()
