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

type dbModeRuntime struct {
	app     *App
	options DBOptions
}

func newDBModeRuntime(app *App, options DBOptions) dbModeRuntime {
	return dbModeRuntime{
		app:     app,
		options: options,
	}
}

func (r dbModeRuntime) run(ctx context.Context) error {
	dbOpts := r.options
	if dbOpts.Action == "" {
		return fmt.Errorf("db mode requires an action (generate/migrate/status/rollback)")
	}

	switch dbOpts.Action {
	case DBActionGenerate:
		return r.runGenerate(ctx)
	case DBActionMigrate:
		return r.runMigrate(ctx)
	case DBActionStatus:
		return r.runStatus(ctx)
	case DBActionRollback:
		return r.runRollback(ctx)
	default:
		return fmt.Errorf("unsupported db action %q", dbOpts.Action)
	}
}

func (r dbModeRuntime) runGenerate(_ context.Context) error {
	opts := r.options
	migrationsDir := r.migrationsDir()

	desc := strings.TrimSpace(opts.Description)
	if desc == "" {
		desc = "migration"
	}

	driver := r.dialect()

	if r.app.infra.logger != nil {
		r.app.infra.logger.Info(fmt.Sprintf("db generate: dialect=%s dir=%s", driver, migrationsDir))
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

	if r.app.infra.logger != nil {
		r.app.infra.logger.Info(fmt.Sprintf("db generate: scripts written to %s", migrationsDir))
	}
	return nil
}

func (r dbModeRuntime) runMigrate(ctx context.Context) error {
	opts := r.options
	if err := r.bootstrap(ctx); err != nil {
		return err
	}
	defer func() { _ = r.app.Shutdown(context.TODO()) }()

	migrator := pkgmigrate.New(r.app.infra.database.DB(), r.dialect(), r.migrationsDir())

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

func (r dbModeRuntime) runStatus(ctx context.Context) error {
	if err := r.bootstrap(ctx); err != nil {
		return err
	}
	defer func() { _ = r.app.Shutdown(context.TODO()) }()

	migrator := pkgmigrate.New(r.app.infra.database.DB(), r.dialect(), r.migrationsDir())

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

func (r dbModeRuntime) runRollback(ctx context.Context) error {
	opts := r.options
	if err := r.bootstrap(ctx); err != nil {
		return err
	}
	defer func() { _ = r.app.Shutdown(context.TODO()) }()

	steps := opts.Steps
	if steps < 1 {
		steps = 1
	}

	migrator := pkgmigrate.New(r.app.infra.database.DB(), r.dialect(), r.migrationsDir())

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

func (r dbModeRuntime) bootstrap(ctx context.Context) error {
	if err := r.app.infrastructureProvisioning().bootstrapDB(ctx); err != nil {
		return err
	}
	if r.app.infra.database == nil {
		return fmt.Errorf("db mode requires database.enabled = true")
	}
	return nil
}

func (r dbModeRuntime) migrationsDir() string {
	return databaseMigrationsDir(r.app.cfg)
}

func (r dbModeRuntime) dialect() string {
	return databaseDialect(r.app.cfg)
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
