package middleware

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rin721/go-scaffold2/types/constants"
)

var traceCounter atomic.Uint64

// TraceID 为每个请求注入链路追踪 ID。
func TraceID(cfg MiddlewareConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := cfg.TraceHeader
		if header == "" {
			header = constants.HeaderTraceID
		}

		traceID := c.GetHeader(header)
		if traceID == "" {
			traceID = fmt.Sprintf("%d-%d", time.Now().UTC().UnixNano(), traceCounter.Add(1))
		}

		c.Set(constants.ContextKeyTraceID, traceID)
		c.Header(header, traceID)
		c.Next()
	}
}
