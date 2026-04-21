# Current Task

Status: Completed

Completed On: 2026-04-21

Completed Task: Introduce declarative lifecycle composition for business and delivery slices so mode coordinators delegate slice-local sequencing instead of relying on handwritten step lists.

Completion Notes:
- Business runtime bootstrap is now resolved through a slice-owned lifecycle registry instead of a hard-coded step array.
- Delivery runtime lifecycle is now resolved declaratively for:
  - bootstrap
  - start
  - reload
  - shutdown
- Delivery capabilities now declare runtime concerns as named components:
  - `router`
  - `http-server`
- Business capabilities now declare runtime concerns as named components:
  - `business-modules`
- Dependency-aware lifecycle ordering is now reused across:
  - infrastructure capabilities
  - business slice capabilities
  - delivery slice capabilities
- Unit coverage was added for:
  - delivery bootstrap dependency ordering
  - start-phase hook selection
- Validation passed with:
  - `go test ./internal/app ./internal/router`
  - `go test ./...`
  - `go build ./...`

Next Recommended Task:
- Move mode execution toward declarative mode plans so runtime coordinators select named slice profiles instead of calling slice lifecycle methods directly.
