# ADR 012: Use Narrow App-Owned Provisioning Views Instead of Passing Root `App`

Date: 2026-04-20

Status: Accepted

## Context

After runtime containers were introduced, the app layer still had an important remaining leak:

- module providers still received the whole root `App`
- delivery assembly still read multiple runtime slices directly from the root app shell

That meant the internal structure was cleaner, but not yet fully reflected in the wiring API.

## Decision

1. Introduce `businessProvisioning` as the narrow business assembly view.
2. Introduce `deliveryProvisioning` as the narrow delivery assembly view.
3. Make module providers accept `businessProvisioning` instead of `*App`.
4. Make business seeders accept `businessProvisioning` instead of `*App`.
5. Make router and HTTP server assembly read from `deliveryProvisioning` instead of pulling directly from the full root app shell.

## Consequences

1. Module providers no longer depend on unrelated app state.
2. Seeding logic now depends on the same business-owned assembly surface as module registration.
3. Delivery assembly has a clearer boundary around:
   - handlers
   - middleware-facing runtime services
   - HTTP server config
4. The root `App` remains the composition shell, but fewer internals need to know its full shape.

## Guardrails

- Do not pass `*App` into module providers when a narrower provisioning view is sufficient.
- Do not let delivery assembly read arbitrary root-app state when a dedicated delivery view can express the required contract.
- Future app-layer refactors should prefer explicit app-owned provisioning views over reintroducing broad root-object coupling.
