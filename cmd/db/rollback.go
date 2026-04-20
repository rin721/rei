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

const flagSteps = "steps"

// RollbackCmd 实现 "db rollback" 子命令。
type RollbackCmd struct{}

// Command 返回 "db rollback" 命令定义。
func (c *RollbackCmd) Command() *cli.Command {
	return &cli.Command{
		Use:   "rollback",
		Short: "回滚最近一次（或 N 次）数据库迁移",
		Long: `db rollback 执行最近已应用版本的 down 脚本，撤销迁移并删除锁表记录。

  - 若目标版本无 .down.sql 文件，命令将报错退出
  - --steps N 指定回滚步数（默认 1）

示例：
  rei db rollback
  rei db rollback --steps 2
  rei db rollback --yes
  rei db rollback --dry-run`,
		LocalFlags: []cli.FlagDef{
			{Name: flagSteps, Kind: "int", DefaultInt: 1, Usage: "回滚步数（默认 1）"},
		},
		RunE: func(ctx context.Context, f cli.FlagSet, _ []string) error {
			return runDBRollback(ctx, f, os.Stdout) //nolint:forbidigo
		},
	}
}

func runDBRollback(ctx context.Context, f cli.FlagSet, out io.Writer) error {
	dryRun := f.GetBool(flags.FlagDryRun)
	steps := f.GetInt(flagSteps)
	if steps < 1 {
		steps = 1
	}

	if !f.GetBool(flags.FlagYes) && !dryRun {
		confirmed, err := prompt.Confirm(fmt.Sprintf("确认回滚最近 %d 次迁移？", steps))
		if err != nil {
			return fmt.Errorf("交互确认失败: %w", err)
		}
		if !confirmed {
			fmt.Fprintln(out, "已取消回滚。")
			return nil
		}
	}

	application, err := internalapp.New(internalapp.Options{
		Mode:       internalapp.ModeDB,
		ConfigPath: f.GetString(flags.FlagConfig),
		DryRun:     dryRun,
		DBOptions: internalapp.DBOptions{
			Action: internalapp.DBActionRollback,
			Steps:  steps,
			DryRun: dryRun,
		},
	})
	if err != nil {
		return fmt.Errorf("创建应用实例失败: %w", err)
	}

	if err := application.Run(ctx); err != nil {
		if errors.Is(err, prompt.ErrAborted) {
			fmt.Fprintln(out, "已取消回滚。")
			return nil
		}
		return fmt.Errorf("db rollback 失败: %w", err)
	}

	if !dryRun {
		fmt.Fprintln(out, "✓ 回滚完成。")
	}
	return nil
}
