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

func (a *App) registerReloadHooks() {
	a.configManager.RegisterReloadHook("app", a.reloadComponents)
}

func (a *App) startReload(ctx context.Context) error {
	return a.configManager.Start(ctx)
}

func (a *App) reloadComponents(ctx context.Context, oldCfg, newCfg config.Config) error {
	if err := a.infrastructureProvisioningWithConfig(newCfg).reload(ctx); err != nil {
		return err
	}

	a.cfg = newCfg.Clone()
	_ = oldCfg
	return nil
}

func (p infrastructureProvisioning) reload(ctx context.Context) error {
	return runRuntimeReloaders(ctx, "reload runtime components", p.runtimeReloaders())
}

func (p infrastructureProvisioning) runtimeReloaders() []runtimeReloader {
	return []runtimeReloader{
		newRuntimeReloader("logger", func(_ context.Context) error {
			return p.reloadLogger()
		}),
		newRuntimeReloader("cache", func(_ context.Context) error {
			return p.reloadCache()
		}),
		newRuntimeReloader("database", func(_ context.Context) error {
			return p.reloadDatabase()
		}),
		newRuntimeReloader("executor", func(_ context.Context) error {
			return p.reloadExecutor()
		}),
		newRuntimeReloader("http server", func(_ context.Context) error {
			return p.reloadHTTPServer()
		}),
		newRuntimeReloader("storage", func(_ context.Context) error {
			return p.reloadStorage()
		}),
	}
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

func (p infrastructureProvisioning) reloadHTTPServer() error {
	if p.delivery.httpServer == nil {
		return nil
	}
	return p.delivery.httpServer.Reload(toHTTPServerConfig(p.cfg.Server))
}

func (p infrastructureProvisioning) reloadStorage() error {
	if p.infra.storage == nil {
		return nil
	}
	return p.infra.storage.Reload(context.Background(), toStorageConfig(p.cfg.Storage))
}
