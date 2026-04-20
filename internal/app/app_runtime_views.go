package app

import (
	"fmt"
	"net/http"
	"time"

	"github.com/rin721/rei/internal/middleware"
	"github.com/rin721/rei/internal/repository"
	pkgcache "github.com/rin721/rei/pkg/cache"
	pkgcrypto "github.com/rin721/rei/pkg/crypto"
	pkgdatabase "github.com/rin721/rei/pkg/database"
	pkgdbtx "github.com/rin721/rei/pkg/dbtx"
	pkghttpserver "github.com/rin721/rei/pkg/httpserver"
	pkgi18n "github.com/rin721/rei/pkg/i18n"
	pkgjwt "github.com/rin721/rei/pkg/jwt"
	pkglogger "github.com/rin721/rei/pkg/logger"
	pkgrbac "github.com/rin721/rei/pkg/rbac"
	pkgutils "github.com/rin721/rei/pkg/utils"
	"github.com/rin721/rei/types/constants"
)

type businessProvisioning struct {
	runtime         *businessRuntime
	database        *pkgdatabase.Database
	dbtx            *pkgdbtx.Manager
	cache           *pkgcache.MemoryCache
	crypto          *pkgcrypto.Service
	jwt             *pkgjwt.Manager
	rbac            *pkgrbac.Manager
	idGen           *pkgutils.IDGenerator
	refreshTokenTTL time.Duration
}

func (a *App) businessProvisioning() businessProvisioning {
	return businessProvisioning{
		runtime:         &a.business,
		database:        a.infra.database,
		dbtx:            a.infra.dbtx,
		cache:           a.infra.cache,
		crypto:          a.infra.crypto,
		jwt:             a.infra.jwt,
		rbac:            a.infra.rbac,
		idGen:           a.infra.idGen,
		refreshTokenTTL: time.Duration(a.cfg.JWT.RefreshTokenTTLHours) * time.Hour,
	}
}

func (p businessProvisioning) Validate() error {
	if p.database == nil {
		return fmt.Errorf("database is required")
	}
	if p.dbtx == nil {
		return fmt.Errorf("dbtx is required")
	}
	if p.cache == nil {
		return fmt.Errorf("cache is required")
	}
	if p.crypto == nil {
		return fmt.Errorf("crypto service is required")
	}
	if p.jwt == nil {
		return fmt.Errorf("jwt manager is required")
	}
	if p.rbac == nil {
		return fmt.Errorf("rbac manager is required")
	}
	if p.idGen == nil {
		return fmt.Errorf("id generator is required")
	}
	return nil
}

func (p businessProvisioning) RepositorySet() *repository.Set {
	return repository.NewSet(p.database.DB(), p.dbtx)
}

type deliveryProvisioning struct {
	business     *businessRuntime
	runtime      *deliveryRuntime
	logger       pkglogger.Logger
	i18n         pkgi18n.I18n
	jwt          *pkgjwt.Manager
	rbac         *pkgrbac.Manager
	executor     pkghttpserver.AsyncSubmitter
	serverConfig pkghttpserver.Config
	cors         middleware.CORSConfig
}

func (a *App) deliveryProvisioning() deliveryProvisioning {
	return deliveryProvisioning{
		business:     &a.business,
		runtime:      &a.delivery,
		logger:       a.infra.logger,
		i18n:         a.infra.i18n,
		jwt:          a.infra.jwt,
		rbac:         a.infra.rbac,
		executor:     newExecutorAsyncSubmitter(a.infra.executor),
		serverConfig: toHTTPServerConfig(a.cfg.Server),
		cors: middleware.CORSConfig{
			Enabled:      a.cfg.CORS.Enabled,
			AllowOrigins: append([]string(nil), a.cfg.CORS.AllowOrigins...),
			AllowMethods: append([]string(nil), a.cfg.CORS.AllowMethods...),
			AllowHeaders: append([]string(nil), a.cfg.CORS.AllowHeaders...),
		},
	}
}

func (p deliveryProvisioning) HTTPHandler() http.Handler {
	if p.runtime != nil && p.runtime.routerEngine != nil {
		return p.runtime.routerEngine
	}
	return http.NotFoundHandler()
}

func (p deliveryProvisioning) MiddlewareConfig() middleware.MiddlewareConfig {
	return middleware.MiddlewareConfig{
		AppName:      constants.ApplicationName,
		Logger:       p.logger,
		I18n:         p.i18n,
		JWT:          p.jwt,
		RBAC:         p.rbac,
		CORS:         p.cors,
		TraceHeader:  "X-Trace-ID",
		LocaleHeader: "Accept-Language",
	}
}
