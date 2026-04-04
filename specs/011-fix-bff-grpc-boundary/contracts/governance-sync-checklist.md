# Governance Sync Checklist

These sync items are part of the completion gate for `011-fix-bff-grpc-boundary`.

## Memory sync

- [ ] Update `.specify/memory/architecture-diagram.md` to show the final `Frontend -> BFF -> payments gRPC -> payments repository -> PostgreSQL` path.
- [ ] Update `.specify/memory/bff-flows.md` so payment-cycle, history, and reconciliation flows point to downstream payments RPCs instead of any temporary fallback state.
- [ ] Update `.specify/memory/payments-service-flows.md` with the actual payments gRPC transport and ownership notes.
- [ ] Update `/memories/repo/bff-service-boundary-conventions.md` with the final post-implementation rule and flow notes.

## Instruction sync

- [ ] Re-check `.github/instructions/architecture.instructions.md` for any additional wording needed after the new payments gRPC transport lands.
- [ ] Re-check `.github/instructions/project-structure.instructions.md` so the BFF and payments transport structure stays documented correctly.
- [ ] Re-check `.github/instructions/ai-behavior.instructions.md` if implementation reveals any additional deterministic BFF ownership rule that should be preserved.

## Verification sync

- [ ] Keep route-level integration tests in `backend/tests/integration/bff/` with snake_case names and BDD + AAA structure.
- [ ] Keep cross-service tests in `backend/tests/integration/cross_service/`.
- [ ] Record fresh evidence for `go test ./...` and short service startup checks before closing the feature.
