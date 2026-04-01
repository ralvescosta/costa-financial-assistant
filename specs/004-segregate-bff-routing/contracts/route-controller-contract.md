# Route-Controller Contract

## Purpose

Define the transport-layer contract for separating BFF route registration from controller behavior while preserving the current public API and middleware semantics.

## Route Contract

Every BFF route module must:

- Own all `huma.Register(...)` declarations for its resource area.
- Implement a shared registration contract that accepts a Huma API and the middleware collaborators required for the resource's operations.
- Register the exact public `OperationID`, HTTP method, path, summary, description, and tag metadata already exposed by the BFF unless a separate approved change explicitly updates the API.
- Preserve middleware ordering for each route, with auth middleware evaluated before project-role enforcement.

## Controller Contract

Every BFF controller module must:

- Remain a concrete struct with injected dependencies.
- Own request parsing, service/client orchestration, response shaping, and error mapping for its handlers.
- Satisfy only the narrow controller capability interfaces consumed by its route module.
- Avoid owning any `huma.Register(...)` declarations after this refactor.

## Shared Contract Rules

- Shared behavior must use Go-idiomatic composition: a minimal common contract plus small capability-specific interfaces.
- Reusable default behavior belongs in an embeddable base struct or helper methods, not in a broad interface with unused methods.
- Route modules must depend on interfaces, not concrete controller implementations, even when concrete structs are wired through Dig.

## Dependency Injection Contract

- `backend/cmd/bff/container.go` remains the sole place that wires concrete controllers and route modules.
- Route modules receive controller dependencies through Dig rather than creating or looking up controllers locally.
- Middleware constructors or factories are passed in the same registration flow used by the current BFF server bootstrap.

## Compatibility Contract

- No public route path changes.
- No public HTTP method changes.
- No public operation metadata loss.
- No changes to the authorization or project-isolation intent of existing routes.
- No business logic introduced into route modules.