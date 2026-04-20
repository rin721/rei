package app

import (
	"fmt"

	"github.com/rin721/rei/internal/config"
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

	infra    infrastructureRuntime
	business businessRuntime
	delivery deliveryRuntime
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
