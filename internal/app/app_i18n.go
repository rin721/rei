package app

import (
	"fmt"

	pkgi18n "github.com/rin721/rei/pkg/i18n"
)

func (p infrastructureProvisioning) initI18n() error {
	if p.infra.i18n != nil {
		return nil
	}

	manager, err := pkgi18n.New(toI18nConfig(p.cfg.I18n))
	if err != nil {
		return fmt.Errorf("init i18n: %w", err)
	}

	p.infra.i18n = manager
	return nil
}
