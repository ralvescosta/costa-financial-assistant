# BFF Endpoint Flows

## Scope

This document maps all current BFF HTTP endpoints and their flow through:
- Auth and authorization middleware
- Inter-service calls and protocol (HTTP/gRPC)
- Data interactions (PostgreSQL)
- Cache interactions (in-memory JWKS cache)
- Redis and RabbitMQ interactions (when present)

Notes:
- All endpoints are registered via Huma on Echo.
- Auth middleware uses Bearer JWT + JWKS key lookup.
- JWKS cache is in-memory inside BFF (not Redis).
- For endpoints that call downstream services via gRPC, DB details happen in those services.
- **006 boundary (enforced)**: controllers are pure HTTP adapters — they validate view contracts and call one BFF service method. All downstream gRPC orchestration lives in `internals/bff/services/`. HTTP contracts (request/response structs) are owned exclusively by `transport/http/views/`. Route modules own all `huma.Register(...)` calls.
- **AppError boundary rule (008)**: all downstream/native failures are translated to `AppError` before crossing service boundaries; BFF services log native errors once with structured context and propagate only sanitized `AppError` contracts.

## Shared auth and guard pattern (applies to all endpoints)

```mermaid
flowchart LR
    C[Client] -->|HTTPS + Bearer JWT| BFF[BFF Echo + Huma]
    BFF --> AUTH[Auth Middleware]
    AUTH --> EK[Extract KID from JWT]
    EK --> CACHE{JWKS key in cache?}
    CACHE -->|Yes| VERIFY[Validate JWT signature + claims]
    CACHE -->|No| ID[Identity Service via gRPC GetJwksMetadata]
    ID --> JWKSC[Populate in-memory JWKS cache]
    JWKSC --> VERIFY
    VERIFY --> GUARD[ProjectGuard role check read_only/update/write]
    GUARD --> CTRL[Endpoint Controller Handler]
```

---

## GET /api/v1/projects/current

```mermaid
flowchart LR
    C[Client] -->|HTTPS GET| BFF[BFF /projects/current]
    BFF --> AUTH[Auth + ProjectGuard read_only]
    AUTH --> PC[ProjectsController.handleGetCurrent]
    PC -->|BFF service| SVC[ProjectsService.GetCurrentProject]
    SVC -->|gRPC| ONB[Onboarding Service GetProject]
    ONB -->|SQL| ONBDB[(PostgreSQL onboarding: projects)]
    ONBDB --> ONB
    ONB --> SVC
    SVC --> PC
    PC --> C
```

Protocol: HTTPS -> BFF service -> gRPC
Data store: PostgreSQL (onboarding service)
Redis: none in this path
RabbitMQ: none in this path

## GET /api/v1/projects/members

```mermaid
flowchart LR
    C[Client] -->|HTTPS GET| BFF[BFF /projects/members]
    BFF --> AUTH[Auth + ProjectGuard read_only]
    AUTH --> PC[ProjectsController.handleListMembers]
    PC -->|BFF service| SVC[ProjectsService.ListProjectMembers]
    SVC -->|gRPC| ONB[Onboarding Service ListProjectMembers]
    ONB -->|SQL| ONBDB[(PostgreSQL onboarding: project_members)]
    ONBDB --> ONB
    ONB --> SVC
    SVC --> PC
    PC --> C
```

Protocol: HTTPS -> BFF service -> gRPC
Data store: PostgreSQL (onboarding service)
Redis: none in this path
RabbitMQ: none in this path

## POST /api/v1/projects/members/invite

```mermaid
flowchart LR
    C[Client] -->|HTTPS POST| BFF[BFF /projects/members/invite]
    BFF --> AUTH[Auth + ProjectGuard write]
    AUTH --> PC[ProjectsController.handleInvite]
    PC -->|gRPC| ONB[Onboarding Service InviteProjectMember]
    ONB -->|SQL read| UDB[(PostgreSQL onboarding: users by email)]
    ONB -->|SQL write| MDB[(PostgreSQL onboarding: project_members)]
    UDB --> ONB
    MDB --> ONB
    ONB --> PC
    PC --> C
```

Protocol: HTTPS -> gRPC
Data store: PostgreSQL (onboarding service)
Redis: none in this path
RabbitMQ: none in this path

## PATCH /api/v1/projects/members/{memberId}/role

