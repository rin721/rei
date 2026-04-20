# ADR 010: Move `internal/app` Lifecycle Logic Behind Explicit Runtime Boundaries

Date: 2026-04-20

Status: Accepted

## Context

After bootstrap phases and module-level providers were established, `internal/app` still had two concentrated lifecycle responsibilities:

- inline shutdown ordering in `app.go`
- inline hot-reload sequencing in `reload.go`

The result was that the app layer had become cleaner as a composition root, but long-lived runtime resources were still coordinated through ad hoc hand-written sequences.

## Decision

1. Move shutdown behavior into `internal/app/app_shutdown.go`.
2. Represent shutdown as explicit named steps so cleanup order is visible without keeping the logic inside `app.go`.
3. Treat runtime hot reload as a reloader registry rather than one long inline method.
4. Introduce `app_runtime_bindings.go` as the single place that synchronizes executor bindings shared by logger and HTTP server.
5. Keep `app.go` focused on app construction and state, not lifecycle sequencing.

## Consequences

1. Lifecycle behavior is now easier to evolve without reopening the main app container file.
2. Runtime resource coupling is more explicit:
   - executor binding is a shared runtime concern
   - reload and shutdown are independent orchestration paths
3. `internal/app` is closer to a true app shell with localized boundaries for startup, reload, shutdown, seeding, and module registration.

## Guardrails

- Do not move module business behavior into shutdown or reload files.
- Do not reintroduce long inline lifecycle chains into `app.go`.
- When a new long-lived runtime resource is added, register it through the explicit lifecycle boundary instead of attaching cleanup/reload logic ad hoc in multiple places.
