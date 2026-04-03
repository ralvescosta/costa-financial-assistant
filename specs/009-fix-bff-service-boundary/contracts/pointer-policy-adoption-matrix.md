# Pointer Policy Adoption Matrix

Policy threshold:
- Use pointer semantics when struct has reference-like fields, or
- Use pointer semantics when value size is greater than 3 machine words.

| Module | Boundary Scope | Policy Status | Exception Record | Notes |
|--------|----------------|---------------|------------------|-------|
| bff | service contracts and mapper boundaries | Complete | None | Verified in interfaces and services tests |
| files | service/repository contracts | Complete | None | Verified upload boundary pointer input |
| bills | service/repository contracts | Complete | None | Verified mark-paid pointer return boundary |
| identity | service boundaries | Complete | None | Verified JWT/JWKS pointer-return boundaries |
| onboarding | service boundaries | Complete | None | Verified project/member pointer-return boundaries |
| payments | service boundaries | Complete | None | Verified cycle preference pointer-return boundaries |

## Status Legend

- Not started: no scoped boundary updates committed.
- In progress: one or more boundaries updated, validation pending.
- Complete: all scoped boundaries comply or have approved exceptions.
