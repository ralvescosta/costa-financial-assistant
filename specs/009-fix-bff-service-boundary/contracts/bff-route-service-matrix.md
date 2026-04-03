# BFF Route to Service Matrix

| Resource | Route Module | Controller | Service | Service Contract Type | Mapper Owner |
|----------|--------------|------------|---------|-----------------------|--------------|
| documents | backend/internals/bff/transport/http/routes/documents_routes.go | backend/internals/bff/transport/http/controllers/documents_controller.go | backend/internals/bff/services/documents_service.go | Pending refactor audit | transport/http/controllers/mappers |
| history | backend/internals/bff/transport/http/routes/history_routes.go | backend/internals/bff/transport/http/controllers/history_controller.go | backend/internals/bff/services/history_service.go | Pending refactor audit | transport/http/controllers/mappers |
| payments | backend/internals/bff/transport/http/routes/payments_routes.go | backend/internals/bff/transport/http/controllers/payments_controller.go | backend/internals/bff/services/payments_service.go | Pending refactor audit | transport/http/controllers/mappers |
| projects | backend/internals/bff/transport/http/routes/projects_routes.go | backend/internals/bff/transport/http/controllers/projects_controller.go | backend/internals/bff/services/projects_service.go | Pending refactor audit | transport/http/controllers/mappers |
| reconciliation | backend/internals/bff/transport/http/routes/reconciliation_routes.go | backend/internals/bff/transport/http/controllers/reconciliation_controller.go | backend/internals/bff/services/reconciliation_service.go | Pending refactor audit | transport/http/controllers/mappers |
| settings | backend/internals/bff/transport/http/routes/settings_routes.go | backend/internals/bff/transport/http/controllers/settings_controller.go | backend/internals/bff/services/settings_service.go | Pending refactor audit | transport/http/controllers/mappers |

## Notes

- This matrix is the single tracking source for boundary migration status.
- Update the Service Contract Type column to `proto`, `service-owned`, or `mixed` after each resource audit.
