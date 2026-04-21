# ADR 020: Phase-Scoped Capability Dependencies

Date: 2026-04-21

## Status

Accepted

## Context

The infrastructure capability registry already used dependency-aware ordering, but each capability still had only one shared dependency list.

That meant:

- bootstrap, reload, and shutdown all inherited the same prerequisite graph
- a dependency needed by one lifecycle phase could accidentally constrain another phase
- the registry could not express the minimal graph that each lifecycle hook actually requires

## Decision

### 1. Capability prerequisites are now lifecycle-phase specific

Infrastructure capabilities may now declare separate dependency metadata for:

- bootstrap
- reload
- shutdown

### 2. Ordering now resolves dependencies per hook

Profile resolution now reads only the dependency set that belongs to the lifecycle hook being executed.

Current behavior:

- bootstrap uses `bootstrapDependencies`
- reload uses `reloadDependencies`
- shutdown uses `shutdownDependencies`

### 3. Shared ordering is no longer forced across phases

Independent lifecycle hooks on the same capability may now evolve without implicitly constraining each other.

This keeps bootstrap, reload, and shutdown closer to the real prerequisite graph of the resources they coordinate.

## Consequences

- Lifecycle ordering is now more precise because each phase declares only its own prerequisites.
- The registry is a better fit for future infrastructure concerns whose reload/shutdown dependencies differ from bootstrap.
- Tests now cover phase-scoped dependency behavior so hook-specific prerequisites do not silently leak into other lifecycle paths.
- The next refinement is to apply the same declarative lifecycle composition style to business and delivery slices where it meaningfully reduces root-app orchestration knowledge.
