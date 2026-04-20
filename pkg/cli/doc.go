// Package cli 提供基于 Cobra 的分层 CLI 框架核心能力。
//
// 架构分层：
//
//	pkg/cli/
//	├── exitcode/   退出码常量（Success / Execution / Usage）
//	├── flags/      全局持久 Flag 定义与读取工具
//	├── prompt/     Survey 交互式提示封装（Confirm / Select / Input）
//	├── middleware/  Cobra PersistentPreRunE 钩子链工具
//	├── command.go  Registrar 接口（子命令扩展点）
//	├── root.go     BuildRootCmd 根命令工厂
//	└── app.go      App 入口（Run / 信号处理 / 退出码映射）
//
// 典型使用流程：
//
//  1. 实现 cli.Registrar 接口，构建各子命令
//  2. 调用 cli.BuildRootCmd 创建根命令，注入 Registrar
//  3. 通过 cli.NewApp 包装根命令
//  4. 在 main 中调用 app.Run(ctx, os.Args[1:])，并以返回值作为退出码
package cli
