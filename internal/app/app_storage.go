package app

import (
	"fmt"

	pkgstorage "github.com/rin721/rei/pkg/storage"
)

func (a *App) initStorage() error {
	if a.storage != nil {
		return nil
	}

	store, err := pkgstorage.New(toStorageConfig(a.cfg.Storage))
	if err != nil {
		return fmt.Errorf("init storage: %w", err)
	}

	a.storage = store
	return nil
}
