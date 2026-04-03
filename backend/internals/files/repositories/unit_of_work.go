package repositories

import (
	"context"
	"database/sql"

	"github.com/ralvescosta/costa-financial-assistant/backend/internals/files/interfaces"
	apperrors "github.com/ralvescosta/costa-financial-assistant/backend/pkgs/errors"
)

// PostgresUnitOfWork implements UnitOfWork using a *sql.DB connection pool.
type PostgresUnitOfWork struct {
	db *sql.DB
}

// NewUnitOfWork constructs a PostgresUnitOfWork.
func NewUnitOfWork(db *sql.DB) interfaces.UnitOfWork {
	return &PostgresUnitOfWork{db: db}
}

// Begin starts a new database transaction.
func (u *PostgresUnitOfWork) Begin(ctx context.Context) (*sql.Tx, error) {
	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, apperrors.TranslateError(err, "repository")
	}
	return tx, nil
}

// Commit commits the transaction.
func (u *PostgresUnitOfWork) Commit(tx *sql.Tx) error {
	if err := tx.Commit(); err != nil {
		return apperrors.TranslateError(err, "repository")
	}
	return nil
}

// Rollback rolls back the transaction. Always call in a defer after Begin.
func (u *PostgresUnitOfWork) Rollback(tx *sql.Tx) error {
	if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
		return apperrors.TranslateError(err, "repository")
	}
	return nil
}
