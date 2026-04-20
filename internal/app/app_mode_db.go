package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/rin721/rei/internal/config"
	"github.com/rin721/rei/internal/models"
	pkgmigrate "github.com/rin721/rei/pkg/migrate"
)

type DBAction string

const (
	DBActionGenerate DBAction = "generate"
	DBActionMigrate  DBAction = "migrate"
	DBActionStatus   DBAction = "status"
	DBActionRollback DBAction = "rollback"
)

type DBOptions struct {
	Action      DBAction
	Version     string
	Description string
	WithCRUD    bool
	Steps       int
	DryRun      bool
}

func (a *App) runModeDB(ctx context.Context) error {
	dbOpts := a.options.DBOptions
	if dbOpts.Action == "" {
		return fmt.Errorf("db mode requires an action (generate/migrate/status/rollback)")
	}

	switch dbOpts.Action {
	case DBActionGenerate:
		return a.runDBGenerate(ctx, dbOpts)
	case DBActionMigrate:
		return a.runDBMigrate(ctx, dbOpts)
	case DBActionStatus:
		return a.runDBStatus(ctx, dbOpts)
	case DBActionRollback:
		return a.runDBRollback(ctx, dbOpts)
	default:
		return fmt.Errorf("unsupported db action %q", dbOpts.Action)
	}
}

func (a *App) runDBGenerate(_ context.Context, opts DBOptions) error {
	migrationsDir := databaseMigrationsDir(a.cfg)

	desc := strings.TrimSpace(opts.Description)
	if desc == "" {
		desc = "migration"
	}

	driver := databaseDialect(a.cfg)

	if a.infra.logger != nil {
		a.infra.logger.Info(fmt.Sprintf("db generate: dialect=%s dir=%s", driver, migrationsDir))
	}

	genOpts := pkgmigrate.GenerateOptions{
		Models:      models.All(),
		OutputDir:   migrationsDir,
		Version:     opts.Version,
		Description: desc,
		Dialect:     driver,
		WithCRUD:    opts.WithCRUD,
	}

	if opts.DryRun {
		fmt.Printf("[dry-run] db generate: would write scripts to %s\n", migrationsDir)
		fmt.Printf("  dialect: %s\n  version: %s\n  desc:    %s\n", driver, opts.Version, desc)
		return nil
	}

	if err := pkgmigrate.Generate(genOpts); err != nil {
		return fmt.Errorf("db generate failed: %w", err)
	}

	if a.infra.logger != nil {
		a.infra.logger.Info(fmt.Sprintf("db generate: scripts written to %s", migrationsDir))
	}
	return nil
}

func (a *App) runDBMigrate(ctx context.Context, opts DBOptions) error {
	if err := a.bootstrapDB(ctx); err != nil {
		return err
	}
	defer func() { _ = a.Shutdown(context.TODO()) }()

	migrator := pkgmigrate.New(a.infra.database.DB(), databaseDialect(a.cfg), databaseMigrationsDir(a.cfg))

	executed, err := migrator.Migrate(ctx, opts.DryRun)
	if err != nil {
		return fmt.Errorf("db migrate failed: %w", err)
	}

	if len(executed) == 0 {
		fmt.Println("No pending migrations.")
		return nil
	}

	prefix := ""
	if opts.DryRun {
		prefix = "[dry-run] "
	}
	for _, version := range executed {
		fmt.Printf("%sApplied migration: %s\n", prefix, version)
	}
	return nil
}

func (a *App) runDBStatus(ctx context.Context, _ DBOptions) error {
	if err := a.bootstrapDB(ctx); err != nil {
		return err
	}
	defer func() { _ = a.Shutdown(context.TODO()) }()

	migrator := pkgmigrate.New(a.infra.database.DB(), databaseDialect(a.cfg), databaseMigrationsDir(a.cfg))

	status, err := migrator.Status(ctx)
	if err != nil {
		return fmt.Errorf("db status failed: %w", err)
	}

	fmt.Printf("%-20s  %-20s  %-26s  %s\n", "VERSION", "DESCRIPTION", "APPLIED_AT", "STATUS")
	fmt.Println(strings.Repeat("-", 82))

	for _, rec := range status.Applied {
		fmt.Printf("%-20s  %-20s  %-26s  applied\n",
			rec.Version, rec.Description, rec.AppliedAt.Format("2006-01-02T15:04:05Z"))
	}
	for _, mig := range status.Pending {
		fmt.Printf("%-20s  %-20s  %-26s  pending\n",
			mig.Version, mig.Description, "(pending)")
	}

	total := len(status.Applied) + len(status.Pending)
	fmt.Printf("\n  Applied: %d  Pending: %d  Total: %d\n",
		len(status.Applied), len(status.Pending), total)
	return nil
}

func (a *App) runDBRollback(ctx context.Context, opts DBOptions) error {
	if err := a.bootstrapDB(ctx); err != nil {
		return err
	}
	defer func() { _ = a.Shutdown(context.TODO()) }()

	steps := opts.Steps
	if steps < 1 {
		steps = 1
	}

	migrator := pkgmigrate.New(a.infra.database.DB(), databaseDialect(a.cfg), databaseMigrationsDir(a.cfg))

	rolledBack, err := migrator.Rollback(ctx, steps, opts.DryRun)
	if err != nil {
		return fmt.Errorf("db rollback failed: %w", err)
	}

	if len(rolledBack) == 0 {
		fmt.Println("Nothing to rollback.")
		return nil
	}

	prefix := ""
	if opts.DryRun {
		prefix = "[dry-run] "
	}
	for _, version := range rolledBack {
		fmt.Printf("%sRolled back migration: %s\n", prefix, version)
	}
	return nil
}

func (a *App) bootstrapDB(ctx context.Context) error {
	if err := a.bootstrapDBInfrastructure(ctx); err != nil {
		return err
	}
	if a.infra.database == nil {
		return fmt.Errorf("db mode requires database.enabled = true")
	}
	return nil
}

func databaseMigrationsDir(cfg config.Config) string {
	dir := strings.TrimSpace(cfg.Database.MigrationsDir)
	if dir == "" {
		return "scripts/migrations"
	}
	return dir
}

func databaseDialect(cfg config.Config) string {
	driver := strings.TrimSpace(cfg.Database.Driver)
	if driver == "" {
		return "sqlite"
	}
	return driver
}
