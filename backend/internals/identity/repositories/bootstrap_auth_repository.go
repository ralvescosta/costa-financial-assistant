package repositories

import (
	"context"
	"database/sql"
	"fmt"

	goerrors "errors"

	"go.opentelemetry.io/otel"
	"go.uber.org/zap"

	apperrors "github.com/ralvescosta/costa-financial-assistant/backend/pkgs/errors"
)

var identityTracer = otel.Tracer("identity/repositories")

// BootstrapAuthRecord contains the seeded login data needed by the identity service.
type BootstrapAuthRecord struct {
	UserID       string
	ProjectID    string
	Username     string
	Email        string
	PasswordHash string
	Role         string
	ProjectName  string
}

// BootstrapAuthRepository resolves the seeded owner user for the login flow.
type BootstrapAuthRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewBootstrapAuthRepository constructs a BootstrapAuthRepository.
func NewBootstrapAuthRepository(db *sql.DB, logger *zap.Logger) *BootstrapAuthRepository {
	return &BootstrapAuthRepository{db: db, logger: logger}
}

// FindBootstrapUser returns the seeded bootstrap user by username or email.
func (r *BootstrapAuthRepository) FindBootstrapUser(ctx context.Context, username string) (*BootstrapAuthRecord, error) {
	ctx, span := identityTracer.Start(ctx, "BootstrapAuthRepository.FindBootstrapUser")
	defer span.End()

	const q = `
		SELECT
			u.id::text,
			COALESCE(u.project_id::text, ''),
			COALESCE(u.username, ''),
			COALESCE(u.email, ''),
			COALESCE(u.password_hash, ''),
			COALESCE(u.role, ''),
			COALESCE(p.name, '')
		FROM users u
		LEFT JOIN projects p ON p.id = u.project_id
		WHERE LOWER(COALESCE(u.username, u.email)) = LOWER($1)
		LIMIT 1`

	record := &BootstrapAuthRecord{}
	if err := r.db.QueryRowContext(ctx, q, username).Scan(
		&record.UserID,
		&record.ProjectID,
		&record.Username,
		&record.Email,
		&record.PasswordHash,
		&record.Role,
		&record.ProjectName,
	); err != nil {
		if goerrors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.NewCatalogError(apperrors.ErrResourceNotFound).WithError(err)
		}
		r.logger.Error("bootstrap_auth_repository.find_bootstrap_user: query failed",
			zap.String("username", username),
			zap.Error(err))
		return nil, fmt.Errorf("identity auth lookup failed: %w", apperrors.TranslateError(err, "repository"))
	}

	return record, nil
}
