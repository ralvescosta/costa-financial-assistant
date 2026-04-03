# Data Model: Enforce Service Boundary Contracts

## Entity: TransportViewContract

- Purpose: HTTP-specific request/response shape used only in BFF transport layer.
- Fields:
  - `direction` (enum): request | response.
  - `resource` (string): route resource domain (documents, history, etc.).
  - `schema_fields` (list): protocol-facing fields and validation metadata.
  - `validation_tags` (map): runtime validation rules applied in transport.
- Invariants:
  - Must not be imported into `backend/internals/bff/services`.
  - Must remain under `backend/internals/bff/transport/http/views/`.

## Entity: ServiceContract

- Purpose: Service-layer input/output contract independent from HTTP transport concerns.
- Fields:
  - `contract_name` (string): stable service contract identifier.
  - `origin` (enum): proto_domain_message | service_owned_struct.
  - `payload_fields` (list): orchestration-facing fields.
  - `nullability_rules` (map): required/optional expectations for service logic.
- Invariants:
  - Can be implemented by proto message reuse or service-owned struct.
  - Must not include HTTP-only semantics (headers, status metadata, route params formatting details).

## Entity: BoundaryMapper

- Purpose: Transport-owned mapping component that converts view contracts to service contracts and back.
- Fields:
  - `input_type` (reference): TransportViewContract.
  - `output_type` (reference): ServiceContract.
  - `mapping_rules` (list): deterministic field transformations.
  - `error_mapping` (list): deterministic mapper failure outputs.
- Invariants:
  - Mapping must occur before service invocation and before HTTP response return.
  - Services must receive already-transformed transport-neutral inputs.

## Entity: PointerConventionRule

- Purpose: Repository-wide rule governing pointer vs value semantics for structs crossing function boundaries.
- Fields:
  - `reference_like_field_detected` (bool): true when struct contains slice/map/chan/func/interface/pointer-containing composites.
  - `size_words` (integer): machine-word size estimate for struct.
  - `must_use_pointer` (bool): enforced pointer condition.
  - `exception_allowed` (bool): true only when documented exception exists.
- Invariants:
  - `must_use_pointer = true` when `reference_like_field_detected = true` OR `size_words > 3`.
  - Value semantics require an explicit exception record.

## Entity: PointerExceptionRecord

- Purpose: Explicitly documented exception to default pointer convention.
- Fields:
  - `struct_name` (string)
  - `boundary_location` (string)
  - `justification` (enum): immutable_small_value | safety_copy | compatibility_bridge.
  - `approver_note` (string)
- Invariants:
  - Every value-semantic exception must include non-empty justification.
  - Exception must be reviewable in code and/or governance docs.

## Relationships

- `BoundaryMapper` transforms `TransportViewContract` <-> `ServiceContract`.
- `PointerConventionRule` evaluates each struct boundary transition.
- `PointerExceptionRecord` overrides `PointerConventionRule` only when explicitly documented.

## State Transitions

1. HTTP request enters route/controller using `TransportViewContract`.
2. Boundary mapper transforms request into `ServiceContract`.
3. Service executes orchestration using transport-neutral contracts only.
4. Service output contract returns to transport layer.
5. Boundary mapper transforms output to response `TransportViewContract`.
6. For every cross-boundary struct in backend signatures, pointer policy is evaluated and applied (or exception documented).
