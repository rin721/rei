package app

import (
	"context"
	"fmt"
)

type modeRuntime interface {
	run(context.Context) error
}

func (a *App) newModeRuntime() (modeRuntime, error) {
	switch a.options.Mode {
	case ModeServer:
		return newServerModeRuntime(a), nil
	case ModeDB:
		return newDBModeRuntime(a, a.options.DBOptions), nil
	default:
		return nil, fmt.Errorf("unsupported app mode %q", a.options.Mode)
	}
}
