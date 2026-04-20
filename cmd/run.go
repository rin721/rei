package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// run 是程序的执行入口，负责：
//  1. 构建 CLI 应用（根命令 + 子命令注册）
//  2. 监听系统终止信号（SIGINT / SIGTERM），绑定到 context
//  3. 执行 CLI 并返回退出码
//
// 与 main 分离有助于测试：测试可直接调用 run(args)。
func run(args []string) int {
	// 监听操作系统终止信号，保证优雅关闭
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 构建 CLI App（含根命令 + 所有子命令 + flags + 中间件）
	app := buildApp()

	return app.Run(ctx, args)
}
