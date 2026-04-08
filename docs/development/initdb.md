# InitDB Workflow

`initdb` mode currently performs these steps:

1. Load config and initialize the database connection.
2. Build DDL for scaffold tables and the RBAC policy table.
3. Write the generated SQL script into `scripts/initdb/`.
4. Execute the SQL when not in `--dry-run`.
5. Write `.initdb.lock` to prevent repeated initialization.

Useful commands:

```bash
go run ./cmd/server initdb --dry-run
go run ./cmd/server initdb
```

The default generated file name is `initdb.<driver>.sql`, for example `initdb.sqlite.sql`.
