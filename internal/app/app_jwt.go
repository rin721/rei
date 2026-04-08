package app

import (
	"fmt"

	pkgjwt "github.com/rei0721/go-scaffold2/pkg/jwt"
)

func (a *App) initJWT() error {
	if a.jwt != nil || !a.cfg.JWT.Enabled {
		return nil
	}

	manager, err := pkgjwt.New(toJWTConfig(a.cfg.JWT))
	if err != nil {
		return fmt.Errorf("init jwt: %w", err)
	}

	a.jwt = manager
	return nil
}
