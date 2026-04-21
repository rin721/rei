package app

import (
	"context"
)

type serverModeRuntime struct {
	app *App
}

func newServerModeRuntime(app *App) serverModeRuntime {
	return serverModeRuntime{app: app}
}

func (a *App) bootstrapServer(ctx context.Context) error {
	return newServerModeRuntime(a).bootstrap(ctx)
}

func (r serverModeRuntime) run(ctx context.Context) (err error) {
	if err := r.bootstrap(ctx); err != nil {
		return err
	}
	defer func() {
		shutdownErr := r.app.Shutdown(context.TODO())
		if err == nil {
			err = shutdownErr
		}
	}()

	if r.app.options.DryRun {
		return nil
	}

	if err := r.app.deliveryProvisioning().start(ctx); err != nil {
		return err
	}
	if err := r.app.infrastructureProvisioning().startReload(ctx); err != nil {
		return err
	}

	<-ctx.Done()
	return nil
}

func (r serverModeRuntime) bootstrap(ctx context.Context) error {
	if err := r.app.infrastructureProvisioning().bootstrapServer(ctx); err != nil {
		return err
	}
	if err := r.app.businessProvisioning().bootstrap(ctx); err != nil {
		return err
	}
	if err := r.app.deliveryProvisioning().bootstrap(ctx); err != nil {
		return err
	}

	r.app.infrastructureProvisioning().registerReloadHooks(r.app.reloadComponents)
	return nil
}
