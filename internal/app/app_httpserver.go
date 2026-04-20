package app

import (
	"fmt"

	pkghttpserver "github.com/rin721/rei/pkg/httpserver"
)

func (a *App) initHTTPServer() error {
	if a.httpServer != nil {
		return nil
	}

	server := pkghttpserver.New(toHTTPServerConfig(a.cfg.Server), a.httpHandler())
	a.httpServer = server
	a.syncExecutorBindings()
	if a.logger != nil {
		a.logger.Info(fmt.Sprintf("http server prepared on %s", a.cfg.Server.Host))
	}

	return nil
}
