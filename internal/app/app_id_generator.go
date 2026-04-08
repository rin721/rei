package app

import (
	"fmt"

	pkgutils "github.com/rei0721/go-scaffold2/pkg/utils"
)

func (a *App) initIDGenerator() error {
	if a.idGen != nil {
		return nil
	}

	generator, err := pkgutils.NewIDGenerator(1)
	if err != nil {
		return fmt.Errorf("init id generator: %w", err)
	}

	a.idGen = generator
	return nil
}
