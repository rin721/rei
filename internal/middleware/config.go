package middleware

import (
	pkgi18n "github.com/rin721/go-scaffold2/pkg/i18n"
	pkgjwt "github.com/rin721/go-scaffold2/pkg/jwt"
	pkglogger "github.com/rin721/go-scaffold2/pkg/logger"
	pkgrbac "github.com/rin721/go-scaffold2/pkg/rbac"
)

// CORSConfig 描述中间件层使用的跨域配置。
type CORSConfig struct {
	Enabled      bool
	AllowOrigins []string
	AllowMethods []string
	AllowHeaders []string
}

// MiddlewareConfig 描述路由层所需的中间件依赖。
type MiddlewareConfig struct {
	AppName      string
	Logger       *pkglogger.Logger
	I18n         *pkgi18n.Manager
	JWT          *pkgjwt.Manager
	RBAC         *pkgrbac.Manager
	CORS         CORSConfig
	TraceHeader  string
	LocaleHeader string
}
