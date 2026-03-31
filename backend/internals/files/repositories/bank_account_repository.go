package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"

	filesv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/files/v1"
)

// ErrBankAccountNotFound is returned when a queried bank account does not exist in the project scope.
var ErrBankAccountNotFound = errors.New("bank account not found")

// ErrDuplicateBankAccount is returned when a label already exists in the project.
var ErrDuplicateBankAccount = errors.New("bank account label already exists in this project")

// ErrBankAccountInUse is returned when the bank account is referenced by one or more statement records.
var ErrBankAccountInUse = errors.New("bank account is referenced by statement records and cannot be deleted")

// BankAccountRepository defines the project-scoped persistence contract for bank accounts.
type BankAccountRepository interface {
	Create(ctx context.Context, account *filesv1.BankAccount) (*filesv1.BankAccount, error)
	ListByProject(ctx context.Context, projectID string) ([]*filesv1.BankAccount, error)
	FindByProjectAndID(ctx context.Context, projectID, id string) (*filesv1.BankAccount, error)
	Delete(ctx context.Context, projectID, id string) error
}

// PostgresBankAccountRepository implements BankAccountRepository using PostgreSQL.
type PostgresBankAccountRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewBankAccountRepository constructs a PostgresBankAccountRepository.
func NewBankAccountRepository(db *sql.DB, logger *zap.Logger) BankAccountRepository {
	return &PostgresBankAccountRepository{db: db, logger: logger}
}

// Create inserts a new bank account row for the project.
func (r *PostgresBankAccountRepository) Create(ctx context.Context, account *filesv1.BankAccount) (*filesv1.BankAccount, error) {
	ctx, span := tracer.Start(ctx, "bank_account.create")
	defer span.End()
	span.SetAttributes(attribute.String("project_id", account.ProjectId))

	const query = `
		INSERT INTO bank_accounts (project_id, label, created_by)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at`

	var id, createdAt, updatedAt string
	err := r.db.QueryRowContext(ctx, query,
		account.ProjectId,
		account.Label,
		nullableUUID(account.CreatedBy),
	).Scan(&id, &createdAt, &updatedAt)
	if err != nil {
		span.RecordError(err)
		if isDuplicateBankAccountConstraint(err) {
			return nil, ErrDuplicateBankAccount
		}
		r.logger.Error("bank_account.create: insert failed",
			zap.String("project_id", account.ProjectId),
			zap.Error(err))
		return nil, fmt.Errorf("bank account repository: create: %w", err)
	}

	account.Id = id
	account.CreatedAt = createdAt
	account.UpdatedAt = updatedAt
	return account, nil
}

// ListByProject returns all bank accounts scoped to the given project, ordered by label.
func (r *PostgresBankAccountRepository) ListByProject(ctx context.Context, projectID string) ([]*filesv1.BankAccount, error) {
	ctx, span := tracer.Start(ctx, "bank_account.listByProject")
	defer span.End()
	span.SetAttributes(attribute.String("project_id", projectID))

	const query = `
		SELECT id, project_id, label, COALESCE(created_by::text,''), created_at, updated_at
		FROM bank_accounts
		WHERE project_id = $1
		ORDER BY label ASC`

	rows, err := r.db.QueryContext(ctx, query, projectID)
	if err != nil {
		span.RecordError(err)
		r.logger.Error("bank_account.listByProject: query failed",
			zap.String("project_id", projectID),
			zap.Error(err))
		return nil, fmt.Errorf("bank account repository: listByProject: %w", err)
	}
	defer rows.Close()

	var accounts []*filesv1.BankAccount
	for rows.Next() {
		a, err := scanBankAccount(rows)
		if err != nil {
			span.RecordError(err)
			r.logger.Error("bank_account.listByProject: scan failed", zap.Error(err))
			return nil, fmt.Errorf("bank account repository: listByProject scan: %w", err)
		}
		accounts = append(accounts, a)
	}
	if err := rows.Err(); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("bank account repository: listByProject rows: %w", err)
	}
	return accounts, nil
}

// FindByProjectAndID returns the bank account matching (projectID, id), or ErrBankAccountNotFound.
func (r *PostgresBankAccountRepository) FindByProjectAndID(ctx context.Context, projectID, id string) (*filesv1.BankAccount, error) {
	ctx, span := tracer.Start(ctx, "bank_account.findByProjectAndID")
	defer span.End()
	span.SetAttributes(
		attribute.String("project_id", projectID),
		attribute.String("bank_account_id", id),
	)

	const query = `
		SELECT id, project_id, label, COALESCE(created_by::text,''), created_at, updated_at
		FROM bank_accounts
		WHERE project_id = $1 AND id = $2`

	row := r.db.QueryRowContext(ctx, query, projectID, id)
	account, err := scanBankAccount(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrBankAccountNotFound
		}
		span.RecordError(err)
		r.logger.Error("bank_account.findByProjectAndID: query failed",
			zap.String("project_id", projectID),
			zap.String("bank_account_id", id),
			zap.Error(err))
		return nil, fmt.Errorf("bank account repository: findByProjectAndID: %w", err)
	}
	return account, nil
}

// Delete removes a bank account. Returns ErrBankAccountInUse if referenced by statement records.
func (r *PostgresBankAccountRepository) Delete(ctx context.Context, projectID, id string) error {
	ctx, span := tracer.Start(ctx, "bank_account.delete")
	defer span.End()
	span.SetAttributes(
		attribute.String("project_id", projectID),
		attribute.String("bank_account_id", id),
	)

	// Guard: check for referencing statement_records
	const checkQuery = `
		SELECT COUNT(*) FROM statement_records WHERE bank_account_id = $1`
	var count int
	if err := r.db.QueryRowContext(ctx, checkQuery, id).Scan(&count); err != nil {
		span.RecordError(err)
		return fmt.Errorf("bank account repository: delete attribution check: %w", err)
	}
	if count > 0 {
		return ErrBankAccountInUse
	}

	const query = `DELETE FROM bank_accounts WHERE project_id = $1 AND id = $2`
	res, err := r.db.ExecContext(ctx, query, projectID, id)
	if err != nil {
		span.RecordError(err)
		r.logger.Error("bank_account.delete: exec failed",
			zap.String("project_id", projectID),
			zap.String("bank_account_id", id),
			zap.Error(err))
		return fmt.Errorf("bank account repository: delete: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrBankAccountNotFound
	}
	return nil
}

// ── helpers ──────────────────────────────────────────────────────────────────

type bankAccountScanner interface {
	Scan(dest ...any) error
}

func scanBankAccount(row bankAccountScanner) (*filesv1.BankAccount, error) {
	var a filesv1.BankAccount
	var createdAt, updatedAt time.Time
	if err := row.Scan(&a.Id, &a.ProjectId, &a.Label, &a.CreatedBy, &createdAt, &updatedAt); err != nil {
		return nil, err
	}
	a.CreatedAt = createdAt.UTC().Format(time.RFC3339)
	a.UpdatedAt = updatedAt.UTC().Format(time.RFC3339)
	return &a, nil
}

func isDuplicateBankAccountConstraint(err error) bool {
	if err == nil {
		return false
	}
	return isDuplicateConstraint(err) || strings.Contains(err.Error(), "uq_bank_accounts_project_label")
}

func nullableUUID(s string) any {
	if s == "" {
		return nil
	}
	return s
}
