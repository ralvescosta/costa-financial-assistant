# Route, Controller, and Service Contract

## Purpose

Define the ownership boundary for active BFF HTTP routes after the refactor so route registration, HTTP contracts, controller behavior, and service orchestration each live in exactly one layer.

## Layer Responsibilities

| Layer | Owns | Must Not Own |
|------|------|--------------|
| `transport/http/routes` | `huma.Register(...)`, operation metadata, middleware assignment, binding each route to the correct controller capability | Request/response struct definitions, business rules, downstream client orchestration |
| `transport/http/views` | All HTTP request structs, response structs, nested JSON payload structs, binding tags, validator tags, and transport-facing schema metadata | Route registration, service orchestration, repository access |
| `transport/http/controllers` | Reading claims/context, validating bound view contracts, invoking BFF services, mapping service results to HTTP responses, translating service errors to HTTP errors | Direct gRPC client calls, repository calls, business rules, Huma route registration, ownership of transport structs |
| `services` | Downstream gRPC orchestration, repository-backed business/application workflows, transport-neutral request/response models | Huma or Echo types, route metadata, direct HTTP response shaping |

## Controller Contract Rules

- Every controller method must accept and return HTTP view types rather than controller-owned transport structs.
- Controllers must validate incoming view contracts before calling the service layer.
- Controllers must extract authenticated project context and pass transport-neutral data to the service layer.
- Controllers must not import generated gRPC client packages directly once the refactor is complete.
- Controllers must not instantiate or query repositories directly.

## Service Contract Rules

- Each active route group must have a corresponding BFF service contract or clearly named shared service contract.
- Service contracts must expose transport-agnostic operations oriented around application behavior, not HTTP details.
- Services may depend on downstream client interfaces already defined in `backend/internals/bff/interfaces/` and on existing payments service interfaces where appropriate.
- Service results returned to controllers must be mapped into view responses by controllers, not by route modules.

## View Contract Rules

- Every HTTP request and response struct used by active BFF routes must live in `backend/internals/bff/transport/http/views/`.
- Every field requiring runtime validation must define a `validate` tag.
- View structs must retain the tags needed for Huma binding and OpenAPI generation.
- Nested body payloads must also live in the views package instead of anonymous controller-owned structs.

## Route Capability Rules

- Route capability interfaces in `routes/contracts.go` must import `views`, not `controllers`, for request and response contracts.
- Route modules must stay resource-scoped and preserve the current method/path/operation inventory.
- Middleware order must remain auth before project-guard for operations that currently use both.

## Dependency Injection Rules

- `backend/cmd/bff/container.go` must wire validators, BFF services, controllers, and route modules through Dig.
- Controllers must be provided as route capability interfaces.
- Route modules remain the only transport objects directly invoked for Huma registration from container wiring.
