package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

// Exit code 常量与通用命令行约定保持一致。
const (
	// ExitCodeSuccess 表示命令执行成功。
	ExitCodeSuccess = 0
	// ExitCodeExecution 表示命令执行失败。
	ExitCodeExecution = 1
	// ExitCodeUsage 表示命令参数或调用方式错误。
	ExitCodeUsage = 64
)

// Runner 定义命令执行函数签名。
type Runner func(context.Context, []string) error

// Command 描述一个稳定 CLI 命令。
type Command struct {
	Name        string
	Description string
	Run         Runner
}

// App 管理命令注册、帮助信息与执行入口。
type App struct {
	name     string
	summary  string
	commands map[string]Command
	stdout   io.Writer
	stderr   io.Writer
}

// New 创建一个新的 CLI 容器。
func New(name, summary string) *App {
	return &App{
		name:     name,
		summary:  summary,
		commands: make(map[string]Command),
		stdout:   os.Stdout,
		stderr:   os.Stderr,
	}
}

// SetOutput 设置标准输出与错误输出，便于测试与嵌入调用。
func (a *App) SetOutput(stdout, stderr io.Writer) {
	if stdout != nil {
		a.stdout = stdout
	}

	if stderr != nil {
		a.stderr = stderr
	}
}

// Register 注册一个稳定命令。
func (a *App) Register(command Command) error {
	if strings.TrimSpace(command.Name) == "" {
		return errors.New("command name is required")
	}

	if command.Run == nil {
		return fmt.Errorf("command %q requires a runner", command.Name)
	}

	if _, exists := a.commands[command.Name]; exists {
		return fmt.Errorf("command %q already registered", command.Name)
	}

	a.commands[command.Name] = command
	return nil
}

// Run 根据输入参数执行对应命令，并返回进程退出码。
func (a *App) Run(ctx context.Context, args []string) int {
	if ctx == nil {
		ctx = context.Background()
	}

	if len(args) == 0 {
		fmt.Fprint(a.stdout, a.Usage())
		return ExitCodeUsage
	}

	command, exists := a.commands[args[0]]
	if !exists {
		fmt.Fprintf(a.stderr, "unknown command: %s\n\n%s", args[0], a.Usage())
		return ExitCodeUsage
	}

	if err := command.Run(ctx, args[1:]); err != nil {
		fmt.Fprintf(a.stderr, "%v\n", err)
		return ExitCodeExecution
	}

	return ExitCodeSuccess
}

// Usage 返回格式化后的帮助文本。
func (a *App) Usage() string {
	var builder strings.Builder

	builder.WriteString(a.name)
	if a.summary != "" {
		builder.WriteString(" - ")
		builder.WriteString(a.summary)
	}
	builder.WriteString("\n\nUsage:\n  ")
	builder.WriteString(a.name)
	builder.WriteString(" <command> [arguments]\n\nCommands:\n")

	names := make([]string, 0, len(a.commands))
	for name := range a.commands {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		command := a.commands[name]
		builder.WriteString("  ")
		builder.WriteString(command.Name)
		if command.Description != "" {
			builder.WriteString("\t")
			builder.WriteString(command.Description)
		}
		builder.WriteString("\n")
	}

	return builder.String()
}
