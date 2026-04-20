package app

import (
	"fmt"

	pkgutils "github.com/rin721/rei/pkg/utils"
)

func (p infrastructureProvisioning) initIDGenerator() error {
	if p.infra.idGen != nil {
		return nil
	}

	generator, err := pkgutils.NewIDGenerator(1)
	if err != nil {
		return fmt.Errorf("init id generator: %w", err)
	}

	p.infra.idGen = generator
	return nil
}
