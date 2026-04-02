# Data Model: BFF HTTP Boundary Separation

## Entities

### 1. HTTP View Contract

- **Purpose**: Represents the request or response payload transmitted across the BFF HTTP boundary.
- **Core Fields / Attributes**:
  - `resource_name`: documents, projects, settings, payments, reconciliation, or history
  - `transport_role`: request, response, or nested payload fragment
  - `binding_tags`: path/query/json/body metadata used by Huma binding
  - `validation_tags`: `validator` rules required for controller-side contract checks
  - `schema_metadata`: documentation fields that must remain available for OpenAPI generation
- **Relationships**:
  - Consumed by one or more controller handler methods.
  - Referenced by route capability contracts.

### 2. Controller Capability

- **Purpose**: Represents the HTTP-facing handler methods a route module requires from a controller.
- **Core Fields / Attributes**:
  - `capability_name`: resource-scoped identifier
  - `supported_handlers`: list of handler methods for that route group
  - `input_views`: view contracts accepted by each handler
  - `output_views`: view contracts returned by each handler
- **Relationships**:
  - Implemented by one controller module.
  - Consumed by one route module.

### 3. Controller Module

- **Purpose**: Owns only HTTP-layer responsibilities for one BFF resource area.
- **Core Fields / Attributes**:
  - `name`: resource-aligned controller name
  - `service_dependency`: BFF service interface used for business orchestration
  - `validator_dependency`: runtime validator used to check bound views
  - `handler_methods`: methods that read claims/context, validate views, call services, and translate service results into HTTP responses
- **Relationships**:
  - Implements one controller capability.
  - Depends on one or more BFF service contracts.

### 4. BFF Service Contract

- **Purpose**: Defines the transport-agnostic operations controllers invoke to satisfy a route.
- **Core Fields / Attributes**:
  - `service_name`: resource-aligned name
  - `operation_set`: business/application operations exposed to controllers
  - `downstream_dependencies`: gRPC clients, repositories, or cross-service interfaces used internally
  - `request_model`: transport-neutral input expected from controllers
  - `response_model`: transport-neutral result returned to controllers
- **Relationships**:
  - Implemented by one concrete BFF service.
  - Called by one controller module.

### 5. Route Module

- **Purpose**: Owns the Huma registration metadata for one HTTP resource area.
- **Core Fields / Attributes**:
  - `name`: route module name
  - `route_inventory`: method/path/operation metadata collection
  - `controller_dependency`: injected controller capability
  - `middleware_stack`: auth and project-guard assignments for each operation
- **Relationships**:
  - Implements the shared route contract.
  - Delegates each registered operation to one controller capability handler.

### 6. Route Operation

- **Purpose**: Represents one active BFF HTTP endpoint that must remain behaviorally stable through the refactor.
- **Core Fields / Attributes**:
  - `operation_id`
  - `method`
  - `path`
  - `resource`
  - `middleware_stack`
  - `primary_handler`
- **Relationships**:
  - Owned by one route module.
  - Bound to one controller handler.
  - Covered by one or more route coverage scenarios.

### 7. Route Coverage Record

- **Purpose**: Records how a route operation is validated by integration tests.
- **Core Fields / Attributes**:
  - `operation_id`
  - `method`
  - `path`
  - `primary_suite`
  - `structural_suites`
  - `coverage_status`
- **Relationships**:
  - References one route operation.
  - Aggregated by the route coverage matrix artifact.

## Relationships Summary

- An **HTTP View Contract** is the sole transport payload type used between routes and controllers.
- A **Controller Module** implements a **Controller Capability** and depends on a **BFF Service Contract**.
- A **BFF Service Contract** hides downstream gRPC clients, repository interactions, and business orchestration from controllers.
- A **Route Module** owns many **Route Operations** and delegates them to one **Controller Capability**.
- Every **Route Operation** must map to at least one **Route Coverage Record** in the coverage matrix.

## Lifecycle

1. A route module registers a Huma operation and binds it to a controller capability method.
2. Huma binds request data into an HTTP view contract owned by `transport/http/views`.
3. The controller validates the bound view and request context, then translates the request into a service call.
4. The BFF service performs orchestration through downstream gRPC clients or repositories and returns a transport-neutral result.
5. The controller maps the service result back into an HTTP view response and returns it through Huma.
6. Integration tests and the route coverage matrix confirm that each route remains registered, documented, and behaviorally reachable.
