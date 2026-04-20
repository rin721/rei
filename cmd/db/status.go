package db

import (
	"context"
	"fmt"
	"io"
	"os"

	internalapp "github.com/rin721/rei/internal/app"
	"github.com/rin721/rei/pkg/cli"
	"github.com/rin721/rei/pkg/cli/flags"
)

// StatusCmd 实现 "db status" 子命令。
type StatusCmd struct{}

// Command 返回 "db status" 命令定义。
func (c *StatusCmd) Command() *cli.Command {
	return &cli.Command{
		Use:   "status",
		Short: "查看数据库迁移版本状态（applied / pending）",
		Long: `db status 连接数据库，输出所有迁移版本的状态表格：

  VERSION           DESCRIPTION         APPLIED_AT              STATUS
  20260420_001      init_schema         2026-04-20T00:30:00Z    applied
  20260421_002      add_user_avatar     (pending)               pending

示例：
  rei db status`,
		RunE: func(ctx context.Context, f cli.FlagSet, _ []string) error {
			return runDBStatus(ctx, f, os.Stdout) //nolint:forbidigo
		},
	}
}

func runDBStatus(ctx context.Context, f cli.FlagSet, _ io.Writer) error {
	application, err := internalapp.New(internalapp.Options{
		Mode:       internalapp.ModeDB,
		ConfigPath: f.GetString(flags.FlagConfig),
		DBOptions: internalapp.DBOptions{
			Action: internalapp.DBActionStatus,
		},
	})
	if err != nil {
		return fmt.Errorf("创建应用实例失败: %w", err)
	}

	if err := application.Run(ctx); err != nil {
		return fmt.Errorf("db status 失败: %w", err)
	}
	return nil
}
