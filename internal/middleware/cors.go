package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORS 处理跨域响应头与预检请求。
func CORS(cfg MiddlewareConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !cfg.CORS.Enabled {
			c.Next()
			return
		}

		if len(cfg.CORS.AllowOrigins) > 0 {
			c.Header("Access-Control-Allow-Origin", strings.Join(cfg.CORS.AllowOrigins, ","))
		}
		if len(cfg.CORS.AllowMethods) > 0 {
			c.Header("Access-Control-Allow-Methods", strings.Join(cfg.CORS.AllowMethods, ","))
		}
		if len(cfg.CORS.AllowHeaders) > 0 {
			c.Header("Access-Control-Allow-Headers", strings.Join(cfg.CORS.AllowHeaders, ","))
		}

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
