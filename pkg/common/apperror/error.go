package apperror

import (
	"fmt"
	"net/http"
)

type ErrorType string

const (
	TypeNotFound           ErrorType = "NOT_FOUND"           // 资源不存在 (不可重试)
	TypeInvalidInput       ErrorType = "INVALID_INPUT"       // 参数错误 (不可重试)
	TypeConflict           ErrorType = "CONFLICT"            // 冲突，如库存不足 (通常不可重试，除非是并发锁)
	TypeInternal           ErrorType = "INTERNAL"            // 内部错误 (部分可重试)
	TypeServiceUnavailable ErrorType = "SERVICE_UNAVAILABLE" // 下游挂了 (可重试)
	TypeTimeout            ErrorType = "TIMEOUT"             // 超时 (可重试)
)

// AppError 是我们自定义的错误结构
type AppError struct {
	Type    ErrorType
	Code    int    // 具体的业务错误码，如 10001
	Message string // 错误描述
	Cause   error  // 原始错误
}

func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %d: %s | Cause: %v", e.Type, e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %d: %s", e.Type, e.Code, e.Message)
}

// 工厂方法
func New(errType ErrorType, code int, msg string, cause error) *AppError {
	return &AppError{
		Type:    errType,
		Code:    code,
		Message: msg,
		Cause:   cause,
	}
}

// Helper functions for common errors

func NotFound(msg string, cause error) *AppError {
	return New(TypeNotFound, 40400, msg, cause)
}

func InvalidInput(msg string, cause error) *AppError {
	return New(TypeInvalidInput, 40000, msg, cause)
}

func Conflict(msg string, cause error) *AppError {
	return New(TypeConflict, 40900, msg, cause)
}

func Internal(msg string, cause error) *AppError {
	return New(TypeInternal, 50000, msg, cause)
}

func ServiceUnavailable(msg string, cause error) *AppError {
	return New(TypeServiceUnavailable, 50300, msg, cause)
}

func Timeout(msg string, cause error) *AppError {
	return New(TypeTimeout, 50400, msg, cause)
}

// 快速判断是否可重试
func IsRetryable(err error) bool {
	if e, ok := err.(*AppError); ok {
		switch e.Type {
		case TypeServiceUnavailable, TypeTimeout, TypeInternal:
			return true
		}
	}
	return false
}

// 转换为 HTTP 状态码
func (e *AppError) HTTPStatus() int {
	switch e.Type {
	case TypeNotFound:
		return http.StatusNotFound
	case TypeInvalidInput:
		return http.StatusBadRequest
	case TypeConflict:
		return http.StatusConflict
	case TypeServiceUnavailable:
		return http.StatusServiceUnavailable
	case TypeTimeout:
		return http.StatusGatewayTimeout
	default:
		return http.StatusInternalServerError
	}
}
