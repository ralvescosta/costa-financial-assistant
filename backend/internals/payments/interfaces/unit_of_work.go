package interfaces

import (
	"context"
	"database/sql"
)

// UnitOfWork manages transactional boundaries for multi-step database writes.
// It is implemented by repositories.PostgresUnitOfWork.
type UnitOfWork interface {
	Begin(ctx context.Context) (*sql.Tx, error)
	Commit(tx *sql.Tx) error
	Rollback(tx *sql.Tx) error
}
