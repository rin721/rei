package db

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	internalapp "github.com/rin721/rei/internal/app"
	"github.com/rin721/rei/pkg/cli"
	"github.com/rin721/rei/pkg/cli/flags"
	"github.com/rin721/rei/pkg/cli/prompt"
)

// MigrateCmd 实现 "db migrate" 子命令。
type MigrateCmd struct{}

// Command 返回 "db migrate" 命令定义。
func (c *MigrateCmd) Command() *cli.Command {
	return &cli.Command{
		Use:   "migrate",
		Short: "执行所有待执行的数据库迁移",
		Long: `db migrate 连接数据库，按版本顺序执行所有待执行的 up 脚本。

执行步骤：
  1. 确保 schema_migrations 锁表存在
  2. 扫描 migrations_dir，过滤已执行版本
  3. 在单事务中执行每条 pending 迁移，写入执行记录

示例：
  rei db migrate
  rei db migrate --yes           # 跳过确认
  rei db migrate --dry-run       # 仅打印将执行的 SQL`,
		RunE: func(ctx context.Context, f cli.FlagSet, _ []string) error {
			return runDBMigrate(ctx, f, os.Stdout) //nolint:forbidigo
		},
	}
}

func runDBMigrate(ctx context.Context, f cli.FlagSet, out io.Writer) error {
	dryRun := f.GetBool(flags.FlagDryRun)

	if !f.GetBool(flags.FlagYes) && !dryRun {
		confirmed, err := prompt.Confirm("确认执行数据库迁移？")
		if err != nil {
			return fmt.Errorf("交互确认失败: %w", err)
		}
		if !confirmed {
			fmt.Fprintln(out, "已取消数据库迁移。")
			return nil
		}
	}

	application, err := internalapp.New(internalapp.Options{
		Mode:       internalapp.ModeDB,
		ConfigPath: f.GetString(flags.FlagConfig),
		DryRun:     dryRun,
		DBOptions: internalapp.DBOptions{
			Action: internalapp.DBActionMigrate,
			DryRun: dryRun,
		},
	})
	if err != nil {
		return fmt.Errorf("创建应用实例失败: %w", err)
	}

	if err := application.Run(ctx); err != nil {
		if errors.Is(err, prompt.ErrAborted) {
			fmt.Fprintln(out, "已取消数据库迁移。")
			return nil
		}
		return fmt.Errorf("db migrate 失败: %w", err)
	}

	if !dryRun {
		fmt.Fprintln(out, "✓ 数据库迁移完成。")
	}
	return nil
}
