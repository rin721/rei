package app

import (
	"fmt"

	pkgdbtx "github.com/rin721/go-scaffold2/pkg/dbtx"
)

func (a *App) initDBTx() error {
	if a.dbtx != nil || a.database == nil {
		return nil
	}

	manager, err := pkgdbtx.New(a.database.DB())
	if err != nil {
		return fmt.Errorf("init dbtx: %w", err)
	}

	a.dbtx = manager
	return nil
}
