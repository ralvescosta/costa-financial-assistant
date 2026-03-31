// Package interfaces defines the canonical service and repository contracts for the payments domain.
// These interfaces consolidate the key contracts used by the payments service and are used as mock targets in tests.
package interfaces

import (
	"context"
	"time"
)

// CyclePreference holds the project-scoped preferred payment day configuration.
type CyclePreference struct {
	ID                  string
	ProjectID           string
	PreferredDayOfMonth int
	UpdatedBy           string
	UpdatedAt           time.Time
}

// PaymentCycleService defines the contract for managing payment cycle preferences per project.
// It is implemented by services.PaymentCycleService.
type PaymentCycleService interface {
	// GetCyclePreference returns the preferred payment day for the given project.
	// Returns nil, nil if no preference has been configured.
	GetCyclePreference(ctx context.Context, projectID string) (*CyclePreference, error)

	// UpsertCyclePreference creates or updates the preferred payment day for the project.
	// dayOfMonth must be between 1 and 28 inclusive.
	UpsertCyclePreference(ctx context.Context, projectID string, dayOfMonth int, updatedBy string) (*CyclePreference, error)
}

// PaymentCycleRepository defines the persistence contract for payment cycle preferences.
// It is implemented by repositories.PostgresPaymentCycleRepository.
type PaymentCycleRepository interface {
	// GetByProjectID returns the cycle preference for the project, or nil if absent.
	GetByProjectID(ctx context.Context, projectID string) (*CyclePreference, error)

	// Upsert creates or updates the cycle preference record and returns the persisted state.
	Upsert(ctx context.Context, projectID string, dayOfMonth int, updatedBy string) (*CyclePreference, error)
}
