package cli

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// adapt 将 *Command 适配为 *cobra.Command。
//
// 转换规则：
//   - Use / Short / Long 直接映射
//   - RunE 包装为 cobra 的 func(*cobra.Command, []string) error，
//     在调用时将命令的合并 FlagSet（persistent + local）包装为 FlagSet 传入
//   - SubCommands 递归适配
func adapt(cmd *Command) *cobra.Command {
	c := &cobra.Command{
		Use:   cmd.Use,
		Short: cmd.Short,
		Long:  cmd.Long,
	}

	// 注册命令级别的 local flags
	for _, fd := range cmd.LocalFlags {
		switch fd.Kind {
		case "bool":
			c.Flags().Bool(fd.Name, fd.DefaultBool, fd.Usage)
		case "int":
			c.Flags().Int(fd.Name, fd.DefaultInt, fd.Usage)
		default: // "string" 或未指定
			c.Flags().String(fd.Name, fd.DefaultString, fd.Usage)
		}
	}

	if cmd.RunE != nil {
		runE := cmd.RunE // 捕获，避免闭包捕获循环变量
		c.RunE = func(cobraCmd *cobra.Command, args []string) error {
			ctx := cobraCmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			// 使用命令自身的合并 FlagSet（cobra 已自动把 persistent flags 合并进来）
			// 直接使用 Flags()（包含 local + inherited persistent flags）
			fs := adaptFlagSet(cobraCmd.Flags())
			return runE(ctx, fs, args)
		}
	}

	for _, sub := range cmd.SubCommands {
		c.AddCommand(adapt(sub))
	}

	return c
}

// adaptFlagSet 将 *pflag.FlagSet 包装为 FlagSet 接口。
func adaptFlagSet(pf *pflag.FlagSet) FlagSet {
	return &cobraFlagSet{fs: pf}
}

// cobraFlagSet 实现 FlagSet，底层委托给 pflag.FlagSet。
type cobraFlagSet struct {
	fs *pflag.FlagSet
}

func (c *cobraFlagSet) GetString(name string) string {
	v, _ := c.fs.GetString(name)
	return v
}

func (c *cobraFlagSet) GetBool(name string) bool {
	v, _ := c.fs.GetBool(name)
	return v
}

func (c *cobraFlagSet) GetInt(name string) int {
	v, _ := c.fs.GetInt(name)
	return v
}

func (c *cobraFlagSet) GetStringSlice(name string) []string {
	v, _ := c.fs.GetStringSlice(name)
	return v
}
