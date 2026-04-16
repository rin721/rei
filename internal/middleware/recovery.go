package middleware

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	apperrors "github.com/rin721/go-scaffold2/types/errors"
)

// Recovery 捕获 panic 并返回统一错误结构。
func Recovery(cfg MiddlewareConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if recovered := recover(); recovered != nil {
				if cfg.Logger != nil {
					cfg.Logger.Error(fmt.Sprintf("panic recovered: %v", recovered))
				}
				writeFailure(c, http.StatusInternalServerError, apperrors.Internal("internal error"))
			}
		}()

		c.Next()
	}
}
