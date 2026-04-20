package app

import (
	"fmt"
	"net/http"

	"github.com/rin721/rei/internal/config"
	"github.com/rin721/rei/internal/handler"
	pkgcache "github.com/rin721/rei/pkg/cache"
	pkgcrypto "github.com/rin721/rei/pkg/crypto"
	pkgdatabase "github.com/rin721/rei/pkg/database"
	pkgdbtx "github.com/rin721/rei/pkg/dbtx"
	pkgexecutor "github.com/rin721/rei/pkg/executor"
	pkghttpserver "github.com/rin721/rei/pkg/httpserver"
	pkgi18n "github.com/rin721/rei/pkg/i18n"
	pkgjwt "github.com/rin721/rei/pkg/jwt"
	pkglogger "github.com/rin721/rei/pkg/logger"
	pkgrbac "github.com/rin721/rei/pkg/rbac"
	pkgstorage "github.com/rin721/rei/pkg/storage"
	pkgutils "github.com/rin721/rei/pkg/utils"
)

type Options struct {
	Mode       Mode
	ConfigPath string
	DryRun     bool
	DBOptions  DBOptions
}

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

func (a *App) Config() config.Config {
	return a.cfg.Clone()
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
