package app

import (
	"context"
	"fmt"
)

// Mode 描述应用运行模式。
type Mode string

const (
	// ModeServer 表示长期运行的 HTTP 服务模式。
	ModeServer Mode = "server"
	// ModeInitDB 表示一次性的 initdb 模式。
	ModeInitDB Mode = "initdb"
)

// Run 根据模式执行应用。
func (a *App) Run(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	switch a.options.Mode {
	case ModeServer:
		return a.runModeServer(ctx)
	case ModeInitDB:
		return a.runModeInitDB(ctx)
	default:
		return fmt.Errorf("unsupported app mode %q", a.options.Mode)
	}
}
