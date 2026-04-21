package app

import (
	"context"
	"fmt"
)

func (a *App) initBusiness() error {
	return a.businessProvisioning().bootstrap(context.Background())
}

func (p businessProvisioning) bootstrap(ctx context.Context) error {
	steps, err := p.bootstrapSteps()
	if err != nil {
		return err
	}
	return runBootstrap(ctx, "bootstrap business runtime", steps)
}

func (p businessProvisioning) bootstrapSteps() ([]bootstrapStep, error) {
	return p.lifecycle().bootstrapSteps(businessCapabilityProfileRuntime)
}

func (p businessProvisioning) initRuntime() error {
	if p.runtime != nil && p.runtime.handlers != nil {
		return nil
	}
	if err := p.Validate(); err != nil {
		return fmt.Errorf("init business: %w", err)
	}

	repos := p.RepositorySet()
	if err := p.seed(context.Background(), repos); err != nil {
		return fmt.Errorf("seed business data: %w", err)
	}

	modules, err := p.provideModules(repos)
	if err != nil {
		return err
	}

	p.runtime.handlers = modules.Handlers()
	return nil
}
