// Package migrate 提供两阶段数据库迁移能力：
//
//   - Generate（离线）：从 Go Model 反射生成版本化的 up/down SQL 脚本
//   - Migrate（在线）：扫描脚本目录，按版本顺序执行未执行的迁移
//   - Status：查看已执行/待执行版本状态
//   - Rollback：执行 down 脚本撤销最近一次迁移
//
// 锁机制：使用数据库表 schema_migrations 替代文件锁，支持多实例部署。
package migrate

import "time"

// Migration 描述一个迁移版本（对应一对 up/down 文件）。
type Migration struct {
	// Version 是排序键，格式为 YYYYMMDD_NNN（例如 20260420_001）。
	Version string
	// Description 是版本描述，来自文件名（例如 init_schema）。
	Description string
	// UpFile 是 .up.sql 文件的绝对路径。
	UpFile string
	// DownFile 是 .down.sql 文件的绝对路径（可为空）。
	DownFile string
	// Checksum 是 up 文件内容的 SHA256 十六进制字符串（首次扫描时计算）。
	Checksum string
}

// MigrationRecord 对应数据库 schema_migrations 表的一行记录。
type MigrationRecord struct {
	Version     string    `gorm:"column:version;primaryKey;size:64"`
	Description string    `gorm:"column:description;size:255"`
	AppliedAt   time.Time `gorm:"column:applied_at;not null"`
	Checksum    string    `gorm:"column:checksum;size:64;not null"`
}

// TableName 实现 gorm.Tabler。
func (MigrationRecord) TableName() string { return "schema_migrations" }

// MigrationStatus 汇总了迁移系统的当前状态。
type MigrationStatus struct {
	// Applied 是已执行的版本列表（按执行时间排序）。
	Applied []MigrationRecord
	// Pending 是待执行的版本列表（按版本号排序）。
	Pending []*Migration
}

// GenerateOptions 描述离线脚本生成的选项。
type GenerateOptions struct {
	// Models 是要生成 DDL 的 Go Model 指针列表（例如 []any{&User{}, &Role{}}）。
	Models []any
	// OutputDir 是脚本输出目录（例如 "scripts/migrations"）。
	OutputDir string
	// Version 是版本号；留空时自动生成 YYYYMMDD_NNN 格式。
	Version string
	// Description 是版本描述，用于文件名（例如 "init_schema"）。
	Description string
	// Dialect 是 SQL 方言（"mysql" / "postgres" / "sqlite" / "sqlserver"）。
	Dialect string
	// WithCRUD 为 true 时，额外生成 CRUD SQL 参考文档到 scripts/crud/。
	WithCRUD bool
}
