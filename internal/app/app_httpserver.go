package app

import (
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
