# ADR 018: Infrastructure Capability Registry

Date: 2026-04-21

## Status

Accepted

## Context

The app shell had already moved mode orchestration into dedicated runtime coordinators, but infrastructure lifecycle composition still depended on several hard-coded ordered lists spread across:

- bootstrap
- reload
- shutdown

That meant adding, removing, or reordering one infrastructure concern still required editing multiple files and duplicated the identity of each infrastructure concern across lifecycle phases.

## Decision

### 1. Infrastructure concerns are now first-class capabilities

`internal/app` now defines infrastructure concerns once as named capabilities.

Each capability may provide one or more lifecycle hooks:

- bootstrap
- reload
- shutdown

### 2. Lifecycle composition now happens through named profiles

Infrastructure lifecycle composition now uses declarative profiles instead of repeated hard-coded step arrays.

Current profiles:

- `server-bootstrap`
- `db-bootstrap`
- `runtime-reload`
- `runtime-shutdown`

### 3. Capability identity is centralized

The registry is now the single place where infrastructure capability identity lives.

This means lifecycle wiring for concerns like `logger`, `database`, `executor`, and `storage` is no longer re-declared independently in bootstrap, reload, and shutdown files.

## Consequences

- Adding a new infrastructure concern now has a clearer extension point.
- Mode runtimes compose infrastructure behavior through profiles instead of knowing concrete step arrays.
- Infrastructure lifecycle behavior is easier to audit because capability identity and lifecycle hooks live together.
- Ordering is still profile-driven today; the next refinement is to derive order from declared capability prerequisites rather than relying on profile sequence alone.
