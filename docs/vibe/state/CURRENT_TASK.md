# Current Task

Status: Completed

Completed On: 2026-04-20

Completed Task: Isolate `internal/app` lifecycle concerns so reload, shutdown, and runtime cross-resource bindings are managed through narrower app-owned boundaries.

Completion Notes:
- `internal/app/app_shutdown.go` now owns shutdown ordering through explicit named steps instead of a long inline cleanup chain in `app.go`.
- `internal/app/reload.go` now treats runtime reload as a registry of reloaders:
  - logger
  - cache
  - database
  - executor
  - http server
  - storage
- `internal/app/app_runtime_bindings.go` now isolates the executor-to-runtime binding point shared by logger and HTTP server.
- `internal/app/app.go` has been reduced to app construction and state only; lifecycle orchestration no longer lives in the main container file.
- This makes lifecycle behavior more local and prepares the app layer for a future split of long-lived runtime state into narrower containers.
- Validation passed with:
  - `go test ./internal/app ./internal/router`
  - `go test ./...`

Next Recommended Task:
- Continue shrinking `internal/app` by grouping long-lived runtime state into narrower app-owned containers, such as infrastructure, delivery, and business runtime slices.
