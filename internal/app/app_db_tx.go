package app

import (
	"fmt"

	pkgdbtx "github.com/rin721/rei/pkg/dbtx"
)

func (p infrastructureProvisioning) initDBTx() error {
	if p.infra.dbtx != nil || p.infra.database == nil {
		return nil
	}

	manager, err := pkgdbtx.New(p.infra.database.DB())
	if err != nil {
		return fmt.Errorf("init dbtx: %w", err)
	}

	p.infra.dbtx = manager
	return nil
}
