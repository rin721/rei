package app

import (
	"fmt"

	pkghttpserver "github.com/rin721/go-scaffold2/pkg/httpserver"
)

func (a *App) initHTTPServer() error {
	if a.httpServer != nil {
		return nil
	}

	server := pkghttpserver.New(toHTTPServerConfig(a.cfg.Server), a.httpHandler())
	if a.executor != nil {
		server.SetExecutor(a.executor)
	}

	a.httpServer = server
	if a.logger != nil {
		a.logger.Info(fmt.Sprintf("http server prepared on %s", a.cfg.Server.Host))
	}

	return nil
}
