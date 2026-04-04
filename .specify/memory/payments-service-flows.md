# Payments Service RPC Flows

## Scope

This document maps Payments service gRPC flows and boundary policy expectations.

Notes:
- Payments service is the ownership boundary for cycle preference, history, and reconciliation data and must preserve tenant/project scope.
- BFF must consume payments capabilities through the Payments gRPC boundary only; it must not inject payments repositories or storage-backed payment services in-process.
- Final state: the BFF now consumes cycle preference, history analytics, and reconciliation capabilities through `payments.v1.PaymentsService` over gRPC only.
- Feature 012 extends the authenticated gateway path so these requests also receive `common.v1.Session`, while dashboard/history list-style flows keep deterministic `Pagination` defaults from the BFF.
- Repository/service/transport boundaries propagate `AppError` contracts only.
- Pointer-threshold policy applies on modified boundaries: pointer signatures are default for large/reference-like structs.
- Any intentional value-semantics exception must be documented in the feature-level pointer exception contract artifact.

## Shared gRPC flow

```mermaid
flowchart LR
    C[Caller e.g. BFF] -->|gRPC| G[Payments gRPC Server]
    G --> SVC[Payments Service Layer]
    SVC --> REPO[Payments Repositories]
    REPO --> DB[(PostgreSQL payments schema)]
    DB --> REPO
    REPO --> SVC
    SVC --> G
    G --> C
```

## Integration summary matrix

| Flow | Protocol | PostgreSQL | Redis | RabbitMQ |
|---|---|---|---|---|
| Payment-cycle preference queries | gRPC | Yes | No | No |
| History analytics (`timeline`, `categories`, `compliance`) | gRPC | Yes | No | No |
| Reconciliation projections and manual links | gRPC | Yes | No | Optional event consumers |

## Policy checklist

- Use pointer signatures for modified cross-layer contracts unless documented exception exists.
- Log native dependency errors once at boundary translation points.
- Return sanitized `AppError` values across service and transport boundaries.
