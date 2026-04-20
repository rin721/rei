package app

import (
	"fmt"

	pkgjwt "github.com/rin721/rei/pkg/jwt"
)

func (p infrastructureProvisioning) initJWT() error {
	if p.infra.jwt != nil || !p.cfg.JWT.Enabled {
		return nil
	}

	manager, err := pkgjwt.New(toJWTConfig(p.cfg.JWT))
	if err != nil {
		return fmt.Errorf("init jwt: %w", err)
	}

	p.infra.jwt = manager
	return nil
}
