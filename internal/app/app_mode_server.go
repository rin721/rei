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
	if err := a.initLogger(); err != nil {
		return err
	}
	if err := a.initI18n(); err != nil {
		return err
	}
	if err := a.initIDGenerator(); err != nil {
		return err
	}
	if err := a.initCache(); err != nil {
		return err
	}
	if err := a.initDatabase(ctx); err != nil {
		return err
	}
	if err := a.initDBTx(); err != nil {
		return err
	}
	if err := a.initExecutor(); err != nil {
		return err
	}
	if err := a.initCrypto(); err != nil {
		return err
	}
	if err := a.initJWT(); err != nil {
		return err
	}
	if err := a.initStorage(); err != nil {
		return err
	}
	if err := a.initRBAC(); err != nil {
		return err
	}
	if err := a.initBusiness(); err != nil {
		return err
	}
	if err := a.initRouter(); err != nil {
		return err
	}
	if err := a.initHTTPServer(); err != nil {
		return err
	}

	a.registerReloadHooks()
	return nil
}
