package clients

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"
	"vv-ecommerce/pkg/common/apperror"
	"vv-ecommerce/pkg/common/response"

	"github.com/sony/gobreaker"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// NewCircuitBreaker creates a standard circuit breaker with given name
func NewCircuitBreaker(name string) *gobreaker.CircuitBreaker {
	st := gobreaker.Settings{
		Name:        name,
		MaxRequests: 1, // Half-open state allows 1 request
		Interval:    0, // Do not clear counts automatically
		Timeout:     30 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			// Trip if 5 consecutive failures
			return counts.ConsecutiveFailures >= 5
		},
		IsSuccessful: func(err error) bool {
			// 4xx errors should not trip the circuit breaker
			if err == nil {
				return true
			}
			var appErr *apperror.AppError
			if errors.As(err, &appErr) {
				if appErr.Code >= 400 && appErr.Code < 500 {
					return true
				}
			}
			return false
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			// In production, we should log this event or emit metrics
			fmt.Printf("CircuitBreaker '%s' changed state from %s to %s\n", name, from, to)
		},
	}
	return gobreaker.NewCircuitBreaker(st)
}

// NewHTTPClient returns an http.Client with OpenTelemetry instrumentation
func NewHTTPClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout:   timeout,
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}
}

// HandleHTTPError 解析 HTTP 错误响应
// 提取为公共函数，避免重复代码
func HandleHTTPError(resp *http.Response) error {
	// 尝试解析标准错误响应
	var res response.ErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err == nil && res.Code != 0 {
		return apperror.New(apperror.ErrorType(res.Type), res.Code, res.Message, nil)
	}

	// Fallback: 根据状态码推断
	switch resp.StatusCode {
	case http.StatusBadRequest:
		return apperror.InvalidInput("invalid input", nil)
	case http.StatusNotFound:
		return apperror.NotFound("resource not found", nil)
	case http.StatusConflict:
		return apperror.Conflict("resource conflict", nil)
	case http.StatusRequestTimeout, http.StatusGatewayTimeout, http.StatusServiceUnavailable:
		return apperror.ServiceUnavailable("service unavailable", nil)
	default:
		return apperror.Internal(fmt.Sprintf("upstream service error: %d", resp.StatusCode), nil)
	}
}

// WrapClientError 处理 http.Client.Do/Get/Post 返回的错误
// 专门检查 context error 和网络错误
func WrapClientError(err error, message string) error {
	if err == nil {
		return nil
	}

	// 检查 Context 超时
	if errors.Is(err, context.DeadlineExceeded) || os.IsTimeout(err) {
		return apperror.Timeout(message, err)
	}

	// 检查 Context 取消
	if errors.Is(err, context.Canceled) {
		// 如果是主动取消，通常也视为一种中断，暂定为 Internal 或 Timeout
		// 但在微服务调用中，通常意味着上游超时导致取消了下游调用
		return apperror.Timeout(message+" (canceled)", err)
	}

	// 其他网络错误视为服务不可用 (可重试)
	return apperror.ServiceUnavailable(message, err)
}
