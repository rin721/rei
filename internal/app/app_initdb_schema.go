package app

import (
	"fmt"
	"strings"

	"github.com/rei0721/go-scaffold2/internal/config"
	pkgsqlgen "github.com/rei0721/go-scaffold2/pkg/sqlgen"
)

func buildInitDBTables(cfg config.Config) ([]pkgsqlgen.Table, error) {
	driver := initDBDriver(cfg)
	if driver == "" {
		return nil, fmt.Errorf("initdb driver is required")
	}

	tables := []pkgsqlgen.Table{
		{
			Name: "users",
			Columns: []pkgsqlgen.Column{
				{Name: "id", Type: typeFor(driver, "id"), Nullable: false, PrimaryKey: true},
				{Name: "created_at", Type: typeFor(driver, "timestamp"), Nullable: false},
				{Name: "updated_at", Type: typeFor(driver, "timestamp"), Nullable: false},
				{Name: "deleted_at", Type: typeFor(driver, "timestamp"), Nullable: true},
				{Name: "username", Type: typeFor(driver, "string_64"), Nullable: false, Unique: true},
				{Name: "email", Type: typeFor(driver, "string_128"), Nullable: true},
				{Name: "display_name", Type: typeFor(driver, "string_128"), Nullable: false},
				{Name: "password_hash", Type: typeFor(driver, "string_255"), Nullable: false},
				{Name: "status", Type: typeFor(driver, "string_32"), Nullable: false, Default: quoted("active")},
			},
		},
		{
			Name: "roles",
			Columns: []pkgsqlgen.Column{
				{Name: "id", Type: typeFor(driver, "id"), Nullable: false, PrimaryKey: true},
				{Name: "created_at", Type: typeFor(driver, "timestamp"), Nullable: false},
				{Name: "updated_at", Type: typeFor(driver, "timestamp"), Nullable: false},
				{Name: "deleted_at", Type: typeFor(driver, "timestamp"), Nullable: true},
				{Name: "name", Type: typeFor(driver, "string_64"), Nullable: false, Unique: true},
				{Name: "description", Type: typeFor(driver, "string_255"), Nullable: true},
			},
		},
		{
			Name: "user_roles",
			Columns: []pkgsqlgen.Column{
				{Name: "id", Type: typeFor(driver, "id"), Nullable: false, PrimaryKey: true},
				{Name: "created_at", Type: typeFor(driver, "timestamp"), Nullable: false},
				{Name: "updated_at", Type: typeFor(driver, "timestamp"), Nullable: false},
				{Name: "deleted_at", Type: typeFor(driver, "timestamp"), Nullable: true},
				{Name: "user_id", Type: typeFor(driver, "id"), Nullable: false},
				{Name: "role_name", Type: typeFor(driver, "string_64"), Nullable: false},
			},
			UniqueConstraints: []pkgsqlgen.UniqueConstraint{
				{Columns: []string{"user_id", "role_name"}},
			},
		},
		{
			Name: "policies",
			Columns: []pkgsqlgen.Column{
				{Name: "id", Type: typeFor(driver, "id"), Nullable: false, PrimaryKey: true},
				{Name: "created_at", Type: typeFor(driver, "timestamp"), Nullable: false},
				{Name: "updated_at", Type: typeFor(driver, "timestamp"), Nullable: false},
				{Name: "deleted_at", Type: typeFor(driver, "timestamp"), Nullable: true},
				{Name: "subject", Type: typeFor(driver, "string_64"), Nullable: false},
				{Name: "object", Type: typeFor(driver, "string_255"), Nullable: false},
				{Name: "action", Type: typeFor(driver, "string_32"), Nullable: false},
			},
			UniqueConstraints: []pkgsqlgen.UniqueConstraint{
				{Columns: []string{"subject", "object", "action"}},
			},
		},
		{
			Name: "samples",
			Columns: []pkgsqlgen.Column{
				{Name: "id", Type: typeFor(driver, "id"), Nullable: false, PrimaryKey: true},
				{Name: "created_at", Type: typeFor(driver, "timestamp"), Nullable: false},
				{Name: "updated_at", Type: typeFor(driver, "timestamp"), Nullable: false},
				{Name: "deleted_at", Type: typeFor(driver, "timestamp"), Nullable: true},
				{Name: "name", Type: typeFor(driver, "string_128"), Nullable: false, Unique: true},
				{Name: "description", Type: typeFor(driver, "string_255"), Nullable: true},
				{Name: "enabled", Type: typeFor(driver, "boolean"), Nullable: false, Default: booleanLiteral(driver, true)},
			},
		},
	}

	if cfg.RBAC.Enabled {
		policyTable := strings.TrimSpace(cfg.RBAC.PolicyTable)
		if policyTable == "" {
			policyTable = "casbin_rule"
		}
		tables = append(tables, pkgsqlgen.Table{
			Name: policyTable,
			Columns: []pkgsqlgen.Column{
				{Name: "id", Type: typeFor(driver, "integer"), Nullable: false, PrimaryKey: true},
				{Name: "ptype", Type: typeFor(driver, "string_100"), Nullable: false},
				{Name: "v0", Type: typeFor(driver, "string_100"), Nullable: true},
				{Name: "v1", Type: typeFor(driver, "string_100"), Nullable: true},
				{Name: "v2", Type: typeFor(driver, "string_100"), Nullable: true},
				{Name: "v3", Type: typeFor(driver, "string_100"), Nullable: true},
				{Name: "v4", Type: typeFor(driver, "string_100"), Nullable: true},
				{Name: "v5", Type: typeFor(driver, "string_100"), Nullable: true},
			},
		})
	}

	return tables, nil
}

func typeFor(driver, kind string) string {
	switch strings.ToLower(driver) {
	case "postgres", "postgresql":
		switch kind {
		case "id":
			return "VARCHAR(32)"
		case "integer":
			return "INTEGER"
		case "timestamp":
			return "TIMESTAMPTZ"
		case "boolean":
			return "BOOLEAN"
		case "string_32":
			return "VARCHAR(32)"
		case "string_64":
			return "VARCHAR(64)"
		case "string_100":
			return "VARCHAR(100)"
		case "string_128":
			return "VARCHAR(128)"
		case "string_255":
			return "VARCHAR(255)"
		default:
			return "TEXT"
		}
	case "mysql":
		switch kind {
		case "id":
			return "VARCHAR(32)"
		case "integer":
			return "INT"
		case "timestamp":
			return "DATETIME"
		case "boolean":
			return "BOOLEAN"
		case "string_32":
			return "VARCHAR(32)"
		case "string_64":
			return "VARCHAR(64)"
		case "string_100":
			return "VARCHAR(100)"
		case "string_128":
			return "VARCHAR(128)"
		case "string_255":
			return "VARCHAR(255)"
		default:
			return "TEXT"
		}
	default:
		switch kind {
		case "integer":
			return "INTEGER"
		case "timestamp":
			return "DATETIME"
		case "boolean":
			return "BOOLEAN"
		default:
			return "TEXT"
		}
	}
}

func quoted(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}

func booleanLiteral(driver string, value bool) string {
	switch strings.ToLower(driver) {
	case "postgres", "postgresql":
		if value {
			return "TRUE"
		}
		return "FALSE"
	default:
		if value {
			return "1"
		}
		return "0"
	}
}
