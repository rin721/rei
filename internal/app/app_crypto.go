package app

import (
	"fmt"

	pkgcrypto "github.com/rin721/go-scaffold2/pkg/crypto"
)

func (a *App) initCrypto() error {
	if a.crypto != nil {
		return nil
	}

	service, err := pkgcrypto.New(pkgcrypto.Config{})
	if err != nil {
		return fmt.Errorf("init crypto: %w", err)
	}

	a.crypto = service
	return nil
}
