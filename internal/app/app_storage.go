package app

import (
	"fmt"

	pkgstorage "github.com/rin721/rei/pkg/storage"
)

func (p infrastructureProvisioning) initStorage() error {
	if p.infra.storage != nil {
		return nil
	}

	store, err := pkgstorage.New(toStorageConfig(p.cfg.Storage))
	if err != nil {
		return fmt.Errorf("init storage: %w", err)
	}

	p.infra.storage = store
	return nil
}
