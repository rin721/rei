package app

import (
	"context"
	"fmt"
)

type Mode string

const (
	ModeServer Mode = "server"
	ModeDB     Mode = "db"
)

func (a *App) Run(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	switch a.options.Mode {
	case ModeServer:
		return a.runModeServer(ctx)
	case ModeDB:
		return a.runModeDB(ctx)
	default:
		return fmt.Errorf("unsupported app mode %q", a.options.Mode)
	}
}
