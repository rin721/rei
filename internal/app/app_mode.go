package app

import (
	"context"
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

	runtime, err := a.newModeRuntime()
	if err != nil {
		return err
	}

	return runtime.run(ctx)
}
