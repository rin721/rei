# go-scaffold2

`go-scaffold2` is a phased scaffold for a modular Go backend aimed at medium-to-large web services.

The repository now completes Phase 0 through Phase 7:

- repository skeleton, root engineering files, and safe examples
- shared constants, errors, result envelope, and base contracts
- reusable `pkg/*` infrastructure packages
- `internal/config` config loading, env override, and hot reload
- `internal/app` application container for `server` and `initdb`
- `internal/middleware` and `internal/router` for the full HTTP chain
- `internal/models`, `internal/repository`, `internal/service`, and `internal/handler`
- working auth, user, rbac, and sample business routes
- executable `initdb` SQL generation, lock-file protection, and aligned docs
- final quality sweep with `go test ./...` and `go vet ./...`

## Current Stage

The scaffold is now in the Phase 7 acceptance-ready state. The planned runtime, business, and `initdb` workflows are all wired and verified.

## Module Path

```text
github.com/rei0721/go-scaffold2
```

## Quick Start

```bash
go list ./...
go test ./...
go vet ./...
go run ./cmd/server server --dry-run
go run ./cmd/server initdb --dry-run
go run ./cmd/server initdb
go run ./cmd/server
```

The default example config uses safe local SQLite settings so the scaffold can run without an external database.

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

## InitDB Output

Running `initdb` writes:

- `scripts/initdb/initdb.<driver>.sql`
- `scripts/initdb/.initdb.lock` after a successful non-dry-run execution

## Notes

- YAML keys use `snake_case`.
- Environment variables use `UPPER_SNAKE_CASE`.
- The API envelope is fixed to `code`, `message`, `data`, `traceId`, and `serverTime`.
- Example configuration only contains placeholders and safe defaults.
- Password hashes never appear in JSON responses.
