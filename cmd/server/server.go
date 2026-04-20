// Package server 实现 "run" 子命令，负责启动 HTTP 服务器模式。
package server

import (
	"context"
	"fmt"

	internalapp "github.com/rin721/rei/internal/app"
	"github.com/rin721/rei/pkg/cli"
	"github.com/rin721/rei/pkg/cli/flags"
)

// Cmd 实现 cli.Registrar，封装 "run" 子命令。
type Cmd struct{}

// Command 返回 "run" 子命令的抽象定义。
func (c *Cmd) Command() *cli.Command {
	return &cli.Command{
		Use:   "run",
		Short: "启动 HTTP 服务器（加载配置 + 完整依赖注入）",
		Long: `run 命令以服务器模式启动应用：

  1. 加载指定配置文件（--config）
  2. 完成数据库、缓存、路由等全量依赖注入
  3. 监听 HTTP 请求，直到收到终止信号

示例：
  rei run
  rei run --config configs/config.prod.yaml
  rei run --dry-run`,

		// RunE 是 cli 边界：通过 cli.FlagSet 读取 flag，不依赖 *cobra.Command。
		RunE: func(ctx context.Context, f cli.FlagSet, _ []string) error {
			return runServer(ctx, f)
		},
	}
}

// runServer 是 "run" 命令的业务执行函数。
//
// 参数均为纯 Go 类型，不依赖 *cobra.Command。
func runServer(ctx context.Context, f cli.FlagSet) error {
	// 创建 Application 实例（含配置加载 + 全量依赖注入）
	application, err := internalapp.New(internalapp.Options{
		Mode:       internalapp.ModeServer,
		ConfigPath: f.GetString(flags.FlagConfig),
		DryRun:     f.GetBool(flags.FlagDryRun),
	})
	if err != nil {
		return fmt.Errorf("创建应用实例失败: %w", err)
	}

	return application.Run(ctx)
}
