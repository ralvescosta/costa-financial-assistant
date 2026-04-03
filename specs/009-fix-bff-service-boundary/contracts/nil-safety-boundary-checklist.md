# Nil Safety Boundary Checklist

## Purpose

Define nil/empty handling checks for mapper and service boundaries introduced by
transport-to-service contract separation.

## Mapper Boundary Checks

- [ ] Nil input transport view returns deterministic validation/app error.
- [ ] Empty optional collections map to stable empty output (not panic).
- [ ] Nil proto/service response maps to deterministic HTTP-safe output.
- [ ] Pointer fields are guarded before dereference.

## Service Boundary Checks

- [ ] Nil input contract rejected with deterministic error category.
- [ ] Nil downstream response handled without panic.
- [ ] Optional pointer fields preserve semantics across mapper boundaries.
- [ ] Unknown/null downstream errors translate to sanitized AppError.

## Test Evidence

- [ ] Documents mapper nil/empty tests added.
- [ ] History mapper nil/empty tests added.
- [ ] Payments/projects/reconciliation/settings mapper nil/empty tests added.
- [ ] Service boundary tests include nil contract cases where applicable.
