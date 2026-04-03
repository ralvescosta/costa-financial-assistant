// Package repositories implements the persistence layer for the payments domain.
package repositories

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"go.uber.org/zap"

	"github.com/ralvescosta/costa-financial-assistant/backend/internals/payments/interfaces"
	apperrors "github.com/ralvescosta/costa-financial-assistant/backend/pkgs/errors"
)

// ErrCyclePreferenceNotFound is returned when no row is found for GetByProjectID.
var ErrCyclePreferenceNotFound = errors.New("payment cycle preference not found")

// PostgresPaymentCycleRepository implements interfaces.PaymentCycleRepository.
type PostgresPaymentCycleRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewPaymentCycleRepository constructs a PostgresPaymentCycleRepository.
func NewPaymentCycleRepository(db *sql.DB, logger *zap.Logger) interfaces.PaymentCycleRepository {
	return &PostgresPaymentCycleRepository{db: db, logger: logger}
}

// GetByProjectID returns the cycle preference for the project, or nil if absent.
func (r *PostgresPaymentCycleRepository) GetByProjectID(ctx context.Context, projectID string) (*interfaces.CyclePreference, error) {
	const q = `
		SELECT id, project_id, preferred_day_of_month, updated_by, updated_at
		FROM payment_cycle_preferences
		WHERE project_id = $1`

	var pref interfaces.CyclePreference
	var updatedAt time.Time

	err := r.db.QueryRowContext(ctx, q, projectID).Scan(
		&pref.ID,
		&pref.ProjectID,
		&pref.PreferredDayOfMonth,
		&pref.UpdatedBy,
		&updatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		r.logger.Error("payment cycle repo: get by project failed",
			zap.String("project_id", projectID),
			zap.Error(err))
		return nil, apperrors.TranslateError(err, "repository")
	}

	pref.UpdatedAt = updatedAt
	return &pref, nil
}

// Upsert creates or updates the cycle preference record and returns the persisted state.
func (r *PostgresPaymentCycleRepository) Upsert(ctx context.Context, projectID string, dayOfMonth int, updatedBy string) (*interfaces.CyclePreference, error) {
	const q = `
		INSERT INTO payment_cycle_preferences (project_id, preferred_day_of_month, updated_by, updated_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (project_id) DO UPDATE SET
			preferred_day_of_month = EXCLUDED.preferred_day_of_month,
			updated_by             = EXCLUDED.updated_by,
			updated_at             = NOW()
		RETURNING id, project_id, preferred_day_of_month, updated_by, updated_at`

	var pref interfaces.CyclePreference
	var updatedAt time.Time

	err := r.db.QueryRowContext(ctx, q, projectID, dayOfMonth, updatedBy).Scan(
		&pref.ID,
		&pref.ProjectID,
		&pref.PreferredDayOfMonth,
		&pref.UpdatedBy,
		&updatedAt,
	)
	if err != nil {
		r.logger.Error("payment cycle repo: upsert failed",
			zap.String("project_id", projectID),
			zap.Int("day_of_month", dayOfMonth),
			zap.Error(err))
		return nil, apperrors.TranslateError(err, "repository")
	}

	pref.UpdatedAt = updatedAt
	return &pref, nil
}
