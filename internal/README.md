# Internal

This directory stores project-private implementation packages.

Current modules:

- `internal/config` for typed config domains, env expansion, override, and hot reload
- `internal/app` for the runtime container, mode wiring, and `initdb`
- `internal/middleware` for the shared Gin middleware chain
- `internal/router` for route mounting and middleware ordering
- `internal/handler` for request binding and envelope responses
- `internal/service` for business rules and transaction boundaries
- `internal/repository` for persistence access
- `internal/models` for GORM entities and table definitions
