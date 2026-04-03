# Pointer Policy Exceptions

## Purpose

Document all approved value-semantics exceptions to the default pointer policy.

Default pointer policy:
- Use pointers when struct contains reference-like fields, or
- Use pointers when struct size is greater than 3 machine words.

## Exception Record Template

- Struct: 
- Boundary location: 
- Reason: immutable_small_value | safety_copy | compatibility_bridge
- Justification:
- Evidence (tests or compatibility note):
- Approved by:
- Date:

## Approved Exceptions

No approved exceptions for the audited US2 modules (bff, files, bills, identity, onboarding, payments).

## Audit Notes

- 2026-04-03: Verified pointer-based struct boundaries in targeted service contracts.
- 2026-04-03: Added policy compliance tests under bff/files/payments service test packages.
