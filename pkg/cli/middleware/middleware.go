// Package middleware 提供 Cobra PersistentPreRunE 钩子链工具。
//
// Cobra 的 PersistentPreRunE 只支持单个函数，本包提供 ChainPreRunE
// 将多个钩子按顺序组合，实现类似中间件的前置处理流程。
//
// 典型用法：
//
//	root.PersistentPreRunE = middleware.ChainPreRunE(
//	    middleware.LogStartHook,
//	    middleware.ValidateConfigHook,
//	)
package middleware

import (
	"fmt"

	"github.com/spf13/cobra"
)

// PreRunE 与 cobra.Command.PersistentPreRunE 签名一致。
type PreRunE func(cmd *cobra.Command, args []string) error

// ChainPreRunE 将多个 PreRunE 函数串联为一个，按传入顺序执行。
// 任一函数返回非 nil 错误时，链条中断并直接返回该错误。
func ChainPreRunE(hooks ...PreRunE) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		for _, hook := range hooks {
			if err := hook(cmd, args); err != nil {
				return err
			}
		}
		return nil
	}
}

// LogStartHook 是一个开箱即用的前置钩子，在命令执行前打印启动信息。
// 输出格式：▶ <命令全路径> 正在启动...
func LogStartHook(cmd *cobra.Command, _ []string) error {
	fmt.Fprintf(cmd.OutOrStdout(), "▶ %s 正在启动...\n", cmd.CommandPath())
	return nil
}
