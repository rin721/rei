package cli

import (
	"github.com/spf13/cobra"

	"github.com/rin721/rei/pkg/cli/flags"
	"github.com/rin721/rei/pkg/cli/middleware"
)

// BuildRootCmd 创建并返回配置好的根命令（root cobra.Command）。
//
// 根命令会自动完成以下初始化：
//   - 设置应用名称（Use）与描述（Short / Long）
//   - 注册全局持久 Flag（--config / --dry-run / --yes）
//   - 挂载默认 PersistentPreRunE 钩子（LogStartHook）
//   - 依次调用所有 Registrar，将子命令适配并挂载到根命令
//
// 参数：
//   - name：命令名称，通常与二进制文件名一致
//   - summary：命令简短描述，显示在 --help 第一行
//   - registrars：实现 Registrar 接口的子命令列表
//
// 返回值：
//   - *cobra.Command 配置完毕的根命令，可直接传入 NewApp 使用
func BuildRootCmd(name, summary string, registrars ...Registrar) *cobra.Command {
	root := &cobra.Command{
		Use:   name,
		Short: summary,
		Long: summary + "\n\n" +
			"使用 <命令> --help 查看各子命令详细说明。",

		// SilenceErrors 避免 cobra 重复打印错误（由 App.Run 统一处理）
		SilenceErrors: true,
		// SilenceUsage 避免命令出错时自动打印完整 Usage，保持输出简洁
		SilenceUsage: true,

		// 根命令本身不执行任何逻辑，仅作为子命令容器
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},

		// PersistentPreRunE 在所有子命令执行前触发
		PersistentPreRunE: middleware.ChainPreRunE(
			middleware.LogStartHook,
		),
	}

	// 注册全局持久 Flag（子命令自动继承）
	flags.RegisterPersistent(root)

	// 将每个 Registrar 返回的 *cli.Command 适配为 *cobra.Command 并挂载
	for _, r := range registrars {
		root.AddCommand(adapt(r.Command()))
	}

	return root
}
