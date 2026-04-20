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
	return runBootstrap(ctx, "bootstrap server infrastructure", []bootstrapStep{
		newBootstrapTask("logger", a.initLogger),
		newBootstrapTask("i18n", a.initI18n),
		newBootstrapTask("id generator", a.initIDGenerator),
		newBootstrapTask("cache", a.initCache),
		newBootstrapStep("database", a.initDatabase),
		newBootstrapTask("database transaction manager", a.initDBTx),
		newBootstrapTask("executor", a.initExecutor),
		newBootstrapTask("crypto", a.initCrypto),
		newBootstrapTask("jwt", a.initJWT),
		newBootstrapTask("storage", a.initStorage),
		newBootstrapTask("rbac", a.initRBAC),
	})
}

func (a *App) bootstrapBusinessRuntime(ctx context.Context) error {
	return runBootstrap(ctx, "bootstrap business runtime", []bootstrapStep{
		newBootstrapTask("business modules", a.initBusiness),
	})
}

func (a *App) bootstrapDeliveryRuntime(ctx context.Context) error {
	return runBootstrap(ctx, "bootstrap delivery runtime", []bootstrapStep{
		newBootstrapTask("router", a.initRouter),
		newBootstrapTask("http server", a.initHTTPServer),
	})
}

func (a *App) bootstrapDBInfrastructure(ctx context.Context) error {
	return runBootstrap(ctx, "bootstrap db infrastructure", []bootstrapStep{
		newBootstrapTask("logger", a.initLogger),
		newBootstrapStep("database", a.initDatabase),
	})
}
