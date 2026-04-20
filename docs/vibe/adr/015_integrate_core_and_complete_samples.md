# ADR 015: Integrate Core Infrastructure APIs And Complete Isolated Package Samples

Date: 2026-04-21

## Status

Accepted

## Context

The project replaced the implementations behind four infrastructure packages:

- `pkg/logger`
- `pkg/i18n`
- `pkg/executor`
- `pkg/storage`

Those packages now expose different construction and lifecycle APIs than the deleted legacy code. The application shell and middleware still referenced removed types and signatures, which left the repository in a broken build state.

In parallel, the `pkg/` tree still contained useful but isolated helpers that had no business-facing usage example inside `internal/`, especially `pkg/sqlgen` and `pkg/yaml2go`.

## Decision

### 1. Core infrastructure is integrated through the new package interfaces

The app shell must consume the rewritten packages through their new public contracts:

- `pkg/logger.Logger`
- `pkg/i18n.I18n`
- `pkg/executor.Manager`
- `pkg/storage.Storage`

`internal/app` runtime state and provisioning views now hold these interfaces directly instead of pointer-to-interface or removed concrete-manager types.

### 2. App-owned config mappers absorb compatibility work

We keep the current app config surface stable for now, and translate it inside `internal/app/config_helpers.go` into the new package-native config shapes.

This keeps the migration localized at the composition root and avoids leaking package churn into unrelated business modules.

### 3. HTTP server async execution uses an app-owned adapter

`pkg/httpserver` expects an `AsyncSubmitter`, while the rewritten executor package exposes `Manager.Execute(poolName, task)`.

The composition root now owns a small adapter from `pkg/executor.Manager` to `pkg/httpserver.AsyncSubmitter`, using a single default executor pool for HTTP server background work.

### 4. Middleware depends on interfaces, not concrete package internals

`internal/middleware` now accepts:

- `pkglogger.Logger`
- `pkgi18n.I18n`

Locale negotiation is now explicit request-header handling with fallback logic instead of calling the removed legacy `PickLocale` helper.

### 5. Isolated pkg helpers get business-facing demos through the sample module

The `sample` module is now the standard home for forward-looking business demos of otherwise isolated infrastructure/tooling helpers.

This round adds injectable demos for:

- `pkg/sqlgen`
- `pkg/yaml2go`

Those demos are exposed through the sample use case and routed by `GET /api/v1/samples/tooling`.

## Implemented Changes

### Core integration

- Rewired `internal/app` logger, i18n, executor, and storage initialization to the new constructors and reload shapes.
- Replaced removed executor assumptions with `pkgexecutor.NewManager(...)`.
- Replaced removed storage reload signature with `Reload(ctx, config)`.
- Removed the old logger executor binding assumption and kept executor binding only where it is still semantically valid: the HTTP server.

### Sample demos

- Added `sample.Tooling(...)` to the sample use case contract.
- Added DI-backed demo providers for:
  - offline schema preview with `pkg/sqlgen`
  - YAML-to-struct scaffolding with `pkg/yaml2go`
- Added prominent TODO guidance in the sample demo implementations to explain when future developers should extend each demo.
- Added route coverage so the demos are exercised through the delivery layer, not left as dead internal helpers.

## Explicit Non-Goals / Exclusions

- `pkg/cli*` is CLI infrastructure and should not be forced into HTTP/business sample flows.
- `pkg/i18nold` is legacy/deprecated code and should not receive new forward-looking business integrations.
- No new domain entity was introduced in this round.

## Consequences

- The repository build is green again under the rewritten core infrastructure packages.
- The composition root remains the only place where app-level compatibility shims are allowed.
- The sample module now has an explicit secondary role as the home for business-facing demos of useful but otherwise isolated internal packages.
