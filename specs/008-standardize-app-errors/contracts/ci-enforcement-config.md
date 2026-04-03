# CI Enforcement Config: AppError Non-Leakage Gate

## Goal

Prevent regressions where raw dependency errors cross backend layer boundaries.

## Required CI Checks

1. `go test ./pkgs/errors/...`
2. `go test ./tests/integration/cross_service -run 'Test(AppErrorPropagationAcrossLayers_T012|NoRawErrorsLeakBoundary_T012|RetryabilityPreservation_T012|AsyncErrorNoSensitiveDataLeak_T061|UnknownFallback.*_T038|BoundaryLogging_.*_T064)'`
3. `go test ./internals/...`

Implemented in `.github/workflows/backend-app-error-gate.yml`.

## Suggested GitHub Actions Step

```yaml
- name: AppError boundary contract checks
  working-directory: backend
  run: |
    set -euo pipefail
    go test ./pkgs/errors/...
    go test ./tests/integration/cross_service -run 'Test(AppErrorPropagationAcrossLayers_T012|NoRawErrorsLeakBoundary_T012|RetryabilityPreservation_T012|AsyncErrorNoSensitiveDataLeak_T061|UnknownFallback.*_T038|BoundaryLogging_.*_T064)'
    go test ./internals/...
```

## Failure Conditions

- Any test failure in `pkgs/errors`.
- Any test failure in cross-service AppError contract suites.
- Any compile/test failure in backend `internals` packages.

## Enforcement Evidence

- Implementation cycle executed targeted suites successfully before phase closure.
- Contract tests now cover:
  - propagation type (`AppError`),
  - unknown fallback determinism,
  - async sanitization,
  - cross-service boundary logging traces.
