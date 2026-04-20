package app

import "github.com/rin721/rei/internal/config"

type infrastructureProvisioning struct {
	configManager *config.Manager
	cfg           config.Config
	infra         *infrastructureRuntime
	delivery      *deliveryRuntime
}

func (a *App) infrastructureProvisioning() infrastructureProvisioning {
	return a.infrastructureProvisioningWithConfig(a.cfg)
}

func (a *App) infrastructureProvisioningWithConfig(cfg config.Config) infrastructureProvisioning {
	return infrastructureProvisioning{
		configManager: a.configManager,
		cfg:           cfg.Clone(),
		infra:         &a.infra,
		delivery:      &a.delivery,
	}
}
