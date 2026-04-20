package app

import (
	"net/http"

	"github.com/rin721/rei/internal/handler"
	pkgcache "github.com/rin721/rei/pkg/cache"
	pkgcrypto "github.com/rin721/rei/pkg/crypto"
	pkgdatabase "github.com/rin721/rei/pkg/database"
	pkgdbtx "github.com/rin721/rei/pkg/dbtx"
	pkgexecutor "github.com/rin721/rei/pkg/executor"
	pkghttpserver "github.com/rin721/rei/pkg/httpserver"
	pkgi18n "github.com/rin721/rei/pkg/i18n"
	pkgjwt "github.com/rin721/rei/pkg/jwt"
	pkglogger "github.com/rin721/rei/pkg/logger"
	pkgrbac "github.com/rin721/rei/pkg/rbac"
	pkgstorage "github.com/rin721/rei/pkg/storage"
	pkgutils "github.com/rin721/rei/pkg/utils"
)

type infrastructureRuntime struct {
	logger   pkglogger.Logger
	i18n     pkgi18n.I18n
	idGen    *pkgutils.IDGenerator
	cache    *pkgcache.MemoryCache
	database *pkgdatabase.Database
	dbtx     *pkgdbtx.Manager
	executor pkgexecutor.Manager
	crypto   *pkgcrypto.Service
	jwt      *pkgjwt.Manager
	storage  pkgstorage.Storage
	rbac     *pkgrbac.Manager
}

type businessRuntime struct {
	handlers *handler.Bundle
}

type deliveryRuntime struct {
	httpServer   *pkghttpserver.Server
	routerEngine http.Handler
}
