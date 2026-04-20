# Development

Recommended local workflow:

1. Run `go test ./...`.
2. Use `go run ./cmd run --dry-run` to verify container wiring.
3. Use `go run ./cmd db migrate --dry-run` to inspect pending schema changes.
4. Use `go run ./cmd run` for the full local runtime.

The default example config is SQLite-based so local verification does not require MySQL, Postgres, or Redis.
