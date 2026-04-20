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

	return a.infrastructureProvisioning().shutdown(ctx)
}

func (p infrastructureProvisioning) shutdown(ctx context.Context) error {
	return runShutdownSteps(ctx, "shutdown runtime", p.shutdownSteps())
}

func (p infrastructureProvisioning) shutdownSteps() []shutdownStep {
	return []shutdownStep{
		newShutdownTask("config manager", p.stopConfigManager),
		newShutdownStep("http server", p.shutdownHTTPServer),
		newShutdownTask("storage", p.closeStorage),
		newShutdownStep("executor", p.shutdownExecutor),
		newShutdownTask("cache", p.closeCache),
		newShutdownTask("database", p.closeDatabase),
		newShutdownTask("rbac", p.closeRBAC),
		newShutdownTask("logger", p.flushLogger),
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

func (p infrastructureProvisioning) stopConfigManager() error {
	if p.configManager == nil {
		return nil
	}
	return p.configManager.Stop()
}

func (p infrastructureProvisioning) shutdownHTTPServer(ctx context.Context) error {
	if p.delivery.httpServer == nil {
		return nil
	}
	return p.delivery.httpServer.Shutdown(ctx)
}

func (p infrastructureProvisioning) closeStorage() error {
	if p.infra.storage == nil {
		return nil
	}
	return p.infra.storage.Close()
}

func (p infrastructureProvisioning) shutdownExecutor(ctx context.Context) error {
	_ = ctx
	if p.infra.executor == nil {
		return nil
	}
	p.infra.executor.Shutdown()
	return nil
}

func (p infrastructureProvisioning) closeCache() error {
	if p.infra.cache == nil {
		return nil
	}
	return p.infra.cache.Close()
}

func (p infrastructureProvisioning) closeDatabase() error {
	if p.infra.database == nil {
		return nil
	}
	return p.infra.database.Close()
}

func (p infrastructureProvisioning) closeRBAC() error {
	if p.infra.rbac == nil {
		return nil
	}
	return p.infra.rbac.Close()
}

func (p infrastructureProvisioning) flushLogger() error {
	if p.infra.logger == nil {
		return nil
	}
	return p.infra.logger.Sync()
}
