# ADR 014: Let Business and Delivery Slices Own Their Local Bootstrap Orchestration

Date: 2026-04-20

Status: Accepted

## Context

After `infrastructureProvisioning` took ownership of infrastructure lifecycle sequencing, the root `App` still directly coordinated two slice-local flows:

- business assembly
- delivery assembly

That left the app shell thinner than before, but it still knew more business and delivery bootstrap detail than necessary.

## Decision

1. Let `businessProvisioning` own business bootstrap orchestration.
2. Let `deliveryProvisioning` own delivery bootstrap orchestration.
3. Move business assembly flow behind the business slice:
   - dependency validation
   - repository-set creation
   - seeding
   - module provider assembly
   - handler bundle attachment
4. Move delivery assembly flow behind the delivery slice:
   - router creation
   - HTTP server creation
   - executor-aware HTTP server setup
5. Keep the root `App` responsible for coordinating slices, not for spelling out each slice’s internal bootstrap sequence.

## Consequences

1. Business and delivery orchestration is now more local to the slices that own the corresponding runtime state.
2. The root app shell is closer to a high-level coordinator for:
   - infrastructure
   - business
   - delivery
3. Future slice refactors can evolve their bootstrap sequence with less pressure to reopen root-app orchestration code.

## Guardrails

- Do not move slice-local bootstrap details back into the root `App` when a provisioning slice can own them.
- Do not pass the full root `App` into slice-local bootstrap logic where a narrower slice view is sufficient.
- Continue treating the root app shell as the coordinator of slices, not the implementation site for slice internals.
