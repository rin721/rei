# Development

Recommended local workflow:

1. Run `go test ./...`.
2. Use `go run ./cmd/server run --dry-run` to verify container wiring.
3. Use `go run ./cmd/server initdb --dry-run` to generate SQL without applying it.
4. Use `go run ./cmd/server` for the full local runtime.

The default example config is SQLite-based so local verification does not require MySQL, Postgres, or Redis.
