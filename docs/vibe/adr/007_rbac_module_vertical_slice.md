# ADR 007: Use RBAC as the Third Backward-Compatible Vertical Slice

Date: 2026-04-20

Status: Accepted

## Context

After the `user` and `auth` vertical slices, `rbac` was the last core business module still carrying direct leakage from persistence and infrastructure into the usecase layer.

Previously, `internal/service/rbac` mixed:

- HTTP DTOs from `types`
- direct references to `internal/models`
- repository implementation details
- explicit transaction callbacks using `gorm.DB`
- direct coupling to the runtime RBAC manager

inside the same usecase implementation.

## Decision

1. Refactor `rbac` as the third backward-compatible vertical slice.
2. Introduce pure RBAC domain entities under `internal/domain/rbac`:
   - `Role`
   - `RoleBinding`
   - `Policy`
3. Move RBAC usecase dependencies behind RBAC-specific ports:
   - user lookup
   - role persistence
   - role-binding persistence
   - policy persistence
   - transaction boundary
   - runtime enforcer behavior
4. Let `internal/handler/rbac_handler.go` own all translation between HTTP DTOs and RBAC usecase commands/results.
5. Introduce `internal/adapter/rbac/module.go` as the bridge from the RBAC usecase to legacy repositories and the existing in-memory RBAC manager.
6. Reduce the old shared `internal/service/base.go` to the only still-valid shared contract: `IDProvider`.

## Consequences

1. `internal/service/rbac` is now isolated from `internal/models`, repository details, DTO types, and `gorm.DB`.
2. The existing RBAC manager is still reused, but only through an explicit RBAC-module adapter boundary.
3. `user`, `auth`, and `rbac` now share the same migration pattern:
   `handler boundary -> usecase contracts -> domain entities -> adapter bridge -> legacy implementation`
4. The old shared service-layer abstraction bucket has been narrowed significantly, making future module boundaries easier to reason about.

## Guardrails

- Do not reintroduce `internal/models` or repository implementations into `internal/service/rbac`.
- Do not let `types.*` DTOs cross into `internal/service/rbac`.
- Do not reintroduce `gorm.DB` into RBAC usecase callbacks.
- Future RBAC integrations should extend the RBAC adapter or RBAC ports instead of leaking runtime manager details inward.
