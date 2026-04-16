package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/rin721/go-scaffold2/internal/config"
	"github.com/rin721/go-scaffold2/internal/handler"
	pkgcache "github.com/rin721/go-scaffold2/pkg/cache"
	pkgcrypto "github.com/rin721/go-scaffold2/pkg/crypto"
	pkgdatabase "github.com/rin721/go-scaffold2/pkg/database"
	pkgdbtx "github.com/rin721/go-scaffold2/pkg/dbtx"
	pkgexecutor "github.com/rin721/go-scaffold2/pkg/executor"
	pkghttpserver "github.com/rin721/go-scaffold2/pkg/httpserver"
	pkgi18n "github.com/rin721/go-scaffold2/pkg/i18n"
	pkgjwt "github.com/rin721/go-scaffold2/pkg/jwt"
	pkglogger "github.com/rin721/go-scaffold2/pkg/logger"
	pkgrbac "github.com/rin721/go-scaffold2/pkg/rbac"
	pkgstorage "github.com/rin721/go-scaffold2/pkg/storage"
	pkgutils "github.com/rin721/go-scaffold2/pkg/utils"
)

// Options 描述应用容器初始化选项。
type Options struct {
	Mode       Mode
	ConfigPath string
	DryRun     bool
}

// App 负责装配基础设施并管理不同运行模式。
type App struct {
	options       Options
	configManager *config.Manager
	cfg           config.Config

	logger       *pkglogger.Logger
	i18n         *pkgi18n.Manager
	idGen        *pkgutils.IDGenerator
	cache        *pkgcache.MemoryCache
	database     *pkgdatabase.Database
	dbtx         *pkgdbtx.Manager
	executor     *pkgexecutor.Manager
	crypto       *pkgcrypto.Service
	jwt          *pkgjwt.Manager
	storage      *pkgstorage.Storage
	rbac         *pkgrbac.Manager
	httpServer   *pkghttpserver.Server
	routerEngine http.Handler
	handlers     *handler.Bundle
}

// New 创建一个新的应用容器。
func New(options Options) (*App, error) {
	options = normalizeOptions(options)

	manager := config.NewManager(config.Options{
		Path:          options.ConfigPath,
		WatchInterval: defaultConfigWatchInterval,
	})

	cfg, err := manager.Load()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	return &App{
		options:       options,
		configManager: manager,
		cfg:           cfg,
	}, nil
}

// Config 返回当前配置快照。
func (a *App) Config() config.Config {
	return a.cfg.Clone()
}

// Shutdown 按稳定顺序释放资源。
func (a *App) Shutdown(ctx context.Context) error {
	var errs []error
	if err := a.configManager.Stop(); err != nil {
		errs = append(errs, err)
	}

	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), defaultShutdownTimeout)
		defer cancel()
	}

	if a.httpServer != nil {
		if err := a.httpServer.Shutdown(ctx); err != nil {
			errs = append(errs, err)
		}
	}
	if a.storage != nil {
		if err := a.storage.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if a.executor != nil {
		if err := a.executor.Shutdown(ctx); err != nil {
			errs = append(errs, err)
		}
	}
	if a.cache != nil {
		if err := a.cache.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if a.database != nil {
		if err := a.database.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if a.rbac != nil {
		if err := a.rbac.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if a.logger != nil {
		if err := a.logger.Sync(); err != nil {
			errs = append(errs, err)
		}
	}

	return joinErrors(errs...)
}

func normalizeOptions(options Options) Options {
	if options.Mode == "" {
		options.Mode = ModeServer
	}
	if options.ConfigPath == "" {
		options.ConfigPath = resolveConfigPath()
	}
	return options
}

func joinErrors(errs ...error) error {
	filtered := make([]error, 0, len(errs))
	for _, err := range errs {
		if err != nil {
			filtered = append(filtered, err)
		}
	}
	if len(filtered) == 0 {
		return nil
	}
	return fmt.Errorf("%v", filtered)
}
