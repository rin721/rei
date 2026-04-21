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

func (p infrastructureProvisioning) bootstrapServer(ctx context.Context) error {
	steps, err := p.capabilities().bootstrapSteps(infrastructureProfileServerBootstrap, p)
	if err != nil {
		return err
	}
	return runBootstrap(ctx, "bootstrap server infrastructure", steps)
}

func (p infrastructureProvisioning) bootstrapDB(ctx context.Context) error {
	steps, err := p.capabilities().bootstrapSteps(infrastructureProfileDBBootstrap, p)
	if err != nil {
		return err
	}
	return runBootstrap(ctx, "bootstrap db infrastructure", steps)
}
