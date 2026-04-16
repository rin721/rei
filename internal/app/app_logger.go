package app

import (
	"fmt"

	pkglogger "github.com/rin721/rei/pkg/logger"
)

func (a *App) initLogger() error {
	if a.logger != nil {
		return nil
	}

	log, err := pkglogger.New(toLoggerConfig(a.cfg.Logger))
	if err != nil {
		return fmt.Errorf("init logger: %w", err)
	}

	a.logger = log
	return nil
}
