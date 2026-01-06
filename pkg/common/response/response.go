package response

import (
	"net/http"
	"vv-ecommerce/pkg/common/apperror"

	"github.com/gin-gonic/gin"
)

// Success 200 OK，直接返回数据对象
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, data)
}

// ErrorResponse 定义统一的错误返回结构
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Type    string `json:"type"`
}

// Error 处理错误返回
// 如果是 AppError，则使用其定义的 Status 和 Info
// 否则默认返回 500
func Error(c *gin.Context, err error) {
	if err == nil {
		c.JSON(http.StatusOK, gin.H{"success": true})
		return
	}

	if appErr, ok := err.(*apperror.AppError); ok {
		c.JSON(appErr.HTTPStatus(), ErrorResponse{
			Code:    appErr.Code,
			Message: appErr.Message,
			Type:    string(appErr.Type),
		})
		return
	}

	// 默认兜底
	c.JSON(http.StatusInternalServerError, ErrorResponse{
		Code:    50000,
		Message: err.Error(),
		Type:    "INTERNAL_ERROR",
	})
}
