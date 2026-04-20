# pkg/cli

`pkg/cli` wraps Cobra behind repository-owned abstractions so command code only depends on `cli.Command`, `cli.FlagSet`, and related helpers.

## Current command shape

- `cmd/server` registers the runtime command
- `cmd/db` registers `generate`, `migrate`, `status`, and `rollback`
- `BuildRootCmd` assembles the root command and global flags
- business command handlers receive `context.Context` and `cli.FlagSet` instead of Cobra types

## Global flags

- `--config`
- `--dry-run`
- `--yes` / `-y`

## Example

```go
type Cmd struct{}

func (c *Cmd) Command() *cli.Command {
	return &cli.Command{
		Use:   "example",
		Short: "example command",
		RunE: func(ctx context.Context, f cli.FlagSet, _ []string) error {
			_ = f.GetString(flags.FlagConfig)
			_ = f.GetBool(flags.FlagDryRun)
			return nil
		},
	}
}
```
