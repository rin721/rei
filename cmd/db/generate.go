package db

import (
	"context"
	"fmt"
	"io"
	"os"

	internalapp "github.com/rin721/rei/internal/app"
	"github.com/rin721/rei/pkg/cli"
	"github.com/rin721/rei/pkg/cli/flags"
)

const (
	flagVersion  = "version"
	flagDesc     = "desc"
	flagWithCRUD = "with-crud"
)

// GenerateCmd implements `db generate`.
type GenerateCmd struct{}

func (c *GenerateCmd) Command() *cli.Command {
	return &cli.Command{
		Use:   "generate",
		Short: "Generate versioned migration scripts from Go models",
		Long: `db generate reflects the registered models into migration scripts:
  - {version}_{desc}.up.sql
  - {version}_{desc}.down.sql

Scripts are written to database.migrations_dir (default scripts/migrations).`,
		LocalFlags: []cli.FlagDef{
			{Name: flagVersion, Kind: "string", Usage: "override the generated migration version"},
			{Name: flagDesc, Kind: "string", Usage: "migration description used in the file name"},
			{Name: flagWithCRUD, Kind: "bool", Usage: "also generate CRUD SQL reference files in scripts/crud/"},
		},
		RunE: func(ctx context.Context, f cli.FlagSet, _ []string) error {
			return runDBGenerate(ctx, f, os.Stdout) //nolint:forbidigo
		},
	}
}

func runDBGenerate(ctx context.Context, f cli.FlagSet, out io.Writer) error {
	version := f.GetString(flagVersion)
	desc := f.GetString(flagDesc)
	withCRUD := f.GetBool(flagWithCRUD)
	dryRun := f.GetBool(flags.FlagDryRun)

	application, err := internalapp.New(internalapp.Options{
		Mode:       internalapp.ModeDB,
		ConfigPath: f.GetString(flags.FlagConfig),
		DryRun:     dryRun,
		DBOptions: internalapp.DBOptions{
			Action:      internalapp.DBActionGenerate,
			Version:     version,
			Description: desc,
			WithCRUD:    withCRUD,
			DryRun:      dryRun,
		},
	})
	if err != nil {
		return fmt.Errorf("create application: %w", err)
	}

	if err := application.Run(ctx); err != nil {
		return fmt.Errorf("db generate failed: %w", err)
	}

	if !dryRun {
		fmt.Fprintln(out, "Migration scripts generated.")
	}
	return nil
}
