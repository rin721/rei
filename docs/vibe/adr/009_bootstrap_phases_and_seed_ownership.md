# ADR 009: Use Explicit Bootstrap Phases and Module-Owned Seeders in `internal/app`

Date: 2026-04-20

Status: Accepted

## Context

After introducing module-level providers, `internal/app` was still carrying two kinds of orchestration noise:

- long mode-specific initialization chains in the server and DB startup paths
- centralized business seeding logic that still knew too much about individual modules

That meant the composition root was becoming cleaner, but startup and seed ownership were still harder to evolve safely.

## Decision

1. Introduce an explicit bootstrap-step mechanism in `internal/app/app_bootstrap.go`.
2. Split server startup into three named phases:
   - server infrastructure
   - business runtime
   - delivery runtime
3. Route DB mode through the same bootstrap-step mechanism so mode-specific initialization is described as phases instead of ad hoc init chains.
4. Turn `app_seed.go` into a seed registry only.
5. Move module seed ownership into dedicated seeders:
   - `app_seed_rbac.go`
   - `app_seed_sample.go`

## Consequences

1. Mode files are now orchestration entry points instead of long lists of infrastructure initialization calls.
2. Startup order is still explicit, but it is now grouped by runtime intent instead of by one monolithic function body.
3. Seed logic can evolve per module without reopening the main app bootstrap file.
4. The app layer still owns orchestration, but it carries less direct module-specific detail than before.

## Guardrails

- Do not move module business logic into bootstrap phases; phases should only describe startup order.
- Do not let `app_seed.go` grow back into a centralized business seed implementation.
- New seed responsibilities should be added as dedicated module seeders when they are module-specific.
- New modes should prefer the shared bootstrap-step mechanism instead of adding one-off init chains.
