# Current Task

Status: Completed

Completed On: 2026-04-21

Completed Task: Force-integrate the rewritten `pkg/logger`, `pkg/i18n`, `pkg/executor`, and `pkg/storage` APIs across the app shell, restore green builds, and backfill business-facing demos for isolated pkg helpers.

Completion Notes:
- `internal/app` now integrates the rewritten core infrastructure packages through the new API shapes instead of the removed legacy types:
  - `pkg/logger.Logger`
  - `pkg/i18n.I18n`
  - `pkg/executor.Manager`
  - `pkg/storage.Storage`
- `internal/middleware` now consumes logger and i18n through interfaces rather than pointer-to-interface anti-patterns, and locale selection is handled with explicit request-header fallback logic.
- `internal/app` now maps existing app config into the new infrastructure package configs:
  - logger config maps onto `Output`, `FilePath`, `MaxSize`, `MaxBackups`, `MaxAge`
  - i18n config maps onto `DefaultLanguage`, `SupportedLanguages`, `MessagesDir`
  - executor config maps onto explicit pool configs for `pkgexecutor.NewManager`
  - storage config maps onto `FSType`, `BasePath`, and watch settings
- HTTP server executor integration now uses an app-owned adapter from `pkg/executor.Manager` to `pkg/httpserver.AsyncSubmitter` instead of assuming the executor package exposes the old submitter shape.
- `sample` module now owns business-facing toolkit demos for previously isolated pkg helpers:
  - `pkg/sqlgen`
  - `pkg/yaml2go`
- Explicit exclusions were identified and documented for components that should not be forced into business demos:
  - `pkg/cli*` remains CLI-only infrastructure
  - `pkg/i18nold` remains deprecated legacy code and is not a valid forward-looking business dependency
- Validation passed with:
  - `go build ./...`
  - `go test ./...`

Next Recommended Task:
- Continue shrinking the root app shell by pushing runtime start/reload/shutdown registration further into slice-owned orchestration boundaries where cross-slice coordination is still centralized in `internal/app`.
