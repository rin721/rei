# ADR 006: Use Auth as the Second Backward-Compatible Vertical Slice

Date: 2026-04-20

Status: Accepted

## Context

After the `user` module vertical slice, the next highest-leverage target was `auth`, because it previously mixed:

- HTTP DTOs
- GORM models
- JWT implementation details
- cache session handling
- transaction manager details

inside the same usecase implementation.

That made `auth` the most obvious example of infrastructure leakage into business logic.

## Decision

1. Refactor `auth` as the second vertical slice while keeping external API compatibility.
2. Keep `internal/domain/user.User` as the core user entity reused by auth flows.
3. Move auth-facing infrastructure dependencies behind auth-specific ports:
   - token issuing and refresh validation
   - refresh-token session storage
   - transaction boundary
   - user/role/role-binding persistence
4. Let `internal/handler/auth_handler.go` own all DTO translation between `types/user` and the auth usecase.
5. Introduce `internal/adapter/auth/module.go` as the bridge from the auth usecase to legacy repository, JWT, cache, and transaction implementations.

## Consequences

1. `internal/service/auth` is now isolated from concrete JWT claims types and `gorm.DB`.
2. `auth` no longer directly constructs or consumes `internal/models.User` or `internal/models.UserRole`.
3. The usecase now depends on auth-specific contracts rather than the old shared infrastructure interfaces.
4. The new adapter package establishes the pattern for wrapping non-repository infrastructure as module-local adapters.

## Guardrails

- Do not reintroduce `pkg/jwt` imports into `internal/service/auth`.
- Do not reintroduce `gorm.DB` into auth usecase method signatures or transaction callbacks.
- Do not let `types/user` DTOs cross into `internal/service/auth`.
- Future auth behavior changes should prefer extending auth ports and adapters instead of leaking implementation types inward.
