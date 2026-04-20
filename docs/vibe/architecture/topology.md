# Topology

Updated On: 2026-04-21

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
    D --> H["http server"]
```

- `app.go` now carries state and construction only.
- Reload and shutdown are no longer embedded as long inline sequences in the main app container file.
- Cross-resource binding is now localized instead of being repeated in multiple init and reload paths.
- Executor binding now targets the HTTP server only; logger no longer assumes the removed legacy async hook.

## Core Infrastructure Integration

```mermaid
flowchart LR
    A["internal/config/*.go"] --> B["internal/app/config_helpers.go"]
    B --> C["pkg/logger.Logger"]
    B --> D["pkg/i18n.I18n"]
    B --> E["pkg/executor.Manager"]
    B --> F["pkg/storage.Storage"]
    E --> G["app_runtime_bindings.go<br/>executorAsyncSubmitter"]
    G --> H["pkg/httpserver.AsyncSubmitter"]
    D --> I["internal/middleware/i18n.go"]
    C --> J["internal/middleware/logger.go<br/>internal/middleware/recovery.go"]
```

- Compatibility mapping for rewritten infrastructure packages is now isolated in the app composition root.
- Middleware depends on the new logger and i18n interfaces instead of removed concrete manager types.
- The HTTP server receives executor capability through an app-owned adapter rather than a direct package-level type match.

## App Runtime Containers

```mermaid
flowchart LR
    A["app.go"] --> B["a.infra<br/>infrastructureRuntime"]
    A --> C["a.business<br/>businessRuntime"]
    A --> D["a.delivery<br/>deliveryRuntime"]
    B --> E["logger/i18n/idGen/cache/database/dbtx/executor/crypto/jwt/storage/rbac"]
    C --> F["handler bundle"]
    D --> G["router engine"]
    D --> H["http server"]
```

- Runtime state is no longer stored as one flat field list on `App`.
- Infrastructure, business, and delivery concerns now have explicit ownership boundaries inside the app shell.
- Root-app orchestration still exists, but the state it coordinates is now grouped by runtime role.

## App Provisioning Views

```mermaid
flowchart LR
    A["app.go"] --> B["businessProvisioning"]
    A --> C["deliveryProvisioning"]
    B --> D["repository set"]
    B --> E["module providers"]
    B --> F["business seeders"]
    C --> G["router setup"]
    C --> H["http server setup"]
```

- Module providers and seeders now receive a narrow business assembly surface instead of the full root app object.
- Delivery assembly now reads from a dedicated delivery view instead of directly traversing root-app runtime state.
- Root-app knowledge is now more localized at the composition boundary.

## Infrastructure Provisioning

```mermaid
flowchart LR
    A["app.go"] --> B["infrastructureProvisioning"]
    B --> C["server bootstrap steps"]
    B --> D["db bootstrap steps"]
    B --> E["runtime reloaders"]
    B --> F["shutdown steps"]
    B --> G["logger/i18n/idGen/cache/database/dbtx/executor/crypto/jwt/storage/rbac"]
```

- Root-app orchestration no longer manually chains the infrastructure helper list.
- Infrastructure lifecycle registration now lives behind an app-owned provisioning boundary.
- The infrastructure slice now owns both its long-lived state and the lifecycle sequencing around that state.

## Business And Delivery Bootstrap

```mermaid
flowchart LR
    A["app.go"] --> B["businessProvisioning.bootstrap"]
    A --> C["deliveryProvisioning.bootstrap"]
    B --> D["Validate"]
    B --> E["RepositorySet"]
    B --> F["seeders"]
    B --> G["module providers"]
    B --> H["handler bundle"]
    C --> I["router init"]
    C --> J["http server init"]
```

- Business and delivery slices now own their local bootstrap sequences.
- Root-app orchestration coordinates slices, but no longer spells out those slice-local assembly details.
- Slice bootstrap is now aligned with slice-owned runtime state.

## Sample Toolkit Demos

```mermaid
flowchart LR
    A["GET /api/v1/samples/tooling"] --> B["handler/sample_handler.go"]
    B --> C["service/sample.Tooling"]
    C --> D["ToolkitDemo[]"]
    D --> E["sqlgenToolkitDemo"]
    D --> F["yaml2goToolkitDemo"]
    E --> G["pkg/sqlgen"]
    F --> H["pkg/yaml2go"]
```

- The sample module now provides business-facing demos for useful pkg helpers that were previously isolated from `internal/`.
- Current demo-backed packages:
  - `pkg/sqlgen`
  - `pkg/yaml2go`
- Explicitly excluded from business demos:
  - `pkg/cli*` because it is CLI-only infrastructure
  - `pkg/i18nold` because it is deprecated legacy code

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
- Completed app-layer state step: runtime containers for infrastructure, business, and delivery concerns
- Completed app-layer provisioning step: narrow business/delivery views instead of broad root-app coupling
- Completed app-layer infrastructure step: infrastructure lifecycle orchestration delegated to `infrastructureProvisioning`
- Completed app-layer slice orchestration step: business and delivery bootstrap delegated to slice-owned provisioning flows
- Completed platform integration step: rewritten logger/i18n/executor/storage packages are now wired through the current composition root
- Completed isolated tooling step: sample-module demos now cover `pkg/sqlgen` and `pkg/yaml2go`
- Next recommended focus: extract slice-local registration for runtime start/reload/shutdown where those paths still cross back through the root app shell.

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
