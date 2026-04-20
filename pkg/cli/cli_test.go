package cli_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/rin721/rei/pkg/cli"
	"github.com/rin721/rei/pkg/cli/exitcode"
)

// fakeRegistrar 实现 cli.Registrar，用于测试中注册一个简单的 fake 子命令。
type fakeRegistrar struct {
	name    string
	runFunc func(ctx context.Context, f cli.FlagSet, args []string) error
}

func (f *fakeRegistrar) Command() *cli.Command {
	return &cli.Command{
		Use:   f.name,
		Short: f.name + " command (test)",
		RunE:  f.runFunc,
	}
}

// TestAppRunSuccess 验证已注册子命令执行成功时返回 exitcode.Success。
func TestAppRunSuccess(t *testing.T) {
	t.Parallel()

	root := cli.BuildRootCmd("test", "test cli",
		&fakeRegistrar{
			name: "greet",
			runFunc: func(ctx context.Context, f cli.FlagSet, args []string) error {
				return nil
			},
		},
	)

	var stdout, stderr bytes.Buffer
	app := cli.NewApp(root)
	app.SetOutput(&stdout, &stderr)

	code := app.Run(context.Background(), []string{"greet"})
	if code != exitcode.Success {
		t.Fatalf("Run() code = %d, want %d; stderr = %q", code, exitcode.Success, stderr.String())
	}
}

// TestAppRunUnknownCommand 验证未知子命令：
//   - cobra 遇到无效子命令会返回错误
//   - App.Run 应输出错误信息到 stderr 并返回非零退出码
func TestAppRunUnknownCommand(t *testing.T) {
	t.Parallel()

	// 注册一个真实子命令，使 cobra 能正确区分"有子命令但命令名不对"的场景
	root := cli.BuildRootCmd("test", "test cli",
		&fakeRegistrar{
			name: "known",
			runFunc: func(ctx context.Context, f cli.FlagSet, args []string) error {
				return nil
			},
		},
	)

	var stdout, stderr bytes.Buffer
	app := cli.NewApp(root)
	app.SetOutput(&stdout, &stderr)

	code := app.Run(context.Background(), []string{"nonexistent"})
	if code == exitcode.Success {
		t.Fatalf("Run() with unknown command returned success (code=0); stderr=%q", stderr.String())
	}
}

// TestAppRunNoArgs 验证无参数时执行根命令帮助，返回 exitcode.Success（显示帮助不视为错误）。
func TestAppRunNoArgs(t *testing.T) {
	t.Parallel()

	root := cli.BuildRootCmd("test", "test cli")

	var stdout, stderr bytes.Buffer
	app := cli.NewApp(root)
	app.SetOutput(&stdout, &stderr)

	// 无参数执行根命令，cobra 会调用 RunE（打印帮助），返回 Success
	code := app.Run(context.Background(), []string{})
	if code != exitcode.Success {
		t.Fatalf("Run() with no args code = %d, want %d", code, exitcode.Success)
	}
}

// TestBuildRootCmdMultipleRegistrars 验证多个 Registrar 能同时注册成功。
func TestBuildRootCmdMultipleRegistrars(t *testing.T) {
	t.Parallel()

	ran := make(map[string]bool)

	makeReg := func(name string) *fakeRegistrar {
		return &fakeRegistrar{
			name: name,
			runFunc: func(ctx context.Context, f cli.FlagSet, args []string) error {
				ran[name] = true
				return nil
			},
		}
	}

	root := cli.BuildRootCmd("test", "test cli",
		makeReg("cmd1"),
		makeReg("cmd2"),
	)

	var stdout, stderr bytes.Buffer
	app := cli.NewApp(root)
	app.SetOutput(&stdout, &stderr)

	for _, name := range []string{"cmd1", "cmd2"} {
		code := app.Run(context.Background(), []string{name})
		if code != exitcode.Success {
			t.Fatalf("Run(%q) code = %d, want %d", name, code, exitcode.Success)
		}
	}

	for _, name := range []string{"cmd1", "cmd2"} {
		if !ran[name] {
			t.Errorf("command %q was not executed", name)
		}
	}
}
