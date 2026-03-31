// Package interfaces defines the canonical service and repository contracts for the files domain.
package interfaces

import (
	"context"
	"database/sql"

	filesv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/files/v1"
)

// BillRecordRepository defines the persistence contract for extracted bill data.
// It is implemented by repositories.PostgresBillRecordRepository and consumed by ExtractionService.
type BillRecordRepository interface {
	Create(ctx context.Context, tx *sql.Tx, record *filesv1.BillRecord) (*filesv1.BillRecord, error)
	FindByProjectAndDocumentID(ctx context.Context, projectID, documentID string) (*filesv1.BillRecord, error)
}

// StatementRecordRepository defines the persistence contract for extracted statement data.
// It is implemented by repositories.PostgresStatementRecordRepository and consumed by ExtractionService.
type StatementRecordRepository interface {
	Create(ctx context.Context, tx *sql.Tx, record *filesv1.StatementRecord) (*filesv1.StatementRecord, error)
	FindByProjectAndDocumentID(ctx context.Context, projectID, documentID string) (*filesv1.StatementRecord, error)
}
