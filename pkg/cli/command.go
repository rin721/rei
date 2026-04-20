package cli

import "context"

// RunE 是命令执行函数的统一签名。
//
//   - ctx  由框架通过 cobra ExecuteContext 注入，携带生命周期信号
//   - f    flag 读取接口（屏蔽底层 pflag/cobra，业务代码只感知此接口）
//   - args 位置参数（cobra 解析后的非 flag 部分）
type RunE func(ctx context.Context, f FlagSet, args []string) error

// FlagDef 描述一个命令级别的 local flag 定义。
//
// 支持三种类型：string / bool / int。
type FlagDef struct {
	// Name 是 flag 名称（例如 "version"）。
	Name string
	// Usage 是 flag 说明文字。
	Usage string
	// DefaultString 是 string 类型 flag 的默认值。
	DefaultString string
	// DefaultBool 是 bool 类型 flag 的默认值。
	DefaultBool bool
	// DefaultInt 是 int 类型 flag 的默认值。
	DefaultInt int
	// Kind 区分 flag 类型："string" / "bool" / "int"，默认 "string"。
	Kind string
}

// Command 是 cli 包对 CLI 命令的抽象表示。
//
// 业务代码通过此类型定义命令，不直接创建 cobra.Command。
// 框架内部会将 *Command 转换为 *cobra.Command（见 adapt.go）。
//
// 示例：
//
//	func (c *Cmd) Command() *cli.Command {
//	    return &cli.Command{
//	        Use:   "run",
//	        Short: "启动 HTTP 服务器",
//	        RunE:  runServer,
//	    }
//	}
type Command struct {
	// Use 是命令名称（cobra Use 字段）。
	Use string
	// Short 是命令的单行描述，显示在 --help 列表中。
	Short string
	// Long 是命令的详细描述，显示在命令自身的 --help 中。
	Long string
	// RunE 是命令的执行函数；为 nil 时命令仅作为分组容器。
	RunE RunE
	// SubCommands 允许嵌套子命令，无需引入 cobra 类型。
	SubCommands []*Command
	// LocalFlags 声明此命令专属的 local flags（不影响子命令）。
	// 框架适配时会将这些 flag 注册到对应 cobra.Command.Flags() 上。
	LocalFlags []FlagDef
}

// Registrar 是子命令的注册接口（适配器 Port）。
//
// 实现者返回 *cli.Command，由框架内部将其转换为 *cobra.Command 并挂载。
// 业务代码实现此接口时无需导入 cobra 包。
//
// 示例：
//
//	type ServerCmd struct{}
//
//	func (s *ServerCmd) Command() *cli.Command {
//	    return &cli.Command{Use: "run", RunE: runServer}
//	}
type Registrar interface {
	// Command 返回该注册器对应的命令定义。
	Command() *Command
}
