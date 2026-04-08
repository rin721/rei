package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger 记录基础请求日志。
func Logger(cfg MiddlewareConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		if cfg.Logger == nil {
			return
		}

		cfg.Logger.Info(fmt.Sprintf(
			"method=%s path=%s status=%d duration=%s trace_id=%s",
			c.Request.Method,
			c.FullPath(),
			c.Writer.Status(),
			time.Since(start).String(),
			traceIDFromContext(c),
		))
	}
}
