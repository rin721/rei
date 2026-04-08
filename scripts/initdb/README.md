# InitDB Scripts

Generated SQL scripts for the scaffold `initdb` workflow are written here.

- `initdb.sqlite.sql` is generated when the default local config is used.
- `.initdb.lock` is created after a successful non-dry-run initialization.
- `--dry-run` generates the SQL script but skips database execution and lock creation.
