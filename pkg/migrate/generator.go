package migrate

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	pkgsqlgen "github.com/rin721/rei/pkg/sqlgen"
)

// Generate reflects Go models into versioned migration scripts.
func Generate(opts GenerateOptions) error {
	if len(opts.Models) == 0 {
		return fmt.Errorf("generate: models is required")
	}
	if opts.OutputDir == "" {
		opts.OutputDir = "scripts/migrations"
	}
	if opts.Description == "" {
		opts.Description = "migration"
	}

	if err := os.MkdirAll(opts.OutputDir, 0o755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}

	version := strings.TrimSpace(opts.Version)
	if version == "" {
		date := time.Now().Format("20060102")
		var err error
		version, err = nextVersion(opts.OutputDir, date)
		if err != nil {
			return fmt.Errorf("compute next version: %w", err)
		}
	}

	dialect := dialectFromString(opts.Dialect)
	gen := pkgsqlgen.New(&pkgsqlgen.Config{Dialect: dialect})

	upStatements := make([]string, 0, len(opts.Models))
	downStatements := make([]string, 0, len(opts.Models))

	for _, model := range opts.Models {
		upSQL, err := gen.Table(model)
		if err != nil {
			return fmt.Errorf("generate CREATE TABLE for %T: %w", model, err)
		}
		upSQL = strings.Replace(upSQL, "CREATE TABLE ", "CREATE TABLE IF NOT EXISTS ", 1)

		if strings.ToLower(strings.TrimSpace(opts.Dialect)) == "sqlite" || opts.Dialect == "" {
			tableSQL, indexStatements := normalizeSQLiteTableSQL(upSQL, tableNameOf(model))
			upStatements = append(upStatements, tableSQL)
			upStatements = append(upStatements, indexStatements...)
		} else {
			upStatements = append(upStatements, upSQL)
		}

		downSQL, err := gen.Drop(model)
		if err != nil {
			return fmt.Errorf("generate DROP TABLE for %T: %w", model, err)
		}
		downStatements = append([]string{downSQL}, downStatements...)
	}

	baseName := fmt.Sprintf("%s_%s", version, opts.Description)
	upPath := filepath.Join(opts.OutputDir, baseName+".up.sql")
	downPath := filepath.Join(opts.OutputDir, baseName+".down.sql")

	upContent := renderHeader(version, opts.Description, "up") + pkgsqlgen.RenderScript(upStatements)
	if err := os.WriteFile(upPath, []byte(upContent), 0o644); err != nil {
		return fmt.Errorf("write up script: %w", err)
	}

	downContent := renderHeader(version, opts.Description, "down") + pkgsqlgen.RenderScript(downStatements)
	if err := os.WriteFile(downPath, []byte(downContent), 0o644); err != nil {
		return fmt.Errorf("write down script: %w", err)
	}

	if opts.WithCRUD {
		if err := generateCRUD(gen, opts.Models, opts.OutputDir); err != nil {
			return fmt.Errorf("generate crud docs: %w", err)
		}
	}

	return nil
}

func renderHeader(version, desc, direction string) string {
	return fmt.Sprintf(
		"-- Migration: %s_%s (%s)\n-- Generated: %s\n\n",
		version, desc, direction,
		time.Now().UTC().Format(time.RFC3339),
	)
}

func generateCRUD(gen *pkgsqlgen.Generator, models []any, migrationsDir string) error {
	crudDir := filepath.Join(filepath.Dir(migrationsDir), "crud")
	if err := os.MkdirAll(crudDir, 0o755); err != nil {
		return fmt.Errorf("create crud dir: %w", err)
	}

	for _, model := range models {
		insertSQL, err := gen.Create(model)
		if err != nil {
			insertSQL = "-- INSERT not generated: " + err.Error()
		}

		selectSQL, err := gen.Find(model)
		if err != nil {
			selectSQL = "-- SELECT not generated: " + err.Error()
		}

		content := fmt.Sprintf(
			"-- CRUD reference for %T\n\n-- INSERT\n%s\n\n-- SELECT\n%s\n",
			model, insertSQL, selectSQL,
		)

		outPath := filepath.Join(crudDir, tableNameOf(model)+".sql")
		if err := os.WriteFile(outPath, []byte(content), 0o644); err != nil {
			return fmt.Errorf("write crud for %T: %w", model, err)
		}
	}

	return nil
}

func tableNameOf(model any) string {
	type tabler interface{ TableName() string }
	if t, ok := model.(tabler); ok {
		return t.TableName()
	}
	return "unknown"
}

func dialectFromString(d string) pkgsqlgen.Dialect {
	switch strings.ToLower(strings.TrimSpace(d)) {
	case "postgres", "postgresql":
		return pkgsqlgen.PostgreSQL
	case "mysql":
		return pkgsqlgen.MySQL
	case "sqlserver", "mssql":
		return pkgsqlgen.SQLServer
	default:
		return pkgsqlgen.SQLite
	}
}

func normalizeSQLiteTableSQL(sql, table string) (string, []string) {
	lines := strings.Split(sql, "\n")
	tableLines := make([]string, 0, len(lines))
	indexStatements := []string{}

	for _, line := range lines {
		trimmed := strings.TrimSpace(strings.TrimSuffix(line, ","))
		upper := strings.ToUpper(trimmed)

		switch {
		case strings.HasPrefix(upper, "UNIQUE INDEX "):
			if stmt := sqliteIndexStatement(trimmed, table, true); stmt != "" {
				indexStatements = append(indexStatements, stmt)
			}
		case strings.HasPrefix(upper, "INDEX "):
			if stmt := sqliteIndexStatement(trimmed, table, false); stmt != "" {
				indexStatements = append(indexStatements, stmt)
			}
		default:
			tableLines = append(tableLines, line)
		}
	}

	return fixTrailingComma(strings.Join(tableLines, "\n")), indexStatements
}

func sqliteIndexStatement(line, table string, unique bool) string {
	prefix := "INDEX"
	createPrefix := "CREATE INDEX IF NOT EXISTS"
	if unique {
		prefix = "UNIQUE INDEX"
		createPrefix = "CREATE UNIQUE INDEX IF NOT EXISTS"
	}

	rest := strings.TrimSpace(strings.TrimPrefix(line, prefix))
	openIdx := strings.Index(rest, "(")
	if openIdx <= 0 {
		return ""
	}

	name := strings.TrimSpace(rest[:openIdx])
	columns := strings.TrimSpace(rest[openIdx:])
	if name == "" || !strings.HasPrefix(columns, "(") {
		return ""
	}

	return fmt.Sprintf(`%s %s ON "%s" %s`, createPrefix, name, table, columns)
}

func fixTrailingComma(sql string) string {
	lines := strings.Split(sql, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		t := strings.TrimSpace(lines[i])
		if t == "" || strings.HasPrefix(t, "--") {
			continue
		}
		if t == ");" || t == ")" {
			for j := i - 1; j >= 0; j-- {
				tj := strings.TrimSpace(lines[j])
				if tj == "" {
					continue
				}
				if strings.HasSuffix(tj, ",") {
					lines[j] = strings.TrimRight(lines[j], " \t")
					lines[j] = lines[j][:len(lines[j])-1]
				}
				break
			}
			break
		}
	}
	return strings.Join(lines, "\n")
}
