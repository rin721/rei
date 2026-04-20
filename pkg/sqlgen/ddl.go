package sqlgen

import (
	"fmt"
	"strings"
)

// ============================================================================
// DDL 生成方法
// ============================================================================

// Table 生成 CREATE TABLE 语句
func (g *Generator) Table(model interface{}) (string, error) {
	if err := g.parseModel(model); err != nil {
		return "", err
	}

	fields := parseStructFields(g.ctx.ModelType, g.ctx.ModelValue, g.dialect)
	if len(fields) == 0 {
		return "", ErrInvalidModel
	}

	return g.buildCreateTable(g.ctx.TableName, fields), nil
}

// Drop 生成 DROP TABLE 语句
func (g *Generator) Drop(model interface{}) (string, error) {
	if err := g.parseModel(model); err != nil {
		return "", err
	}

	return g.buildDropTable(g.ctx.TableName), nil
}

// Migrate 返回迁移构建器
func (g *Generator) Migrate(model interface{}) *MigrateBuilder {
	ng := g.clone()
	_ = ng.parseModel(model)

	return &MigrateBuilder{
		generator: ng,
		tableName: ng.ctx.TableName,
		fields:    parseStructFields(ng.ctx.ModelType, ng.ctx.ModelValue, ng.dialect),
	}
}

// ============================================================================
// DDLFromSchema 从 Schema 生成 DDL
// ============================================================================

// DDLFromSchema 从 Schema 结构体生成 CREATE TABLE 语句
// ifNotExists 为 true 时生成 CREATE TABLE IF NOT EXISTS
func (g *Generator) DDLFromSchema(schema *Schema, ifNotExists bool) (string, error) {
	if schema == nil || schema.TableName == "" {
		return "", ErrInvalidModel
	}

	fields := schemaToFieldInfo(schema)

	var sb strings.Builder
	quotedTable := g.dialect.Quote(schema.TableName)

	if ifNotExists {
		sb.WriteString("CREATE TABLE IF NOT EXISTS ")
	} else {
		sb.WriteString("CREATE TABLE ")
	}
	sb.WriteString(quotedTable)
	sb.WriteString(" (\n")

	var columnDefs []string
	var primaryKeys []string

	for _, field := range fields {
		colDef := g.buildColumnDef(field)
		columnDefs = append(columnDefs, "  "+colDef)

		if field.Tag.PrimaryKey {
			primaryKeys = append(primaryKeys, g.dialect.Quote(field.ColumnName))
		}
	}

	sb.WriteString(strings.Join(columnDefs, ",\n"))

	// PRIMARY KEY 约束
	if len(primaryKeys) > 0 {
		sb.WriteString(",\n  PRIMARY KEY (")
		sb.WriteString(strings.Join(primaryKeys, ", "))
		sb.WriteString(")")
	}

	// UNIQUE 约束 (来自 Schema.Indexes)
	for _, idx := range schema.Indexes {
		if idx.Unique && len(idx.Columns) > 0 {
			quotedCols := make([]string, len(idx.Columns))
			for i, col := range idx.Columns {
				quotedCols[i] = g.dialect.Quote(col)
			}
			idxName := idx.Name
			if idxName == "" {
				idxName = "uk_" + schema.TableName + "_" + strings.Join(idx.Columns, "_")
			}
			sb.WriteString(fmt.Sprintf(",\n  CONSTRAINT %s UNIQUE (%s)",
				g.dialect.Quote(idxName),
				strings.Join(quotedCols, ", ")))
		}
	}

	sb.WriteString("\n)")

	// 引擎子句 (MySQL)
	if engine := g.dialect.EngineClause(); engine != "" {
		sb.WriteString(" ")
		sb.WriteString(engine)
	}

	sb.WriteString(";")

	return sb.String(), nil
}

// schemaToFieldInfo 将 Schema 转换为 FieldInfo 列表
func schemaToFieldInfo(schema *Schema) []FieldInfo {
	fields := make([]FieldInfo, 0, len(schema.Fields))
	for _, f := range schema.Fields {
		tag := &ParsedTag{
			Column:        f.Column.Name,
			PrimaryKey:    f.Column.PrimaryKey,
			AutoIncrement: f.Column.AutoIncrement,
			NotNull:       f.Column.NotNull,
			Default:       f.Column.Default,
			Comment:       f.Column.Comment,
			Size:          f.Column.Size,
		}
		fields = append(fields, FieldInfo{
			Name:       f.Name,
			ColumnName: f.Column.Name,
			SQLType:    f.Column.Type,
			Tag:        tag,
		})
	}
	return fields
}

// RenderScript 将多条 SQL 语句渲染为脚本内容
func RenderScript(statements []string) string {
	if len(statements) == 0 {
		return ""
	}
	return strings.Join(statements, "\n\n") + "\n"
}

// ============================================================================
// CREATE TABLE 构建
// ============================================================================

