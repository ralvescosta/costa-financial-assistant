# Quickstart: Adopt Integration Test Standard

## Prerequisites
- Docker and Docker Compose available
- Go toolchain installed (matching repository version)
- Access to repository root

## 1. Identify test ownership
1. Determine if the scenario belongs to one backend service or spans multiple services.
2. Place the new file in one of:
   - `backend/tests/integration/<service>/`
   - `backend/tests/integration/cross_service/`

## 2. Create canonical filename
1. Describe the primary behavior outcome.
2. Convert to behavior-based snake_case.
3. Ensure the file ends with `_test.go`.

Example:
- Behavior: create bill succeeds
- Filename: `create_bill_success_test.go`

## 3. Author scenario using required structure
Use table-driven `t.Run` with explicit Given/When/Then fields.

```go
func TestCreateBill(t *testing.T) {
	type scenario struct {
		name  string
		given string
		when  string
		then  string
	}

	scenarios := []scenario{
		{
			name:  "Given valid request When creating bill Then persists and returns success",
			given: "a valid authenticated user and project context",
			when:  "the create bill endpoint is called",
			then:  "the bill is stored and response is successful",
		},
	}

	for _, tc := range scenarios {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			// setup dependencies, fixtures, and inputs

			// Act
			// execute transport-level call

			// Assert
			// verify response, persistence, and side effects
		})
	}
}
```

## 4. Use approved libraries
- Required:
  - `testing`
  - `github.com/stretchr/testify`
  - `github.com/testcontainers/testcontainers-go`

## 5. Keep DB lifecycle deterministic
- Use `TestMain` for suite-level ephemeral DB/container lifecycle.
- Apply migrations before tests execute.
- Ensure cleanup always runs.

## 6. Validate locally
From repository root:

```bash
cd backend
go test ./tests/integration/... -tags=integration
```

If your service has a dedicated target, also run project Make targets used by CI.

## 7. Review checklist before opening PR
1. Correct folder placement
2. Behavior-based snake_case filename
3. Table-driven BDD scenarios with explicit given/when/then
4. AAA readability in each scenario
5. Approved stack only
6. Deterministic setup and teardown

## 8. Compliance review checklist
Run these checks before requesting review:

1. Structural compliance check:

```bash
cd backend
./scripts/validate_integration_test_conventions.sh
```

2. Integration package validation:

```bash
cd backend
go test ./tests/integration/... -tags=integration
```

3. Traceability update:
	- update `specs/005-standardize-integration-tests/migration-mapping.md`
	- update `specs/005-standardize-integration-tests/migration-baseline.md`
4. Governance sync:
	- `.specify/memory/constitution.md`
	- `.github/instructions/testing.instructions.md`
	- `.github/instructions/ai-behavior.instructions.md`
