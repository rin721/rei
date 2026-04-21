package app

import (
	"context"
	"fmt"

	"github.com/rin721/rei/internal/config"
)

type runtimeReloader struct {
	name   string
	reload func(context.Context) error
}

func newRuntimeReloader(name string, reload func(context.Context) error) runtimeReloader {
	return runtimeReloader{
		name:   name,
		reload: reload,
	}
}

func (p infrastructureProvisioning) registerReloadHooks(reload func(context.Context, config.Config, config.Config) error) {
	if p.configManager == nil {
		return
	}
	p.configManager.RegisterReloadHook("app", reload)
}

func (p infrastructureProvisioning) startReload(ctx context.Context) error {
	if p.configManager == nil {
		return nil
	}
	return p.configManager.Start(ctx)
}

func (a *App) reloadComponents(ctx context.Context, oldCfg, newCfg config.Config) error {
	if err := a.infrastructureProvisioningWithConfig(newCfg).reload(ctx); err != nil {
		return err
	}
	if err := a.deliveryProvisioningWithConfig(newCfg).reload(ctx); err != nil {
		return err
	}

	a.cfg = newCfg.Clone()
	_ = oldCfg
	return nil
}

func (p infrastructureProvisioning) reload(ctx context.Context) error {
	reloaders, err := p.runtimeReloaders()
	if err != nil {
		return err
	}
	return runRuntimeReloaders(ctx, "reload runtime components", reloaders)
}

func (p infrastructureProvisioning) runtimeReloaders() ([]runtimeReloader, error) {
	return p.capabilities().reloaders(infrastructureProfileRuntimeReload, p)
}

func runRuntimeReloaders(ctx context.Context, phase string, reloaders []runtimeReloader) error {
	for _, reloader := range reloaders {
		if err := reloader.reload(ctx); err != nil {
			return fmt.Errorf("%s: %s: %w", phase, reloader.name, err)
		}
	}
	return nil
}

func (p infrastructureProvisioning) reloadLogger() error {
	if p.infra.logger == nil {
		return nil
	}
	return p.infra.logger.Reload(toLoggerConfig(p.cfg.Logger))
}

func (p infrastructureProvisioning) reloadCache() error {
	if p.infra.cache == nil {
		return nil
	}
	return p.infra.cache.Reload(toCacheConfig(p.cfg.Redis))
}

func (p infrastructureProvisioning) reloadDatabase() error {
	if p.infra.database == nil {
		return nil
	}
	if !p.cfg.Database.Enabled {
		return fmt.Errorf("disabling database is not supported once initialized")
	}
	return p.infra.database.Reload(toDatabaseConfig(p.cfg.Database))
}

func (p infrastructureProvisioning) reloadExecutor() error {
	if p.infra.executor == nil {
		return nil
	}
	if !p.cfg.Executor.Enabled {
		return fmt.Errorf("disabling executor is not supported once initialized")
	}
	if err := p.infra.executor.Reload(toExecutorConfig(p.cfg.Executor)); err != nil {
		return err
	}
	p.syncExecutorBindings()
	return nil
}

func (p deliveryProvisioning) reload(ctx context.Context) error {
	reloaders, err := p.runtimeReloaders()
	if err != nil {
		return err
	}
	return runRuntimeReloaders(ctx, "reload delivery runtime", reloaders)
}

func (p deliveryProvisioning) runtimeReloaders() ([]runtimeReloader, error) {
	return p.lifecycle().reloaders(deliveryCapabilityProfileRuntime)
}

func (p deliveryProvisioning) reloadHTTPServer() error {
	if p.runtime == nil || p.runtime.httpServer == nil {
		return nil
	}
	return p.runtime.httpServer.Reload(p.serverConfig)
}

func (p infrastructureProvisioning) reloadStorage(ctx context.Context) error {
	if p.infra.storage == nil {
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	return p.infra.storage.Reload(ctx, toStorageConfig(p.cfg.Storage))
}
