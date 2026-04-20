// Package flags 定义应用级全局 CLI Flag，供框架内部统一注册与读取。
//
// 所有 Flag 均注册为 persistent（持久）标志，挂载在根命令上，
// 子命令可直接继承，无需重复声明。
//
// # 分层设计
//
// 该包是 pkg/cli 框架的内部组件，负责：
//   - 声明全局 flag 名常量（供 cobraadapter 读取时引用）
//   - 在根命令上注册全局持久 flag（RegisterPersistent）
//
// 业务代码读取 flag 应通过 [cli.FlagSet] 接口（由框架在调用 RunE 时注入），
// 而不是直接调用本包的函数。
package flags

import "github.com/spf13/cobra"

const (
	// FlagConfig 是配置文件路径的 flag 名。
	FlagConfig = "config"
	// FlagDryRun 是干跑模式的 flag 名。
	FlagDryRun = "dry-run"
	// FlagYes 是跳过交互确认的 flag 名。
	FlagYes = "yes"
)

// RegisterPersistent 将全局持久 Flag 注册到指定命令（通常为根命令）。
//
// 由 BuildRootCmd 在内部调用，子命令自动继承这些 Flag，无需重复声明。
func RegisterPersistent(cmd *cobra.Command) {
	pf := cmd.PersistentFlags()

	pf.String(
		FlagConfig, "",
		"配置文件路径（默认读取 configs/config.yaml）",
	)
	pf.Bool(
		FlagDryRun, false,
		"干跑模式：仅初始化组件，完成后立即退出，不执行实际业务",
	)
	pf.BoolP(
		FlagYes, "y", false,
		"跳过交互确认提示，自动回答 yes（适用于 CI/脚本环境）",
	)
}
