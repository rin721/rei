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

	return joinErrors(
		a.deliveryProvisioning().shutdown(ctx),
		a.infrastructureProvisioning().shutdown(ctx),
	)
}

func (p infrastructureProvisioning) shutdown(ctx context.Context) error {
	steps, err := p.shutdownSteps()
	if err != nil {
		return err
	}
	return runShutdownSteps(ctx, "shutdown infrastructure runtime", steps)
}

func (p deliveryProvisioning) shutdown(ctx context.Context) error {
	steps, err := p.shutdownSteps()
	if err != nil {
		return err
	}
	return runShutdownSteps(ctx, "shutdown delivery runtime", steps)
}

func (p infrastructureProvisioning) shutdownSteps() ([]shutdownStep, error) {
	return p.capabilities().shutdownSteps(infrastructureProfileRuntimeShutdown, p)
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

func (p deliveryProvisioning) shutdownSteps() ([]shutdownStep, error) {
	return p.lifecycle().shutdownSteps(deliveryCapabilityProfileRuntime)
}

func (p deliveryProvisioning) shutdownHTTPServer(ctx context.Context) error {
	if p.runtime == nil || p.runtime.httpServer == nil {
		return nil
	}
	return p.runtime.httpServer.Shutdown(ctx)
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
