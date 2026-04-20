package app

import (
	"context"
	"fmt"
)

type shutdownStep struct {
	name string
	run  func(context.Context) error
}

func newShutdownStep(name string, run func(context.Context) error) shutdownStep {
	return shutdownStep{
		name: name,
		run:  run,
	}
}

func newShutdownTask(name string, run func() error) shutdownStep {
	return newShutdownStep(name, func(context.Context) error {
		return run()
	})
}

func (a *App) Shutdown(ctx context.Context) error {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), defaultShutdownTimeout)
		defer cancel()
	}

	return runShutdownSteps(ctx, "shutdown runtime", a.shutdownSteps())
}

func (a *App) shutdownSteps() []shutdownStep {
	return []shutdownStep{
		newShutdownTask("config manager", a.stopConfigManager),
		newShutdownStep("http server", a.shutdownHTTPServer),
		newShutdownTask("storage", a.closeStorage),
		newShutdownStep("executor", a.shutdownExecutor),
		newShutdownTask("cache", a.closeCache),
		newShutdownTask("database", a.closeDatabase),
		newShutdownTask("rbac", a.closeRBAC),
		newShutdownTask("logger", a.flushLogger),
	}
}

func runShutdownSteps(ctx context.Context, phase string, steps []shutdownStep) error {
	var errs []error
	for _, step := range steps {
		if err := step.run(ctx); err != nil {
			errs = append(errs, fmt.Errorf("%s: %s: %w", phase, step.name, err))
		}
	}
	return joinErrors(errs...)
}

func (a *App) stopConfigManager() error {
	if a.configManager == nil {
		return nil
	}
	return a.configManager.Stop()
}

func (a *App) shutdownHTTPServer(ctx context.Context) error {
	if a.httpServer == nil {
		return nil
	}
	return a.httpServer.Shutdown(ctx)
}

func (a *App) closeStorage() error {
	if a.storage == nil {
		return nil
	}
	return a.storage.Close()
}

func (a *App) shutdownExecutor(ctx context.Context) error {
	if a.executor == nil {
		return nil
	}
	return a.executor.Shutdown(ctx)
}

func (a *App) closeCache() error {
	if a.cache == nil {
		return nil
	}
	return a.cache.Close()
}

func (a *App) closeDatabase() error {
	if a.database == nil {
		return nil
	}
	return a.database.Close()
}

func (a *App) closeRBAC() error {
	if a.rbac == nil {
		return nil
	}
	return a.rbac.Close()
}

func (a *App) flushLogger() error {
	if a.logger == nil {
		return nil
	}
	return a.logger.Sync()
}
