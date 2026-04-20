# ADR 011: Group `internal/app` State into Runtime Containers

Date: 2026-04-20

Status: Accepted

## Context

After startup phases, module providers, seed ownership, and lifecycle boundaries were separated, the `App` type was still carrying most long-lived runtime state as one flat field list.

That shape made the code harder to read because:

- infrastructure resources
- business runtime state
- delivery runtime state

were all still attached directly to the same root object.

## Decision

1. Introduce explicit runtime containers in `internal/app/app_runtime.go`.
2. Group app-owned state into:
   - `infrastructureRuntime`
   - `businessRuntime`
   - `deliveryRuntime`
3. Keep `App` as the root shell, but have it own the runtime slices through:
   - `a.infra`
   - `a.business`
   - `a.delivery`
4. Update initialization, module wiring, reload, shutdown, and delivery assembly to reference the appropriate runtime slice instead of flat fields.

## Consequences

1. The app layer now expresses runtime ownership more clearly.
2. Delivery concerns are easier to spot because router and HTTP server state live together.
3. Business runtime state is narrowed to business-owned runtime artifacts such as the handler bundle.
4. Infrastructure responsibilities remain substantial, but they are now explicitly grouped and therefore easier to split further later.

## Guardrails

- Do not reintroduce new long-lived runtime fields directly on `App` unless they truly belong to the app root.
- New runtime state should be attached to the matching container whenever possible.
- Future app-layer refactors should evolve the runtime slices or extract providers around them, instead of flattening the structure again.
