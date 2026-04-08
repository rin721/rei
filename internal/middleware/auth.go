package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rei0721/go-scaffold2/types/constants"
	apperrors "github.com/rei0721/go-scaffold2/types/errors"
)

// Auth 校验 Bearer Token，并将用户身份写入上下文。
func Auth(cfg MiddlewareConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if cfg.JWT == nil {
			writeFailure(c, http.StatusInternalServerError, apperrors.Internal("jwt manager unavailable"))
			return
		}

		authHeader := c.GetHeader(authorizationHeader)
		if !strings.HasPrefix(authHeader, bearerPrefix) {
			writeFailure(c, http.StatusUnauthorized, apperrors.Unauthorized("missing bearer token"))
			return
		}

		token := strings.TrimSpace(strings.TrimPrefix(authHeader, bearerPrefix))
		claims, err := cfg.JWT.ValidateToken(token)
		if err != nil {
			writeFailure(c, http.StatusUnauthorized, apperrors.Unauthorized("invalid token"))
			return
		}

		c.Set(constants.ContextKeyUserID, claims.Subject)
		c.Set(constants.ContextKeyJWTClaims, claims)
		c.Next()
	}
}
