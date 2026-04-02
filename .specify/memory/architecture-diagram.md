# Costa Financial Assistant - Architecture Diagram

## System Architecture Overview

```mermaid
graph TB
    subgraph "Client Layer"
        Web["🌐 React Frontend<br/>(Vite + Tailwind)"]
    end

    subgraph "API Gateway & BFF"
        BFF["BFF Service<br/>(Echo + Huma)<br/>HTTP/REST + OpenAPI"]
    end

    subgraph "Core Business Services"
        Files["📁 Files Service<br/>(gRPC)<br/>Document Storage<br/>& Classification"]
        Bills["💰 Bills Service<br/>(gRPC)<br/>Bill Extraction<br/>& Analysis"]
        Payments["💳 Payments Service<br/>(gRPC)<br/>Payment Tracking<br/>& Reconciliation"]
        Onboarding["🚀 Onboarding Service<br/>(gRPC)<br/>User Registration<br/>& Setup"]
        Identity["🔐 Identity Service<br/>(gRPC)<br/>JWKS + Token<br/>Validation"]
    end

    subgraph "Data Layer"
        PG["🗄️ PostgreSQL<br/>Multi-tenant DB<br/>- Documents<br/>- Bills<br/>- Transactions<br/>- Users"]
        Redis["⚡ Redis<br/>Cache-Aside<br/>- JWT Cache<br/>- Query Cache<br/>- Session Data"]
        S3["📦 S3-Compatible<br/>Object Storage<br/>- PDF Files<br/>- Extracted Data"]
    end

    subgraph "Message Queue"
        RabbitMQ["🐰 RabbitMQ<br/>Async Processing<br/>- Document Processing<br/>- Reconciliation Queue<br/>- Payment Events"]
    end

    subgraph "Observability"
        OTEL["📊 OpenTelemetry<br/>- Tracing<br/>- Metrics<br/>- Logging"]
    end

    subgraph "Infrastructure"
        Migrations["🔄 Migrations Service<br/>Schema Management<br/>& Versioning"]
    end

    %% HTTP/REST connections
    Web -->|HTTP REST| BFF
    BFF -->|OpenAPI| OTEL

    %% gRPC service connections
    BFF -->|gRPC| Files
    BFF -->|gRPC| Bills
    BFF -->|gRPC| Payments
    BFF -->|gRPC| Onboarding
    BFF -->|gRPC| Identity

    %% Inter-service gRPC
    Bills -->|gRPC| Identity
    Payments -->|gRPC| Identity
    Onboarding -->|gRPC| Identity
    Files -->|gRPC| Identity

    %% Database connections
    BFF --> PG
    Files --> PG
    Bills --> PG
    Payments --> PG
    Onboarding --> PG
    Identity --> PG
    Migrations --> PG

    %% Cache connections
    BFF --> Redis
    Identity --> Redis
    Bills --> Redis
    Payments --> Redis

    %% Object storage
    Files --> S3
    Bills --> S3

    %% Message queue
    Files -->|Publish| RabbitMQ
    Bills -->|Consumer| RabbitMQ
    Payments -->|Consumer| RabbitMQ
    RabbitMQ -->|Events| Payments
    RabbitMQ -->|Events| Bills

    %% Observability
    Files -.->|Traces & Metrics| OTEL
    Bills -.->|Traces & Metrics| OTEL
    Payments -.->|Traces & Metrics| OTEL
    Onboarding -.->|Traces & Metrics| OTEL
    Identity -.->|Traces & Metrics| OTEL

    style Web fill:#e1f5f7
    style BFF fill:#fff3e0
    style PG fill:#f3e5f5
    style Redis fill:#e8f5e9
    style S3 fill:#fce4ec
    style RabbitMQ fill:#e0f2f1
    style OTEL fill:#f1f8e9
```

## Service Responsibilities

