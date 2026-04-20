package app

import (
	"context"
	"fmt"
)

type bootstrapStep struct {
	name string
	run  func(context.Context) error
}

func newBootstrapStep(name string, run func(context.Context) error) bootstrapStep {
	return bootstrapStep{
		name: name,
		run:  run,
	}
}

func newBootstrapTask(name string, run func() error) bootstrapStep {
	return newBootstrapStep(name, func(context.Context) error {
		return run()
	})
}

func runBootstrap(ctx context.Context, phase string, steps []bootstrapStep) error {
	for _, step := range steps {
		if err := step.run(ctx); err != nil {
			return fmt.Errorf("%s: %s: %w", phase, step.name, err)
		}
	}
	return nil
}

func (a *App) bootstrapServerInfrastructure(ctx context.Context) error {
	return a.infrastructureProvisioning().bootstrapServer(ctx)
}

func (a *App) bootstrapBusinessRuntime(ctx context.Context) error {
	return a.businessProvisioning().bootstrap(ctx)
}

func (a *App) bootstrapDeliveryRuntime(ctx context.Context) error {
	return a.deliveryProvisioning().bootstrap(ctx)
}

func (a *App) bootstrapDBInfrastructure(ctx context.Context) error {
	return a.infrastructureProvisioning().bootstrapDB(ctx)
}

func (p infrastructureProvisioning) bootstrapServer(ctx context.Context) error {
	return runBootstrap(ctx, "bootstrap server infrastructure", p.serverBootstrapSteps())
}

func (p infrastructureProvisioning) serverBootstrapSteps() []bootstrapStep {
	return []bootstrapStep{
		newBootstrapTask("logger", p.initLogger),
		newBootstrapTask("i18n", p.initI18n),
		newBootstrapTask("id generator", p.initIDGenerator),
		newBootstrapTask("cache", p.initCache),
		newBootstrapStep("database", p.initDatabase),
		newBootstrapTask("database transaction manager", p.initDBTx),
		newBootstrapTask("executor", p.initExecutor),
		newBootstrapTask("crypto", p.initCrypto),
		newBootstrapTask("jwt", p.initJWT),
		newBootstrapTask("storage", p.initStorage),
		newBootstrapTask("rbac", p.initRBAC),
	}
}

func (p infrastructureProvisioning) bootstrapDB(ctx context.Context) error {
	return runBootstrap(ctx, "bootstrap db infrastructure", p.dbBootstrapSteps())
}

func (p infrastructureProvisioning) dbBootstrapSteps() []bootstrapStep {
	return []bootstrapStep{
		newBootstrapTask("logger", p.initLogger),
		newBootstrapStep("database", p.initDatabase),
	}
}
