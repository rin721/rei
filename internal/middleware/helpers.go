package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rin721/rei/types/constants"
	apperrors "github.com/rin721/rei/types/errors"
	"github.com/rin721/rei/types/result"
)

func writeSuccess(c *gin.Context, status int, data any) {
	c.JSON(status, result.Success(data, traceIDFromContext(c), time.Now()))
}

func writeFailure(c *gin.Context, status int, err error) {
	c.AbortWithStatusJSON(status, result.Failure(err, traceIDFromContext(c), time.Now()))
}

func statusFromError(err error) int {
	switch apperrors.CodeOf(err) {
	case apperrors.CodeBadRequest:
		return http.StatusBadRequest
	case apperrors.CodeUnauthorized:
		return http.StatusUnauthorized
	case apperrors.CodeForbidden:
		return http.StatusForbidden
	case apperrors.CodeNotFound:
		return http.StatusNotFound
	case apperrors.CodeNotImplemented:
		return http.StatusNotImplemented
	default:
		return http.StatusInternalServerError
	}
}

func traceIDFromContext(c *gin.Context) string {
	if traceID := c.GetString(constants.ContextKeyTraceID); traceID != "" {
		return traceID
	}
	return ""
}
