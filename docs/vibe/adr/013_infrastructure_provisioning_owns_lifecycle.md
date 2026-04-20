# ADR 013: Let `infrastructureProvisioning` Own Infrastructure Lifecycle Orchestration

Date: 2026-04-20

Status: Accepted

## Context

After runtime containers and narrow business/delivery provisioning views were introduced, one large concentration still remained:

- the root `App` still assembled the infrastructure lifecycle by directly chaining
  - `init*`
  - `reload*`
  - `close*`
  helpers

That meant infrastructure state had been grouped, but infrastructure orchestration still lived too high in the app shell.

## Decision

1. Introduce `infrastructureProvisioning` as the app-owned infrastructure lifecycle view.
2. Route server-mode bootstrap and DB-mode bootstrap through infrastructure-owned bootstrap step lists.
3. Route runtime reload through infrastructure-owned reloader registration.
4. Route shutdown through infrastructure-owned shutdown step registration.
5. Move infrastructure helper receivers for long-lived infrastructure resources onto `infrastructureProvisioning`.

## Consequences

1. The root `App` now delegates infrastructure lifecycle orchestration instead of spelling it out inline.
2. Infrastructure lifecycle behavior is now local to the infrastructure slice.
3. Cross-resource concerns such as executor binding remain explicit, but they are owned by the infrastructure boundary instead of the root shell.
4. The app shell is closer to a pure coordinator for slice-level providers rather than a place where every infrastructure helper is chained manually.

## Guardrails

- Do not reintroduce root-level chains of infrastructure `init*`, `reload*`, or `close*` helpers in `App`.
- New infrastructure resources should be registered through `infrastructureProvisioning` rather than attached ad hoc to root lifecycle methods.
- Continue to use narrow provisioning views instead of passing `*App` where a slice-owned contract is sufficient.
