# Pointer Exceptions

This file tracks any approved value-semantics exceptions on backend boundaries touched by `011-fix-bff-grpc-boundary`.

## Default policy

Use pointer semantics on modified backend boundaries when a struct:
- contains reference-like fields, or
- exceeds the repository threshold of more than 3 machine words.

## Planned exceptions for this feature

At planning time, **no pointer-policy exceptions are approved**.

| Struct / Type | Boundary location | Reason for value semantics | Approval note |
|---|---|---|---|
| _None yet_ | — | — | Add only if implementation proves an explicit exception is justified |

## Rule

If an exception is introduced during implementation, it must be added here in the same feature cycle with a concrete rationale.
