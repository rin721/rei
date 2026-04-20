package app

import (
	"fmt"

	pkgrbac "github.com/rin721/rei/pkg/rbac"
)

func (p infrastructureProvisioning) initRBAC() error {
	if p.infra.rbac != nil || !p.cfg.RBAC.Enabled {
		return nil
	}

	cfg, err := toRBACConfig(p.cfg.RBAC)
	if err != nil {
		return fmt.Errorf("prepare rbac config: %w", err)
	}

	manager, err := pkgrbac.New(cfg)
	if err != nil {
		return fmt.Errorf("init rbac: %w", err)
	}

	p.infra.rbac = manager
	return nil
}
