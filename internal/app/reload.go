package app

import (
	"context"
	"fmt"

	"github.com/rin721/go-scaffold2/internal/config"
)

func (a *App) registerReloadHooks() {
	a.configManager.RegisterReloadHook("app", a.reloadComponents)
}

func (a *App) startReload(ctx context.Context) error {
	return a.configManager.Start(ctx)
}

func (a *App) reloadComponents(_ context.Context, oldCfg, newCfg config.Config) error {
	if a.logger != nil {
		if err := a.logger.Reload(toLoggerConfig(newCfg.Logger)); err != nil {
			return fmt.Errorf("reload logger: %w", err)
		}
	}
	if a.cache != nil {
		if err := a.cache.Reload(toCacheConfig(newCfg.Redis)); err != nil {
			return fmt.Errorf("reload cache: %w", err)
		}
	}
	if a.database != nil {
		if !newCfg.Database.Enabled {
			return fmt.Errorf("hot reload does not support disabling database once initialized")
		}
		if err := a.database.Reload(toDatabaseConfig(newCfg.Database)); err != nil {
			return fmt.Errorf("reload database: %w", err)
		}
	}
	if a.executor != nil {
		if !newCfg.Executor.Enabled {
			return fmt.Errorf("hot reload does not support disabling executor once initialized")
		}
		if err := a.executor.Reload(toExecutorConfig(newCfg.Executor)); err != nil {
			return fmt.Errorf("reload executor: %w", err)
		}
		if a.logger != nil {
			a.logger.SetExecutor(a.executor)
		}
		if a.httpServer != nil {
			a.httpServer.SetExecutor(a.executor)
		}
	}
	if a.httpServer != nil {
		if err := a.httpServer.Reload(toHTTPServerConfig(newCfg.Server)); err != nil {
			return fmt.Errorf("reload http server: %w", err)
		}
	}
	if a.storage != nil {
		if err := a.storage.Reload(toStorageConfig(newCfg.Storage)); err != nil {
			return fmt.Errorf("reload storage: %w", err)
		}
	}

	a.cfg = newCfg.Clone()
	_ = oldCfg
	return nil
}