```mermaid
flowchart LR
    C[Client] -->|HTTPS PATCH| BFF[BFF /projects/members/{memberId}/role]
    BFF --> AUTH[Auth + ProjectGuard write]
    AUTH --> PC[ProjectsController.handleUpdateRole]
    PC -->|gRPC| ONB[Onboarding Service UpdateProjectMemberRole]
    ONB -->|SQL update| MDB[(PostgreSQL onboarding: project_members)]
    MDB --> ONB
    ONB --> PC
    PC --> C
```

Protocol: HTTPS -> gRPC
Data store: PostgreSQL (onboarding service)
Redis: none in this path
RabbitMQ: none in this path

---

## POST /api/v1/documents/upload

```mermaid
flowchart LR
    C[Client] -->|HTTPS POST raw PDF| BFF[BFF /documents/upload]
    BFF --> AUTH[Auth + ProjectGuard update]
    AUTH --> DOC[DocumentsController.handleUpload]
    DOC --> HASH[Compute SHA-256 in BFF]
    HASH -->|gRPC| FS[Files Service UploadDocument]
    FS -->|SQL read| DEDUP[(PostgreSQL files: find by project_id + file_hash)]
    FS -->|SQL write tx| DOCS[(PostgreSQL files: documents)]
    DEDUP --> FS
    DOCS --> FS
    FS --> DOC
    DOC --> C
```

Protocol: HTTPS -> gRPC
Data store: PostgreSQL (files service)
Redis: none in this path
RabbitMQ: none in this path

## POST /api/v1/documents/{documentId}/classify

```mermaid
flowchart LR
    C[Client] -->|HTTPS POST| BFF[BFF /documents/{documentId}/classify]
    BFF --> AUTH[Auth + ProjectGuard update]
    AUTH --> DOC[DocumentsController.handleClassify]
    DOC -->|gRPC| FS[Files Service ClassifyDocument]
    FS -->|SQL update tx| DOCS[(PostgreSQL files: documents.kind)]
    DOCS --> FS
    FS --> DOC
    DOC --> C
```

Protocol: HTTPS -> gRPC
Data store: PostgreSQL (files service)
Redis: none in this path
RabbitMQ: none in this endpoint path

## GET /api/v1/documents

```mermaid
flowchart LR
    C[Client] -->|HTTPS GET| BFF[BFF /documents]
    BFF --> AUTH[Auth + ProjectGuard read_only]
    AUTH --> DOC[DocumentsController.handleList]
    DOC -->|gRPC| FS[Files Service ListDocuments]
    FS -->|SQL read| DOCS[(PostgreSQL files: documents)]
    DOCS --> FS
    FS --> DOC
    DOC --> C
```

Protocol: HTTPS -> gRPC
Data store: PostgreSQL (files service)
Redis: none in this path
RabbitMQ: none in this path

## GET /api/v1/documents/{documentId}

```mermaid
flowchart LR
    C[Client] -->|HTTPS GET| BFF[BFF /documents/{documentId}]
    BFF --> AUTH[Auth + ProjectGuard read_only]
    AUTH --> DOC[DocumentsController.handleGet]
    DOC -->|gRPC| FS[Files Service GetDocument]
    FS --> EXT[ExtractionService.GetDocumentDetail]
    EXT -->|SQL read| DOCS[(PostgreSQL files: documents)]
    EXT -->|SQL read optional| BILL[(PostgreSQL files: bill_records)]
    EXT -->|SQL read optional| STMT[(PostgreSQL files: statement_records + transaction_lines)]
    DOCS --> EXT
    BILL --> EXT
    STMT --> EXT
    EXT --> FS
    FS --> DOC
    DOC --> C
```

Protocol: HTTPS -> gRPC
Data store: PostgreSQL (files service)
Redis: none in this path
RabbitMQ: none in this path

---

## GET /api/v1/bank-accounts

```mermaid
flowchart LR
    C[Client] -->|HTTPS GET| BFF[BFF /bank-accounts]
    BFF --> AUTH[Auth + ProjectGuard read_only]
    AUTH --> SC[SettingsController.handleList]
    SC -->|gRPC| FS[Files Service ListBankAccounts]
    FS -->|SQL read| BA[(PostgreSQL files: bank_accounts)]
    BA --> FS
    FS --> SC
    SC --> C
```

Protocol: HTTPS -> gRPC
Data store: PostgreSQL (files service)
Redis: none in this path
RabbitMQ: none in this path

## POST /api/v1/bank-accounts

