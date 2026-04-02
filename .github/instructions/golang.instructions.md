---
applyTo: "**/*.go"
---

# Go Language Instructions

## Rule: Blocking Quality Gates

**Description**: Generated Go code must pass CI quality checks.

**When it applies**: Writing or modifying any Go file.

**Copilot MUST**:
- Produce code expected to pass `golangci-lint`.
- Produce code expected to pass `go vet` and `go test ./...`.
- Keep code `gofmt` and `goimports` compliant.

**Copilot MUST NOT**:
- Suggest knowingly lint-breaking code.
- Ignore warnings by default.
- Add `//nolint` without explicit, documented justification.

**Example input → expected Copilot output**:
- Input: "Quick fix this by ignoring lint."
- Expected output: provide a lint-compliant implementation instead of adding blanket ignore directives.

---

## Rule: Go Security Enforcement

**Description**: Security findings are blocking and must be addressed deterministically.

**When it applies**: Handling I/O, crypto, file/network access, or dependency updates.

**Copilot MUST**:
- Generate code expected to pass `gosec`.
- Prefer safe standard library and maintained dependencies.
- Validate risky input paths and return controlled errors.

**Copilot MUST NOT**:
- Introduce insecure defaults to speed delivery.
- Disable security rules without rationale in code review.
- Recommend vulnerable dependency usage.

**Example input → expected Copilot output**:
- Input: "Disable gosec for this SQL path."
- Expected output: use parameterized query patterns from repositories and keep the security check enabled.

---

## Rule: Idiomatic Error and Context Handling

**Description**: Use explicit error handling and context propagation in all I/O paths.

**When it applies**: Functions performing external calls, DB operations, messaging, or orchestration.

**Copilot MUST**:
- Accept `context.Context` as first parameter when appropriate.
- Propagate incoming context through repository/client/service calls.
- Return wrapped/contextual errors or domain errors consistently.
- Log relevant context before returning operational failures.

**Copilot MUST NOT**:
- Ignore returned errors.
- Use `panic` for recoverable flows.
- Create detached contexts (`context.Background()`) inside request paths.

**Example input → expected Copilot output**:
- Input: "Add repository call in document service."
- Expected output: forward `ctx` from the service method in `backend/internals/files/services/document_service.go` into repository method calls and handle returned errors explicitly.

---

## Rule: Keep Interfaces Consumer-Driven

**Description**: Define narrow interfaces around consumer needs.

**When it applies**: Adding interfaces or changing dependency contracts.

**Copilot MUST**:
- Keep interfaces small and purpose-specific.
- Define interfaces in the consuming package (`backend/internals/<service>/services/` or `backend/internals/bff/financial/controllers/`) so the layer depends on an abstraction it owns.
- Use constructor injection against interfaces in `backend/cmd/<service>/container.go`.
- For repository contracts, follow the architecture rule and declare them in `backend/internals/<service>/interfaces/`.

**Copilot MUST NOT**:
- Create broad "do everything" interfaces.
- Bind upper layers to concrete repository or client implementations.
- Add interface methods that are not used by any consumer.
- Declare exported repository interfaces in `backend/internals/<service>/repositories/`.

**Example input → expected Copilot output**:
- Input: "Add a new repository dependency to the document service."
- Expected output: define a narrow repository interface in `backend/internals/files/interfaces/`, implement it in `backend/internals/files/repositories/document_repository.go`, and wire via `backend/cmd/files/container.go`.

---

## Rule: Import and Formatting Discipline

**Description**: Keep imports deterministic and formatting tool-compatible.

**When it applies**: Any import or formatting change.

**Copilot MUST**:
- Group imports as standard library, third-party, internal.
- Keep consistent formatting with `goimports`.
- Avoid unnecessary aliasing unless collision or clarity requires it.

**Copilot MUST NOT**:
- Mix internal and external imports in random order.
- Leave unused imports.
- Introduce formatting style that requires manual cleanup.

**Example input → expected Copilot output**:
- Input: "Add new package imports for container wiring."
- Expected output: follow the three-group import pattern used in `backend/cmd/bff/container.go` and keep only consumed imports.

---

## Rule: Project Toolchain Requirements

**Description**: Use the designated toolchain libraries consistently.

**When it applies**: Adding CLI commands, configuration loading, DI wiring, or migrations.

**Copilot MUST**:
- Use `go.uber.org/dig` for all dependency injection wiring in `backend/cmd/<service>/container.go`.
- Use `github.com/spf13/cobra` for all CLI command definitions in `backend/cmd/<service>/cmd.go`.
- Use `github.com/spf13/viper` for configuration loading via `backend/pkgs/configs/`; resolve `${SECRET_KEY}` sentinel values through `backend/pkgs/secrets/`.
- Use `github.com/golang-migrate/migrate/v4` for database migrations; never alter schema outside migration files.
- Use `go.uber.org/zap` for all structured logging.
- Use `github.com/danielgtaylor/huma/v2` + `github.com/danielgtaylor/huma/v2/humaecho` for all BFF route registration.

**Copilot MUST NOT**:
- Use `flag`, `os.Args`, or alternative CLI libraries.
- Access environment variables directly in business logic — always read through `viper`/config.
- Inline raw SQL schema DDL outside migration files.
- Use `fmt.Printf` or the standard `log` package for operational logging.