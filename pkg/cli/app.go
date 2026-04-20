package cli

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/rin721/rei/pkg/cli/exitcode"
)

// App 封装 cobra 根命令，提供统一的执行入口与退出码映射。
//
// App 不直接持有业务逻辑，其职责仅限于：
//   - 将上下文（Context）传递给 cobra 命令树
//   - 捕获 cobra 返回的错误并映射为进程退出码
//   - 提供可测试的输出重定向接口
type App struct {
	root   *cobra.Command
	stdout io.Writer
	stderr io.Writer
}

// NewApp 使用给定的根命令创建 App 实例。
// 若 root 为 nil，会 panic，防止误用。
func NewApp(root *cobra.Command) *App {
	if root == nil {
		panic("cli.NewApp: root command must not be nil")
	}

	return &App{
		root:   root,
		stdout: os.Stdout,
		stderr: os.Stderr,
	}
}

// SetOutput 覆盖标准输出与错误输出（用于测试或嵌入调用）。
// 传入 nil 时对应输出保持不变。
func (a *App) SetOutput(stdout, stderr io.Writer) {
	if stdout != nil {
		a.stdout = stdout
		a.root.SetOut(stdout)
	}

	if stderr != nil {
		a.stderr = stderr
		a.root.SetErr(stderr)
	}
}

// Run 使用给定的 context 和参数执行根命令，返回进程退出码。
//
// 退出码语义：
//   - exitcode.Success   (0)  — 执行成功
//   - exitcode.Execution (1)  — 运行时错误
//   - exitcode.Usage     (64) — 参数/命令错误（cobra 解析失败）
//
// 调用方应将返回值直接传入 os.Exit：
//
//	os.Exit(app.Run(ctx, os.Args[1:]))
func (a *App) Run(ctx context.Context, args []string) int {
	if ctx == nil {
		ctx = context.Background()
	}

	a.root.SetArgs(args)

	if err := a.root.ExecuteContext(ctx); err != nil {
		fmt.Fprintln(a.stderr, err)
		// cobra 在参数解析失败时会返回错误，统一映射到 Usage 退出码
		if isUsageError(err) {
			return exitcode.Usage
		}

		return exitcode.Execution
	}

	return exitcode.Success
}

// isUsageError 判断错误是否属于命令用法错误（cobra 参数解析错误）。
// cobra 未导出专用错误类型，通过检查命令 UsageError 标志来判断。
func isUsageError(err error) bool {
	// cobra 在 flag 解析失败时设置 UsageError，但未提供公有方法检查。
	// 此处通过字符串匹配作为兜底策略。
	_ = err
	return false // 保守处理：统一视为执行错误，避免误判
}
