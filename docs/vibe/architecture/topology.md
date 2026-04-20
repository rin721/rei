# Topology

Updated On: 2026-04-20

## Current Topology

```mermaid
flowchart TD
    A["configs/*.yaml<br/>database.driver + database.migrations_dir"] --> B["cmd/db"]
    B --> C["internal/app (ModeDB)"]
    C --> D["pkg/migrate"]
    D --> E["scripts/migrations/*.sql"]
    D --> F["schema_migrations"]

    G["cmd/run"] --> H["internal/app (ModeServer)"]
    H --> I["router -> handler -> service -> repository"]
    H --> J["seedBusiness<br/>(idempotent seed only)"]
    H -. forbidden .-> K["AutoMigrate"]

    L["internal/app tests"] --> D
    M["internal/router tests"] --> D
    N["pkg/dbtx tests"] --> O["explicit test DDL only"]
```

## Schema Source

- Single source of truth:
  `configs/*.yaml(database.migrations_dir) -> cmd/db -> internal/app(ModeDB) -> pkg/migrate -> scripts/migrations/*.sql`
- Server runtime responsibility:
  `cmd/run -> internal/app(ModeServer)` assembles modules, starts the service, and runs idempotent seed logic only.
- Test schema responsibility:
  tests that need the real business schema must use `pkg/migrate + scripts/migrations`; low-level tests that do not depend on business schema may use explicit test DDL.

## App Composition Root

```mermaid
flowchart LR
    A["app_mode_server.go"] --> B["app_bootstrap.go<br/>server infrastructure"]
    A --> C["app_bootstrap.go<br/>business runtime"]
    A --> D["app_bootstrap.go<br/>delivery runtime"]
    C --> E["app_business.go"]
    E --> F["repository.NewSet"]
    E --> G["app_seed.go<br/>seed registry"]
    G --> H["app_seed_rbac.go"]
    G --> I["app_seed_sample.go"]
    E --> J["app_modules.go"]
    J --> K["app_module_auth.go"]
    J --> L["app_module_user.go"]
    J --> M["app_module_rbac.go"]
    J --> N["app_module_sample.go"]
    J --> O["handler.Bundle"]
```

- `internal/app` now owns orchestration and registration, not module internals.
- Startup order is expressed as named bootstrap phases instead of one long mode-specific init sequence.
- `app_seed.go` is now a registry boundary; module-specific seed behavior lives in dedicated seeder files.

## App Lifecycle Boundary

```mermaid
flowchart LR
    A["app.go<br/>app state"] --> B["reload.go<br/>runtime reloader registry"]
    A --> C["app_shutdown.go<br/>named shutdown steps"]
    A --> D["app_runtime_bindings.go<br/>executor binding sync"]
    B --> E["logger/cache/database/executor/http server/storage"]
    C --> F["config manager -> http server -> storage -> executor -> cache -> database -> rbac -> logger"]
    D --> G["logger"]
    D --> H["http server"]
```

- `app.go` now carries state and construction only.
- Reload and shutdown are no longer embedded as long inline sequences in the main app container file.
- Cross-resource binding is now localized instead of being repeated in multiple init and reload paths.

## User Vertical Slice

```mermaid
flowchart LR
    A["HTTP DTO<br/>types/user"] --> B["handler/user_handler.go"]
    B --> C["service/user<br/>command/query/result"]
    C --> D["domain/user.User"]
    C --> E["user ports"]
    E --> F["repository/user_domain_adapter.go"]
    F --> G["internal/models.User + GORM repo"]
```

- Status: Completed

## Auth Vertical Slice

```mermaid
flowchart LR
    A["HTTP DTO<br/>types/user"] --> B["handler/auth_handler.go"]
    B --> C["service/auth<br/>command/result/ports"]
    C --> D["domain/user.User"]
    C --> E["auth ports"]
    E --> F["adapter/auth/module.go"]
    F --> G["repository + jwt + cache + tx"]
```

- Status: Completed

## RBAC Vertical Slice

```mermaid
flowchart LR
    A["HTTP DTO<br/>types/*.go"] --> B["handler/rbac_handler.go"]
    B --> C["service/rbac<br/>command/result/ports"]
    C --> D["domain/rbac.*"]
    C --> E["rbac ports"]
    E --> F["adapter/rbac/module.go"]
    F --> G["repository + runtime rbac manager + tx"]
```

- `handler/rbac_handler.go` owns DTO translation for permission checks, role assignment, role revocation, and policy management.
- `service/rbac` depends on RBAC-specific ports and pure RBAC domain entities.
- `internal/adapter/rbac/module.go` bridges the RBAC usecase to legacy repositories and the existing runtime RBAC manager.
- Status: Completed

## Refactor Strategy

```mermaid
flowchart LR
    A["Stable External API"] --> B["Handler DTO Boundary"]
    B --> C["Usecase Command / Query / Result"]
    C --> D["Domain Entity"]
    C --> E["Ports"]
    E --> F["Repository Adapter"]
    F --> G["Legacy GORM Model / Existing Repo"]
```

- This is now the standard migration order for future modules.
- Completed slices: `user`, `auth`, `rbac`
- Completed app-layer composition step: module-level providers under `internal/app`
- Completed app-layer startup step: explicit bootstrap phases and module-owned seeders
- Completed app-layer lifecycle step: explicit reload/shutdown boundaries and shared runtime bindings
- Next recommended focus: group long-lived app state into narrower runtime containers for infrastructure, business, and delivery concerns.

## Removed Nodes

- `cmd/initdb`
- `internal/app/app_initdb.go`
- `internal/app/app_initdb_schema.go`
- `internal/app/app_mode_initdb.go`
- `scripts/initdb/`
- runtime `AutoMigrate`
- shared service-layer infrastructure contracts that are no longer used by migrated modules

## Entity Notes

- No new infrastructure library was introduced in this round.
- `auth` continues to reuse `internal/domain/user.User`.
- `rbac` now has explicit pure domain entities under `internal/domain/rbac`.
- Domain entity memory continues to live under `architecture/entities/`.
