package app

import (
	"fmt"

	pkgcrypto "github.com/rin721/rei/pkg/crypto"
)

func (p infrastructureProvisioning) initCrypto() error {
	if p.infra.crypto != nil {
		return nil
	}

	service, err := pkgcrypto.New(pkgcrypto.Config{})
	if err != nil {
		return fmt.Errorf("init crypto: %w", err)
	}

	p.infra.crypto = service
	return nil
}
