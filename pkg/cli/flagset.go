package cli

// FlagSet 是 cli 包对命令行 Flag 读取能力的抽象接口。
//
// 业务代码通过此接口读取 flag，不感知底层 cobra/pflag 实现细节。
// 框架在调用 RunE 之前，会将 cobra 的 pflag.FlagSet 包装为此接口传入。
//
// 示例：
//
//	func runServer(ctx context.Context, f cli.FlagSet, _ []string) error {
//	    configPath := f.GetString("config")
//	    dryRun     := f.GetBool("dry-run")
//	    // ...
//	}
type FlagSet interface {
	// GetString 返回指定名称的 string flag 值；flag 不存在时返回空字符串。
	GetString(name string) string
	// GetBool 返回指定名称的 bool flag 值；flag 不存在时返回 false。
	GetBool(name string) bool
	// GetInt 返回指定名称的 int flag 值；flag 不存在时返回 0。
	GetInt(name string) int
	// GetStringSlice 返回指定名称的 []string flag 值；flag 不存在时返回 nil。
	GetStringSlice(name string) []string
}
