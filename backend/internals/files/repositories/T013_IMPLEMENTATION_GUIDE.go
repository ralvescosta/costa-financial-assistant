package repositories

// T013_IMPLEMENTATION_REFERENCE
//
// This file demonstrates the error translation pattern required for T013-T017.
// Each repository method should follow this pattern:
//
// BEFORE (current code):
//   func (r *PostgresDocumentRepository) FindByProjectAndID(ctx context.Context, projectID, id string) (*filesv1.Document, error) {
//       ...
//       if err != nil {
//           r.logger.Error("document.findByProjectAndID: query failed", zap.Error(err))
//           return nil, fmt.Errorf("document repository: findByProjectAndID: %w", err)
//       }
//   }
//
// AFTER (with appError translation):
//   import apperrors "github.com/ralvescosta/costa-financial-assistant/backend/pkgs/errors"
//
//   func (r *PostgresDocumentRepository) FindByProjectAndID(ctx context.Context, projectID, id string) (*filesv1.Document, error) {
//       ...
//       if err != nil {
//           if errors.Is(err, sql.ErrNoRows) {
//               return nil, apperrors.ErrResourceNotFound
//           }
//           // CRITICAL: Log native error ONCE at translation boundary
//           r.logger.Error("repository: database error in FindByProjectAndID", zap.Error(err))
//           // CRITICAL: Translate to AppError and return
//           appErr := apperrors.TranslateError(err, "repository")
//           return nil, appErr
//       }
//   }
//
// KEY RULES:
// 1. Import: apperrors "github.com/ralvescosta/costa-financial-assistant/backend/pkgs/errors"
// 2. Log native error ONCE using zap.Error(err) with context-specific message
// 3. Translate to AppError using apperrors.TranslateError(err, "repository")
// 4. Return AppError, NOT fmt.Errorf or wrapped error
// 5. Special case: sql.ErrNoRows → return apperrors.ErrResourceNotFound directly (no logging needed, not an error)
//
// APPLY THIS PATTERN TO:
// - document_repository.go: Create, FindByProjectAndHash, FindByProjectAndID, UpdateKind, UpdateAnalysisStatus, UpdateFailure, ListByProject, DeleteByProjectAndID
// All other repository files follow the same pattern.

// CHECKLIST FOR T013-T017 IMPLEMENTATION:
//
// Files Service (T013):
//   [ ] Add import apperrors
//   [ ] Update: Create (handle duplicate constraint → apperrors.ErrResourceAlreadyExists)
//   [ ] Update: FindByProjectAndHash
//   [ ] Update: FindByProjectAndID
//   [ ] Update: UpdateKind
//   [ ] Update: UpdateAnalysisStatus
//   [ ] Update: UpdateFailure
//   [ ] Update: ListByProject
//   [ ] Update: DeleteByProjectAndID
//   [ ] Test locally: go test ./internals/files/repositories/...
//
// Bills Service (T014):
//   [ ] Add import apperrors to payment_repository.go
//   [ ] Apply same pattern to all public methods
//   [ ] Test locally: go test ./internals/bills/repositories/...
//
// Onboarding Service (T015):
//   [ ] Add import apperrors to project_members_repository.go  
//   [ ] Apply same pattern to all public methods
//   [ ] Test locally: go test ./internals/onboarding/repositories/...
//
// Payments Service (T016-T017):
//   [ ] Add import apperrors to payment_cycle_repository.go
//   [ ] Add import apperrors to reconciliation_repository.go
//   [ ] Apply same pattern to all public methods
//   [ ] Test locally: go test ./internals/payments/repositories/...
