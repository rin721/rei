package app

import (
	"context"
)

func (a *App) runModeServer(ctx context.Context) (err error) {
	if err := a.bootstrapServer(ctx); err != nil {
		return err
	}
	defer func() {
		shutdownErr := a.Shutdown(context.TODO())
		if err == nil {
			err = shutdownErr
		}
	}()

	if a.options.DryRun {
		return nil
	}

	if err := a.httpServer.Start(); err != nil {
		return err
	}
	if err := a.startReload(ctx); err != nil {
		return err
	}

	<-ctx.Done()
	return nil
}

func (a *App) bootstrapServer(ctx context.Context) error {
	if err := a.bootstrapServerInfrastructure(ctx); err != nil {
		return err
	}
	if err := a.bootstrapBusinessRuntime(ctx); err != nil {
		return err
	}
	if err := a.bootstrapDeliveryRuntime(ctx); err != nil {
		return err
	}

	a.registerReloadHooks()
	return nil
}
