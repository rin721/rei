// Package db 实现 "db" 命令组，提供数据库迁移管理能力。
//
// 子命令：
//   - db generate  离线生成版本化迁移脚本
//   - db migrate   执行待执行的迁移
//   - db status    查看迁移版本状态
//   - db rollback  回滚最近一次迁移
package db

import "github.com/rin721/rei/pkg/cli"

// Cmd 实现 cli.Registrar，封装 "db" 命令组。
type Cmd struct{}

// Command 返回 "db" 命令组的定义，RunE 为 nil（纯容器命令）。
func (c *Cmd) Command() *cli.Command {
	return &cli.Command{
		Use:   "db",
		Short: "数据库管理命令组（迁移脚本生成/执行/状态/回滚）",
		Long: `db 命令组提供两阶段数据库迁移能力：

  阶段一（离线）：
    rei db generate   从 Go Model 反射生成版本化的 up/down SQL 脚本

  阶段二（在线）：
    rei db migrate    按版本顺序执行所有待执行迁移，记录到 schema_migrations 锁表
    rei db status     查看已执行与待执行版本状态
    rei db rollback   执行 down 脚本，撤销最近一次迁移

示例：
  rei db generate --desc init_schema
  rei db migrate --yes
  rei db status
  rei db rollback`,
		SubCommands: []*cli.Command{
			(&GenerateCmd{}).Command(),
			(&MigrateCmd{}).Command(),
			(&StatusCmd{}).Command(),
			(&RollbackCmd{}).Command(),
		},
	}
}
