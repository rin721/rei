# rei

`rei` is a modular Go backend scaffold for medium-to-large web services.

## Current Architecture

- `cmd/` exposes the public CLI: `run` and `db {generate|migrate|status|rollback}`
- `internal/config` loads typed config with env override and hot reload
- `internal/app` assembles runtime and database-management entrypoints
- `internal/models`, `internal/repository`, `internal/service`, and `internal/handler` implement the current business stack
- `scripts/migrations/` is the only schema history source

## Module Path

```text
github.com/rin721/rei
```

## Quick Start

```bash
go list ./...
go test ./...
go vet ./...
go run ./cmd run --dry-run
go run ./cmd db migrate --dry-run
go run ./cmd run
```

The default example config uses local SQLite settings so the scaffold can run without external services.

## Migration Workflow

- Generate versioned SQL with `go run ./cmd db generate --desc <name>`
- Review scripts in `scripts/migrations/`
- Apply pending migrations with `go run ./cmd db migrate`
- Inspect state with `go run ./cmd db status`

## HTTP Routes

- `GET /health`
- `GET /api/v1/samples`
- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/logout`
- `POST /api/v1/auth/change-password`
- `POST /api/v1/auth/refresh`
- `GET /api/v1/users/me`
- `PUT /api/v1/users/me`
- `GET /api/v1/rbac/check`
- `POST /api/v1/rbac/roles/assign`
- `POST /api/v1/rbac/roles/revoke`
- `GET /api/v1/rbac/users/:user_id/roles`
- `GET /api/v1/rbac/roles/:role/users`
- `POST /api/v1/rbac/policies`
- `DELETE /api/v1/rbac/policies`
- `GET /api/v1/rbac/policies`

## Notes

- YAML keys use `snake_case`
- Environment variables use `UPPER_SNAKE_CASE`
- API responses use `code`, `message`, `data`, `traceId`, and `serverTime`
- Password hashes never appear in JSON responses
