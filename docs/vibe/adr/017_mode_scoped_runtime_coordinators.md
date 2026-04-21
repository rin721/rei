# ADR 017: Mode-Scoped Runtime Coordinators

Date: 2026-04-21

## Status

Accepted

## Context

The app shell had already delegated most bootstrap and lifecycle work to provisioning slices, but mode entrypoints still contained orchestration logic directly inside root-app methods.

In particular:

- server mode still manually stitched together bootstrap, start, reload-loop startup, and shutdown sequencing
- db mode still manually stitched together action dispatch, infrastructure bootstrap, migration commands, and cleanup sequencing
- `App.Run(...)` still knew about mode-specific execution functions instead of delegating to explicit runtime objects

This meant `internal/app` had slice-owned provisioning, but mode orchestration itself was still not modeled as a first-class concept.

## Decision

### 1. Modes are now represented by dedicated runtime coordinators

The app shell now resolves a mode runtime through `newModeRuntime()` and delegates execution through a small `modeRuntime` interface.

Current implementations:

- `serverModeRuntime`
- `dbModeRuntime`

### 2. Server mode owns server orchestration as one runtime object

`serverModeRuntime` now coordinates:

- infrastructure bootstrap
- business bootstrap
- delivery bootstrap
- reload-hook registration
- delivery start
- infrastructure reload-loop start
- deferred shutdown

### 3. DB mode owns DB orchestration as one runtime object

`dbModeRuntime` now coordinates:

- db action validation
- db infrastructure bootstrap
- migration command execution
- shutdown cleanup
- config-derived migration metadata such as dialect and migrations directory

### 4. Root app shell stops carrying dead mode wrappers

The root shell keeps:

- shared app state
- shared config
- mode runtime selection
- shared shutdown entrypoint

It no longer needs dead forwarding helpers for mode execution that add no domain value.

## Consequences

- Mode orchestration is now an explicit concept instead of being implicit in root-app methods.
- Future server/db mode changes can evolve behind runtime-specific objects without widening `App`.
- The next abstraction seam is now clearer: named infrastructure capability registration so mode runtimes compose capabilities declaratively rather than depending on hard-coded bootstrap step lists.
