# Data Model: BFF Route-Controller Segregation

## Entities

### 1. Route Contract

- **Purpose**: Defines the common behavior every BFF route module must expose.
- **Core Fields / Attributes**:
  - `resource_name`: stable identifier for the route group
  - `operations`: collection of route operations owned by the module
  - `controller_dependency`: injected controller capability set used by the module
  - `middleware_dependencies`: auth and role/project guard collaborators required during registration
- **Relationships**:
  - One route contract is implemented by many route modules.
  - Each route module delegates request execution to one controller capability set.

### 2. Controller Capability Contract

- **Purpose**: Represents the narrow set of handler behaviors a route module needs from a controller.
- **Core Fields / Attributes**:
  - `capability_name`: documents, projects, settings, payments, reconciliation, or history
  - `supported_handlers`: route-specific handler methods with Huma-compatible signatures
  - `shared_behavior`: optional helper behavior supplied through an embeddable base struct
- **Relationships**:
  - One controller module can satisfy one or more capability contracts.
  - A route module depends only on the capability contract it consumes.

### 3. Controller Module

- **Purpose**: Concrete behavior owner for one BFF resource area.
- **Core Fields / Attributes**:
  - `name`: resource-aligned controller name
  - `dependencies`: logger, gRPC clients, repositories, and services already used by handlers
  - `handler_methods`: concrete request handlers invoked by the route module
- **Relationships**:
  - Implements a controller capability contract.
  - Is injected into one corresponding route module.

### 4. Route Module

- **Purpose**: Concrete registration owner for a BFF resource route group.
- **Core Fields / Attributes**:
  - `name`: resource-aligned route module name
  - `route_inventory`: method/path/operation metadata entries owned by the module
  - `controller_dependency`: injected controller capability implementation
  - `registration_method`: common route-contract method that attaches operations to Huma
- **Relationships**:
  - Implements the route contract.
  - Registers many route operations.
  - Delegates each route operation to one controller handler.

### 5. Route Operation

- **Purpose**: Represents a single Huma endpoint that must remain stable through the refactor.
- **Core Fields / Attributes**:
  - `operation_id`
  - `method`
  - `path`
  - `summary`
  - `description`
  - `tags`
  - `middleware_stack`
  - `handler_binding`
- **Relationships**:
  - Owned by one route module.
  - Bound to one controller handler.
  - Covered by one or more integration scenarios.

### 6. Route Coverage Scenario

- **Purpose**: Represents an integration assertion proving one route is registered and behaves as expected.
- **Core Fields / Attributes**:
  - `scenario_name`: BDD-formatted test case name
  - `suite_name`: resource-scoped integration file or suite
  - `route_reference`: method + path + operation ID under test
  - `expected_result`: status code, middleware outcome, or response contract asserted by the test
- **Relationships**:
  - Each route operation must have at least one route coverage scenario.
  - Many scenarios can exist inside one resource-scoped integration suite.

### 7. Route Coverage Matrix

- **Purpose**: Maintained mapping artifact connecting route operations to integration coverage.
- **Core Fields / Attributes**:
  - `resource`
  - `operation_id`
  - `method`
  - `path`
  - `coverage_suite`
  - `coverage_status`
- **Relationships**:
  - Aggregates all route operations.
  - References route coverage scenarios that validate each operation.

## Relationships Summary

- A **Controller Module** implements one or more **Controller Capability Contracts**.
- A **Route Module** implements the **Route Contract** and depends on a **Controller Capability Contract**.
- A **Route Module** owns many **Route Operations**.
- Every **Route Operation** must map to at least one **Route Coverage Scenario**.
- The **Route Coverage Matrix** aggregates all **Route Operations** and links them to their coverage scenarios.

## Lifecycle

1. Define controller capability contracts for a resource.
2. Implement or adapt the concrete controller module to satisfy those contracts.
3. Register the resource's Huma operations inside a dedicated route module.
4. Wire the route module through the BFF Dig container.
5. Add or update integration scenarios until every route operation appears in the coverage matrix.