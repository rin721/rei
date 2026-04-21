# ADR 019: Dependency-Driven Capability Ordering

Date: 2026-04-21

## Status

Accepted

## Context

The infrastructure capability registry already centralized capability identity and profile composition, but profile membership was still expressed as an ordered list.

That meant:

- dependency order was still largely implicit
- shutdown ordering was not derived from dependency direction
- graph integrity problems such as cycles could not be detected explicitly

## Decision

### 1. Capabilities now declare prerequisites

Infrastructure capabilities may now declare dependency metadata directly in the registry.

This metadata represents lifecycle prerequisites between named capabilities.

### 2. Lifecycle ordering is now dependency-aware

Profile resolution now performs dependency-aware topological ordering for lifecycle execution.

Current behavior:

- bootstrap: topological order
- reload: topological order
- shutdown: reverse topological order

### 3. Profile order is now only a tie-breaker

Profile membership still selects which capabilities participate in a lifecycle path, but it is no longer the sole ordering mechanism.

When two capabilities are otherwise independent, profile order is used only to keep ordering deterministic.

### 4. Capability graph integrity is now validated

Profile resolution now fails fast when:

- a profile references an unknown capability
- the selected capability graph contains a dependency cycle

## Consequences

- Lifecycle order is now closer to actual infrastructure prerequisites.
- Shutdown semantics are safer because dependents stop before prerequisites.
- The next refinement is to split dependencies by lifecycle phase so bootstrap, reload, and shutdown can each express the minimal graph they really require.
