package app

import (
	"context"
	"net/http"

	"github.com/rin721/rei/internal/router"
)

func (a *App) initRouter() error {
	return a.deliveryProvisioning().initRouter()
}

func (a *App) httpHandler() http.Handler {
	return a.deliveryProvisioning().HTTPHandler()
}

func (p deliveryProvisioning) bootstrap(ctx context.Context) error {
	steps, err := p.bootstrapSteps()
	if err != nil {
		return err
	}
	return runBootstrap(ctx, "bootstrap delivery runtime", steps)
}

func (p deliveryProvisioning) bootstrapSteps() ([]bootstrapStep, error) {
	return p.lifecycle().bootstrapSteps(deliveryCapabilityProfileRuntime)
}

func (p deliveryProvisioning) initRouter() error {
	if p.runtime != nil && p.runtime.routerEngine != nil {
		return nil
	}

	engine := router.New(p.business.handlers).Setup(p.MiddlewareConfig())
	p.runtime.routerEngine = engine
	return nil
}