```mermaid
flowchart LR
    C[Client] -->|HTTPS POST| BFF[BFF /bank-accounts]
    BFF --> AUTH[Auth + ProjectGuard update]
    AUTH --> SC[SettingsController.handleCreate]
    SC -->|gRPC| FS[Files Service CreateBankAccount]
    FS -->|SQL write| BA[(PostgreSQL files: bank_accounts)]
    BA --> FS
    FS --> SC
    SC --> C
```

Protocol: HTTPS -> gRPC
Data store: PostgreSQL (files service)
Redis: none in this path
RabbitMQ: none in this path

## DELETE /api/v1/bank-accounts/{bankAccountId}

```mermaid
flowchart LR
    C[Client] -->|HTTPS DELETE| BFF[BFF /bank-accounts/{bankAccountId}]
    BFF --> AUTH[Auth + ProjectGuard update]
    AUTH --> SC[SettingsController.handleDelete]
    SC -->|gRPC| FS[Files Service DeleteBankAccount]
    FS -->|SQL check refs + delete| BA[(PostgreSQL files: bank_accounts/statement_records)]
    BA --> FS
    FS --> SC
    SC --> C
```

Protocol: HTTPS -> gRPC
Data store: PostgreSQL (files service)
Redis: none in this path
RabbitMQ: none in this path

---

## GET /api/v1/bills/payment-dashboard

```mermaid
flowchart LR
    C[Client] -->|HTTPS GET| BFF[BFF /bills/payment-dashboard]
    BFF --> AUTH[Auth + ProjectGuard read_only]
    AUTH --> PC[PaymentsController.handleGetDashboard]
    PC -->|gRPC| BS[Bills Service GetPaymentDashboard]
    BS -->|SQL read| BDB[(PostgreSQL bills: bill_records + bill_types)]
    BDB --> BS
    BS --> PC
    PC --> C
```

Protocol: HTTPS -> gRPC
Data store: PostgreSQL (bills service)
Redis: none in this path
RabbitMQ: none in this path

## POST /api/v1/bills/{billId}/mark-paid

```mermaid
flowchart LR
    C[Client] -->|HTTPS POST| BFF[BFF /bills/{billId}/mark-paid]
    BFF --> AUTH[Auth + ProjectGuard update]
    AUTH --> PC[PaymentsController.handleMarkPaid]
    PC -->|gRPC| BS[Bills Service MarkBillPaid]
    BS --> IDEMP{Idempotency key exists?}
    IDEMP -->|Yes| GETB[Read existing bill]
    IDEMP -->|No| UPD[Mark bill paid]
    BS -->|SQL read/write| BDB[(PostgreSQL bills: idempotency_keys + bill_records)]
    BDB --> IDEMP
    BDB --> GETB
    BDB --> UPD
    BS --> PC
    PC --> C
```

Protocol: HTTPS -> gRPC
Data store: PostgreSQL (bills service)
Redis: none in this path (idempotency is DB-backed)
RabbitMQ: none in this path

## GET /api/v1/payment-cycle/preferred-day

```mermaid
flowchart LR
    C[Client] -->|HTTPS GET| BFF[BFF /payment-cycle/preferred-day]
    BFF --> AUTH[Auth + ProjectGuard read_only]
    AUTH --> PC[PaymentsController.handleGetPreferredDay]
    PC --> SVC[PaymentCycleService.GetCyclePreference]
    SVC -->|SQL read| PDB[(PostgreSQL via BFF payments repo: payment_cycle_preferences)]
    PDB --> SVC
    SVC --> PC
    PC --> C
```

Protocol: HTTPS + in-process service call
Data store: PostgreSQL (direct from BFF)
Redis: none in this path
RabbitMQ: none in this path

## PUT /api/v1/payment-cycle/preferred-day

```mermaid
flowchart LR
    C[Client] -->|HTTPS PUT| BFF[BFF /payment-cycle/preferred-day]
    BFF --> AUTH[Auth + ProjectGuard update]
    AUTH --> PC[PaymentsController.handleSetPreferredDay]
    PC --> SVC[PaymentCycleService.UpsertCyclePreference]
    SVC -->|SQL upsert| PDB[(PostgreSQL via BFF payments repo: payment_cycle_preferences)]
    PDB --> SVC
    SVC --> PC
    PC --> C
```

Protocol: HTTPS + in-process service call
Data store: PostgreSQL (direct from BFF)
Redis: none in this path
RabbitMQ: none in this path

