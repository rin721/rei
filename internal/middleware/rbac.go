package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rin721/go-scaffold2/types/constants"
	apperrors "github.com/rin721/go-scaffold2/types/errors"
)

// PermissionResolver 用于从请求中解析权限对象和动作。
type PermissionResolver func(*gin.Context) (object string, action string)

// RBAC 执行权限校验。
func RBAC(cfg MiddlewareConfig, resolver PermissionResolver) gin.HandlerFunc {
	return func(c *gin.Context) {
		if cfg.RBAC == nil {
			writeFailure(c, http.StatusInternalServerError, apperrors.Internal("rbac manager unavailable"))
			return
		}

		subject := c.GetString(constants.ContextKeyUserID)
		if subject == "" {
			writeFailure(c, http.StatusUnauthorized, apperrors.Unauthorized("missing user identity"))
			return
		}

		object, action := resolvePermission(c, resolver)
		allowed, err := cfg.RBAC.CheckPermission(subject, object, action)
		if err != nil {
			writeFailure(c, http.StatusInternalServerError, apperrors.Internal("permission check failed"))
			return
		}
		if !allowed {
			writeFailure(c, http.StatusForbidden, apperrors.Forbidden("permission denied"))
			return
		}

		c.Next()
	}
}

func resolvePermission(c *gin.Context, resolver PermissionResolver) (string, string) {
	if resolver != nil {
		return resolver(c)
	}

	path := c.FullPath()
	if path == "" {
		path = c.Request.URL.Path
	}

	return path, strings.ToLower(c.Request.Method)
}
