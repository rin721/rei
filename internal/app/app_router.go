package app

import (
	"net/http"

	"github.com/rin721/rei/internal/middleware"
	"github.com/rin721/rei/internal/router"
	"github.com/rin721/rei/types/constants"
)

func (a *App) initRouter() error {
	if a.routerEngine != nil {
		return nil
	}

	engine := router.New(a.handlers).Setup(toMiddlewareConfig(a))
	a.routerEngine = engine
	return nil
}

func (a *App) httpHandler() http.Handler {
	if a.routerEngine != nil {
		return a.routerEngine
	}
	return http.NotFoundHandler()
}

func toMiddlewareConfig(a *App) middleware.MiddlewareConfig {
	return middleware.MiddlewareConfig{
		AppName: constants.ApplicationName,
		Logger:  a.logger,
		I18n:    a.i18n,
		JWT:     a.jwt,
		RBAC:    a.rbac,
		CORS: middleware.CORSConfig{
			Enabled:      a.cfg.CORS.Enabled,
			AllowOrigins: append([]string(nil), a.cfg.CORS.AllowOrigins...),
			AllowMethods: append([]string(nil), a.cfg.CORS.AllowMethods...),
			AllowHeaders: append([]string(nil), a.cfg.CORS.AllowHeaders...),
		},
		TraceHeader:  "X-Trace-ID",
		LocaleHeader: "Accept-Language",
	}
}