---

## GET /api/v1/reconciliation/summary

```mermaid
flowchart LR
    C[Client] -->|HTTPS GET| BFF[BFF /reconciliation/summary]
    BFF --> AUTH[Auth + ProjectGuard read_only]
    AUTH --> RC[ReconciliationController.getSummary]
    RC --> SVC[ReconciliationService.GetSummary]
    SVC -->|SQL read joins| RDB[(PostgreSQL via BFF payments repo: transaction_lines + reconciliation_links + bill_records)]
    RDB --> SVC
    SVC --> RC
    RC --> C
```

Protocol: HTTPS + in-process service call
Data store: PostgreSQL (direct from BFF)
Redis: none in this path
RabbitMQ: none in this path

## POST /api/v1/reconciliation/links

```mermaid
flowchart LR
    C[Client] -->|HTTPS POST| BFF[BFF /reconciliation/links]
    BFF --> AUTH[Auth + ProjectGuard update]
    AUTH --> RC[ReconciliationController.createLink]
    RC --> SVC[ReconciliationService.CreateManualLink]
    SVC -->|SQL insert| LDB[(PostgreSQL via BFF payments repo: reconciliation_links)]
    SVC -->|SQL update| TDB[(PostgreSQL via BFF payments repo: transaction_lines.status)]
    LDB --> SVC
    TDB --> SVC
    SVC --> RC
    RC --> C
```

Protocol: HTTPS + in-process service call
Data store: PostgreSQL (direct from BFF)
Redis: none in this path
RabbitMQ: none in this path

---

## GET /api/v1/history/timeline

```mermaid
flowchart LR
    C[Client] -->|HTTPS GET| BFF[BFF /history/timeline]
    BFF --> AUTH[Auth middleware]
    AUTH --> HC[HistoryController.timeline]
    HC -->|SQL aggregation| HDB[(PostgreSQL via BFF payments repo: bill_records)]
    HDB --> HC
    HC --> C
```

Protocol: HTTPS + direct repository query
Data store: PostgreSQL (direct from BFF)
Redis: none in this path
RabbitMQ: none in this path

## GET /api/v1/history/categories

```mermaid
flowchart LR
    C[Client] -->|HTTPS GET| BFF[BFF /history/categories]
    BFF --> AUTH[Auth middleware]
    AUTH --> HC[HistoryController.categories]
    HC -->|SQL aggregation + join| HDB[(PostgreSQL via BFF payments repo: bill_records + bill_types)]
    HDB --> HC
    HC --> C
```

Protocol: HTTPS + direct repository query
Data store: PostgreSQL (direct from BFF)
Redis: none in this path
RabbitMQ: none in this path

## GET /api/v1/history/compliance

```mermaid
flowchart LR
    C[Client] -->|HTTPS GET| BFF[BFF /history/compliance]
    BFF --> AUTH[Auth middleware]
    AUTH --> HC[HistoryController.compliance]
    HC -->|SQL aggregation| HDB[(PostgreSQL via BFF payments repo: bill_records)]
    HDB --> HC
    HC --> C
```

Protocol: HTTPS + direct repository query
Data store: PostgreSQL (direct from BFF)
Redis: none in this path
RabbitMQ: none in this path

---

## Integration summary matrix

| Endpoint Group | Main downstream interaction | Protocol | PostgreSQL | Redis | RabbitMQ |
|---|---|---|---|---|---|
| Projects | Onboarding service | gRPC | Yes (downstream) | No | No |
| Documents | Files service | gRPC | Yes (downstream) | No | No |
| Bank Accounts | Files service | gRPC | Yes (downstream) | No | No |
| Bills Dashboard/Mark Paid | Bills service | gRPC | Yes (downstream) | No | No |
| Payment Cycle | Payments service in-process | internal call + SQL | Yes (direct in BFF) | No | No |
| Reconciliation | Payments service in-process | internal call + SQL | Yes (direct in BFF) | No | No |
| History | Payments repository in-process | direct SQL | Yes (direct in BFF) | No | No |

## Observed cache/broker specifics

- JWKS cache: in-memory map inside BFF middleware layer, refreshed from identity service via gRPC.
- Redis: no active Redis integration in current BFF endpoint code paths.
- RabbitMQ: no active RabbitMQ interaction in current BFF endpoint code paths.
  - Files module has an analysis consumer for async extraction, but these BFF endpoints do not publish to a queue in current implementation.
