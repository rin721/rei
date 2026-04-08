package errors

import (
	stderrors "errors"
	"fmt"
)

// 基础错误码约定。
const (
	// CodeOK 表示成功。
	CodeOK = 0
	// CodeBadRequest 表示参数错误。
	CodeBadRequest = 1000
	// CodeUnauthorized 表示未认证。
	CodeUnauthorized = 1001
	// CodeForbidden 表示无权限。
	CodeForbidden = 1002
	// CodeNotFound 表示资源不存在。
	CodeNotFound = 1003
	// CodeNotImplemented 表示接口已保留但尚未实现完整业务。
	CodeNotImplemented = 1004
	// CodeInternal 表示内部错误。
	CodeInternal = 1500
)

const (
	defaultSuccessMessage = "success"
	defaultInternalMsg    = "internal error"
)

// AppError 是统一的业务错误结构。
type AppError struct {
	Code    int
	Message string
	Err     error
}

// Error 返回适合日志与 CLI 输出的错误文本。
func (e *AppError) Error() string {
	if e == nil {
		return defaultInternalMsg
	}

	if e.Err == nil {
		return e.Message
	}

	return fmt.Sprintf("%s: %v", e.Message, e.Err)
}

// Unwrap 暴露底层错误，便于 errors.Is / errors.As 链式匹配。
func (e *AppError) Unwrap() error {
	if e == nil {
		return nil
	}

	return e.Err
}

// New 创建一个不带底层错误的业务错误。
func New(code int, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: normalizeMessage(code, message),
	}
}

// Wrap 创建一个带上下文信息的业务错误。
func Wrap(code int, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: normalizeMessage(code, message),
		Err:     err,
	}
}

// BadRequest 返回参数错误。
func BadRequest(message string) *AppError {
	return New(CodeBadRequest, message)
}

// Unauthorized 返回未认证错误。
func Unauthorized(message string) *AppError {
	return New(CodeUnauthorized, message)
}

// Forbidden 返回无权限错误。
func Forbidden(message string) *AppError {
	return New(CodeForbidden, message)
}

// NotFound 返回资源不存在错误。
func NotFound(message string) *AppError {
	return New(CodeNotFound, message)
}

// NotImplemented 返回稳定的未实现错误。
func NotImplemented(message string) *AppError {
	return New(CodeNotImplemented, message)
}

// Internal 返回内部错误。
func Internal(message string) *AppError {
	return New(CodeInternal, message)
}

// CodeOf 解析错误链并返回统一错误码。
func CodeOf(err error) int {
	if err == nil {
		return CodeOK
	}

	var target *AppError
	if stderrors.As(err, &target) {
		return target.Code
	}

	return CodeInternal
}

// MessageOf 解析错误链并返回面向客户端的消息。
func MessageOf(err error) string {
	if err == nil {
		return defaultSuccessMessage
	}

	var target *AppError
	if stderrors.As(err, &target) {
		return normalizeMessage(target.Code, target.Message)
	}

	if err.Error() == "" {
		return defaultInternalMsg
	}

	return err.Error()
}

// IsCode 判断错误链是否匹配指定错误码。
func IsCode(err error, code int) bool {
	return CodeOf(err) == code
}

func normalizeMessage(code int, message string) string {
	if message != "" {
		return message
	}

	switch code {
	case CodeOK:
		return defaultSuccessMessage
	case CodeBadRequest:
		return "bad request"
	case CodeUnauthorized:
		return "unauthorized"
	case CodeForbidden:
		return "forbidden"
	case CodeNotFound:
		return "not found"
	case CodeNotImplemented:
		return "not implemented"
	default:
		return defaultInternalMsg
	}
}
