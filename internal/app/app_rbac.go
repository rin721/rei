package app

import (
	"fmt"

	pkgrbac "github.com/rin721/go-scaffold2/pkg/rbac"
)

func (a *App) initRBAC() error {
	if a.rbac != nil || !a.cfg.RBAC.Enabled {
		return nil
	}

	cfg, err := toRBACConfig(a.cfg.RBAC)
	if err != nil {
		return fmt.Errorf("prepare rbac config: %w", err)
	}

	manager, err := pkgrbac.New(cfg)
	if err != nil {
		return fmt.Errorf("init rbac: %w", err)
	}

	a.rbac = manager
	return nil
}
