package repositories

import (
	"context"
	"database/sql"
	"fmt"
)

// UnitOfWork manages transactional boundaries for multi-step database writes.
type UnitOfWork interface {
	Begin(ctx context.Context) (*sql.Tx, error)
	Commit(tx *sql.Tx) error
	Rollback(tx *sql.Tx) error
}

// PostgresUnitOfWork implements UnitOfWork using a *sql.DB connection pool.
type PostgresUnitOfWork struct {
	db *sql.DB
}

// NewUnitOfWork constructs a PostgresUnitOfWork.
func NewUnitOfWork(db *sql.DB) UnitOfWork {
	return &PostgresUnitOfWork{db: db}
}

// Begin starts a new database transaction.
func (u *PostgresUnitOfWork) Begin(ctx context.Context) (*sql.Tx, error) {
	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("unit of work: begin: %w", err)
	}
	return tx, nil
}

// Commit commits the transaction.
func (u *PostgresUnitOfWork) Commit(tx *sql.Tx) error {
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("unit of work: commit: %w", err)
	}
	return nil
}

// Rollback rolls back the transaction. Always call in a defer after Begin.
func (u *PostgresUnitOfWork) Rollback(tx *sql.Tx) error {
	if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
		return fmt.Errorf("unit of work: rollback: %w", err)
	}
	return nil
}
