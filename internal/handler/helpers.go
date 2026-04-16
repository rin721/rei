package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rin721/go-scaffold2/types/constants"
	apperrors "github.com/rin721/go-scaffold2/types/errors"
	"github.com/rin721/go-scaffold2/types/result"
)

func writeSuccess(c *gin.Context, status int, data any) {
	c.JSON(status, result.Success(data, traceIDFromContext(c), time.Now()))
}

func writeFailure(c *gin.Context, status int, err error) {
	c.AbortWithStatusJSON(status, result.Failure(err, traceIDFromContext(c), time.Now()))
}

func writeBindFailure(c *gin.Context, err error) {
	writeFailure(c, http.StatusBadRequest, apperrors.BadRequest(err.Error()))
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
