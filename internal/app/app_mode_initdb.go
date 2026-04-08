package app

import (
	"context"
	"fmt"
)

func (a *App) runModeInitDB(ctx context.Context) (err error) {
	if err := a.bootstrapInitDB(ctx); err != nil {
		return err
	}
	defer func() {
		shutdownErr := a.Shutdown(nil)
		if err == nil {
			err = shutdownErr
		}
	}()

	return a.runInitDB(ctx)
}

func (a *App) bootstrapInitDB(ctx context.Context) error {
	if err := a.initLogger(); err != nil {
		return err
	}
	if err := a.initDatabase(ctx); err != nil {
		return err
	}
	if a.database == nil {
		return fmt.Errorf("initdb mode requires database.enabled to be true")
	}
	if err := a.initRBAC(); err != nil {
		return err
	}
	return nil
}
