# Architecture

Current runtime layering:

`cmd/server -> internal/app -> internal/router|middleware|handler -> internal/service -> internal/repository -> internal/models`

Key boundaries:

- `internal/app` owns lifecycle, wiring, and mode switching.
- `internal/router` only mounts routes and middleware order.
- `internal/handler` binds input and writes the unified response envelope.
- `internal/service` owns business rules and transaction boundaries.
- `internal/repository` owns persistence details.
- `pkg/*` remains reusable and does not depend on `internal/*`.
