# Data Model: Standardize Backend App Errors

## Entity: ApplicationError

- Purpose: Shared cross-layer error contract used across backend modules.
- Fields:
  - `message` (string): sanitized and stable error message for propagation.
  - `retryable` (bool): retryability intent for future retry orchestration.
  - `cause` (error, optional): wrapped dependency-native error for internal diagnosis.
- Invariants:
  - Must be the only propagated error type across layer boundaries.
  - Message must be safe for upper-layer exposure.
  - `retryable` must be explicitly set by catalog semantics.

## Entity: ErrorCatalogEntry

- Purpose: Reusable predefined error definition in `backend/pkgs/errors/consts.go`.
- Fields:
  - `name` (string): canonical identifier used in code references.
  - `message` (string): stable sanitized message.
  - `retryable` (bool): default retryability classification.
  - `category` (enum): validation, auth, conflict, not_found, dependency_db, dependency_grpc, dependency_network, unknown.
- Invariants:
  - Must exist for each known failure category.
  - Must include a mandatory unknown-fallback entry.

## Entity: ErrorTranslationRule

- Purpose: Deterministic mapping from dependency-native failure to `ErrorCatalogEntry`.
- Fields:
  - `source_layer` (enum): repository, service, transport, rmq_consumer, grpc_handler.
  - `source_error_pattern` (string/condition): matching rule for native dependency error.
  - `target_catalog_entry` (reference): catalog error applied to propagation.
  - `log_required` (bool): whether boundary logging is mandatory before translation.
- Invariants:
  - Translation must occur before crossing layer boundary.
  - Boundary logging must happen once per translated error event.

## Entity: ErrorEventLogRecord

- Purpose: Structured operational event emitted when translating dependency-native errors.
- Fields:
  - `service_name` (string)
  - `layer` (string)
  - `operation` (string)
  - `project_id` (string, optional and sanitized)
  - `error` (native error object logged via `zap.Error`)
  - `catalog_entry` (string)
- Invariants:
  - Must not contain sensitive payload data.
  - Must include enough context to correlate incidents.

## Relationships

- `ErrorTranslationRule` maps native failures to one `ErrorCatalogEntry`.
- `ApplicationError` instances are created from one `ErrorCatalogEntry`.
- `ErrorEventLogRecord` is emitted when an `ErrorTranslationRule` is applied.

## State Transitions

1. Native dependency error occurs.
2. Translation rule is selected at boundary.
3. Structured error event is logged once.
4. Catalog-backed `ApplicationError` is propagated upward.
5. Upper layers handle only `ApplicationError` contract.
