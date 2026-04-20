package sqlgen

import (
	"fmt"
	"strings"
	"time"
)

// ============================================================================
// DELETE 生成方法
// ============================================================================

// Delete 生成 DELETE 语句
// 如果模型包含 DeletedAt 字段且未使用 Unscoped，则生成软删除 (UPDATE)
func (g *Generator) Delete(value interface{}, conds ...interface{}) (string, error) {
	ng := g.clone()

	if err := ng.parseModel(value); err != nil {
		return "", err
	}

	// 处理条件参数
	if len(conds) > 0 {
		ng = ng.processConditions(conds...)
	}

	// 检查是否使用软删除
	if !ng.ctx.Unscoped && ng.config.SoftDelete && hasSoftDelete(ng.ctx.ModelType) {
		return ng.buildSoftDelete()
	}

	return ng.buildHardDelete()
}

// ============================================================================
// 软删除构建
// ============================================================================

func (g *Generator) buildSoftDelete() (string, error) {
	var sb strings.Builder

	now := time.Now().Format("2006-01-02 15:04:05")

	sb.WriteString("UPDATE ")
	sb.WriteString(g.dialect.Quote(g.ctx.TableName))
	sb.WriteString(" SET ")
	sb.WriteString(g.dialect.Quote(DefaultSoftDeleteColumn))
	sb.WriteString("='")
	sb.WriteString(now)
	sb.WriteString("'")

	// WHERE 条件
	whereClause := g.buildDeleteWhereClause()
	if whereClause != "" {
		sb.WriteString(" WHERE ")
		sb.WriteString(whereClause)
	} else if !g.config.AllowEmptyCondition {
		return "", ErrMissingCondition
	}

	sb.WriteString(";")

	return sb.String(), nil
}

// ============================================================================
// 物理删除构建
// ============================================================================

func (g *Generator) buildHardDelete() (string, error) {
	var sb strings.Builder

	sb.WriteString("DELETE FROM ")
	sb.WriteString(g.dialect.Quote(g.ctx.TableName))

	// WHERE 条件
	whereClause := g.buildDeleteWhereClause()
	if whereClause != "" {
		sb.WriteString(" WHERE ")
		sb.WriteString(whereClause)
	} else if !g.config.AllowEmptyCondition {
		return "", ErrMissingCondition
	}

	sb.WriteString(";")

	return sb.String(), nil
}

// buildDeleteWhereClause 构建 DELETE 的 WHERE 子句
func (g *Generator) buildDeleteWhereClause() string {
	var conditions []string

	// 处理用户条件
	for _, cond := range g.ctx.WhereConditions {
		condStr := g.buildCondition(cond)
		if condStr != "" {
			conditions = append(conditions, condStr)
		}
	}

	// 从模型主键获取条件
	if len(conditions) == 0 && g.ctx.ModelValue.IsValid() {
		pk := getPrimaryKeyField(parseStructFields(g.ctx.ModelType, g.ctx.ModelValue, g.dialect))
		if pk != nil && !pk.IsZero {
			conditions = append(conditions, fmt.Sprintf("%s = %s",
				g.dialect.Quote(pk.ColumnName),
				formatValue(pk.Value, g.dialect.Quote)))
		}
	}

	return strings.Join(conditions, " AND ")
}

// ============================================================================
// 批量删除
// ============================================================================

// DeleteInBatches 生成批量删除 SQL
func (g *Generator) DeleteInBatches(value interface{}, batchSize int) ([]string, error) {
	// TODO: 实现批量删除
	sql, err := g.Delete(value)
	if err != nil {
		return nil, err
	}
	return []string{sql}, nil
}

// ============================================================================
// 清空表
// ============================================================================

// Truncate 生成 TRUNCATE 语句
func (g *Generator) Truncate(model interface{}) (string, error) {
	if err := g.parseModel(model); err != nil {
		return "", err
	}

	switch g.dialect.Name() {
	case SQLite:
		// SQLite 不支持 TRUNCATE
		return fmt.Sprintf("DELETE FROM %s;", g.dialect.Quote(g.ctx.TableName)), nil
	default:
		return fmt.Sprintf("TRUNCATE TABLE %s;", g.dialect.Quote(g.ctx.TableName)), nil
	}
}
