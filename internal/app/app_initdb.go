package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rin721/rei/internal/config"
	pkgsqlgen "github.com/rin721/rei/pkg/sqlgen"
	"gorm.io/gorm"
)

func (a *App) runInitDB(ctx context.Context) error {
	if !a.cfg.InitDB.Enabled {
		return nil
	}
	if a.database == nil {
		return fmt.Errorf("initdb requires an initialized database connection")
	}

	if err := os.MkdirAll(a.cfg.InitDB.OutputDir, 0o755); err != nil {
		return fmt.Errorf("create initdb output dir: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(a.cfg.InitDB.LockFile), 0o755); err != nil {
		return fmt.Errorf("create initdb lock dir: %w", err)
	}

	if !a.options.DryRun {
		if _, err := os.Stat(a.cfg.InitDB.LockFile); err == nil {
			if a.logger != nil {
				a.logger.Info("initdb lock file detected, skip repeated initialization")
			}
			return nil
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("stat initdb lock file: %w", err)
		}
	}

	driver := initDBDriver(a.cfg)
	tables, err := buildInitDBTables(a.cfg)
	if err != nil {
		return fmt.Errorf("build initdb tables: %w", err)
	}

	statements, err := pkgsqlgen.GenerateStatements(tables, pkgsqlgen.GenerateOptions{
		IfNotExists:     true,
		IdentifierQuote: identifierQuoteFor(driver),
	})
	if err != nil {
		return fmt.Errorf("generate initdb statements: %w", err)
	}

	scriptPath := filepath.Join(a.cfg.InitDB.OutputDir, fmt.Sprintf("initdb.%s.sql", driver))
	scriptContent := pkgsqlgen.RenderScript(statements)
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0o644); err != nil {
		return fmt.Errorf("write initdb script: %w", err)
	}

	if a.logger != nil {
		a.logger.Info(fmt.Sprintf("initdb SQL script generated at %s", scriptPath))
	}

	if a.options.DryRun {
		if a.logger != nil {
			a.logger.Info("initdb dry-run completed without executing SQL")
		}
		return nil
	}

	if err := a.executeInitDBStatements(ctx, statements); err != nil {
		return fmt.Errorf("execute initdb statements: %w", err)
	}

	lockContent := renderInitDBLock(driver, scriptPath, tables)
	if err := os.WriteFile(a.cfg.InitDB.LockFile, []byte(lockContent), 0o644); err != nil {
		return fmt.Errorf("write initdb lock file: %w", err)
	}

	if a.logger != nil {
		a.logger.Info(fmt.Sprintf("initdb completed and lock file written to %s", a.cfg.InitDB.LockFile))
	}

	return nil
}

func identifierQuoteFor(driver string) string {
	switch strings.ToLower(strings.TrimSpace(driver)) {
	case "postgres", "postgresql":
		return `"`
	default:
		return "`"
	}
}

func (a *App) executeInitDBStatements(ctx context.Context, statements []string) error {
	return a.database.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, statement := range statements {
			if err := tx.Exec(statement).Error; err != nil {
				return fmt.Errorf("execute statement %q: %w", statement, err)
			}
		}
		return nil
	})
}

func initDBDriver(cfg config.Config) string {
	driver := strings.TrimSpace(cfg.InitDB.Driver)
	if driver != "" {
		return strings.ToLower(driver)
	}
	return strings.ToLower(strings.TrimSpace(cfg.Database.Driver))
}

func renderInitDBLock(driver, scriptPath string, tables []pkgsqlgen.Table) string {
	tableNames := make([]string, 0, len(tables))
	for _, table := range tables {
		tableNames = append(tableNames, table.Name)
	}

	return fmt.Sprintf(
		"initialized_at: %s\ndriver: %s\nscript_path: %s\ntables: %s\n",
		time.Now().UTC().Format(time.RFC3339),
		driver,
		scriptPath,
		strings.Join(tableNames, ","),
	)
}
