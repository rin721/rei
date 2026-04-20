package app

import (
	"fmt"

	pkglogger "github.com/rin721/rei/pkg/logger"
)

func (p infrastructureProvisioning) initLogger() error {
	if p.infra.logger != nil {
		return nil
	}

	log, err := pkglogger.New(toLoggerConfig(p.cfg.Logger))
	if err != nil {
		return fmt.Errorf("init logger: %w", err)
	}

	p.infra.logger = log
	return nil
}
