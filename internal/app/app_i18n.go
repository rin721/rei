package app

import (
	"fmt"

	pkgi18n "github.com/rin721/go-scaffold2/pkg/i18n"
)

func (a *App) initI18n() error {
	if a.i18n != nil {
		return nil
	}

	manager, err := pkgi18n.New(toI18nConfig(a.cfg.I18n))
	if err != nil {
		return fmt.Errorf("init i18n: %w", err)
	}

	a.i18n = manager
	return nil
}
