package sqlgen

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var createTablePattern = regexp.MustCompile(`(?is)^CREATE\s+TABLE(?:\s+IF\s+NOT\s+EXISTS)?\s+[` + "`" + `"]?([a-zA-Z0-9_]+)[` + "`" + `"]?\s*\((.*)\)$`)

// Column 描述一列的基本信息。
type Column struct {
	Name          string
	Type          string
	Nullable      bool
	PrimaryKey    bool
	Unique        bool
	Default       string
	AutoIncrement bool
}

// UniqueConstraint 描述多列唯一约束。
type UniqueConstraint struct {
	Columns []string
}

// Table 描述表结构定义。
type Table struct {
	Name              string
	Columns           []Column
	UniqueConstraints []UniqueConstraint
}

// GenerateOptions 描述 SQL 生成选项。
type GenerateOptions struct {
	IfNotExists     bool
	IdentifierQuote string
}

// GenerateCreateTable 根据表结构生成 CREATE TABLE 语句。
func GenerateCreateTable(table Table) (string, error) {
	return GenerateCreateTableWithOptions(table, GenerateOptions{})
}

// GenerateCreateTableWithOptions 根据表结构和选项生成 CREATE TABLE 语句。
func GenerateCreateTableWithOptions(table Table, options GenerateOptions) (string, error) {
	if table.Name == "" {
		return "", errors.New("table name is required")
	}
	if len(table.Columns) == 0 {
		return "", errors.New("table columns are required")
	}

	lines := make([]string, 0, len(table.Columns)+len(table.UniqueConstraints))
	quote := options.IdentifierQuote
	if quote == "" {
		quote = "`"
	}
	for _, column := range table.Columns {
		if column.Name == "" || column.Type == "" {
			return "", errors.New("column name and type are required")
		}

		parts := []string{quoteIdentifier(column.Name, quote), strings.ToUpper(column.Type)}
		if column.AutoIncrement {
			parts = append(parts, "AUTOINCREMENT")
		}
		if column.PrimaryKey {
			parts = append(parts, "PRIMARY KEY")
		}
		if column.Unique {
			parts = append(parts, "UNIQUE")
		}
		if !column.Nullable {
			parts = append(parts, "NOT NULL")
		}
		if column.Default != "" {
			parts = append(parts, "DEFAULT", column.Default)
		}
		lines = append(lines, "  "+strings.Join(parts, " "))
	}

	for _, constraint := range table.UniqueConstraints {
		if len(constraint.Columns) == 0 {
			return "", errors.New("unique constraint columns are required")
		}
		columns := make([]string, 0, len(constraint.Columns))
		for _, column := range constraint.Columns {
			if strings.TrimSpace(column) == "" {
				return "", errors.New("unique constraint column name is required")
			}
			columns = append(columns, quoteIdentifier(column, quote))
		}
		lines = append(lines, "  UNIQUE ("+strings.Join(columns, ", ")+")")
	}

	prefix := "CREATE TABLE"
	if options.IfNotExists {
		prefix += " IF NOT EXISTS"
	}

	return fmt.Sprintf("%s %s (\n%s\n);", prefix, quoteIdentifier(table.Name, quote), strings.Join(lines, ",\n")), nil
}

// GenerateStatements 批量生成 CREATE TABLE 语句。
func GenerateStatements(tables []Table, options GenerateOptions) ([]string, error) {
	if len(tables) == 0 {
		return nil, errors.New("tables are required")
	}

	statements := make([]string, 0, len(tables))
	for _, table := range tables {
		statement, err := GenerateCreateTableWithOptions(table, options)
		if err != nil {
			return nil, fmt.Errorf("generate create table for %q: %w", table.Name, err)
		}
		statements = append(statements, statement)
	}

	return statements, nil
}

// RenderScript 将多条 SQL 语句渲染为脚本内容。
func RenderScript(statements []string) string {
	if len(statements) == 0 {
		return ""
	}

	return strings.Join(statements, "\n\n") + "\n"
}

// ParseCreateTable 将简化 CREATE TABLE 语句解析为表结构。
func ParseCreateTable(statement string) (Table, error) {
	trimmed := strings.TrimSpace(statement)
	trimmed = strings.TrimSuffix(trimmed, ";")

	matches := createTablePattern.FindStringSubmatch(trimmed)
	if len(matches) != 3 {
		return Table{}, errors.New("unsupported create table statement")
	}

	table := Table{Name: matches[1]}
	for _, rawColumn := range splitDefinitions(matches[2]) {
		rawColumn = strings.TrimSpace(rawColumn)
		if rawColumn == "" {
			continue
		}

		upper := strings.ToUpper(rawColumn)
		if strings.HasPrefix(upper, "UNIQUE ") || strings.HasPrefix(upper, "PRIMARY KEY ") {
			continue
		}

		fields := strings.Fields(rawColumn)
		if len(fields) < 2 {
			return Table{}, fmt.Errorf("invalid column definition %q", rawColumn)
		}

		column := Column{
			Name:     strings.Trim(fields[0], "`\""),
			Type:     strings.ToLower(fields[1]),
			Nullable: !strings.Contains(upper, "NOT NULL"),
			Unique:   strings.Contains(upper, "UNIQUE"),
		}
		if strings.Contains(upper, "PRIMARY KEY") {
			column.PrimaryKey = true
		}

		table.Columns = append(table.Columns, column)
	}

	if len(table.Columns) == 0 {
		return Table{}, errors.New("no columns parsed")
	}

	return table, nil
}

func splitDefinitions(content string) []string {
	var (
		items    []string
		builder  strings.Builder
		depth    int
		inQuotes bool
	)

	for _, r := range content {
		switch r {
		case '\'':
			inQuotes = !inQuotes
		case '(':
			if !inQuotes {
				depth++
			}
		case ')':
			if !inQuotes && depth > 0 {
				depth--
			}
		case ',':
			if !inQuotes && depth == 0 {
				items = append(items, builder.String())
				builder.Reset()
				continue
			}
		}

		builder.WriteRune(r)
	}

	if builder.Len() > 0 {
		items = append(items, builder.String())
	}

	return items
}

func quoteIdentifier(name, quote string) string {
	return quote + name + quote
}
