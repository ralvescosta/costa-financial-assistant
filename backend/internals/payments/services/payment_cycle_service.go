// Package services implements use-case logic for the payments domain.
package services

import (
	"context"
	"database/sql"
	"errors"

	"go.uber.org/zap"

	"github.com/ralvescosta/costa-financial-assistant/backend/internals/payments/interfaces"
	apperrors "github.com/ralvescosta/costa-financial-assistant/backend/pkgs/errors"
)

// ErrCyclePreferenceNotFound is returned when no payment cycle preference has been
// configured for the project.
var ErrCyclePreferenceNotFound = errors.New("payment cycle preference not found")

// paymentCycleRepository is a local alias so callers only import the interfaces package.
type paymentCycleRepository = interfaces.PaymentCycleRepository

// PaymentCycleService implements interfaces.PaymentCycleService.
type PaymentCycleService struct {
	repo   paymentCycleRepository
	db     *sql.DB
	logger *zap.Logger
}

// NewPaymentCycleService constructs a PaymentCycleService.
func NewPaymentCycleService(repo paymentCycleRepository, db *sql.DB, logger *zap.Logger) interfaces.PaymentCycleService {
	return &PaymentCycleService{repo: repo, db: db, logger: logger}
}

// GetCyclePreference returns the preferred payment day for the given project.
// Returns nil, nil if no preference has been configured.
func (s *PaymentCycleService) GetCyclePreference(ctx context.Context, projectID string) (*interfaces.CyclePreference, error) {
	pref, err := s.repo.GetByProjectID(ctx, projectID)
	if err != nil {
		s.logger.Error("cycle_service: get preference failed",
			zap.String("project_id", projectID),
			zap.Error(err))
		if appErr := apperrors.AsAppError(err); appErr != nil {
			return nil, appErr
		}
		return nil, apperrors.TranslateError(err, "service")
	}
	return pref, nil
}

// UpsertCyclePreference creates or updates the preferred payment day for the project.
// dayOfMonth must be between 1 and 28 inclusive.
func (s *PaymentCycleService) UpsertCyclePreference(ctx context.Context, projectID string, dayOfMonth int, updatedBy string) (*interfaces.CyclePreference, error) {
	if dayOfMonth < 1 || dayOfMonth > 28 {
		return nil, apperrors.NewCatalogError(apperrors.ErrValidationError)
	}

	pref, err := s.repo.Upsert(ctx, projectID, dayOfMonth, updatedBy)
	if err != nil {
		s.logger.Error("cycle_service: upsert preference failed",
			zap.String("project_id", projectID),
			zap.Int("day_of_month", dayOfMonth),
			zap.Error(err))
		if appErr := apperrors.AsAppError(err); appErr != nil {
			return nil, appErr
		}
		return nil, apperrors.TranslateError(err, "service")
	}

	s.logger.Info("cycle_service: preference upserted",
		zap.String("project_id", projectID),
		zap.Int("day_of_month", dayOfMonth))

	return pref, nil
}
