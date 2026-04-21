package app

import (
	"context"
	"fmt"

	pkghttpserver "github.com/rin721/rei/pkg/httpserver"
)

func (a *App) initHTTPServer() error {
	return a.deliveryProvisioning().initHTTPServer()
}

func (p deliveryProvisioning) initHTTPServer() error {
	if p.runtime != nil && p.runtime.httpServer != nil {
		return nil
	}

	server := pkghttpserver.New(p.serverConfig, p.HTTPHandler())
	if p.executor != nil {
		server.SetExecutor(p.executor)
	}
	p.runtime.httpServer = server
	if p.logger != nil {
		p.logger.Info(fmt.Sprintf("http server prepared on %s", p.serverConfig.Address))
	}

	return nil
}

func (p deliveryProvisioning) start(ctx context.Context) error {
	starters, err := p.lifecycle().starters(deliveryCapabilityProfileRuntime)
	if err != nil {
		return err
	}
	return runRuntimeStarters(ctx, "start delivery runtime", starters)
}

func (p deliveryProvisioning) startHTTPServer(ctx context.Context) error {
	_ = ctx
	if p.runtime == nil || p.runtime.httpServer == nil {
		return fmt.Errorf("http server is not initialized")
	}
	return p.runtime.httpServer.Start()
}
