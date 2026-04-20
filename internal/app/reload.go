package app

import (
	"context"
	"fmt"

	"github.com/rin721/rei/internal/config"
)

type runtimeReloader struct {
	name   string
	reload func(context.Context, config.Config) error
}

func newRuntimeReloader(name string, reload func(context.Context, config.Config) error) runtimeReloader {
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
	if err := runRuntimeReloaders(ctx, "reload runtime components", newCfg, a.runtimeReloaders()); err != nil {
		return err
	}

	a.cfg = newCfg.Clone()
	_ = oldCfg
	return nil
}

func (a *App) runtimeReloaders() []runtimeReloader {
	return []runtimeReloader{
		newRuntimeReloader("logger", func(_ context.Context, cfg config.Config) error {
			return a.reloadLogger(cfg)
		}),
		newRuntimeReloader("cache", func(_ context.Context, cfg config.Config) error {
			return a.reloadCache(cfg)
		}),
		newRuntimeReloader("database", func(_ context.Context, cfg config.Config) error {
			return a.reloadDatabase(cfg)
		}),
		newRuntimeReloader("executor", func(_ context.Context, cfg config.Config) error {
			return a.reloadExecutor(cfg)
		}),
		newRuntimeReloader("http server", func(_ context.Context, cfg config.Config) error {
			return a.reloadHTTPServer(cfg)
		}),
		newRuntimeReloader("storage", func(_ context.Context, cfg config.Config) error {
			return a.reloadStorage(cfg)
		}),
	}
}

func runRuntimeReloaders(ctx context.Context, phase string, cfg config.Config, reloaders []runtimeReloader) error {
	for _, reloader := range reloaders {
		if err := reloader.reload(ctx, cfg); err != nil {
			return fmt.Errorf("%s: %s: %w", phase, reloader.name, err)
		}
	}
	return nil
}

func (a *App) reloadLogger(newCfg config.Config) error {
	if a.logger == nil {
		return nil
	}
	return a.logger.Reload(toLoggerConfig(newCfg.Logger))
}

func (a *App) reloadCache(newCfg config.Config) error {
	if a.cache == nil {
		return nil
	}
	return a.cache.Reload(toCacheConfig(newCfg.Redis))
}

func (a *App) reloadDatabase(newCfg config.Config) error {
	if a.database == nil {
		return nil
	}
	if !newCfg.Database.Enabled {
		return fmt.Errorf("disabling database is not supported once initialized")
	}
	return a.database.Reload(toDatabaseConfig(newCfg.Database))
}

func (a *App) reloadExecutor(newCfg config.Config) error {
	if a.executor == nil {
		return nil
	}
	if !newCfg.Executor.Enabled {
		return fmt.Errorf("disabling executor is not supported once initialized")
	}
	if err := a.executor.Reload(toExecutorConfig(newCfg.Executor)); err != nil {
		return err
	}
	a.syncExecutorBindings()
	return nil
}

func (a *App) reloadHTTPServer(newCfg config.Config) error {
	if a.httpServer == nil {
		return nil
	}
	return a.httpServer.Reload(toHTTPServerConfig(newCfg.Server))
}

func (a *App) reloadStorage(newCfg config.Config) error {
	if a.storage == nil {
		return nil
	}
	return a.storage.Reload(toStorageConfig(newCfg.Storage))
}
