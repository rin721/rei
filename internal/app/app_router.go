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
	return runBootstrap(ctx, "bootstrap delivery runtime", p.bootstrapSteps())
}

func (p deliveryProvisioning) bootstrapSteps() []bootstrapStep {
	return []bootstrapStep{
		newBootstrapTask("router", p.initRouter),
		newBootstrapTask("http server", p.initHTTPServer),
	}
}

func (p deliveryProvisioning) initRouter() error {
	if p.runtime != nil && p.runtime.routerEngine != nil {
		return nil
	}

	engine := router.New(p.business.handlers).Setup(p.MiddlewareConfig())
	p.runtime.routerEngine = engine
	return nil
}
