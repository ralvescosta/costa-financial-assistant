# Files Service RPC Flows

## Scope

This document maps all current Files service gRPC RPCs and their flow through:
- gRPC server validation and handlers
- Document, extraction, and bank-account service orchestration
- Repository and transactional data interactions (PostgreSQL)
- Analysis pipeline context (analysis jobs and async consumer)
- Redis and RabbitMQ interactions (when present)

Notes:
- Files exposes gRPC APIs consumed primarily by BFF.
- Project isolation is enforced with `project_id` across all repository calls.
- Async extraction is handled by an RMQ consumer in this service, but these RPCs do not directly publish to RabbitMQ in current implementation.

## Shared gRPC service pattern (applies to all RPCs)

```mermaid
flowchart LR
    C[Caller e.g. BFF] -->|gRPC| GRPC[Files gRPC Server]
    GRPC --> VAL[Validate project context and required fields]
    VAL --> SVC[Document / Extraction / BankAccount Service]
    SVC --> REPO[Repositories + UnitOfWork where needed]
    REPO --> DB[(PostgreSQL files schema)]
    DB --> REPO
    REPO --> SVC
    SVC --> GRPC
    GRPC --> C
```

---

## RPC UploadDocument

```mermaid
flowchart LR
    C[Caller] -->|gRPC UploadDocument| G[grpc.Server.UploadDocument]
    G --> SVC[DocumentService.UploadDocument]
    SVC --> DEDUP[DocumentRepository.FindByProjectAndHash]
    SVC --> TX[UnitOfWork Begin/Commit]
    TX --> CREATE[DocumentRepository.Create]
    DEDUP --> DOCS[(documents)]
    CREATE --> DOCS
    DOCS --> DEDUP
    DOCS --> CREATE
    SVC --> G
    G --> C
```

Protocol: gRPC
Data store: PostgreSQL (files service: documents)
Redis: none in this path
RabbitMQ: none in this RPC path

## RPC ClassifyDocument

```mermaid
flowchart LR
    C[Caller] -->|gRPC ClassifyDocument| G[grpc.Server.ClassifyDocument]
    G --> SVC[DocumentService.ClassifyDocument]
    SVC --> TX[UnitOfWork Begin/Commit]
    TX --> UPD[DocumentRepository.UpdateKind]
    UPD --> DOCS[(documents.kind + analysis_status)]
    DOCS --> UPD
    SVC --> G
    G --> C
```

Protocol: gRPC
Data store: PostgreSQL (files service: documents)
Redis: none in this path
RabbitMQ: none in this RPC path

## RPC GetDocument

```mermaid
flowchart LR
    C[Caller] -->|gRPC GetDocument| G[grpc.Server.GetDocument]
    G --> EXT[ExtractionService.GetDocumentDetail]
    EXT --> DOC[DocumentRepository.FindByProjectAndID]
    DOC --> DOCS[(documents)]
    EXT --> BILL[BillRecordRepository.FindByProjectAndDocumentID]
    BILL --> BR[(bill_records)]
    EXT --> STMT[StatementRecordRepository.FindByProjectAndDocumentID]
    STMT --> SR[(statement_records + transaction_lines)]
    DOCS --> DOC
    BR --> BILL
    SR --> STMT
    EXT --> G
    G --> C
```

Protocol: gRPC
Data store: PostgreSQL (files service: documents + bill_records + statement_records + transaction_lines)
Redis: none in this path
RabbitMQ: none in this RPC path

## RPC ListDocuments

```mermaid
flowchart LR
    C[Caller] -->|gRPC ListDocuments| G[grpc.Server.ListDocuments]
    G --> SVC[DocumentService.ListDocuments]
    SVC --> REPO[DocumentRepository.ListByProject]
    REPO --> DOCS[(documents)]
    DOCS --> REPO
    REPO --> SVC
    SVC --> G
    G --> C
```

Protocol: gRPC
Data store: PostgreSQL (files service: documents)
Redis: none in this path
RabbitMQ: none in this RPC path

## RPC CreateBankAccount

```mermaid
flowchart LR
    C[Caller] -->|gRPC CreateBankAccount| G[grpc.Server.CreateBankAccount]
    G --> SVC[BankAccountService.CreateBankAccount]
    SVC --> REPO[BankAccountRepository.Create]
    REPO --> BA[(bank_accounts)]
    BA --> REPO
    REPO --> SVC
    SVC --> G
    G --> C
```

Protocol: gRPC
Data store: PostgreSQL (files service: bank_accounts)
Redis: none in this path
RabbitMQ: none in this path

## RPC ListBankAccounts

```mermaid
flowchart LR
    C[Caller] -->|gRPC ListBankAccounts| G[grpc.Server.ListBankAccounts]
    G --> SVC[BankAccountService.ListBankAccounts]
    SVC --> REPO[BankAccountRepository.ListByProject]
    REPO --> BA[(bank_accounts)]
    BA --> REPO
    REPO --> SVC
    SVC --> G
    G --> C
```

Protocol: gRPC
Data store: PostgreSQL (files service: bank_accounts)
Redis: none in this path
RabbitMQ: none in this path

## RPC DeleteBankAccount

```mermaid
flowchart LR
    C[Caller] -->|gRPC DeleteBankAccount| G[grpc.Server.DeleteBankAccount]
    G --> SVC[BankAccountService.DeleteBankAccount]
    SVC --> REPO[BankAccountRepository.Delete]
    REPO --> CHECK[Reference check]
    CHECK --> SR[(statement_records.bank_account_id)]
    REPO --> DEL[(bank_accounts)]
    SR --> CHECK
    DEL --> REPO
    REPO --> SVC
    SVC --> G
    G --> C
```

Protocol: gRPC
Data store: PostgreSQL (files service: bank_accounts + statement_records)
Redis: none in this path
RabbitMQ: none in this path

---

## Integration summary matrix

| RPC | Main interaction | Protocol | PostgreSQL | Redis | RabbitMQ |
|---|---|---|---|---|---|
| UploadDocument | Project-scoped dedup + create document | gRPC | Yes | No | No |
| ClassifyDocument | Update document kind and status-related fields | gRPC | Yes | No | No |
| GetDocument | Document detail with optional extracted records | gRPC | Yes | No | No |
| ListDocuments | Document list with keyset pagination | gRPC | Yes | No | No |
| CreateBankAccount | Create project bank account label | gRPC | Yes | No | No |
| ListBankAccounts | List project bank account labels | gRPC | Yes | No | No |
| DeleteBankAccount | Check references then delete label | gRPC | Yes | No | No |

## Observed cache/broker specifics

- Analysis pipeline tables exist in service (`analysis_jobs`, extracted records) and are updated by extraction flows.
- RabbitMQ: Files has an analysis consumer (`transport/rmq/analysis_consumer.go`) for async processing, but these gRPC RPCs do not publish queue messages directly in current implementation.
- Redis: no active Redis integration in Files RPC paths.
