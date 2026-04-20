# ADR 008: Split `internal/app` into Module-Level Providers

Date: 2026-04-20

Status: Accepted

## Context

After the `user`, `auth`, and `rbac` vertical slices were established, `internal/app` was still acting as a large centralized wiring unit.

Previously, one business bootstrap file mixed:

- infrastructure precondition checks
- repository-set construction
- seed orchestration
- module-specific dependency wiring
- handler bundle assembly

inside the same code path.

That structure made the application layer harder to evolve because every module refactor still had to reopen the same large bootstrap file.

## Decision

1. Keep `internal/app` as the composition root.
2. Narrow `internal/app/app_business.go` to bootstrap orchestration only.
3. Move module-specific dependency registration into dedicated provider files:
   - `app_module_auth.go`
   - `app_module_user.go`
   - `app_module_rbac.go`
   - `app_module_sample.go`
4. Introduce `app_modules.go` as the app-local registry that:
   - requests each module provider
   - collects the module outputs
   - builds the handler bundle
5. Move seed logic into `app_seed.go` so it no longer lives in the same file as business composition logic.

## Consequences

1. `internal/app` now reads as a composition root instead of a module implementation bucket.
2. Each module can evolve its adapter or usecase dependencies with a smaller, more localized wiring surface.
3. Future module additions can follow the same provider pattern without reopening a monolithic bootstrap function.
4. Seed logic is still in the app layer for now, but it is physically separated from the composition boundary and easier to relocate later.

## Guardrails

- Do not move module business logic back into `internal/app`; only module registration belongs there.
- Do not let provider files become a second service layer; they should only translate app-owned infrastructure into module dependencies.
- Keep `internal/app/app_business.go` focused on orchestration, not per-module dependency knowledge.
- When new modules are added, prefer a dedicated provider file over extending a shared switchboard function.