func (g *Generator) buildCreateTable(tableName string, fields []FieldInfo) string {
	var sb strings.Builder
	quotedTable := g.dialect.Quote(tableName)

	sb.WriteString("CREATE TABLE ")
	sb.WriteString(quotedTable)
	sb.WriteString(" (\n")

	var columnDefs []string
	var primaryKeys []string
	var indexes []string

	for _, field := range fields {
		colDef := g.buildColumnDef(field)
		columnDefs = append(columnDefs, "  "+colDef)

		if field.Tag.PrimaryKey {
			primaryKeys = append(primaryKeys, g.dialect.Quote(field.ColumnName))
		}

		if field.Tag.Index != "" {
			indexName := field.Tag.Index
			if indexName == "" || indexName == "true" {
				indexName = "idx_" + tableName + "_" + field.ColumnName
			}
			indexes = append(indexes, fmt.Sprintf("INDEX %s (%s)",
				g.dialect.Quote(indexName), g.dialect.Quote(field.ColumnName)))
		}

		if field.Tag.UniqueIndex != "" {
			indexName := field.Tag.UniqueIndex
			if indexName == "" || indexName == "true" {
				indexName = "uk_" + tableName + "_" + field.ColumnName
			}
			indexes = append(indexes, fmt.Sprintf("UNIQUE INDEX %s (%s)",
				g.dialect.Quote(indexName), g.dialect.Quote(field.ColumnName)))
		}
	}

	sb.WriteString(strings.Join(columnDefs, ",\n"))

	// 添加主键约束
	if len(primaryKeys) > 0 {
		sb.WriteString(",\n  PRIMARY KEY (")
		sb.WriteString(strings.Join(primaryKeys, ", "))
		sb.WriteString(")")
	}

	// 添加索引
	for _, idx := range indexes {
		sb.WriteString(",\n  ")
		sb.WriteString(idx)
	}

	sb.WriteString("\n)")

	// 添加引擎子句 (MySQL)
	if engine := g.dialect.EngineClause(); engine != "" {
		sb.WriteString(" ")
		sb.WriteString(engine)
	}

	sb.WriteString(";")

	return sb.String()
}

// buildColumnDef 构建列定义
func (g *Generator) buildColumnDef(field FieldInfo) string {
	var parts []string

	// 列名
	parts = append(parts, g.dialect.Quote(field.ColumnName))

	// 类型
	parts = append(parts, field.SQLType)

	// NOT NULL
	if field.Tag.NotNull || field.Tag.PrimaryKey {
		parts = append(parts, "NOT NULL")
	}

	// AUTO_INCREMENT
	if field.Tag.AutoIncrement {
		autoIncr := g.dialect.AutoIncrementKeyword()
		if autoIncr != "" {
			parts = append(parts, autoIncr)
		}
	}

	// DEFAULT
	if field.Tag.Default != "" {
		parts = append(parts, fmt.Sprintf("DEFAULT %s", field.Tag.Default))
	}

	// COMMENT (MySQL 特有)
	if field.Tag.Comment != "" && g.dialect.Name() == MySQL {
		parts = append(parts, fmt.Sprintf("COMMENT '%s'", escapeString(field.Tag.Comment)))
	}

	return strings.Join(parts, " ")
}

// ============================================================================
// DROP TABLE 构建
// ============================================================================

func (g *Generator) buildDropTable(tableName string) string {
	return fmt.Sprintf("DROP TABLE IF EXISTS %s;", g.dialect.Quote(tableName))
}

// ============================================================================
// MigrateBuilder 迁移构建器
// ============================================================================

// MigrateBuilder 迁移操作构建器
type MigrateBuilder struct {
	generator  *Generator
	tableName  string
	fields     []FieldInfo
	operations []string
}

// AddColumn 添加列
func (m *MigrateBuilder) AddColumn(columnName string) *MigrateBuilder {
	for _, field := range m.fields {
		if field.Name == columnName || field.ColumnName == columnName {
			colDef := m.generator.buildColumnDef(field)
			m.operations = append(m.operations, fmt.Sprintf("ADD COLUMN %s", colDef))
			break
		}
	}
	return m
}

// DropColumn 删除列
func (m *MigrateBuilder) DropColumn(columnName string) *MigrateBuilder {
	m.operations = append(m.operations, fmt.Sprintf("DROP COLUMN %s",
		m.generator.dialect.Quote(toSnakeCase(columnName))))
	return m
}

// ModifyColumn 修改列类型
func (m *MigrateBuilder) ModifyColumn(columnName, newType string) *MigrateBuilder {
	keyword := "MODIFY COLUMN"
	if m.generator.dialect.Name() == PostgreSQL {
		keyword = "ALTER COLUMN"
	}
	m.operations = append(m.operations, fmt.Sprintf("%s %s %s",
		keyword, m.generator.dialect.Quote(toSnakeCase(columnName)), newType))
	return m
}

// RenameColumn 重命名列
func (m *MigrateBuilder) RenameColumn(oldName, newName string) *MigrateBuilder {
	m.operations = append(m.operations, fmt.Sprintf("RENAME COLUMN %s TO %s",
		m.generator.dialect.Quote(toSnakeCase(oldName)),
		m.generator.dialect.Quote(toSnakeCase(newName))))
	return m
}

// AddIndex 添加索引
func (m *MigrateBuilder) AddIndex(indexName string, columns ...string) *MigrateBuilder {
	quotedCols := make([]string, len(columns))
	for i, col := range columns {
		quotedCols[i] = m.generator.dialect.Quote(toSnakeCase(col))
	}

	m.operations = append(m.operations, fmt.Sprintf("ADD INDEX %s (%s)",
		m.generator.dialect.Quote(indexName), strings.Join(quotedCols, ", ")))
	return m
}

// DropIndex 删除索引
func (m *MigrateBuilder) DropIndex(indexName string) *MigrateBuilder {
	m.operations = append(m.operations, fmt.Sprintf("DROP INDEX %s",
		m.generator.dialect.Quote(indexName)))
	return m
}

// Build 生成 ALTER TABLE 语句
func (m *MigrateBuilder) Build() (string, error) {
	if len(m.operations) == 0 {
		return "", nil
	}

	var sb strings.Builder
	sb.WriteString("ALTER TABLE ")
	sb.WriteString(m.generator.dialect.Quote(m.tableName))
	sb.WriteString("\n  ")
	sb.WriteString(strings.Join(m.operations, ",\n  "))
	sb.WriteString(";")

	return sb.String(), nil
}
