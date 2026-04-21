# ADR 016: Slice-Owned Runtime Lifecycle

Date: 2026-04-21

## Status

Accepted

## Context

The app shell had already split bootstrap responsibilities across infrastructure, business, and delivery provisioning views, but runtime lifecycle behavior was still partially centralized in the root app shell.

Specifically:

- `runModeServer(...)` still called `httpServer.Start()` directly
- reload-hook registration and reload-loop startup still lived on the root `App`
- infrastructure shutdown still reached across slice boundaries to stop the delivery-owned HTTP server
- config reload still treated the delivery HTTP server as part of infrastructure reloading

This left the root shell knowing too much about slice-local lifecycle details.

## Decision

### 1. Delivery owns delivery runtime lifecycle

`deliveryProvisioning` is now the owner of delivery runtime lifecycle actions:

- start HTTP server
- reload HTTP server
- shut down HTTP server

The root app shell may coordinate delivery lifecycle, but it should not perform those delivery-local steps directly.

### 2. Infrastructure owns config-reload wiring

`infrastructureProvisioning` now owns:

- registering app reload hooks with the config manager
- starting the config watcher / reload loop
- reloading infrastructure-owned runtime dependencies
- shutting down infrastructure-owned resources

This keeps config-manager wiring inside the infrastructure slice where it belongs.

### 3. Root shutdown coordinates slices instead of flattening them

`App.Shutdown(...)` now coordinates:

- `deliveryProvisioning.shutdown(...)`
- `infrastructureProvisioning.shutdown(...)`

instead of routing delivery shutdown through infrastructure shutdown steps.

### 4. Config reload fans out by slice

`reloadComponents(...)` now dispatches reload work separately to:

- infrastructure provisioning
- delivery provisioning

This prevents infrastructure reloader registration from becoming a cross-slice catch-all.

## Consequences

- `runModeServer(...)` is thinner and more declarative.
- Delivery lifecycle behavior is easier to reason about and test independently.
- Infrastructure provisioning no longer has to pretend ownership over delivery shutdown/reload internals.
- The next thinning step is now clearer: introduce a mode-scoped runtime coordinator so server/db modes compose provisioning objects rather than stitching lifecycle calls directly in root-app mode handlers.