| Service | Protocol | Purpose | Dependencies |
|---------|----------|---------|--------------|
| **BFF** | Echo HTTP + Huma OpenAPI | API Gateway, user-facing REST endpoints; controllers are pure HTTP adapters; BFF services own all downstream gRPC orchestration; HTTP contracts live in `transport/http/views/` | All gRPC services, Redis, PostgreSQL |
| **Files** | gRPC | PDF document storage, classification (bill vs statement), async processing | PostgreSQL, S3, RabbitMQ, Identity, OpenTelemetry |
| **Bills** | gRPC | Bill extraction, payment status tracking, overdue analysis | PostgreSQL, Redis, RabbitMQ, Identity, OpenTelemetry |
| **Payments** | gRPC | Payment tracking, reconciliation, historical dashboards | PostgreSQL, Redis, RabbitMQ, Identity, OpenTelemetry |
| **Onboarding** | gRPC | User registration, project setup, team management | PostgreSQL, Identity, OpenTelemetry |
| **Identity** | gRPC | JWKS cache, JWT token validation, multi-tenant access control | PostgreSQL, Redis, OpenTelemetry |
| **Migrations** | CLI | Database schema versioning, run-and-exit pattern | PostgreSQL |

## Data Flow Examples

### 1. Document Upload Flow

```mermaid
flowchart LR
        FE[Frontend]
        BFF[BFF POST /upload]
        FS[Files Service gRPC]
        S3[S3 store PDF]
        PG1[PostgreSQL create doc record]
        MQ[RabbitMQ publish DocumentUploaded]
        BILLS[Bills Service Consumer extract data]
        PG2[PostgreSQL store extracted records]

        FE --> BFF --> FS --> S3 --> PG1 --> MQ --> BILLS --> PG2
```

### 2. Payment Dashboard Query Flow

```mermaid
flowchart LR
        FE[Frontend]
        BFF[BFF GET /payments]
        PAY[Payments Service gRPC]
        R[Redis cache-aside check]
        MISS{Cache miss?}
        PG[PostgreSQL]
        AGG[Sum aggregates and return]

        FE --> BFF --> PAY --> R --> MISS
        MISS -->|Yes| PG --> AGG
        MISS -->|No| AGG
```

### 3. Reconciliation Flow

```mermaid
flowchart LR
        MQ[RabbitMQ StatementReceived]
        PAY[Payments Service consumer]
        MATCH[Match statement transactions vs bills]
        UPDATE[Update payment status]
        PG[PostgreSQL persist matches]
        DONE[Publish ReconciliationComplete event]

        MQ --> PAY --> MATCH --> UPDATE --> PG --> DONE
```

## Communication Matrix

| From → To | Method | Purpose |
|-----------|--------|---------|
| Frontend → BFF | HTTP REST | User requests, queries, mutations |
| BFF → All Services | gRPC | Service-to-service orchestration |
| Any Service → PostgreSQL | SQL | CRUD operations, project-scoped queries |
| Any Service → Redis | RESP | Cache reads/writes, JWT cache, session data |
| Files → S3 | S3 API | Document storage/retrieval |
| Services → RabbitMQ | AMQP | Async event publishing/consuming |
| Services → OTEL | gRPC/HTTP | Trace/metric collection |

## Technology Stack by Service

| Layer | Technology | Version |
|-------|-----------|---------|
| **Frontend** | React 18+ | Vite, Tailwind CSS, TanStack Query |
| **BFF** | Go 1.21+ | Echo HTTP framework, Huma (OpenAPI-first) |
| **Services** | Go 1.21+ | gRPC, Protocol Buffers |
| **Database** | PostgreSQL 14+ | Multi-tenant, project-scoped access |
| **Cache** | Redis 7+ | Cache-aside pattern, JWT cache |
| **Storage** | S3-compatible | Minio/AWS S3 |
| **Message Queue** | RabbitMQ 3.11+ | AMQP, durable queues |
| **Observability** | OpenTelemetry | Jaeger (traces), Prometheus (metrics) |
| **Container Orch** | Docker Compose | Local dev, integration tests |

---

## Last Updated
- **Date**: 2026-03-31
- **Version**: 1.0.0
- **Services Count**: 7 (bff, files, bills, payments, identity, onboarding, migrations)
- **Services Status**: All designed, partial implementation

