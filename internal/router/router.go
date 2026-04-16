package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rin721/go-scaffold2/internal/handler"
	"github.com/rin721/go-scaffold2/internal/middleware"
	"github.com/rin721/go-scaffold2/types"
	"github.com/rin721/go-scaffold2/types/constants"
	apperrors "github.com/rin721/go-scaffold2/types/errors"
)

// Router 负责集中装配 Gin 路由与中间件。
type Router struct {
	handlers *handler.Bundle
}

// New 创建 Router，并可选注入业务处理器集合。
func New(bundles ...*handler.Bundle) *Router {
	var bundle *handler.Bundle
	if len(bundles) > 0 {
		bundle = bundles[0]
	}

	return &Router{handlers: bundle}
}

// Setup 创建 Gin Engine，并按约定顺序注册中间件与业务路由。
func (r *Router) Setup(cfg middleware.MiddlewareConfig) *gin.Engine {
	bundle := r.handlers
	if bundle == nil {
		bundle = &handler.Bundle{}
	}

	engine := gin.New()
	engine.Use(
		middleware.I18n(cfg),
		middleware.TraceID(cfg),
		middleware.CORS(cfg),
		middleware.Logger(cfg),
		middleware.Recovery(cfg),
	)

	engine.GET(RouteHealth, func(c *gin.Context) {
		middlewareWriteSuccess(c, http.StatusOK, types.HealthResponse{
			Status:  "ok",
			Service: cfg.AppName,
			Stage:   constants.PhaseTag,
		})
	})

	api := engine.Group("/api/v1")
	api.GET("/samples", wrapHandler(bundle.Sample, (*handler.SampleHandler).List))

	authGroup := api.Group("/auth")
	authGroup.POST("/register", wrapHandler(bundle.Auth, (*handler.AuthHandler).Register))
	authGroup.POST("/login", wrapHandler(bundle.Auth, (*handler.AuthHandler).Login))
	authGroup.POST("/refresh", wrapHandler(bundle.Auth, (*handler.AuthHandler).Refresh))
	authProtected := authGroup.Group("")
	authProtected.Use(middleware.Auth(cfg))
	authProtected.POST("/logout", wrapHandler(bundle.Auth, (*handler.AuthHandler).Logout))
	authProtected.POST("/change-password", wrapHandler(bundle.Auth, (*handler.AuthHandler).ChangePassword))

	userGroup := api.Group("/users")
	userGroup.Use(middleware.Auth(cfg), middleware.RBAC(cfg, nil))
	userGroup.GET("/me", wrapHandler(bundle.User, (*handler.UserHandler).GetMe))
	userGroup.PUT("/me", wrapHandler(bundle.User, (*handler.UserHandler).UpdateMe))

	rbacGroup := api.Group("/rbac")
	rbacGroup.Use(middleware.Auth(cfg), middleware.RBAC(cfg, nil))
	rbacGroup.GET("/check", wrapHandler(bundle.RBAC, (*handler.RBACHandler).Check))
	rbacGroup.POST("/roles/assign", wrapHandler(bundle.RBAC, (*handler.RBACHandler).AssignRole))
	rbacGroup.POST("/roles/revoke", wrapHandler(bundle.RBAC, (*handler.RBACHandler).RevokeRole))
	rbacGroup.GET("/users/:user_id/roles", wrapHandler(bundle.RBAC, (*handler.RBACHandler).GetUserRoles))
	rbacGroup.GET("/roles/:role/users", wrapHandler(bundle.RBAC, (*handler.RBACHandler).GetRoleUsers))
	rbacGroup.POST("/policies", wrapHandler(bundle.RBAC, (*handler.RBACHandler).AddPolicy))
	rbacGroup.DELETE("/policies", wrapHandler(bundle.RBAC, (*handler.RBACHandler).RemovePolicy))
	rbacGroup.GET("/policies", wrapHandler(bundle.RBAC, (*handler.RBACHandler).ListPolicies))

	return engine
}

func wrapHandler[T any](value T, fn func(T, *gin.Context)) gin.HandlerFunc {
	return func(c *gin.Context) {
		var zero T
		if any(value) == any(zero) {
			err := apperrors.Internal("handler is not initialized")
			middlewareWriteFailure(c, middlewareStatusFromError(err), err)
			return
		}

		fn(value, c)
	}
}
