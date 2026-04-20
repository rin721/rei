# Database Migration Workflow

The repository uses `cmd/db + pkg/migrate + scripts/migrations` as the only schema workflow.

Typical steps:

1. Generate a new versioned migration with `go run ./cmd db generate --desc <name>`.
2. Review the generated `.up.sql` and `.down.sql` files in `scripts/migrations/`.
3. Apply pending migrations with `go run ./cmd db migrate`.
4. Inspect applied and pending versions with `go run ./cmd db status`.
5. Roll back the latest migration with `go run ./cmd db rollback` when needed.

Useful commands:

```bash
go run ./cmd db generate --desc init_schema
go run ./cmd db migrate --dry-run
go run ./cmd db migrate
go run ./cmd db status
go run ./cmd db rollback --dry-run
```
