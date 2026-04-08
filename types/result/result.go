package result

import (
	"time"

	apperrors "github.com/rei0721/go-scaffold2/types/errors"
)

// Envelope 定义统一 API 响应外层结构。
type Envelope struct {
	Code       int    `json:"code"`
	Message    string `json:"message"`
	Data       any    `json:"data,omitempty"`
	TraceID    string `json:"traceId,omitempty"`
	ServerTime int64  `json:"serverTime"`
}

// New 构造一个带明确字段的响应 envelope。
func New(code int, message string, data any, traceID string, now time.Time) Envelope {
	now = normalizeTime(now)

	return Envelope{
		Code:       code,
		Message:    message,
		Data:       data,
		TraceID:    traceID,
		ServerTime: now.Unix(),
	}
}

// Success 构造统一成功响应。
func Success(data any, traceID string, now time.Time) Envelope {
	return New(apperrors.CodeOK, apperrors.MessageOf(nil), data, traceID, now)
}

// Failure 构造统一失败响应。
func Failure(err error, traceID string, now time.Time) Envelope {
	return New(apperrors.CodeOf(err), apperrors.MessageOf(err), nil, traceID, now)
}

func normalizeTime(now time.Time) time.Time {
	if now.IsZero() {
		return time.Now().UTC()
	}

	return now.UTC()
}
