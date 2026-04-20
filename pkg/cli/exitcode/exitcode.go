// Package exitcode 定义进程退出码常量，与 POSIX 及通用命令行惯例对齐。
//
// 退出码约定：
//   - 0  成功
//   - 1  执行失败（运行时错误）
//   - 64 用法错误（参数错误、命令未知）
package exitcode

// Success 表示命令执行成功。
const Success = 0

// Execution 表示命令执行过程中发生运行时错误。
const Execution = 1

// Usage 表示命令参数错误或调用方式不符合规范。
const Usage = 64
