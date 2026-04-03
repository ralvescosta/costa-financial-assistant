# BFF Route Behavior Regression Matrix

## Purpose

Track per-resource behavior assertions to ensure boundary refactors preserve
status semantics and response contract meaning.

| Resource | Route Registration Asserted | Reachability Asserted | Status Semantics Asserted | Response Shape Asserted | Notes |
|----------|-----------------------------|------------------------|---------------------------|-------------------------|-------|
| documents | Pending | Pending | Pending | Pending | |
| history | Pending | Pending | Pending | Pending | |
| payments | Pending | Pending | Pending | Pending | |
| projects | Pending | Pending | Pending | Pending | |
| reconciliation | Pending | Pending | Pending | Pending | |
| settings | Pending | Pending | Pending | Pending | |

## Completion Rule

- Mark each column `Pass` only when a committed test asserts that behavior.
- Keep links to test files updated as assertions are added.
