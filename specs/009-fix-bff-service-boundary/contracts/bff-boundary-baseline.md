# BFF Boundary Baseline

## Purpose

Establish the pre-refactor baseline for BFF transport/service boundaries so
changes can be tracked deterministically during implementation.

## Current Active Route Modules

- documents: backend/internals/bff/transport/http/routes/documents_routes.go
- history: backend/internals/bff/transport/http/routes/history_routes.go
- payments: backend/internals/bff/transport/http/routes/payments_routes.go
- projects: backend/internals/bff/transport/http/routes/projects_routes.go
- reconciliation: backend/internals/bff/transport/http/routes/reconciliation_routes.go
- settings: backend/internals/bff/transport/http/routes/settings_routes.go

## Baseline Invariants to Verify

- BFF services must not depend on HTTP view contracts.
- Controllers and mapper helpers own transport <-> service mapping.
- Route modules own registration metadata and capability wiring only.
- API behavior semantics (status and response meaning) are preserved.

## Baseline Evidence Checklist

- [ ] Route modules inventoried.
- [ ] Service-layer imports reviewed for transport leakage.
- [ ] Existing contract shapes (proto vs service-owned) identified.
- [ ] Behavior regression checkpoints defined per resource.
