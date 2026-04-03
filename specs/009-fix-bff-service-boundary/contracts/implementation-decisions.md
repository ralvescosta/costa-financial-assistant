# Implementation Decisions Log

## Decision Template

- Date:
- Decision:
- Scope:
- Rationale:
- Alternatives considered:
- Follow-up actions:

## Decisions

### 2026-04-03 - Initialize boundary and policy tracking artifacts

- Scope: Feature 009 setup phase
- Decision: Create baseline/matrix/log artifacts before code changes
- Rationale: Deterministic tracking of boundary migration and pointer-policy adoption
- Alternatives considered: ad-hoc notes in pull request only (rejected)
- Follow-up actions: keep this log updated per completed implementation phase

### 2026-04-03 - Enforce BFF service-contract ownership boundary

- Scope: BFF controllers/services/routes and boundary tests
- Decision: Keep service-owned contracts in `backend/internals/bff/services/contracts/` and transport view types in `backend/internals/bff/transport/http/views/`; require mapper-layer conversion in `backend/internals/bff/transport/http/controllers/mappers/`
- Rationale: Prevent transport-type leakage into BFF service interfaces and keep controller layer thin and deterministic
- Alternatives considered: direct view usage in service signatures (rejected), controller-local ad-hoc conversion (rejected)
- Follow-up actions: preserve ownership rules in instructions and memory artifacts

### 2026-04-03 - Standardize pointer-threshold policy across modified backend boundaries

- Scope: BFF/files/bills/identity/onboarding/payments boundary signatures and tests
- Decision: Apply pointer semantics by default for large/reference-like structs; document approved value-semantics exceptions in feature contract artifacts
- Rationale: Reduce copy overhead and ensure predictable nil-safety at conversion boundaries
- Alternatives considered: broad value semantics for all contracts (rejected), undocumented per-file exceptions (rejected)
- Follow-up actions: keep pointer-policy coverage in boundary tests and update policy guidance files when boundaries evolve

### 2026-04-03 - Final validation and governance completion gate

- Scope: Feature completion and release-readiness evidence
- Decision: Require full backend regression, integration convention validation, and canonical smoke convention verification before marking feature complete
- Rationale: Ensure behavior safety and deterministic policy propagation beyond code changes
- Alternatives considered: partial test validation only (rejected)
- Follow-up actions: none
